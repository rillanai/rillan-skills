<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.1.0 -->
# Go Test Mode

## Purpose
Use this skill when designing, writing, reviewing, or refactoring tests for a Go codebase.

Apply this skill with:
- `policy.md` for standards that define maintainable Go code
- `workflow.md` for tool-first execution and verification discipline

This mode exists because test work has its own decision rules, fixtures, verification patterns, and failure modes.

## Tool Use
This skill is tool-agnostic and works with Claude Code, Codex, OpenCode, and similar assistants. Map its guidance to whatever file-reading, editing, search, and shell-execution tools your environment exposes.

- Run the tests you write. A new or updated test is not done until `go test` has executed it and the result is known.
- Use structural tools to locate the code under test and its callers; do not guess at signatures or coverage.
- Issue independent tool calls (test discovery, coverage reads, symbol lookups) in parallel.
- Report flaky or timing-sensitive failures with the command and output that produced them, not paraphrased.

## When To Use
Use this skill for:
- writing new tests
- designing a test strategy
- analyzing coverage gaps
- refactoring flaky or brittle tests
- writing regression tests for bugs
- adding integration, e2e, fuzz, benchmark, or golden-file coverage
- reviewing test quality

Do not use this skill for:
- broad production-code review outside test concerns
- documentation work
- security audit work

## Test Philosophy
- Test behavior, not implementation details.
- Test at the right level for the risk.
- Treat tests as production code.
- Prefer deterministic evidence over incidental timing or environment luck.
- A failing test is useful only when it fails for the right reason.

## Choosing Test Type

### Unit Tests
Use for:
- branching logic
- validation
- parsing and transformation
- state-machine behavior
- error-path coverage

### Table-Driven Tests
Use for:
- many input-output combinations
- validation matrices
- parser and formatter coverage
- boundary case expansion

### Integration Tests
Use for:
- database behavior
- HTTP or gRPC integration
- queue or file-system interaction
- multi-package wiring

### End-To-End Tests
Use sparingly for:
- critical request flows
- smoke coverage
- deploy-time or contract confirmation

### Fuzz Tests
Use native Go fuzzing (`func FuzzXxx(f *testing.F)`, `go test -fuzz`) for:
- parsers, decoders, and deserializers
- validators and sanitizers
- anything that handles untrusted or attacker-controlled input
- round-trip invariants (encode→decode, marshal→unmarshal) and any function with a property that must hold for all inputs

Rules:
- Seed the corpus with `f.Add(...)` using known-tricky inputs and any past crash reproducers.
- Assert a property inside `f.Fuzz`, not a fixed output — round-trip equality, "must not panic", invariant preservation. A fuzz target with no property is just a slow no-op.
- The seed corpus runs as ordinary unit tests under a plain `go test` (no `-fuzz`); active mutation only happens with `-fuzz=FuzzXxx`. This means seed cases give regression value on every run for free.
- When the fuzzer finds a failure it writes the input to `testdata/fuzz/FuzzXxx/`. Commit that file — it becomes a permanent regression case replayed by `go test`.
- Active fuzzing is time-boxed, not run-to-completion: drive it with `-fuzztime` (e.g. `-fuzztime=60s` locally, longer on a schedule in CI). See `ci.md` for the CI shape (seed corpus on every PR, time-boxed `-fuzz` on a schedule).
- `-fuzz` matches one target at a time and cannot be combined with multiple packages; run targeted: `go test -run=^$ -fuzz=FuzzXxx -fuzztime=60s ./path/to/pkg`.

### Benchmark Tests
Use for:
- performance-sensitive code
- allocation-sensitive paths
- comparing concrete alternatives

### Golden File Tests
Use for:
- complex rendered output
- code generation output
- CLI formatting
- report generation

### BDD With Ginkgo/Gomega
Use when:
- the repository already uses it
- the behavior benefits from nested shared context

Prefer standard `testing` when:
- table-driven or direct assertions are clearer
- introducing BDD would add ceremony without payoff

## Test Design Rules
- Prefer public-behavior coverage over private implementation coupling.
- Use `t.Run` for named cases and selective execution.
- Use `t.Helper()` in helpers.
- Use `t.Cleanup` for teardown tied to test lifecycle.
- Default to `t.Parallel()`. Call it in every test and subtest unless something concrete makes it unsafe — shared mutable global state, an external resource that can't be concurrently used, or order dependence. Parallelism is the goal; serial is the exception you justify, not the reverse. See `Test Speed And Parallelism` below for the correctness rules that make this safe.
- Prefer fakes and stubs over heavy mock frameworks unless call sequencing is the real behavior under test.
- Do not mock types you do not own; wrap them behind your own interface first when necessary.

