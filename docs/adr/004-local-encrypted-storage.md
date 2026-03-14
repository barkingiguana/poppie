# ADR-004: Encrypted Local File Storage for TOTP Secrets

## Status

Accepted

## Date

2026-03-14

## Context

Poppie stores TOTP secrets — the most sensitive data in any 2FA system. A compromised
secret lets an attacker generate valid codes indefinitely. We need storage that is:
- Secure at rest (encrypted)
- Local-first (no cloud dependency, no network calls)
- Fast to read (sub-millisecond for code generation)
- Simple to back up and restore
- Zero-cost infrastructure

## Decision

Store all TOTP secrets in a single encrypted file (`~/.config/poppie/secrets.enc`)
using AES-256-GCM for authenticated encryption. Derive the encryption key from a
user passphrase using argon2id (memory-hard KDF).

Design:
- **Format**: The vault is a single encrypted blob. Decrypted, it's a protobuf-encoded
  list of Secret messages.
- **Key derivation**: argon2id with parameters tuned for ~0.5s on modern hardware.
  Salt stored alongside the encrypted data (not secret).
- **Backup**: Before each write, copy the existing vault to `.bak`. Atomic write
  via temp file + rename.
- **Unlocking**: The server decrypts the vault on startup and holds secrets in memory.
  The passphrase is not retained.

## Consequences

### Positive

- **Zero infrastructure cost**: No database, no cloud service, no API keys.
- **Simple backup**: One file to copy. Encrypted, so safe to store in cloud backup.
- **Fast reads**: Secrets held in memory after unlock. Code generation is pure compute.
- **Portable**: Works on any OS. No keychain integration required (but could add later).
- **Auditable**: Single encryption scheme, well-understood primitives.

### Negative

- **Single-machine**: Secrets don't sync across devices automatically. Acceptable
  for a dev tool — you typically use poppie on one machine.
- **Memory exposure**: Decrypted secrets live in server process memory. Mitigated by
  Unix socket access control and optional JWT auth.
- **Passphrase UX**: User needs to enter passphrase on server start. Could integrate
  with macOS Keychain or system keyring later.

### Neutral

- Vault format uses protobuf internally — consistent with the gRPC API.
- The argon2id parameters should be documented and tuneable via config.

## Alternatives Considered

### Alternative 1: macOS Keychain / system keyring

- **Pros**: OS-managed security, biometric unlock, no passphrase to remember
- **Cons**: Platform-specific, harder to test, harder to backup/restore, no Linux parity
- **Why rejected**: Could be added as an optional backend later. Starting with
  cross-platform file encryption keeps scope manageable.

### Alternative 2: SQLite with SQLCipher

- **Pros**: Structured queries, battle-tested encryption, concurrent access
- **Cons**: Extra dependency (CGo for SQLCipher), overkill for <1000 secrets,
  harder to distribute as single binary
- **Why rejected**: A TOTP vault is a simple key-value store. SQLite adds complexity
  without proportional benefit.

### Alternative 3: AWS Secrets Manager / DynamoDB

- **Pros**: Cloud-native, managed encryption, multi-device sync
- **Cons**: Network dependency, latency, cost (even if small), requires AWS account
- **Why rejected**: Violates local-first principle. Poppie should work offline.
  Could be added as an optional sync backend in the future.
