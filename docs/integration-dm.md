---
title: dm Integration
nav_order: 4
---

# Integrating poppie with delivery-machine's `dm`

This guide shows how to connect `dm` (Python CLI) to poppie for automated TOTP
code retrieval, replacing manual copy-paste from an authenticator app.

## Overview

```
dm signup / dm users verify
        │
        ├─ stores TOTP secret ──► poppie store --label <domain> --secret <base32>
        │
        └─ needs a code ────────► poppie get --label <domain>
                                      │
                                      └─ returns 6-digit code to stdout
```

## Option A: Shell out to the poppie CLI

The simplest integration. `dm` calls the `poppie` binary as a subprocess.

### Python example

```python
import subprocess


def store_totp_secret(label: str, secret: str) -> str:
    """Store a TOTP secret in poppie and return the verification code."""
    result = subprocess.run(
        ["poppie", "store", "--label", label, "--secret", secret],
        capture_output=True, text=True, check=True,
    )
    # Output: Stored "github.com" — verification code: 123456
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

### Usage in dm

```python
# In dm's signup flow, after receiving the TOTP provisioning URI:
secret = parse_totp_uri(provisioning_uri)  # extract base32 secret
store_totp_secret("delivery-machine", secret)

# Later, when dm needs to verify:
code = get_totp_code("delivery-machine")
response = api.post("/auth/totp/verify", json={"code": code})
```

### Pros and cons

- **Pro**: Zero dependencies — just needs `poppie` on `$PATH`
- **Pro**: Works immediately, no Python gRPC setup
- **Con**: Subprocess overhead (~5ms vs ~0.1ms for gRPC)
- **Verdict**: Good enough for `dm`'s use case (human-speed operations)

## Option B: gRPC client (Python)

For tighter integration or high-frequency calls.

### Setup

```bash
pip install grpcio grpcio-tools

# Generate Python stubs from poppie's proto:
python -m grpc_tools.protoc \
    -I/path/to/poppie/proto \
    --python_out=./dm/generated \
    --grpc_python_out=./dm/generated \
    poppie/poppie.proto
```

### Python example

```python
import grpc
from dm.generated.poppie import poppie_pb2, poppie_pb2_grpc


def get_poppie_client() -> poppie_pb2_grpc.PoppieServiceStub:
    """Connect to the poppie gRPC server via Unix socket."""
    channel = grpc.insecure_channel("unix:///Users/you/.config/poppie/poppie.sock")
    return poppie_pb2_grpc.PoppieServiceStub(channel)


def store_totp_secret(label: str, secret: str) -> str:
    """Store a TOTP secret and return the verification code."""
    client = get_poppie_client()
    response = client.StoreSecret(poppie_pb2.StoreSecretRequest(
        label=label,
        secret=secret,
    ))
    return response.verification_code


def get_totp_code(label: str) -> str:
    """Get a current TOTP code."""
    client = get_poppie_client()
    response = client.GetCode(poppie_pb2.GetCodeRequest(label=label))
    return response.code
```

### Pros and cons

- **Pro**: Sub-millisecond latency, type-safe API
- **Con**: Requires `grpcio` dependency and generated stubs in dm
- **Verdict**: Use this if dm starts making many TOTP calls or needs the speed

## Prerequisites

Both options require a running poppie server:

```bash
# Start the server (will prompt or use env var for passphrase):
export POPPIE_PASSPHRASE="your-vault-passphrase"
poppie server start --daemon

# Verify it's running:
poppie server status
```

## Option C: Python SDK (recommended)

The official Python SDK wraps the gRPC API with a clean interface and version negotiation.

```bash
pip install poppie-sdk
```

```python
from poppie import PoppieClient

with PoppieClient() as client:
    label, verification = client.store_secret("delivery-machine", secret)
    code, valid_for = client.get_code("delivery-machine")
```

See the [Python SDK docs]({% link sdk/python.md %}) for full API reference.

## Recommended integration path

1. Start with **Option C** (Python SDK) — clean API, version negotiation built in
2. Fall back to **Option A** (subprocess) if you want zero Python dependencies
3. Add a poppie availability check in `dm`'s startup
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
