---
title: Versioning
parent: SDKs
nav_order: 3
---

# Version Negotiation Protocol

## How it works

Every gRPC call from an SDK includes two metadata headers:

| Header | Example | Description |
|--------|---------|-------------|
| `x-poppie-sdk-version` | `0.1.0` | SDK version (semver) |
| `x-poppie-sdk-name` | `poppie-go` | SDK identifier |

The server responds with:

| Header | Values | Description |
|--------|--------|-------------|
| `x-poppie-version-status` | `supported`, `deprecated`, `unknown` | Version assessment |
| `x-poppie-deprecation-message` | (string) | Human-readable message (only when deprecated) |

## Status meanings

| Status | Meaning | Action |
|--------|---------|--------|
| `supported` | Your SDK version is compatible | None needed |
| `deprecated` | Your SDK still works but is outdated | Upgrade when convenient |
| `unknown` | Server couldn't determine status (e.g. no headers sent) | Consider upgrading |

## Version comparison

Currently, the server compares **major versions** only. If your SDK's major version is behind the server's, you get `deprecated`.

As poppie is pre-1.0 (`0.x.y`), all current versions share major version `0`, so you won't see deprecation warnings until 1.0 is released.

## Deprecation policy

1. A new major version is released
2. The previous major version is marked `deprecated` — calls still work, but SDK warning handlers fire
3. After a reasonable migration period, the server _may_ start rejecting old clients (not yet implemented)

## Single VERSION file

All version strings are driven from a single `VERSION` file at the repository root:

- **Server**: injected via Go ldflags at build time
- **Go SDK**: `SDKVersion` constant in `sdk/go/version.go`
- **Python SDK**: `SDK_VERSION` in `sdk/python/src/poppie/_version.py`

When releasing, update the `VERSION` file and regenerate SDK version files.

## Clients without SDKs

Tools that call the gRPC API directly (e.g. `grpcurl`) don't send version headers. The server returns `unknown` status, and everything works normally. Version negotiation is entirely opt-in.
