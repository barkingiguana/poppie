---
title: Go SDK
parent: SDKs
nav_order: 1
---

# Go SDK

## Install

```bash
go get github.com/BarkingIguana/poppie/sdk/go
```

## Quick start

```go
package main

import (
    "context"
    "fmt"
    "log"

    poppie "github.com/BarkingIguana/poppie/sdk/go"
)

func main() {
    ctx := context.Background()

    client, err := poppie.New(ctx)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Store a secret
    result, err := client.StoreSecret(ctx, "github.com", "JBSWY3DPEHPK3PXP")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Stored %s, verification: %s\n", result.Label, result.VerificationCode)

    // Get a code
    code, err := client.GetCode(ctx, "github.com")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Code: %s (valid for %ds)\n", code.Code, code.ValidForSeconds)
}
```

## API reference

### `New(ctx, ...Option) (*Client, error)`

Creates a new client connected to the poppie server.

```go
// Default socket path (~/.config/poppie/poppie.sock)
client, err := poppie.New(ctx)

// Custom socket path
client, err := poppie.New(ctx, poppie.WithSocketPath("/tmp/poppie.sock"))

// Custom warning handler
client, err := poppie.New(ctx, poppie.WithWarningHandler(func(w poppie.VersionWarning) {
    log.Printf("poppie: %s — %s", w.Status, w.Message)
}))
```

### `Client.StoreSecret(ctx, label, secret, ...StoreOption) (*StoreSecretResult, error)`

Stores a TOTP secret. Returns the label and a verification code.

```go
result, err := client.StoreSecret(ctx, "github.com", "JBSWY3DPEHPK3PXP",
    poppie.WithAlgorithm(poppie.SHA256),
    poppie.WithDigits(8),
    poppie.WithPeriod(60),
)
```

### `Client.GetCode(ctx, label) (*GetCodeResult, error)`

Generates a current TOTP code.

```go
code, err := client.GetCode(ctx, "github.com")
// code.Code: "483921"
// code.ValidForSeconds: 18
```

### `Client.ListSecrets(ctx) ([]string, error)`

Returns labels of all stored secrets.

### `Client.DeleteSecret(ctx, label) (bool, error)`

Deletes a secret. Returns `true` if a secret was actually removed.

### `Client.Close() error`

Closes the gRPC connection.

## Options

| Option | Description | Default |
|--------|-------------|---------|
| `WithSocketPath(path)` | Unix socket path | `~/.config/poppie/poppie.sock` |
| `WithWarningHandler(fn)` | Version warning callback | `DefaultWarningHandler` (logs via slog) |
| `WithAlgorithm(algo)` | HMAC algorithm (`SHA1`, `SHA256`, `SHA512`) | `SHA1` |
| `WithDigits(n)` | Code digit count | `6` |
| `WithPeriod(s)` | Time step in seconds | `30` |

## Warning handling

By default, the SDK logs version warnings via `slog`. You can provide a custom handler or disable warnings:

```go
// Custom handler
client, _ := poppie.New(ctx, poppie.WithWarningHandler(func(w poppie.VersionWarning) {
    fmt.Fprintf(os.Stderr, "WARNING: %s\n", w.Message)
}))

// Disable warnings
client, _ := poppie.New(ctx, poppie.WithWarningHandler(nil))
```

Only `deprecated` status triggers the warning handler. `supported` and `unknown` are silent.
