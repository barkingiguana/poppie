---
title: SDKs
nav_order: 3
has_children: true
---

# Poppie SDKs

Official client libraries for integrating with the poppie TOTP manager. Both SDKs wrap the gRPC API and add version negotiation, so the server can warn you when your SDK is outdated.

## Which SDK?

| | Go SDK | Python SDK |
|---|--------|------------|
| **Install** | `go get github.com/BarkingIguana/poppie/sdk/go` | `pip install poppie-sdk` |
| **Best for** | Go services, CLI tools | Python scripts, `dm` integration |
| **Version negotiation** | Built-in | Built-in |
| **Context manager** | N/A (`defer client.Close()`) | `with PoppieClient() as client:` |

## Quick comparison

### Go

```go
client, err := poppie.New(ctx)
code, err := client.GetCode(ctx, "github.com")
fmt.Println(code.Code, code.ValidForSeconds)
```

### Python

```python
with PoppieClient() as client:
    code, valid_for = client.get_code("github.com")
    print(code, valid_for)
```

## Version negotiation

Both SDKs automatically send version headers on every gRPC call. The server responds with a status (`supported`, `deprecated`, or `unknown`) and an optional message. See [versioning]({% link sdk/versioning.md %}) for details.
