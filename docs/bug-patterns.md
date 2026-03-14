# Common Bug Patterns

## Go

- **Ignored errors**: `result, _ := doThing()` silently drops failures. Handle every error.
- **Goroutine leaks**: Starting a goroutine without a way to stop it leaks memory. Use `context.Context` cancellation.
- **Race conditions**: Shared state without synchronisation. Run tests with `-race`. Use `sync.Mutex` or channels.
- **Nil pointer on interface**: An interface holding a nil concrete value is not `nil` itself. Check the concrete value.
- **Deferred closure in loop**: `defer f.Close()` in a loop defers everything to function exit. Close explicitly or use a helper function.
- **Time zone bugs in TOTP**: `time.Now()` uses local time. TOTP requires UTC. Always use `time.Now().UTC()`.
- **Slice gotchas**: Slicing doesn't copy — mutations to a slice affect the original. Use `slices.Clone()` when independence matters.

## Protobuf / gRPC

- **Breaking changes**: Renaming or renumbering fields breaks existing clients. Only add new fields; deprecate old ones.
- **Zero values vs absent**: Proto3 doesn't distinguish between zero and unset. Use wrapper types or `optional` keyword if the distinction matters.
- **Streaming without deadlines**: gRPC calls without a context deadline can hang forever. Always set timeouts.
- **Large messages**: gRPC has a 4MB default message limit. Fine for TOTP (secrets are tiny) but be aware.

## Cryptography (Poppie-Specific)

- **Using crypto/rand correctly**: Always use `crypto/rand` for key generation, never `math/rand`.
- **Nonce reuse in AES-GCM**: Reusing a nonce with the same key is catastrophic. Generate a fresh random nonce per encryption.
- **Hardcoded argon2id parameters**: KDF parameters should be configurable and stored with the vault, not compiled in.
- **Logging secrets**: Never log TOTP secrets, encryption keys, or derived key material. Scrub before logging.

## Infrastructure / Cloud

- **Forgetting pagination**: AWS APIs paginate by default. Always handle `NextToken` / `LastEvaluatedKey`.
- **IAM policy too broad**: `"Action": "*"` in dev becomes a security incident in prod. Always scope to specific actions and resources.
- **Missing CloudWatch alarms**: If there's no alarm, the failure is silent. Every deployment should include monitoring.

## General (All Projects)

- **Race conditions in tests**: Tests that depend on timing or ordering are flaky. Use deterministic synchronisation.
- **Hardcoded URLs/IDs**: Environment-specific values must come from config, not code.
- **Missing error context**: `return fmt.Errorf("failed")` — failed what? Use `fmt.Errorf("failed to store secret %q: %w", key, err)`.
- **Off-by-one in TOTP time steps**: The TOTP period is 30 seconds by default. `floor(time / period)` not `round(time / period)`.
- **Log secrets**: Accidentally logging API keys, tokens, or TOTP secrets. Scrub sensitive fields before logging.
