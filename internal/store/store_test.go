package store

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"testing"

	pb "github.com/BarkingIguana/poppie/proto/poppie"
)

func tempVaultPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "secrets.enc")
}

func TestOpenMemory_ReturnsEmptyStore(t *testing.T) {
	s := OpenMemory()
	labels := s.List(context.Background())
	if len(labels) != 0 {
		t.Errorf("expected empty store, got %d secrets", len(labels))
	}
}

func TestOpenMemory_StoreAndGet_WorksWithoutDisk(t *testing.T) {
	ctx := context.Background()
	s := OpenMemory()

	secret := &pb.Secret{
		Label:  "memory-test",
		Secret: "JBSWY3DPEHPK3PXP",
		Digits: 6,
		Period: 30,
	}
	if err := s.Store(ctx, secret); err != nil {
		t.Fatalf("Store: %v", err)
	}

	got, err := s.Get(ctx, "memory-test")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Secret != "JBSWY3DPEHPK3PXP" {
		t.Errorf("secret mismatch: got %s", got.Secret)
	}
}

func TestOpenMemory_Delete_Works(t *testing.T) {
	ctx := context.Background()
	s := OpenMemory()

	if err := s.Store(ctx, &pb.Secret{Label: "del", Secret: "JBSWY3DPEHPK3PXP"}); err != nil {
		t.Fatalf("Store: %v", err)
	}

	deleted, err := s.Delete(ctx, "del")
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !deleted {
		t.Error("expected deleted=true")
	}
	if len(s.List(ctx)) != 0 {
		t.Error("expected empty store after delete")
	}
}

func TestOpen_NewVault_CreatesEmptyStore(t *testing.T) {
	s, err := Open(context.Background(), tempVaultPath(t), "test-passphrase")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	labels := s.List(context.Background())
	if len(labels) != 0 {
		t.Errorf("expected empty vault, got %d secrets", len(labels))
	}
}

func TestStore_NewSecret_Persists(t *testing.T) {
	path := tempVaultPath(t)
	ctx := context.Background()
	passphrase := "test-passphrase"

	s, err := Open(ctx, path, passphrase)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	secret := &pb.Secret{
		Label:  "github.com",
		Secret: "JBSWY3DPEHPK3PXP",
		Digits: 6,
		Period: 30,
	}
	if err := s.Store(ctx, secret); err != nil {
		t.Fatalf("Store: %v", err)
	}

	// Reopen the vault from disk and verify persistence.
	s2, err := Open(ctx, path, passphrase)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}

	labels := s2.List(ctx)
	if len(labels) != 1 || labels[0] != "github.com" {
		t.Errorf("expected [github.com], got %v", labels)
	}
}

