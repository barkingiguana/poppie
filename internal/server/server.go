// Package server implements the poppie gRPC server.
package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/BarkingIguana/poppie/proto/poppie"

	"github.com/BarkingIguana/poppie/internal/store"
	"github.com/BarkingIguana/poppie/internal/totp"
)

// Config holds server configuration.
type Config struct {
	// SocketPath is the Unix socket path for the gRPC server.
	SocketPath string
	// VaultPath is the path to the encrypted vault file.
	VaultPath string
	// Passphrase is the vault encryption passphrase.
	Passphrase string
}

// DefaultConfig returns a Config with standard paths under ~/.config/poppie.
func DefaultConfig() Config {
	home, _ := os.UserHomeDir()
	configDir := filepath.Join(home, ".config", "poppie")
	return Config{
		SocketPath: filepath.Join(configDir, "poppie.sock"),
		VaultPath:  filepath.Join(configDir, "secrets.enc"),
	}
}

// Server is the poppie gRPC server.
type Server struct {
	pb.UnimplementedPoppieServiceServer
	store  *store.Store
	totp   *totp.Generator
	grpc   *grpc.Server
	config Config
	logger *slog.Logger
}

// New creates a new Server with the given configuration.
func New(ctx context.Context, cfg Config, logger *slog.Logger) (*Server, error) {
	st, err := store.Open(ctx, cfg.VaultPath, cfg.Passphrase)
	if err != nil {
		return nil, fmt.Errorf("failed to open vault: %w", err)
	}

	return &Server{
		store:  st,
		totp:   totp.NewGenerator(totp.SystemClock{}),
		config: cfg,
		logger: logger,
	}, nil
}

// NewWithDeps creates a Server with injected dependencies for testing.
func NewWithDeps(st *store.Store, gen *totp.Generator, logger *slog.Logger) *Server {
	return &Server{
		store:  st,
		totp:   gen,
		logger: logger,
	}
}

// Start begins listening on the configured Unix socket.
func (s *Server) Start() error {
	// Remove stale socket if present.
	if err := os.Remove(s.config.SocketPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove stale socket: %w", err)
	}

	dir := filepath.Dir(s.config.SocketPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create socket directory: %w", err)
	}

	lis, err := net.Listen("unix", s.config.SocketPath)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.config.SocketPath, err)
	}

	s.grpc = grpc.NewServer()
	pb.RegisterPoppieServiceServer(s.grpc, s)

	s.logger.Info("server starting", "socket", s.config.SocketPath)
	return s.grpc.Serve(lis)
}

// Stop gracefully stops the gRPC server.
func (s *Server) Stop() {
	if s.grpc != nil {
		s.grpc.GracefulStop()
	}
}

// StoreSecret implements PoppieService.StoreSecret.
func (s *Server) StoreSecret(ctx context.Context, req *pb.StoreSecretRequest) (*pb.StoreSecretResponse, error) {
	if req.Label == "" {
		return nil, status.Error(codes.InvalidArgument, "label is required")
	}
	if req.Secret == "" {
		return nil, status.Error(codes.InvalidArgument, "secret is required")
	}

	params := s.paramsFromRequest(req.Secret, req.Algorithm, req.Digits, req.Period)

	// Validate the secret by generating a test code.
	code, err := s.totp.GenerateCode(params)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid TOTP secret: %v", err)
	}

	secret := &pb.Secret{
		Label:     req.Label,
		Secret:    req.Secret,
		Algorithm: req.Algorithm,
		Digits:    req.Digits,
		Period:    req.Period,
	}

	if err := s.store.Store(ctx, secret); err != nil {
		s.logger.Error("failed to store secret", "label", req.Label, "error", err)
		return nil, status.Errorf(codes.Internal, "failed to store secret: %v", err)
	}

	s.logger.Info("secret stored", "label", req.Label)
	return &pb.StoreSecretResponse{
		Label:            req.Label,
		VerificationCode: code,
	}, nil
}

// GetCode implements PoppieService.GetCode.
func (s *Server) GetCode(ctx context.Context, req *pb.GetCodeRequest) (*pb.GetCodeResponse, error) {
	if req.Label == "" {
		return nil, status.Error(codes.InvalidArgument, "label is required")
	}

	secret, err := s.store.Get(ctx, req.Label)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "secret %q not found", req.Label)
	}

	params := s.paramsFromSecret(secret)
	code, err := s.totp.GenerateCode(params)
	if err != nil {
		s.logger.Error("failed to generate code", "label", req.Label, "error", err)
		return nil, status.Errorf(codes.Internal, "failed to generate code: %v", err)
	}

	period := int(secret.Period)
	if period == 0 {
		period = totp.DefaultPeriod
	}

	return &pb.GetCodeResponse{
		Code:            code,
		ValidForSeconds: s.totp.ValidForSeconds(period),
	}, nil
}

// ListSecrets implements PoppieService.ListSecrets.
func (s *Server) ListSecrets(ctx context.Context, _ *pb.ListSecretsRequest) (*pb.ListSecretsResponse, error) {
	labels := s.store.List(ctx)
	return &pb.ListSecretsResponse{Labels: labels}, nil
}

// DeleteSecret implements PoppieService.DeleteSecret.
func (s *Server) DeleteSecret(ctx context.Context, req *pb.DeleteSecretRequest) (*pb.DeleteSecretResponse, error) {
	if req.Label == "" {
		return nil, status.Error(codes.InvalidArgument, "label is required")
	}

	deleted, err := s.store.Delete(ctx, req.Label)
	if err != nil {
		s.logger.Error("failed to delete secret", "label", req.Label, "error", err)
		return nil, status.Errorf(codes.Internal, "failed to delete secret: %v", err)
	}

	if deleted {
		s.logger.Info("secret deleted", "label", req.Label)
	}

	return &pb.DeleteSecretResponse{Deleted: deleted}, nil
}

func (s *Server) paramsFromRequest(secret string, algo pb.Algorithm, digits, period uint32) totp.Params {
	return totp.Params{
		Secret:    secret,
		Algorithm: convertAlgorithm(algo),
		Digits:    intOrDefault(int(digits), totp.DefaultDigits),
		Period:    intOrDefault(int(period), totp.DefaultPeriod),
	}
}

func (s *Server) paramsFromSecret(secret *pb.Secret) totp.Params {
	return s.paramsFromRequest(secret.Secret, secret.Algorithm, secret.Digits, secret.Period)
}

func convertAlgorithm(algo pb.Algorithm) totp.Algorithm {
	switch algo {
	case pb.Algorithm_ALGORITHM_SHA256:
		return totp.AlgorithmSHA256
	case pb.Algorithm_ALGORITHM_SHA512:
		return totp.AlgorithmSHA512
	default:
		return totp.AlgorithmSHA1
	}
}

func intOrDefault(v, def int) int {
	if v == 0 {
		return def
	}
	return v
}
