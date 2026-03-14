---
title: "005: SDK Version Negotiation via gRPC Metadata"
parent: ADRs
nav_order: 5
---

# ADR-005: SDK Version Negotiation via gRPC Metadata

## Status

Accepted

## Date

2026-03-14

## Context

Poppie provides Go and Python SDKs for programmatic access. As the server evolves,
we need a way to warn SDK clients when their version is outdated or deprecated,
without breaking backward compatibility.

Options considered:
1. Add a `Negotiate` RPC to the proto definition
2. Add version fields to every request/response message
3. Use gRPC metadata headers on every call

## Decision

Use gRPC metadata headers for version negotiation. Clients send
`x-poppie-sdk-version` and `x-poppie-sdk-name` on every call. The server
replies with `x-poppie-version-status` (`supported`, `deprecated`, or
`unknown`) and optionally `x-poppie-deprecation-message`.

This is implemented as a unary server interceptor and a client-side
interceptor in each SDK.

## Consequences

### Positive

- **No proto changes**: Version negotiation stays out of the service contract.
  Older clients that don't send headers still work — they get `unknown` status.
- **No extra RPCs**: Negotiation piggybacks on every call, no `Negotiate` roundtrip.
- **SDK-agnostic**: Any language's gRPC client can participate by setting two headers.
- **Deprecation path**: Server can signal deprecation without rejecting requests,
  giving clients time to upgrade.

### Negative

- **Headers are invisible**: Developers debugging with `grpcurl` won't see the
  negotiation unless they inspect response headers. Documented in SDK guides.
- **Per-call overhead**: Minor — two metadata pairs per request/response.

### Neutral

- If we later need to block old clients (not just warn), the interceptor can
  return an error instead of proceeding. This is a policy change, not an
  architecture change.

## Version Comparison Strategy

Currently: major version comparison only. If the client's major version is
behind the server's, the status is `deprecated`. This will be refined as the
version scheme matures.

A single `VERSION` file at the repository root drives all version strings —
server (via ldflags), Go SDK (constant), and Python SDK (generated file).
