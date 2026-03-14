package poppie_test

import (
	"context"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"google.golang.org/grpc"

	pb "github.com/BarkingIguana/poppie/proto/poppie"

	"github.com/BarkingIguana/poppie/internal/server"
	"github.com/BarkingIguana/poppie/internal/store"
	"github.com/BarkingIguana/poppie/internal/totp"

	poppie "github.com/BarkingIguana/poppie/sdk/go"
)

func startTestServer(t *testing.T) string {
	t.Helper()
	ctx := context.Background()
	vaultPath := filepath.Join(t.TempDir(), "test.enc")

	st, err := store.Open(ctx, vaultPath, "test-pass")
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}

	gen := totp.NewGenerator(totp.SystemClock{})
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	srv := server.NewWithDeps(st, gen, logger)

	socketPath := filepath.Join(t.TempDir(), "test.sock")
	lis, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(server.VersionInterceptor(logger)))
	pb.RegisterPoppieServiceServer(grpcServer, srv)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := grpcServer.Serve(lis); err != nil {
			// Server was stopped, this is expected during cleanup.
		}
	}()

	t.Cleanup(func() {
		grpcServer.GracefulStop()
		wg.Wait()
	})

	return socketPath
}

func TestClient_StoreAndGetCode(t *testing.T) {
	socketPath := startTestServer(t)
	ctx := context.Background()

	client, err := poppie.New(ctx, poppie.WithSocketPath(socketPath))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer client.Close()

	result, err := client.StoreSecret(ctx, "github.com", "JBSWY3DPEHPK3PXP")
	if err != nil {
		t.Fatalf("StoreSecret: %v", err)
	}
	if result.Label != "github.com" {
		t.Errorf("label: got %s, want github.com", result.Label)
	}
	if len(result.VerificationCode) != 6 {
		t.Errorf("expected 6-digit code, got %q", result.VerificationCode)
	}

	code, err := client.GetCode(ctx, "github.com")
	if err != nil {
		t.Fatalf("GetCode: %v", err)
	}
	if len(code.Code) != 6 {
		t.Errorf("expected 6-digit code, got %q", code.Code)
	}
	if code.ValidForSeconds == 0 {
		t.Error("expected ValidForSeconds > 0")
	}
}

func TestClient_StoreWithOptions(t *testing.T) {
	socketPath := startTestServer(t)
	ctx := context.Background()

	client, err := poppie.New(ctx, poppie.WithSocketPath(socketPath))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer client.Close()

	result, err := client.StoreSecret(ctx, "custom.com", "JBSWY3DPEHPK3PXP",
		poppie.WithAlgorithm(poppie.SHA256),
		poppie.WithDigits(8),
		poppie.WithPeriod(60),
	)
	if err != nil {
		t.Fatalf("StoreSecret: %v", err)
	}
	if len(result.VerificationCode) != 8 {
		t.Errorf("expected 8-digit code, got %q", result.VerificationCode)
	}
}

func TestClient_ListSecrets(t *testing.T) {
	socketPath := startTestServer(t)
	ctx := context.Background()

	client, err := poppie.New(ctx, poppie.WithSocketPath(socketPath))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer client.Close()

	labels, err := client.ListSecrets(ctx)
	if err != nil {
		t.Fatalf("ListSecrets: %v", err)
	}
	if len(labels) != 0 {
		t.Errorf("expected 0 labels, got %d", len(labels))
	}

	_, err = client.StoreSecret(ctx, "a.com", "JBSWY3DPEHPK3PXP")
	if err != nil {
		t.Fatalf("StoreSecret: %v", err)
	}

	labels, err = client.ListSecrets(ctx)
	if err != nil {
		t.Fatalf("ListSecrets: %v", err)
	}
	if len(labels) != 1 {
		t.Errorf("expected 1 label, got %d", len(labels))
	}
}

func TestClient_DeleteSecret(t *testing.T) {
	socketPath := startTestServer(t)
	ctx := context.Background()

	client, err := poppie.New(ctx, poppie.WithSocketPath(socketPath))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer client.Close()

	_, err = client.StoreSecret(ctx, "delete-me", "JBSWY3DPEHPK3PXP")
	if err != nil {
		t.Fatalf("StoreSecret: %v", err)
	}

	deleted, err := client.DeleteSecret(ctx, "delete-me")
	if err != nil {
		t.Fatalf("DeleteSecret: %v", err)
	}
	if !deleted {
		t.Error("expected deleted=true")
	}

	deleted, err = client.DeleteSecret(ctx, "delete-me")
	if err != nil {
		t.Fatalf("DeleteSecret: %v", err)
	}
	if deleted {
		t.Error("expected deleted=false for already-deleted secret")
	}
}

func TestClient_WarningHandler(t *testing.T) {
	socketPath := startTestServer(t)
	ctx := context.Background()

	var warnings []poppie.VersionWarning
	handler := func(w poppie.VersionWarning) {
		warnings = append(warnings, w)
	}

	client, err := poppie.New(ctx,
		poppie.WithSocketPath(socketPath),
		poppie.WithWarningHandler(handler),
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer client.Close()

	// Same version — should be supported, no warnings.
	_, err = client.ListSecrets(ctx)
	if err != nil {
		t.Fatalf("ListSecrets: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("expected 0 warnings, got %d", len(warnings))
	}
}

func TestClient_NilWarningHandler(t *testing.T) {
	socketPath := startTestServer(t)
	ctx := context.Background()

	client, err := poppie.New(ctx,
		poppie.WithSocketPath(socketPath),
		poppie.WithWarningHandler(nil),
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer client.Close()

	// Should not panic with nil handler.
	_, err = client.ListSecrets(ctx)
	if err != nil {
		t.Fatalf("ListSecrets: %v", err)
	}
}
