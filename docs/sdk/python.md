---
title: Python SDK
parent: SDKs
nav_order: 2
---

# Python SDK

## Install

```bash
pip install poppie-sdk
```

## Quick start

```python
from poppie import PoppieClient

with PoppieClient() as client:
    # Store a secret
    label, verification = client.store_secret("github.com", "JBSWY3DPEHPK3PXP")
    print(f"Stored {label}, verification: {verification}")

    # Get a code
    code, valid_for = client.get_code("github.com")
    print(f"Code: {code} (valid for {valid_for}s)")
```

## API reference

### `PoppieClient(socket_path=None, warning_handler=None)`

Creates a new client. Supports context manager protocol.

```python
# Default socket path (~/.config/poppie/poppie.sock)
client = PoppieClient()

# Custom socket path
client = PoppieClient(socket_path="/tmp/poppie.sock")

# Custom warning handler
def on_warning(w):
    print(f"poppie: {w.status} — {w.message}")

client = PoppieClient(warning_handler=on_warning)
```

### `store_secret(label, secret, *, algorithm="sha1", digits=0, period=0) -> (str, str)`

Stores a TOTP secret. Returns `(label, verification_code)`.

```python
label, code = client.store_secret(
    "github.com", "JBSWY3DPEHPK3PXP",
    algorithm="sha256", digits=8, period=60,
)
```

### `get_code(label) -> (str, int)`

Generates a current TOTP code. Returns `(code, valid_for_seconds)`.

```python
code, valid_for = client.get_code("github.com")
# code: "483921"
# valid_for: 18
```

### `list_secrets() -> list[str]`

Returns labels of all stored secrets.

### `delete_secret(label) -> bool`

Deletes a secret. Returns `True` if a secret was actually removed.

### `close()`

Closes the gRPC channel. Called automatically when using `with`.

## Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `socket_path` | Unix socket path | `~/.config/poppie/poppie.sock` |
| `warning_handler` | Version warning callback | Logs via `logging` module |
| `algorithm` | `"sha1"`, `"sha256"`, or `"sha512"` | `"sha1"` |
| `digits` | Code digit count (0 = server default) | `0` |
| `period` | Time step in seconds (0 = server default) | `0` |

## Warning handling

By default, the SDK logs version warnings via Python's `logging` module. You can provide a custom handler or disable warnings:

```python
# Custom handler
def on_warning(w):
    print(f"WARNING: {w.message}", file=sys.stderr)

client = PoppieClient(warning_handler=on_warning)

# Disable warnings
client = PoppieClient(warning_handler=None)
```

Only `deprecated` status triggers the warning handler. `supported` and `unknown` are silent.

## VersionWarning

```python
@dataclass
class VersionWarning:
    status: str    # "supported", "deprecated", or "unknown"
    message: str   # Human-readable deprecation message
```
