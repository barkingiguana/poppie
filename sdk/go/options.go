package poppie

import (
	"os"
	"path/filepath"

	pb "github.com/BarkingIguana/poppie/proto/poppie"
)

// Algorithm represents the HMAC algorithm for TOTP generation.
type Algorithm int

const (
	// SHA1 is the default TOTP algorithm (RFC 6238).
	SHA1 Algorithm = iota
	// SHA256 uses SHA-256 for HMAC.
	SHA256
	// SHA512 uses SHA-512 for HMAC.
	SHA512
)

func (a Algorithm) toProto() pb.Algorithm {
	switch a {
	case SHA256:
		return pb.Algorithm_ALGORITHM_SHA256
	case SHA512:
		return pb.Algorithm_ALGORITHM_SHA512
	default:
		return pb.Algorithm_ALGORITHM_SHA1
	}
}

// Option configures a Client.
type Option func(*clientConfig)

type clientConfig struct {
	socketPath     string
	warningHandler WarningHandler
}

func defaultClientConfig() clientConfig {
	home, _ := os.UserHomeDir()
	return clientConfig{
		socketPath:     filepath.Join(home, ".config", "poppie", "poppie.sock"),
		warningHandler: DefaultWarningHandler,
	}
}

// WithSocketPath sets the Unix socket path for the poppie server.
func WithSocketPath(path string) Option {
	return func(c *clientConfig) {
		c.socketPath = path
	}
}

// WithWarningHandler sets the callback for version warnings from the server.
// Pass nil to disable warning handling.
func WithWarningHandler(h WarningHandler) Option {
	return func(c *clientConfig) {
		c.warningHandler = h
	}
}

// StoreOption configures a StoreSecret call.
type StoreOption func(*storeConfig)

type storeConfig struct {
	algorithm pb.Algorithm
	digits    uint32
	period    uint32
}

// WithAlgorithm sets the HMAC algorithm for TOTP generation.
func WithAlgorithm(a Algorithm) StoreOption {
	return func(c *storeConfig) {
		c.algorithm = a.toProto()
	}
}

// WithDigits sets the number of digits in the generated TOTP code.
func WithDigits(d uint32) StoreOption {
	return func(c *storeConfig) {
		c.digits = d
	}
}

// WithPeriod sets the time step in seconds for TOTP generation.
func WithPeriod(p uint32) StoreOption {
	return func(c *storeConfig) {
		c.period = p
	}
}
