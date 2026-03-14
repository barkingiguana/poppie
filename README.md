# Poppie

A CLI-based TOTP manager with a fast gRPC server. Store TOTP secrets, validate codes, and serve fresh codes to other tools automatically.

## Why Poppie?

Authenticator apps are great for humans, but terrible for automation. Poppie bridges the gap — it stores your TOTP secrets securely and serves codes programmatically to tools that need them (like [`dm`](https://github.com/BarkingIguana/delivery-machine)).

```bash
# Store a TOTP secret
poppie store --key deliverymachine.net --secret JBSWY3DPEHPK3PXP

# Get a fresh code
poppie get deliverymachine.net
# => 483921

# Other tools can request codes via gRPC — instant, no shell overhead
```

## Features

- **CLI interface** — human-friendly commands for managing secrets and codes
- **gRPC server** — sub-millisecond code retrieval for tool integrations
- **Encrypted storage** — secrets encrypted at rest with argon2id key derivation
- **Protobuf API** — type-safe, versioned, language-agnostic integration
- **Zero-config start** — sensible defaults, configure only what you need

## Quick Start

```bash
# Install
go install github.com/BarkingIguana/poppie/cmd/poppie@latest
# or: brew install BarkingIguana/tap/poppie (coming soon)

# Store a secret
poppie store --key myservice.example.com --secret YOUR_TOTP_SECRET

# Get a code
poppie get myservice.example.com

# Start the server (for tool integrations)
poppie server start
```

## Integration with `dm`

Poppie is designed to work with delivery-machine's `dm` command:

```bash
# dm stores the TOTP secret during setup
dm signup user@example.com  # → provisions TOTP → stores in poppie

# dm retrieves codes automatically
dm users verify user@example.com  # → asks poppie for code → submits it
```

## Development

```bash
make install    # Download dependencies + install tools
make check      # Run all quality checks
make test       # Unit tests
make bdd        # BDD/Cucumber specs
make proto      # Regenerate protobuf code
make build      # Build binary
make help       # Show all targets
```

## Architecture

See [docs/architecture.md](docs/architecture.md) for system design and [docs/adr/](docs/adr/) for decision records.

## License

TBD
