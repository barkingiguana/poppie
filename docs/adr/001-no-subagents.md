# ADR-001: Use CLAUDE.md + AGENTS.md Instead of Custom Subagents/Skills

## Status

Accepted

## Date

2026-03-13

## Context

Claude Code supports custom subagents and skills that can extend its capabilities.
We needed to decide whether to invest in custom tooling or rely on built-in features
with well-structured context files.

Our projects span Python packages, serverless SaaS, iOS apps, and enterprise monorepos.
Any approach needs to work across all of these without per-project maintenance burden.

## Decision

Use CLAUDE.md (~100 lines) and AGENTS.md (~120 lines) for project context and code
review standards. Do not create custom subagents or skill definitions.

## Consequences

### Positive

- **Lower token cost**: CLAUDE.md loads every session at ~100 lines. Custom skill
  definitions would add overhead on every invocation.
- **Zero maintenance**: CLAUDE.md and AGENTS.md are plain Markdown. No schema changes,
  no API compatibility, no version pinning.
- **Portable**: Works identically across all project types. Copy the files, delete
  what doesn't apply.
- **Transparent**: Anyone can read and understand the configuration. No hidden behaviour
  in skill definitions.
- **Built-in skills are good enough**: Claude Code's built-in commit, review, and
  exploration skills handle the common cases well.

### Negative

- **No automated enforcement**: AGENTS.md is advisory. A custom skill could enforce
  rules programmatically. We accept this trade-off because CI handles enforcement.
- **Less specialised**: A custom skill could have deep domain knowledge. We mitigate
  this by keeping CLAUDE.md dense and specific.

### Neutral

- This decision can be revisited if Claude Code's skill ecosystem matures and the
  maintenance cost drops significantly.

## Alternatives Considered

### Alternative 1: Custom Subagents per Project Type

- **Pros**: Could encode deep project-specific logic (e.g. "always run migrations
  before tests in the SaaS project")
- **Cons**: Maintenance burden across 8+ projects, token overhead, fragile to Claude
  Code updates
- **Why rejected**: The maintenance cost exceeds the benefit. CLAUDE.md's "Known Gotchas"
  section handles project-specific guidance adequately.

### Alternative 2: Shared Skill Library

- **Pros**: Write once, use everywhere
- **Cons**: Versioning across projects, testing burden, risk of over-engineering
- **Why rejected**: A 100-line CLAUDE.md achieves 90% of the value at 10% of the cost.
