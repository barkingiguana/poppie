# Test Philosophy & Structure

## Philosophy

1. **BDD first, TDD underneath**: Cucumber feature files describe what poppie does. Go tests verify how it does it. Start with a Gherkin scenario, then write the unit tests to make it pass.
2. **Red-Green-Refactor**: Write the failing test first. Make it pass with the simplest code. Then refactor. Never skip Red.
3. **Tests are documentation**: A test should tell you what the code does and what it promises. Feature files are readable by anyone. Go tests use `Test<What>_<When>_<Then>`.
4. **Test behaviour, not implementation**: Test what the function returns or causes, not how it does it internally. This lets you refactor freely.
5. **Fast feedback first**: Unit tests run in milliseconds. BDD tests may take a few seconds. Keep the fast ones fast.

## Directory Layout

| Directory | What Lives Here | Speed | When to Use |
|-----------|----------------|-------|-------------|
| `features/` | Gherkin feature files | N/A | Describing user-visible behaviour |
| `features/steps/` | godog step definitions | < 5s | Implementing Gherkin steps |
| `*_test.go` (in-package) | Unit tests, table-driven | < 1s | Testing internal logic |
| `tests/integration/` | Tests needing real file system, server | < 30s | Testing encrypted storage, gRPC |
| `tests/e2e/` | Full CLI invocation tests | < 1m | Testing the binary end-to-end |

## Running Tests

```bash
make check           # Everything: lint + unit tests + BDD
make test            # Go unit tests only (fast)
make bdd             # Cucumber/BDD feature specs
make test-coverage   # Unit tests with coverage report

# Run a specific test
go test -run TestGenerateCode ./internal/totp/

# Run a specific feature
godog run features/store_secret.feature

# Run with race detection
go test -race ./...
```

## Writing Good Tests

### BDD / Gherkin

```gherkin
Feature: Store a TOTP secret
  As a developer
  I want to store a TOTP secret with a key
  So that I can retrieve codes later

  Scenario: Store a valid TOTP secret
    Given the poppie server is running
    When I store a secret "JBSWY3DPEHPK3PXP" with key "example.com"
    Then the secret should be stored successfully
    And I should receive a valid 6-digit code

  Scenario: Reject an invalid TOTP secret
    Given the poppie server is running
    When I store a secret "not-valid-base32" with key "example.com"
    Then I should receive an error "invalid TOTP secret"
```

### Go Unit Tests (Table-Driven)

```go
func TestGenerateCode_ValidSecret_Returns6Digits(t *testing.T) {
    tests := []struct {
        name   string
        secret string
        digits int
    }{
        {"default 6 digits", "JBSWY3DPEHPK3PXP", 6},
        {"8 digit code", "JBSWY3DPEHPK3PXP", 8},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            code, err := totp.GenerateCode(tt.secret, tt.digits, time.Now().UTC())
            if err != nil {
                t.Fatalf("unexpected error: %v", err)
            }
            if len(code) != tt.digits {
                t.Errorf("got %d digits, want %d", len(code), tt.digits)
            }
        })
    }
}
```

### What NOT to Test

- Protobuf-generated code (it's already tested by Google)
- cobra command wiring (test the handlers, not the framework)
- Exact log messages (brittle, low value)
- `crypto/aes` or `crypto/hmac` internals (test your usage of them, not the stdlib)
