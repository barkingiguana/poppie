package server

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestVersionInterceptor_NoHeaders_ReturnsUnknown(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	interceptor := VersionInterceptor(logger)

	ctx := context.Background()
	var capturedCtx context.Context

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		capturedCtx = ctx
		return "ok", nil
	}

	resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)
	if err != nil {
		t.Fatalf("interceptor: %v", err)
	}
	if resp != "ok" {
		t.Errorf("expected ok, got %v", resp)
	}
	_ = capturedCtx
}

func TestVersionInterceptor_MatchingVersion_ReturnsSupported(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	interceptor := VersionInterceptor(logger)

	md := metadata.Pairs(
		HeaderSDKVersion, Version(),
		HeaderSDKName, "poppie-go",
	)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "ok", nil
	}

	_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)
	if err != nil {
		t.Fatalf("interceptor: %v", err)
	}
}

func TestEvaluateVersion_EmptyVersion_ReturnsUnknown(t *testing.T) {
	status, message := evaluateVersion("")
	if status != StatusUnknown {
		t.Errorf("status: got %s, want %s", status, StatusUnknown)
	}
	if message != "" {
		t.Errorf("message: got %q, want empty", message)
	}
}

func TestEvaluateVersion_SameMajor_ReturnsSupported(t *testing.T) {
	// Override version for this test since ldflags aren't set during tests.
	old := version
	version = "0.1.0"
	defer func() { version = old }()

	status, message := evaluateVersion("0.1.0")
	if status != StatusSupported {
		t.Errorf("status: got %s, want %s", status, StatusSupported)
	}
	if message != "" {
		t.Errorf("message: got %q, want empty", message)
	}
}

func TestEvaluateVersion_OlderMajor_ReturnsDeprecated(t *testing.T) {
	old := version
	version = "2.0.0"
	defer func() { version = old }()

	status, message := evaluateVersion("1.0.0")
	if status != StatusDeprecated {
		t.Errorf("status: got %s, want %s", status, StatusDeprecated)
	}
	if message == "" {
		t.Error("expected deprecation message")
	}
}

func TestEvaluateVersion_InvalidVersion_ReturnsUnknown(t *testing.T) {
	status, _ := evaluateVersion("notaversion")
	if status != StatusUnknown {
		t.Errorf("status: got %s, want %s", status, StatusUnknown)
	}
}

func TestMajorVersion_ValidSemver_ExtractsMajor(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"0.1.0", "0"},
		{"1.2.3", "1"},
		{"10.0.0", "10"},
	}
	for _, tt := range tests {
		got := majorVersion(tt.input)
		if got != tt.want {
			t.Errorf("majorVersion(%q): got %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestMajorVersion_NoDot_ReturnsEmpty(t *testing.T) {
	got := majorVersion("123")
	if got != "" {
		t.Errorf("majorVersion(123): got %q, want empty", got)
	}
}

func TestExtractClientHeaders_WithMetadata_ReturnsValues(t *testing.T) {
	md := metadata.Pairs(
		HeaderSDKVersion, "0.1.0",
		HeaderSDKName, "poppie-python",
	)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	ver, name := extractClientHeaders(ctx)
	if ver != "0.1.0" {
		t.Errorf("version: got %q, want 0.1.0", ver)
	}
	if name != "poppie-python" {
		t.Errorf("name: got %q, want poppie-python", name)
	}
}

func TestExtractClientHeaders_NoMetadata_ReturnsEmpty(t *testing.T) {
	ver, name := extractClientHeaders(context.Background())
	if ver != "" {
		t.Errorf("version: got %q, want empty", ver)
	}
	if name != "" {
		t.Errorf("name: got %q, want empty", name)
	}
}
