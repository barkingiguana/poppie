---
title: Integration Guide
nav_order: 4
---

# Integrating poppie with your application

This guide shows how to connect your CLI tool or web app to poppie for automated
TOTP code retrieval, replacing manual copy-paste from an authenticator app.

## Overview

```
your-app signup / your-app verify
        │
        ├─ stores TOTP secret ──► poppie store --label <domain> --secret <base32>
        │
        └─ needs a code ────────► poppie get --label <domain>
                                      │
                                      └─ returns 6-digit code to stdout
```

## Option A: Python SDK (recommended)

The official Python SDK wraps the gRPC API with a clean interface and version negotiation.

```bash
pip install poppie-sdk
```

```python
from poppie import PoppieClient

with PoppieClient() as client:
    # Store a secret during onboarding
    label, verification = client.store_secret("myapp.example.com", secret)

    # Later, retrieve a fresh code
    code, valid_for = client.get_code("myapp.example.com")
```

See the [Python SDK docs]({% link sdk/python.md %}) for full API reference.

### Usage in your app

```python
# During signup, after receiving the TOTP provisioning URI:
secret = parse_totp_uri(provisioning_uri)  # extract base32 secret
client.store_secret("myapp.example.com", secret)

# Later, when you need to verify:
code, _ = client.get_code("myapp.example.com")
response = api.post("/auth/totp/verify", json={"code": code})
```

## Option B: Shell out to the poppie CLI

The simplest integration. Your app calls the `poppie` binary as a subprocess.

### Python example

```python
import subprocess


def store_totp_secret(label: str, secret: str) -> str:
    """Store a TOTP secret in poppie and return the verification code."""
    result = subprocess.run(
        ["poppie", "store", "--label", label, "--secret", secret],
        capture_output=True, text=True, check=True,
    )
    # Output: Stored "myapp.example.com" — verification code: 123456
    return result.stdout.strip().split(": ")[-1]


def get_totp_code(label: str) -> str:
    """Get a current TOTP code from poppie."""
    result = subprocess.run(
        ["poppie", "get", "--label", label],
        capture_output=True, text=True, check=True,
    )
    # Output: 123456 (valid for 18s)
    return result.stdout.strip().split(" ")[0]
```

### Pros and cons

- **Pro**: Zero dependencies — just needs `poppie` on `$PATH`
- **Pro**: Works immediately, no Python gRPC setup
- **Con**: Subprocess overhead (~5ms vs ~0.1ms for gRPC)
- **Verdict**: Good enough for human-speed operations

## Option C: Go SDK

For Go services and CLI tools.

```bash
go get github.com/BarkingIguana/poppie/sdk/go
```

```go
client, err := poppie.New(ctx)
if err != nil {
    log.Fatal(err)
}
defer client.Close()

code, err := client.GetCode(ctx, "myapp.example.com")
fmt.Println(code.Code)
```

See the [Go SDK docs]({% link sdk/go.md %}) for full API reference.

## Prerequisites

All options require a running poppie server:

```bash
# Start the server (will prompt or use env var for passphrase):
export POPPIE_PASSPHRASE="your-vault-passphrase"
poppie server start --daemon

# Verify it's running:
poppie server status
```

## Recommended integration path

1. Start with **Option A** (Python SDK) — clean API, version negotiation built in
2. Fall back to **Option B** (subprocess) if you want zero Python dependencies
3. Add a poppie availability check in your app's startup
4. Fall back to manual code entry if poppie isn't running

## Error handling

```python
import subprocess


def get_totp_code_safe(label: str) -> str | None:
    """Get a TOTP code, returning None if poppie is unavailable."""
    try:
        result = subprocess.run(
            ["poppie", "get", "--label", label],
            capture_output=True, text=True, timeout=5,
        )
        if result.returncode != 0:
            return None
        return result.stdout.strip().split(" ")[0]
    except (FileNotFoundError, subprocess.TimeoutExpired):
        return None
```
