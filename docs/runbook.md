---
title: Runbook
nav_order: 3
---

# Operations Runbook

Poppie is a local-first CLI tool — most "operations" are on your own machine.
This runbook covers distribution, troubleshooting, and future hosted scenarios.

## Installation

### From source

```bash
go install github.com/BarkingIguana/poppie/cmd/poppie@latest
```

### From release binary

```bash
# Download from GitHub releases
curl -sL https://github.com/BarkingIguana/poppie/releases/latest/download/poppie-darwin-arm64 -o poppie
chmod +x poppie
mv poppie /usr/local/bin/
```

### Homebrew (planned)

```bash
brew install BarkingIguana/tap/poppie
```

## Server Management

```bash
# Start the server (foreground)
poppie server start

# Start as background daemon
poppie server start --daemon

# Check server status
poppie server status

# Stop the server
poppie server stop
```

The server listens on a Unix socket at `~/.config/poppie/poppie.sock` by default.

## Troubleshooting

### Server won't start — "address already in use"

**Symptoms**: `poppie server start` fails with bind error.
**Cause**: Stale socket file from a crashed server.
**Fix**:
1. Check if a server is actually running: `poppie server status`
2. If not, remove the stale socket: `rm ~/.config/poppie/poppie.sock`
3. Restart: `poppie server start`

### Codes are wrong / not accepted

**Symptoms**: Generated codes are rejected by the service.
**Cause**: Clock skew — TOTP requires UTC time accurate to ~30 seconds.
**Fix**:
1. Check your system clock: `date -u`
2. Sync if needed: `sudo sntp -sS time.apple.com` (macOS)
3. If the secret was stored incorrectly: `poppie delete <key>` and re-store

### Cannot decrypt vault

**Symptoms**: "decryption failed" error on any operation.
**Cause**: Wrong passphrase or corrupted vault file.
**Fix**:
1. Try your passphrase again (check caps lock, keyboard layout)
2. If the vault is corrupted, restore from backup: `cp ~/.config/poppie/secrets.enc.bak ~/.config/poppie/secrets.enc`
3. Last resort: delete the vault and re-provision secrets from your authenticator app

## Data Locations

| File | Purpose |
|------|---------|
| `~/.config/poppie/secrets.enc` | Encrypted TOTP secret vault |
| `~/.config/poppie/secrets.enc.bak` | Automatic vault backup (before writes) |
| `~/.config/poppie/poppie.sock` | Unix socket for gRPC server |
| `~/.config/poppie/config.yaml` | Optional configuration overrides |

## Backup

```bash
# Manual backup
cp ~/.config/poppie/secrets.enc ~/safe-place/poppie-secrets-$(date +%Y%m%d).enc

# Poppie automatically creates .bak before each vault write
```