func TestStore_DuplicateLabel_ReturnsError(t *testing.T) {
	ctx := context.Background()
	s, err := Open(ctx, tempVaultPath(t), "pass")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	secret := &pb.Secret{Label: "dup", Secret: "JBSWY3DPEHPK3PXP"}
	if err := s.Store(ctx, secret); err != nil {
		t.Fatalf("first Store: %v", err)
	}

	err = s.Store(ctx, &pb.Secret{Label: "dup", Secret: "JBSWY3DPEHPK3PXP"})
	if !errors.Is(err, ErrAlreadyExists) {
		t.Errorf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestGet_ExistingSecret_ReturnsIt(t *testing.T) {
	ctx := context.Background()
	s, err := Open(ctx, tempVaultPath(t), "pass")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	original := &pb.Secret{
		Label:  "example.com",
		Secret: "JBSWY3DPEHPK3PXP",
		Digits: 6,
		Period: 30,
	}
	if err := s.Store(ctx, original); err != nil {
		t.Fatalf("Store: %v", err)
	}

	got, err := s.Get(ctx, "example.com")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if got.Secret != original.Secret {
		t.Errorf("secret mismatch: got %s, want %s", got.Secret, original.Secret)
	}
	if got.LastUsedAt == nil {
		t.Error("expected last_used_at to be set after Get")
	}
}

func TestGet_MissingSecret_ReturnsNotFound(t *testing.T) {
	ctx := context.Background()
	s, err := Open(ctx, tempVaultPath(t), "pass")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	_, err = s.Get(ctx, "nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDelete_ExistingSecret_RemovesIt(t *testing.T) {
	ctx := context.Background()
	s, err := Open(ctx, tempVaultPath(t), "pass")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	if err := s.Store(ctx, &pb.Secret{Label: "del-me", Secret: "JBSWY3DPEHPK3PXP"}); err != nil {
		t.Fatalf("Store: %v", err)
	}

	deleted, err := s.Delete(ctx, "del-me")
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !deleted {
		t.Error("expected deleted=true")
	}

	labels := s.List(ctx)
	if len(labels) != 0 {
		t.Errorf("expected empty vault after delete, got %v", labels)
	}
}

func TestDelete_MissingSecret_ReturnsFalse(t *testing.T) {
	ctx := context.Background()
	s, err := Open(ctx, tempVaultPath(t), "pass")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	deleted, err := s.Delete(ctx, "ghost")
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if deleted {
		t.Error("expected deleted=false for missing secret")
	}
}

func TestList_MultipleSecrets_ReturnsAllLabels(t *testing.T) {
	ctx := context.Background()
	s, err := Open(ctx, tempVaultPath(t), "pass")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	for _, label := range []string{"alpha", "beta", "gamma"} {
		if err := s.Store(ctx, &pb.Secret{Label: label, Secret: "JBSWY3DPEHPK3PXP"}); err != nil {
			t.Fatalf("Store %s: %v", label, err)
		}
	}

	labels := s.List(ctx)
	sort.Strings(labels)
	expected := []string{"alpha", "beta", "gamma"}
	if len(labels) != len(expected) {
		t.Fatalf("expected %d labels, got %d", len(expected), len(labels))
	}
	for i, l := range labels {
		if l != expected[i] {
			t.Errorf("label[%d]: got %s, want %s", i, l, expected[i])
		}
	}
}

func TestOpen_WrongPassphrase_ReturnsError(t *testing.T) {
	ctx := context.Background()
	path := tempVaultPath(t)

	s, err := Open(ctx, path, "correct-horse")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := s.Store(ctx, &pb.Secret{Label: "test", Secret: "JBSWY3DPEHPK3PXP"}); err != nil {
		t.Fatalf("Store: %v", err)
	}

	_, err = Open(ctx, path, "wrong-horse")
	if err == nil {
		t.Fatal("expected error with wrong passphrase, got nil")
	}
}

func TestStore_CreatesBackup(t *testing.T) {
	ctx := context.Background()
	path := tempVaultPath(t)

	s, err := Open(ctx, path, "pass")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	// First store creates the vault file.
	if err := s.Store(ctx, &pb.Secret{Label: "first", Secret: "JBSWY3DPEHPK3PXP"}); err != nil {
		t.Fatalf("Store first: %v", err)
	}

	// Second store should create a backup.
	if err := s.Store(ctx, &pb.Secret{Label: "second", Secret: "JBSWY3DPEHPK3PXP"}); err != nil {
		t.Fatalf("Store second: %v", err)
	}

	bakPath := path + ".bak"
	if _, err := os.Stat(bakPath); err != nil {
		t.Errorf("expected backup file at %s: %v", bakPath, err)
	}
}

func TestOpen_CorruptVaultFile_ReturnsError(t *testing.T) {
	ctx := context.Background()
	path := tempVaultPath(t)

	// Write a file that's too short to contain a valid salt.
	if err := os.WriteFile(path, []byte("short"), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := Open(ctx, path, "pass")
	if err == nil {
		t.Fatal("expected error for corrupt vault file")
	}
}

func TestOpen_GarbageAfterSalt_ReturnsError(t *testing.T) {
	ctx := context.Background()
	path := tempVaultPath(t)

	// Write a file with a valid-length salt but garbage ciphertext.
	data := make([]byte, saltLen+50)
	for i := range data {
		data[i] = byte(i)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := Open(ctx, path, "pass")
	if err == nil {
		t.Fatal("expected error for garbage ciphertext")
	}
}

func TestGet_UpdatesLastUsedAt(t *testing.T) {
	ctx := context.Background()
	s, err := Open(ctx, tempVaultPath(t), "pass")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	if err := s.Store(ctx, &pb.Secret{Label: "ts-test", Secret: "JBSWY3DPEHPK3PXP"}); err != nil {
		t.Fatalf("Store: %v", err)
	}

	sec, err := s.Get(ctx, "ts-test")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if sec.LastUsedAt == nil {
		t.Error("expected last_used_at to be set")
	}
	if sec.CreatedAt == nil {
		t.Error("expected created_at to be set")
	}
}
