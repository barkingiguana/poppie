package totp

import (
	"testing"
	"time"
)

// fixedClock returns a fixed time for deterministic tests.
type fixedClock struct {
	t time.Time
}

func (c fixedClock) Now() time.Time { return c.t }

// RFC 6238 test vectors use the secret "12345678901234567890" (ASCII).
// Base32-encoded: GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ
const rfc6238SecretSHA1 = "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"

func TestGenerateCode_RFC6238Vectors_ReturnsExpectedCodes(t *testing.T) {
	tests := []struct {
		name     string
		time     time.Time
		algo     Algorithm
		secret   string
		digits   int
		expected string
	}{
		{
			name:     "SHA1 at 59s",
			time:     time.Unix(59, 0).UTC(),
			algo:     AlgorithmSHA1,
			secret:   rfc6238SecretSHA1,
			digits:   8,
			expected: "94287082",
		},
		{
			name:     "SHA1 at 1111111109",
			time:     time.Unix(1111111109, 0).UTC(),
			algo:     AlgorithmSHA1,
			secret:   rfc6238SecretSHA1,
			digits:   8,
			expected: "07081804",
		},
		{
			name:     "SHA1 at 1234567890",
			time:     time.Unix(1234567890, 0).UTC(),
			algo:     AlgorithmSHA1,
			secret:   rfc6238SecretSHA1,
			digits:   8,
			expected: "89005924",
		},
		{
			name:     "SHA1 at 2000000000",
			time:     time.Unix(2000000000, 0).UTC(),
			algo:     AlgorithmSHA1,
			secret:   rfc6238SecretSHA1,
			digits:   8,
			expected: "69279037",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewGenerator(fixedClock{t: tt.time})
			code, err := gen.GenerateCodeAt(Params{
				Secret:    tt.secret,
				Algorithm: tt.algo,
				Digits:    tt.digits,
				Period:    DefaultPeriod,
			}, tt.time)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if code != tt.expected {
				t.Errorf("got %s, want %s", code, tt.expected)
			}
		})
	}
}

func TestGenerateCode_DefaultParams_Returns6Digits(t *testing.T) {
	gen := NewGenerator(fixedClock{t: time.Unix(59, 0).UTC()})
	code, err := gen.GenerateCode(Params{
		Secret: rfc6238SecretSHA1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(code) != 6 {
		t.Errorf("expected 6 digit code, got %d digits: %s", len(code), code)
	}
}

func TestGenerateCode_InvalidSecret_ReturnsError(t *testing.T) {
	gen := NewGenerator(SystemClock{})
	_, err := gen.GenerateCode(Params{
		Secret: "!!!not-base32!!!",
	})
	if err == nil {
		t.Fatal("expected error for invalid secret, got nil")
	}
}

func TestGenerateCode_SecretWithSpacesAndLowercase_Normalises(t *testing.T) {
	gen := NewGenerator(fixedClock{t: time.Unix(59, 0).UTC()})

	clean, err := gen.GenerateCode(Params{
		Secret: rfc6238SecretSHA1,
		Digits: 8,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dirty, err := gen.GenerateCode(Params{
		Secret: "gezd gnbv gy3t qojq gezd gnbv gy3t qojq",
		Digits: 8,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if clean != dirty {
		t.Errorf("normalised secret should produce same code: clean=%s dirty=%s", clean, dirty)
	}
}

func TestValidate_CorrectCode_ReturnsTrue(t *testing.T) {
	now := time.Unix(59, 0).UTC()
	gen := NewGenerator(fixedClock{t: now})

	code, err := gen.GenerateCodeAt(Params{
		Secret: rfc6238SecretSHA1,
		Digits: 8,
	}, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	valid, err := gen.ValidateAt(Params{
		Secret: rfc6238SecretSHA1,
		Digits: 8,
	}, code, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !valid {
		t.Error("expected code to be valid")
	}
}

func TestValidate_WrongCode_ReturnsFalse(t *testing.T) {
	gen := NewGenerator(fixedClock{t: time.Unix(59, 0).UTC()})

	valid, err := gen.Validate(Params{
		Secret: rfc6238SecretSHA1,
	}, "000000")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valid {
		t.Error("expected invalid code to return false")
	}
}

func TestValidate_AdjacentTimeStep_ReturnsTrue(t *testing.T) {
	params := Params{
		Secret: rfc6238SecretSHA1,
		Digits: 8,
		Period: DefaultPeriod,
	}

	// Generate code at T=59
	genAt59 := NewGenerator(fixedClock{t: time.Unix(59, 0).UTC()})
	code, err := genAt59.GenerateCodeAt(params, time.Unix(59, 0).UTC())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Validate at T=59+30 (one step later) — should still be accepted (window of ±1)
	genAt89 := NewGenerator(fixedClock{t: time.Unix(89, 0).UTC()})
	valid, err := genAt89.ValidateAt(params, code, time.Unix(89, 0).UTC())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !valid {
		t.Error("expected code from adjacent time step to be valid")
	}
}

func TestValidForSeconds_MidPeriod_ReturnsRemainder(t *testing.T) {
	// At T=45, we're 15 seconds into a 30-second period. 15 seconds remain.
	gen := NewGenerator(fixedClock{t: time.Unix(45, 0).UTC()})
	remaining := gen.ValidForSeconds(DefaultPeriod)
	if remaining != 15 {
		t.Errorf("expected 15 seconds remaining, got %d", remaining)
	}
}

func TestValidForSeconds_StartOfPeriod_Returns30(t *testing.T) {
	gen := NewGenerator(fixedClock{t: time.Unix(30, 0).UTC()})
	remaining := gen.ValidForSeconds(DefaultPeriod)
	if remaining != 30 {
		t.Errorf("expected 30 seconds remaining, got %d", remaining)
	}
}

func TestGenerateCode_SHA256_ProducesCode(t *testing.T) {
	gen := NewGenerator(fixedClock{t: time.Unix(59, 0).UTC()})
	code, err := gen.GenerateCode(Params{
		Secret:    rfc6238SecretSHA1,
		Algorithm: AlgorithmSHA256,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(code) != 6 {
		t.Errorf("expected 6 digit code, got %d digits", len(code))
	}
}

func TestGenerateCode_SHA512_ProducesCode(t *testing.T) {
	gen := NewGenerator(fixedClock{t: time.Unix(59, 0).UTC()})
	code, err := gen.GenerateCode(Params{
		Secret:    rfc6238SecretSHA1,
		Algorithm: AlgorithmSHA512,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(code) != 6 {
		t.Errorf("expected 6 digit code, got %d digits", len(code))
	}
}

func TestValidForSeconds_ZeroPeriod_DefaultsTo30(t *testing.T) {
	gen := NewGenerator(fixedClock{t: time.Unix(45, 0).UTC()})
	remaining := gen.ValidForSeconds(0)
	if remaining != 15 {
		t.Errorf("expected 15 seconds remaining with default period, got %d", remaining)
	}
}

func TestValidate_InvalidSecret_ReturnsError(t *testing.T) {
	gen := NewGenerator(fixedClock{t: time.Unix(59, 0).UTC()})
	_, err := gen.Validate(Params{Secret: "!!!bad!!!"}, "000000")
	if err == nil {
		t.Fatal("expected error for invalid secret in Validate")
	}
}

func TestSystemClock_ReturnsUTC(t *testing.T) {
	clock := SystemClock{}
	now := clock.Now()
	if now.Location() != time.UTC {
		t.Errorf("SystemClock should return UTC, got %v", now.Location())
	}
}
