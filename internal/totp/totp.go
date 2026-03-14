// Package totp implements RFC 6238 Time-Based One-Time Password generation.
package totp

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"hash"
	"math"
	"strings"
	"time"
)

// Algorithm identifies the HMAC hash function used for TOTP generation.
type Algorithm int

const (
	// AlgorithmSHA1 is the default TOTP algorithm per RFC 6238.
	AlgorithmSHA1 Algorithm = iota
	// AlgorithmSHA256 uses SHA-256.
	AlgorithmSHA256
	// AlgorithmSHA512 uses SHA-512.
	AlgorithmSHA512
)

// DefaultDigits is the standard TOTP code length.
const DefaultDigits = 6

// DefaultPeriod is the standard TOTP time step in seconds.
const DefaultPeriod = 30

// Params holds the configuration for TOTP code generation.
type Params struct {
	// Secret is the base32-encoded shared secret.
	Secret string
	// Algorithm is the HMAC hash function to use.
	Algorithm Algorithm
	// Digits is the number of digits in the generated code.
	Digits int
	// Period is the time step in seconds.
	Period int
}

// Clock abstracts time for testability.
type Clock interface {
	Now() time.Time
}

// SystemClock uses the real system time in UTC.
type SystemClock struct{}

// Now returns the current UTC time.
func (SystemClock) Now() time.Time {
	return time.Now().UTC()
}

// Generator produces TOTP codes.
type Generator struct {
	clock Clock
}

// NewGenerator creates a Generator with the given clock.
func NewGenerator(clock Clock) *Generator {
	return &Generator{clock: clock}
}

// GenerateCode produces a TOTP code for the given params at the current time.
func (g *Generator) GenerateCode(params Params) (string, error) {
	return g.GenerateCodeAt(params, g.clock.Now())
}

// GenerateCodeAt produces a TOTP code for the given params at a specific time.
func (g *Generator) GenerateCodeAt(params Params, t time.Time) (string, error) {
	params = withDefaults(params)

	secret, err := decodeSecret(params.Secret)
	if err != nil {
		return "", fmt.Errorf("invalid secret: %w", err)
	}

	counter := uint64(math.Floor(float64(t.UTC().Unix()) / float64(params.Period)))
	code := generateHOTP(secret, counter, params.Algorithm, params.Digits)

	return code, nil
}

// Validate checks whether a code is valid for the given params, allowing
// for a window of +/- 1 time step to account for clock skew.
func (g *Generator) Validate(params Params, code string) (bool, error) {
	return g.ValidateAt(params, code, g.clock.Now())
}

// ValidateAt checks a code against a specific time with a +/- 1 step window.
func (g *Generator) ValidateAt(params Params, code string, t time.Time) (bool, error) {
	params = withDefaults(params)

	secret, err := decodeSecret(params.Secret)
	if err != nil {
		return false, fmt.Errorf("invalid secret: %w", err)
	}

	counter := uint64(math.Floor(float64(t.UTC().Unix()) / float64(params.Period)))

	for i := int64(-1); i <= 1; i++ {
		c := uint64(int64(counter) + i)
		expected := generateHOTP(secret, c, params.Algorithm, params.Digits)
		if hmac.Equal([]byte(expected), []byte(code)) {
			return true, nil
		}
	}

	return false, nil
}

// ValidForSeconds returns how many seconds remain before the current code expires.
func (g *Generator) ValidForSeconds(period int) uint32 {
	if period <= 0 {
		period = DefaultPeriod
	}
	now := g.clock.Now().UTC()
	elapsed := now.Unix() % int64(period)
	return uint32(int64(period) - elapsed)
}

func withDefaults(p Params) Params {
	if p.Digits == 0 {
		p.Digits = DefaultDigits
	}
	if p.Period == 0 {
		p.Period = DefaultPeriod
	}
	return p
}

func decodeSecret(secret string) ([]byte, error) {
	// Base32 secrets are often provided with spaces and lowercase.
	cleaned := strings.ToUpper(strings.ReplaceAll(secret, " ", ""))
	// Add padding if missing.
	if m := len(cleaned) % 8; m != 0 {
		cleaned += strings.Repeat("=", 8-m)
	}
	return base32.StdEncoding.DecodeString(cleaned)
}

func generateHOTP(secret []byte, counter uint64, algo Algorithm, digits int) string {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, counter)

	mac := hmac.New(hashFunc(algo), secret)
	mac.Write(buf)
	sum := mac.Sum(nil)

	// Dynamic truncation per RFC 4226 section 5.4.
	offset := sum[len(sum)-1] & 0x0f
	truncated := binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7fffffff

	mod := uint32(math.Pow10(digits))
	otp := truncated % mod

	return fmt.Sprintf("%0*d", digits, otp)
}

func hashFunc(algo Algorithm) func() hash.Hash {
	switch algo {
	case AlgorithmSHA256:
		return sha256.New
	case AlgorithmSHA512:
		return sha512.New
	default:
		return sha1.New
	}
}
