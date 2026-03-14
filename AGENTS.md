# Code Review Standards

## General (All Languages)

- Every PR must have a clear purpose — one concern per PR
- No commented-out code in production files
- No TODOs without a linked issue
- Error messages must be actionable (say what went wrong AND what to do)
- Prefer explicit over clever — readability beats brevity
- Functions should do one thing. If you need "and" to describe it, split it
- No magic numbers — use named constants

## Go

- Handle every error — no `_` for error returns unless explicitly justified
- Use `context.Context` as first parameter for functions that do IO
- Prefer table-driven tests
- Use `errors.Is` / `errors.As` for error checking, not string comparison
- Struct field tags must be consistent (json, db, yaml)
- No `init()` functions unless absolutely necessary
- Use `slog` for structured logging — never `fmt.Println` for operational output
- Protobuf-generated code lives in its own package, never edited by hand
- Prefer returning errors over panicking — panics are for programmer bugs only
- Use `defer` for cleanup — but be aware of loop gotchas

## Protobuf / gRPC

- All fields must have comments explaining their purpose
- Use well-known types (google.protobuf.Timestamp, Duration) where appropriate
- Service methods should follow the naming: `VerbNoun` (e.g. `StoreSecret`, `GetCode`)
- Breaking changes require a new major version of the proto package

## Infrastructure / IaC

- All infrastructure changes go through code review — no console clicks
- Secrets in AWS Secrets Manager or SSM Parameter Store, never in code or env files
- IAM: least-privilege policies, no wildcard `*` actions in production
- Tag all resources: project, environment, owner

## BDD / Cucumber

- Feature files describe user-visible behaviour, not implementation details
- Scenarios should be understandable by non-developers
- Step definitions should be thin wrappers — business logic lives in the app
- Use scenario outlines for data-driven tests
- One feature file per user-facing capability

## PR Quality Checklist

When reviewing a PR, verify:

- [ ] Tests cover the happy path and at least one error case
- [ ] BDD scenarios updated if user-facing behaviour changed
- [ ] No secrets, credentials, or PII in the diff
- [ ] Breaking changes are called out in the PR description
- [ ] New dependencies are justified (not just for convenience)
- [ ] Error handling is present and produces actionable messages
- [ ] Protobuf changes are backwards-compatible (or version bumped)
