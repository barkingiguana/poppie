package server

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	pb "github.com/BarkingIguana/poppie/proto/poppie"

	"github.com/BarkingIguana/poppie/internal/store"
	"github.com/BarkingIguana/poppie/internal/totp"
)

type fixedClock struct {
	t time.Time
}

func (c fixedClock) Now() time.Time { return c.t }

func testServer(t *testing.T) *Server {
	t.Helper()
	ctx := context.Background()
	vaultPath := filepath.Join(t.TempDir(), "test.enc")

	st, err := store.Open(ctx, vaultPath, "test-pass")
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}

	gen := totp.NewGenerator(fixedClock{t: time.Unix(59, 0).UTC()})
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	return NewWithDeps(st, gen, logger)
}

func TestStoreSecret_ValidRequest_StoresAndReturnsCode(t *testing.T) {
	srv := testServer(t)
	ctx := context.Background()

	resp, err := srv.StoreSecret(ctx, &pb.StoreSecretRequest{
		Label:  "github.com",
		Secret: "JBSWY3DPEHPK3PXP",
		Digits: 6,
		Period: 30,
	})
	if err != nil {
		t.Fatalf("StoreSecret: %v", err)
	}

	if resp.Label != "github.com" {
		t.Errorf("label: got %s, want github.com", resp.Label)
	}
	if len(resp.VerificationCode) != 6 {
		t.Errorf("expected 6-digit code, got %q", resp.VerificationCode)
	}
}

func TestStoreSecret_EmptyLabel_ReturnsError(t *testing.T) {
	srv := testServer(t)
	_, err := srv.StoreSecret(context.Background(), &pb.StoreSecretRequest{
		Secret: "JBSWY3DPEHPK3PXP",
	})
	if err == nil {
		t.Fatal("expected error for empty label")
	}
}

func TestStoreSecret_EmptySecret_ReturnsError(t *testing.T) {
	srv := testServer(t)
	_, err := srv.StoreSecret(context.Background(), &pb.StoreSecretRequest{
		Label: "test",
	})
	if err == nil {
		t.Fatal("expected error for empty secret")
	}
}

func TestStoreSecret_InvalidSecret_ReturnsError(t *testing.T) {
	srv := testServer(t)
	_, err := srv.StoreSecret(context.Background(), &pb.StoreSecretRequest{
		Label:  "test",
		Secret: "!!!invalid!!!",
	})
	if err == nil {
		t.Fatal("expected error for invalid secret")
	}
}

func TestGetCode_ExistingSecret_ReturnsCode(t *testing.T) {
	srv := testServer(t)
	ctx := context.Background()

	_, err := srv.StoreSecret(ctx, &pb.StoreSecretRequest{
		Label:  "example.com",
		Secret: "JBSWY3DPEHPK3PXP",
	})
	if err != nil {
		t.Fatalf("StoreSecret: %v", err)
	}

	resp, err := srv.GetCode(ctx, &pb.GetCodeRequest{Label: "example.com"})
	if err != nil {
		t.Fatalf("GetCode: %v", err)
	}

	if len(resp.Code) != 6 {
		t.Errorf("expected 6-digit code, got %q", resp.Code)
	}
	if resp.ValidForSeconds == 0 {
		t.Error("expected valid_for_seconds > 0")
	}
}

func TestGetCode_MissingSecret_ReturnsNotFound(t *testing.T) {
	srv := testServer(t)
	_, err := srv.GetCode(context.Background(), &pb.GetCodeRequest{Label: "ghost"})
	if err == nil {
		t.Fatal("expected error for missing secret")
	}
}

func TestGetCode_EmptyLabel_ReturnsError(t *testing.T) {
	srv := testServer(t)
	_, err := srv.GetCode(context.Background(), &pb.GetCodeRequest{})
	if err == nil {
		t.Fatal("expected error for empty label")
	}
}

func TestListSecrets_EmptyVault_ReturnsEmpty(t *testing.T) {
	srv := testServer(t)
	resp, err := srv.ListSecrets(context.Background(), &pb.ListSecretsRequest{})
	if err != nil {
		t.Fatalf("ListSecrets: %v", err)
	}
	if len(resp.Labels) != 0 {
		t.Errorf("expected 0 labels, got %d", len(resp.Labels))
	}
}

