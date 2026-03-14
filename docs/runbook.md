---
title: Getting Started
nav_order: 2
---

# Getting Started

## Install

### From source

```bash
go install github.com/BarkingIguana/poppie/cmd/poppie@latest
```

### From release binary

```bash
curl -sL https://github.com/BarkingIguana/poppie/releases/latest/download/poppie-darwin-arm64 -o poppie
chmod +x poppie
mv poppie /usr/local/bin/
```

### Homebrew (planned)

```bash
brew install BarkingIguana/tap/poppie
```

## Start the server

```bash
export POPPIE_PASSPHRASE="your-vault-passphrase"
poppie server start --daemon
```

The server listens on a Unix socket at `~/.config/poppie/poppie.sock`.

## Store a secret

```bash
poppie store --label github.com --secret JBSWY3DPEHPK3PXP
# Stored "github.com" — verification code: 483921
```

## Get a code

```bash
poppie get --label github.com
# 483921 (valid for 18s)
```

## Watch codes live

```bash
poppie get --label github.com --label aws.com --live
```

Displays a continuously updating view with countdown bars. Press Ctrl-C to stop.

## Manage secrets

```bash
# List all stored labels
poppie list

# Delete a secret
poppie delete --label github.com
```

## Server management

```bash
poppie server status    # Check if the server is running
poppie server stop      # Stop the server
```

## Troubleshooting

### Server won't start — "address already in use"

A stale socket from a crashed server. Check with `poppie server status`, then remove
the socket if nothing is running:

```bash
rm ~/.config/poppie/poppie.sock
poppie server start
```

### Codes are wrong / not accepted

TOTP requires UTC time accurate to ~30 seconds. Check your system clock:

```bash
date -u
sudo sntp -sS time.apple.com   # macOS
```

## Data locations

| File | Purpose |
|------|---------|
| `~/.config/poppie/secrets.enc` | Encrypted TOTP secret vault |
| `~/.config/poppie/secrets.enc.bak` | Automatic backup (before each write) |
| `~/.config/poppie/poppie.sock` | Unix socket for gRPC server |