## Test Speed And Parallelism
Fast tests are a quality feature: they get run more often, shorten the feedback loop, and keep CI cheap. Treat suite wall-clock time as a number worth defending — but never by deleting coverage or skipping the race detector. Speed comes from parallelism and avoided work, not from testing less.

- **Parallelize by default.** A package whose tests all call `t.Parallel()` runs its cases concurrently up to `GOMAXPROCS` (override with `-parallel=N`). Across packages, `go test ./...` already runs distinct packages concurrently (`-p=N`, defaults to `GOMAXPROCS`) — so the biggest single lever is making the slow packages internally parallel.
- **Capture the loop variable correctly.** In table-driven parallel subtests, ensure each `t.Run` closure binds its own case. On Go 1.22+ the per-iteration loop variable makes this safe; on older toolchains add `tc := tc` before `t.Run`. A parallel subtest that closes over a shared loop variable tests the wrong case.
- **`t.Setenv` and `t.Parallel()` are mutually exclusive** — `t.Setenv` panics in a parallel test (env is process-global). A test that needs an env var either stays serial or is refactored to inject configuration instead of reading the environment.
- **Make parallel-safe the default design, not an afterthought.** Per-test temp dirs (`t.TempDir()`), unique resource names, no shared package-level mutable vars, no reliance on execution order. Code written this way is parallelizable for free.
- **Run with `-shuffle=on`** in CI and locally. Order-dependent tests are a latent bug; shuffling surfaces them before parallelism does in production.
- **Avoid redundant work.** Use `-count=1` to defeat the cache only when you specifically need a fresh run (the test cache is a speed feature — keep it). Scope expensive suites behind build tags (`-tags integration`) or `testing.Short()` so the fast inner loop stays fast: guard slow cases with `if testing.Short() { t.Skip() }` and run `-short` on the quick tier.
- **Find the long pole.** `go test -v` with timing, or `gotestsum`, shows the slowest tests. Optimizing the slowest 5% of tests usually buys more than micro-tuning the rest.
- **Keep `-race` on despite its cost.** The race detector slows tests ~2–10×, but it catches a bug class nothing else does. Buy the time back with parallelism, sharding, and caching — not by dropping `-race`.

## Concurrency And Lifecycle Testing
- Run concurrent or shared-state tests with the race detector.
- Use explicit synchronization, not `time.Sleep`, for correctness.
- Prefer `testing/synctest` (GA in Go 1.25) for deterministically testing timeouts, tickers, context deadlines, and goroutine coordination: run the bubble with `synctest.Test`, advance the fake clock, and use `synctest.Wait` to block until every goroutine in the bubble is durably idle. This replaces real-time sleeps and the flakiness they cause.
- Test cancellation, timeout, shutdown, and drain behavior when code depends on them.
- Verify goroutine ownership and completion in lifecycle-sensitive code.

## Coverage And Verification
- Coverage quality matters more than percentage.
- Prioritize branches, failure paths, boundary cases, and cancellation behavior.
- Use integration coverage where the risk lies in wiring or external interaction.
- Treat a missing regression test after a bug fix as a gap unless there is a concrete reason it cannot be added.

## Typical Commands
When the repository does not define a stricter workflow, prefer:
- `go test ./...`
- `go test -race ./...` for concurrency or lifecycle-sensitive changes
- `go test -shuffle=on ./...` to flush out order-dependent tests
- `go test -coverprofile=coverage.out ./...` when coverage evidence is needed
- `go test -run=^$ -fuzz=FuzzXxx -fuzztime=60s ./path/to/pkg` to actively fuzz one target (a plain `go test` already replays its seed corpus)
- `go test -short ./...` to run the fast tier, skipping cases guarded by `testing.Short()`
- `ginkgo run -r --race --cover` when Ginkgo is the primary repository workflow

## Anti-Patterns To Reject
- flaky time-based synchronization
- tests coupled to implementation call order without real need
- broad mocks replacing simple fakes
- no regression coverage after bug fixes
- golden files updated without reviewing the diff
- treating high line coverage as proof of meaningful test quality
- serial-by-default tests with no reason for not calling `t.Parallel()`
- `t.Setenv` inside a parallel test (it panics; inject config instead)
- parallel subtests that close over a shared loop variable on pre-1.22 toolchains
- a fuzz target that asserts nothing (no property, no panic check) — it burns CPU for no signal
- discovered fuzz crashers left out of `testdata/fuzz/` so the regression isn't replayed
- buying suite speed by dropping `-race` instead of by parallelizing/sharding

## Invocation Template
Use this skill with a prompt that supplies repository-specific context. Example:

```text
Use Go Test Mode with Go Engineering Policy and Go Engineering Workflow.
Add regression coverage for the queue retry logic in /path/to/repo/internal/worker.
Prefer standard testing and table-driven cases.
Cover cancellation, invalid payloads, and retry exhaustion.
Verify with the relevant package tests and race detection if shared state is involved.
```