func TestListSecrets_WithSecrets_ReturnsLabels(t *testing.T) {
	srv := testServer(t)
	ctx := context.Background()

	for _, label := range []string{"a.com", "b.com"} {
		_, err := srv.StoreSecret(ctx, &pb.StoreSecretRequest{
			Label:  label,
			Secret: "JBSWY3DPEHPK3PXP",
		})
		if err != nil {
			t.Fatalf("StoreSecret %s: %v", label, err)
		}
	}

	resp, err := srv.ListSecrets(ctx, &pb.ListSecretsRequest{})
	if err != nil {
		t.Fatalf("ListSecrets: %v", err)
	}
	if len(resp.Labels) != 2 {
		t.Errorf("expected 2 labels, got %d", len(resp.Labels))
	}
}

func TestDeleteSecret_ExistingSecret_Deletes(t *testing.T) {
	srv := testServer(t)
	ctx := context.Background()

	_, err := srv.StoreSecret(ctx, &pb.StoreSecretRequest{
		Label:  "delete-me",
		Secret: "JBSWY3DPEHPK3PXP",
	})
	if err != nil {
		t.Fatalf("StoreSecret: %v", err)
	}

	resp, err := srv.DeleteSecret(ctx, &pb.DeleteSecretRequest{Label: "delete-me"})
	if err != nil {
		t.Fatalf("DeleteSecret: %v", err)
	}
	if !resp.Deleted {
		t.Error("expected deleted=true")
	}

	list, _ := srv.ListSecrets(ctx, &pb.ListSecretsRequest{})
	if len(list.Labels) != 0 {
		t.Error("expected empty vault after delete")
	}
}

func TestDeleteSecret_MissingSecret_ReturnsFalse(t *testing.T) {
	srv := testServer(t)
	resp, err := srv.DeleteSecret(context.Background(), &pb.DeleteSecretRequest{Label: "nope"})
	if err != nil {
		t.Fatalf("DeleteSecret: %v", err)
	}
	if resp.Deleted {
		t.Error("expected deleted=false for missing secret")
	}
}

func TestDeleteSecret_EmptyLabel_ReturnsError(t *testing.T) {
	srv := testServer(t)
	_, err := srv.DeleteSecret(context.Background(), &pb.DeleteSecretRequest{})
	if err == nil {
		t.Fatal("expected error for empty label")
	}
}

func TestStoreSecret_SHA256_StoresAndReturnsCode(t *testing.T) {
	srv := testServer(t)
	resp, err := srv.StoreSecret(context.Background(), &pb.StoreSecretRequest{
		Label:     "sha256.example",
		Secret:    "JBSWY3DPEHPK3PXP",
		Algorithm: pb.Algorithm_ALGORITHM_SHA256,
		Digits:    8,
		Period:    60,
	})
	if err != nil {
		t.Fatalf("StoreSecret: %v", err)
	}
	if len(resp.VerificationCode) != 8 {
		t.Errorf("expected 8-digit code, got %q", resp.VerificationCode)
	}
}

func TestStoreSecret_SHA512_StoresAndReturnsCode(t *testing.T) {
	srv := testServer(t)
	resp, err := srv.StoreSecret(context.Background(), &pb.StoreSecretRequest{
		Label:     "sha512.example",
		Secret:    "JBSWY3DPEHPK3PXP",
		Algorithm: pb.Algorithm_ALGORITHM_SHA512,
	})
	if err != nil {
		t.Fatalf("StoreSecret: %v", err)
	}
	if len(resp.VerificationCode) != 6 {
		t.Errorf("expected 6-digit code, got %q", resp.VerificationCode)
	}
}

func TestStoreSecret_DuplicateLabel_ReturnsError(t *testing.T) {
	srv := testServer(t)
	ctx := context.Background()

	_, err := srv.StoreSecret(ctx, &pb.StoreSecretRequest{
		Label:  "dup",
		Secret: "JBSWY3DPEHPK3PXP",
	})
	if err != nil {
		t.Fatalf("first StoreSecret: %v", err)
	}

	_, err = srv.StoreSecret(ctx, &pb.StoreSecretRequest{
		Label:  "dup",
		Secret: "JBSWY3DPEHPK3PXP",
	})
	if err == nil {
		t.Fatal("expected error for duplicate label")
	}
}
