# ADR-002: Go + gRPC for CLI and Server

## Status

Accepted

## Date

2026-03-14

## Context

Poppie needs to be both a CLI tool for human interaction and a fast server that other
tools (like `dm` from delivery-machine) can query for TOTP codes programmatically.

Requirements:
- Sub-millisecond code retrieval for tool integrations
- Single binary distribution (no runtime dependencies)
- Type-safe API that works across languages (dm is Python, poppie is its own thing)
- Developer joy — the tech should be fun to build with

## Decision

Use Go for both the CLI and server. Use gRPC with Protocol Buffers for the server API.
Use cobra for CLI command structure.

## Consequences

### Positive

- **Single binary**: `go build` produces one file. No Python venvs, no node_modules.
  Easy to distribute via GitHub releases and Homebrew.
- **Fast startup**: Go binaries start in milliseconds. CLI feels instant.
- **gRPC performance**: Persistent connections, binary serialisation, and HTTP/2
  multiplexing give sub-millisecond latency for `GetCode` calls.
- **Protobuf is fun**: Schema-first API design, generated client stubs in any language,
  versioned contracts. Tech geekyness achieved.
- **Cross-platform**: One codebase builds for macOS (ARM + x86), Linux, Windows.
- **Excellent concurrency**: Goroutines and channels make the server straightforward.

### Negative

- **Different language from dm**: dm is Python. Integration requires a Python gRPC
  client or shelling out to the poppie CLI. We accept this because the protobuf
  contract makes cross-language integration clean.
- **Less rapid prototyping**: Go requires more boilerplate than Python for quick
  experiments. We accept this because the project has a clear scope.
- **Proto toolchain**: Requires `protoc` and Go plugins installed. Mitigated by
  `make install` handling setup automatically.

### Neutral

- Go's error handling is verbose but explicit — aligns with project value of
  "prefer explicit over clever".
- Cucumber/BDD support in Go (godog) is mature but less ecosystem support than
  Ruby/JS. Adequate for our needs.

## Alternatives Considered

### Alternative 1: Python (matching dm's stack)

- **Pros**: Same language as dm, faster prototyping, Craig's existing expertise
- **Cons**: Slower startup, requires runtime (venv), harder to distribute,
  no single binary
- **Why rejected**: Distribution complexity and startup latency matter for a tool
  that other tools call frequently. The protobuf contract eliminates the language
  mismatch concern.

### Alternative 2: Rust

- **Pros**: Even faster, memory safe, excellent CLI ecosystem (clap)
- **Cons**: Steeper learning curve, slower compile times, smaller gRPC ecosystem
- **Why rejected**: Go hits the sweet spot of performance, ecosystem maturity,
  and developer ergonomics for this project's scope.
