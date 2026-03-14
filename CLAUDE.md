# Poppie

## Purpose

Poppie is a CLI-based TOTP manager with a persistent gRPC server.
It stores TOTP secrets, validates codes, and serves fresh codes to other tools
(e.g. `dm` from delivery-machine) on demand — fast enough to feel instant.

## Dev Environment

- Language: Go 1.22+
- Package manager: Go modules
- Key dependencies: cobra (CLI), gRPC + protobuf (server API), godog (BDD/Cucumber)
- Infrastructure: Local-first (encrypted file store). Future: optional AWS Lambda relay.

## Project Structure

```
cmd/poppie/           # CLI entrypoint (cobra)
internal/server/      # gRPC server implementation (+ version interceptor)
internal/totp/        # TOTP generation and validation logic
internal/store/       # Secret storage (encrypted local file)
internal/auth/        # JWT validation for server access
proto/                # Protobuf service definitions
sdk/go/               # Go SDK (client library with version negotiation)
sdk/python/           # Python SDK (client library with version negotiation)
features/             # Cucumber/BDD feature files (Gherkin)
features/steps/       # Step definitions (godog)
tests/                # Additional Go tests
docs/                 # Architecture, runbook, ADRs, SDK docs
```

## Key Architecture Decisions

- Go + gRPC for CLI and server (see docs/adr/002-go-grpc.md)
- BDD with Cucumber/godog driving development (see docs/adr/003-bdd-cucumber.md)
- Encrypted local file storage for TOTP secrets (see docs/adr/004-local-encrypted-storage.md)
- CLAUDE.md + AGENTS.md over custom subagents (see docs/adr/001-no-subagents.md)
- SDK version negotiation via gRPC metadata (see docs/adr/005-sdk-version-negotiation.md)

## Commands

```bash
make install           # Set up local dev environment (go mod download, tools)
make check             # Run all quality checks (lint + test + bdd + SDK tests)
make test              # Run Go unit tests
make bdd               # Run Cucumber/BDD feature specs
make sdk-go-test       # Run Go SDK tests
make sdk-python-test   # Run Python SDK tests
make lint              # Run golangci-lint
make format            # Run gofmt + goimports
make proto             # Regenerate protobuf/gRPC code (Go + Python)
make build             # Build poppie binary
make run               # Run poppie server locally
```

## Code Style

- Handle every error — no `_` for error returns unless explicitly justified
- Use `context.Context` as first parameter for functions that do IO
- Prefer table-driven tests for unit tests, Cucumber for behaviour
- Structured logging via `slog` — never `fmt.Println` for operational output
- All public types and functions need doc comments

## Testing

- Run: `make test` (unit) / `make bdd` (behaviour) / `make check` (all)
- Coverage target: 80% (enforced in CI)
- BDD naming: feature files describe user-visible behaviour in Gherkin
- Unit test naming: `Test<What>_<When>_<Then>` (e.g. `TestGenerateCode_ValidSecret_Returns6Digits`)
- TDD workflow: Red-Green-Refactor. Write the failing test first.

## What's Built

- Protobuf service definition with all RPC methods and messages
- RFC 6238 TOTP engine (SHA1/SHA256/SHA512, configurable digits/period)
- Encrypted vault storage (AES-256-GCM, argon2id KDF, atomic writes, auto-backup)
- gRPC server with Unix socket binding
- CLI commands: store, get, list, delete, server start/stop/status
- BDD feature specs (10 scenarios, 42 steps) covering all operations
- Unit tests for TOTP engine (RFC 6238 vectors) and store
- Go SDK (`sdk/go/`) — client library with version negotiation
- Python SDK (`sdk/python/`) — client library with version negotiation, committed proto stubs
- Version negotiation via gRPC metadata interceptor
- Single VERSION file driving all version strings
- go.work for local multi-module development

## What's In Progress / Left To Do

- Homebrew formula for distribution

## Known Gotchas

- TOTP codes are time-sensitive — always use UTC, never local time
- Protobuf generated code must not be edited — regenerate with `make proto` (generates both Go and Python stubs)
- Python proto stubs need import path fix after generation (handled by `make proto`)
- SDK version constants must be kept in sync with the root VERSION file
- The gRPC server needs a Unix socket or localhost-only binding for security
- Secret encryption key derivation must use a proper KDF (argon2id), not raw passwords
