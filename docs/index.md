---
title: Home
nav_order: 1
---

# Poppie

A CLI-based TOTP manager with a fast gRPC server. Store TOTP secrets, validate codes, and serve fresh codes to other tools automatically.
{: .fs-6 .fw-300 }

---

## Why Poppie?

Authenticator apps are great for humans, but terrible for automation. Poppie bridges the gap — it stores your TOTP secrets securely and serves codes programmatically to tools that need them (like [`dm`](https://github.com/BarkingIguana/delivery-machine)).

```bash
# Store a TOTP secret
poppie store --label deliverymachine.net --secret JBSWY3DPEHPK3PXP

# Get a fresh code
poppie get --label deliverymachine.net
# => 483921 (valid for 18s)
```

## Features

- **CLI interface** — human-friendly commands for managing secrets and codes
- **gRPC server** — sub-millisecond code retrieval for tool integrations
- **Encrypted storage** — AES-256-GCM with argon2id key derivation
- **Protobuf API** — type-safe, versioned, language-agnostic integration
- **Official SDKs** — [Go]({% link sdk/go.md %}) and [Python]({% link sdk/python.md %}) clients with version negotiation
- **Zero-config start** — sensible defaults, configure only what you need

## Quick start

```bash
# Install from source
go install github.com/BarkingIguana/poppie/cmd/poppie@latest

# Start the server
export POPPIE_PASSPHRASE="your-vault-passphrase"
poppie server start

# Store a secret
poppie store --label github.com --secret YOUR_BASE32_SECRET

# Get a code
poppie get --label github.com
```

## Integration with `dm`

Poppie is designed to work with delivery-machine's `dm` command:

```bash
# dm stores the TOTP secret during setup
dm signup user@example.com  # → provisions TOTP → stores in poppie

# dm retrieves codes automatically
dm users verify user@example.com  # → asks poppie for code → submits it
```

See the [integration guide]({% link integration-dm.md %}) for examples, or use the [Python SDK]({% link sdk/python.md %}) directly.
