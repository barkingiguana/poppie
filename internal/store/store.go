// Package store provides encrypted TOTP secret storage backed by a local file.
package store

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/crypto/argon2"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/BarkingIguana/poppie/proto/poppie"
)

const (
	saltLen       = 16
	argon2Time    = 3
	argon2Memory  = 64 * 1024
	argon2Threads = 4
	argon2KeyLen  = 32
)

// ErrNotFound is returned when a secret label does not exist in the vault.
var ErrNotFound = errors.New("secret not found")

// ErrAlreadyExists is returned when trying to store a secret with a duplicate label.
var ErrAlreadyExists = errors.New("secret already exists")

// Store manages encrypted TOTP secrets in a local file.
type Store struct {
	path    string
	key     []byte
	salt    []byte
	mu      sync.RWMutex
	secrets map[string]*pb.Secret
}

// Open creates or opens a vault at the given path, deriving an encryption key
// from the passphrase. If the file does not exist, an empty vault is created.
func Open(_ context.Context, path string, passphrase string) (*Store, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create vault directory %q: %w", dir, err)
	}

	s := &Store{
		path:    path,
		secrets: make(map[string]*pb.Secret),
	}

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		s.salt = make([]byte, saltLen)
		if _, err := io.ReadFull(rand.Reader, s.salt); err != nil {
			return nil, fmt.Errorf("failed to generate salt: %w", err)
		}
		s.key = deriveKey(passphrase, s.salt)
		return s, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read vault %q: %w", path, err)
	}

	if len(data) < saltLen {
		return nil, fmt.Errorf("vault file %q is corrupt: too short", path)
	}

	s.salt = data[:saltLen]
	s.key = deriveKey(passphrase, s.salt)

	plaintext, err := decrypt(s.key, data[saltLen:])
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt vault: %w", err)
	}

	var vault pb.VaultContents
	if err := proto.Unmarshal(plaintext, &vault); err != nil {
		return nil, fmt.Errorf("failed to decode vault contents: %w", err)
	}

	for _, sec := range vault.Secrets {
		s.secrets[sec.Label] = sec
	}

	return s, nil
}

// Store adds a new TOTP secret to the vault and persists to disk.
func (s *Store) Store(_ context.Context, secret *pb.Secret) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.secrets[secret.Label]; exists {
		return fmt.Errorf("label %q: %w", secret.Label, ErrAlreadyExists)
	}

	secret.CreatedAt = timestamppb.Now()
	s.secrets[secret.Label] = secret

	return s.persist()
}

// Get retrieves a secret by label and updates its last_used_at timestamp.
func (s *Store) Get(_ context.Context, label string) (*pb.Secret, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sec, ok := s.secrets[label]
	if !ok {
		return nil, fmt.Errorf("label %q: %w", label, ErrNotFound)
	}

	sec.LastUsedAt = timestamppb.Now()
	if err := s.persist(); err != nil {
		return nil, fmt.Errorf("failed to update last_used_at: %w", err)
	}

	return sec, nil
}

// Delete removes a secret by label and persists to disk.
func (s *Store) Delete(_ context.Context, label string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.secrets[label]; !ok {
		return false, nil
	}

	delete(s.secrets, label)
	if err := s.persist(); err != nil {
		return false, fmt.Errorf("failed to persist after delete: %w", err)
	}

	return true, nil
}

// List returns the labels of all stored secrets.
func (s *Store) List(_ context.Context) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	labels := make([]string, 0, len(s.secrets))
	for label := range s.secrets {
		labels = append(labels, label)
	}
	return labels
}

func (s *Store) persist() error {
	vault := &pb.VaultContents{
		Secrets: make([]*pb.Secret, 0, len(s.secrets)),
	}
	for _, sec := range s.secrets {
		vault.Secrets = append(vault.Secrets, sec)
	}

	plaintext, err := proto.Marshal(vault)
	if err != nil {
		return fmt.Errorf("failed to encode secrets: %w", err)
	}

	ciphertext, err := encrypt(s.key, plaintext)
	if err != nil {
		return fmt.Errorf("failed to encrypt vault: %w", err)
	}

	// Backup existing vault before overwriting.
	if _, statErr := os.Stat(s.path); statErr == nil {
		if err := copyFile(s.path, s.path+".bak"); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
	}

	// Atomic write: temp file + rename.
	data := append(s.salt, ciphertext...)
	tmp := fmt.Sprintf("%s.tmp.%d", s.path, time.Now().UnixNano())
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return fmt.Errorf("failed to write temp vault: %w", err)
	}
	if err := os.Rename(tmp, s.path); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("failed to rename temp vault: %w", err)
	}

	return nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0600)
}

func deriveKey(passphrase string, salt []byte) []byte {
	return argon2.IDKey([]byte(passphrase), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)
}

func encrypt(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func decrypt(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}

	nonce := ciphertext[:gcm.NonceSize()]
	return gcm.Open(nil, nonce, ciphertext[gcm.NonceSize():], nil)
}
