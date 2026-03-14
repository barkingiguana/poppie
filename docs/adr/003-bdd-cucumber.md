# ADR-003: BDD with Cucumber (godog) Drives Development

## Status

Accepted

## Date

2026-03-14

## Context

Craig values BDD with Cucumber and TDD with Red-Green-Refactor. We need a testing
strategy that:
- Describes user-visible behaviour in plain language
- Drives implementation from the outside in
- Produces living documentation
- Works well with Go's testing ecosystem

## Decision

Use Cucumber feature files (Gherkin syntax) with godog as the test runner for
behaviour-driven development. Use standard Go tests for unit-level TDD.

The development flow is:
1. Write a Gherkin scenario describing the desired behaviour
2. Run godog — it fails (Red)
3. Implement step definitions that drive the real code
4. Write unit tests for internal logic (Red-Green-Refactor)
5. All green — commit

## Consequences

### Positive

- **Living documentation**: Feature files describe what poppie does in plain English.
  Non-developers can read and validate them.
- **Outside-in development**: Start with user behaviour, work inward to implementation.
  Prevents building the wrong thing.
- **Two-layer testing**: Cucumber for behaviour, Go tests for logic. Each layer tests
  at its appropriate level of abstraction.
- **Refactor confidence**: Behaviour specs don't break when internals change. Only
  break when actual behaviour changes.

### Negative

- **godog setup**: Requires a specific test runner configuration. Mitigated by
  `make bdd` hiding the complexity.
- **Step definition maintenance**: Gherkin steps need Go implementations. Can accumulate
  if not kept DRY.
- **Slower than pure unit tests**: BDD tests exercise more of the stack. Mitigated by
  keeping unit tests fast and BDD focused on key flows.

### Neutral

- Feature files live in `features/`, step definitions in `features/steps/`. Standard
  Cucumber layout.
- Scenario outlines used for data-driven variations of the same behaviour.

## Alternatives Considered

### Alternative 1: Go tests only (table-driven)

- **Pros**: Simpler toolchain, Go-native, faster
- **Cons**: No plain-language specs, harder to communicate behaviour to non-developers,
  no BDD workflow
- **Why rejected**: Craig specifically values BDD and Cucumber. Go tests alone miss
  the communication and documentation benefits.

### Alternative 2: Behave (Python BDD) calling poppie as a subprocess

- **Pros**: More mature BDD ecosystem, could test the real binary end-to-end
- **Cons**: Cross-language test setup, slow subprocess calls, harder to debug
- **Why rejected**: godog is mature enough and keeps everything in one language.
