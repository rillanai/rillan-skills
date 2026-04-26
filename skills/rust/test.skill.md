---
name: rust-test
description: Use when designing, writing, reviewing, or refactoring Rust tests — unit, integration (`tests/`), doctests, `cargo nextest`, property tests (`proptest`/`quickcheck`), fuzz (`cargo-fuzz`/`afl`), `criterion` benchmarks, snapshot tests (`insta`), `miri` for unsafe, coverage with `cargo llvm-cov`. Run the tests you write before claiming done. Pair with `rust-policy` and `rust-workflow`. Skip for production-code review or documentation.
---

<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 0.1.0 -->
# Rust Test Mode

## Purpose
Use this skill when designing, writing, reviewing, or refactoring tests for a Rust codebase.

Apply this skill with:
- `policy.skill.md` for standards that define maintainable Rust code
- `workflow.skill.md` for tool-first execution and verification discipline

This mode exists because test work has its own decision rules, fixtures, async patterns, property-test tradeoffs, and failure modes — particularly around `unsafe`, async cancellation, and feature-flag matrices.

## Tool Use
This skill is tool-agnostic and works with Claude Code, Codex, OpenCode, and similar assistants. Map its guidance to whatever file-reading, editing, search, and shell-execution tools your environment exposes.

- Run the tests you write. A new or updated test is not done until `cargo test` (or `cargo nextest run`) has executed it and the result is known.
- Use structural tools (`rust-analyzer`) to locate the code under test and its callers; do not guess at signatures, lifetimes, or coverage.
- Issue independent tool calls (test discovery, coverage reads, symbol lookups) in parallel.
- Report flaky, timing-sensitive, or feature-gated failures with the exact command and output that produced them, not paraphrased.

## When To Use
Use this skill for:
- writing new unit, integration, or doc tests
- designing a test strategy
- analyzing coverage gaps with `cargo llvm-cov`
- refactoring flaky or brittle tests
- writing regression tests for bugs
- adding integration, e2e, property-based, fuzz, benchmark, snapshot, or `miri` coverage
- reviewing test quality

Do not use this skill for:
- broad production-code review outside test concerns
- documentation work (use `rust-docs`)
- security audit work (use `security-review` or `rust-audit`)

## Test Philosophy
- Test behavior, not implementation details.
- Test at the right level for the risk.
- Treat tests as production code (clippy applies, lifetimes apply, naming applies).
- Prefer deterministic evidence over incidental timing or environment luck.
- A failing test is useful only when it fails for the right reason.
- Use the type system to make invalid inputs unrepresentable; the tests then cover behavior, not invariant rediscovery.

## Test Layout
Cargo recognizes several test locations; use them deliberately:

- **Unit tests** — `#[cfg(test)] mod tests { ... }` inside the source file. Use for testing private items and tight white-box coverage of the same module.
- **Integration tests** — `tests/<name>.rs` at the crate root. Each file is its own crate that imports the public API only. Use for end-to-end behavior, multi-module wiring, and to enforce that the public surface is genuinely usable.
- **Doc tests** — examples in `///` comments are compiled and run by `cargo test`. Use for short, didactic snippets that double as documentation.
- **Examples** — `examples/<name>.rs` are not run by `cargo test` but are compiled by `cargo build --examples`. Add a smoke-test step in CI to actually compile them.
- **Benchmarks** — `benches/<name>.rs` with `criterion`. Stable Rust does not support the built-in `#[bench]`; use `criterion`.

## Choosing Test Type

### Unit Tests
Use for:
- branching logic
- validation, parsing, transformation
- state-machine behavior
- error-path coverage
- private helper coverage that integration tests cannot reach

### Integration Tests (`tests/`)
Use for:
- end-to-end public-API behavior
- multi-module or multi-crate wiring
- HTTP / gRPC / database integration (with `testcontainers` or in-process fakes)
- ensuring the public surface is genuinely consumable from outside the crate

### Doc Tests
Use for:
- short, didactic examples that double as API documentation
- hidden setup with `# `-prefixed lines
- panic-on-error or compile-fail demonstrations (`compile_fail` annotation)

Skip for:
- complex tests where the example would be longer than the prose

### Property Tests (`proptest`, `quickcheck`)
Use for:
- parsers, encoders, validators
- round-trip invariants (`encode → decode → equal`)
- algebraic properties (associativity, commutativity)
- fuzzing without the heavy `cargo-fuzz` setup
- shrinking-driven minimal failing cases

### Fuzz Tests (`cargo-fuzz` with libFuzzer, or `afl.rs`)
Use for:
- parsers and validators consuming untrusted input
- deserializers (`serde_json`, `bincode`)
- protocol decoders
- code where memory safety bugs would be catastrophic

Run fuzz targets continuously in CI on a budget (5–15 min per PR; longer scheduled runs).

### Benchmarks (`criterion`)
Use for:
- performance-sensitive code
- allocation-sensitive paths (pair with `cargo bench` + heap profiling)
- comparing concrete alternatives (algorithm A vs. B)
- detecting performance regressions across commits (criterion's stored baselines)

Do not gate PRs on absolute timing — benchmark machines vary. Track relative regressions via stored baselines or external trend tools.

### Snapshot Tests (`insta`)
Use for:
- complex rendered output (CLI formatting, code generation, reports)
- error message stability
- serializer output (when format stability is the contract)

Always review the snapshot diff before accepting; `cargo insta review` is a workflow tool, not a rubber stamp.

### `miri` (Unsafe / UB Detection)
Use for:
- every `unsafe` block in the codebase, at least once
- pointer arithmetic, `transmute`, `from_raw`/`into_raw`
- FFI boundaries (within reason — `miri` cannot run foreign code)
- any test exercising the unsafe code path

`cargo +nightly miri test` is slow but catches a class of bugs nothing else does.

## Test Design Rules
- Prefer public-behavior coverage over private implementation coupling.
- Use `#[test]` with `#[should_panic(expected = "...")]` to test panic messages, not raw `#[should_panic]`.
- Use `Result<(), Box<dyn std::error::Error>>` (or `anyhow::Result`) as the test return type to use `?` for setup; failed tests still report cleanly.
- Use `assert_eq!`, `assert_ne!`, and `assert_matches!` (from `assert_matches` crate or std once stabilized); custom messages help diagnostics.
- Prefer `pretty_assertions` for large struct diffs in test output.
- Do not name a test `test_<symbol>`; the function name is the test name. Name it after the behavior (`empty_vec_returns_default`, `retries_exhausted_returns_error`).
- Use `mod tests { ... }` blocks with `use super::*;` for unit tests; this gives access to private items.
- Use `#[cfg(test)]` on test-only helpers to avoid leaking them into production builds.

## Async Tests
- Annotate with `#[tokio::test]` (or the runtime's equivalent: `#[async_std::test]`, `#[smol_potat::test]`).
- Use `#[tokio::test(flavor = "multi_thread", worker_threads = N)]` when the test exercises multi-threaded scheduling.
- Use `tokio::time::pause()` / `advance(...)` for tests that wait on timeouts or intervals — never `tokio::time::sleep` real durations in tests.
- Use `tokio::test` with the `start_paused = true` flag for fully deterministic time control.
- Test cancellation explicitly: drop a `CancellationToken`, drop the `JoinHandle`, or `select!` against a cancel branch.

## Concurrency And Lifecycle Testing
- Run tests under `cargo test` once with default scheduling, then re-run under `cargo test -- --test-threads=1` if interleaving issues are suspected.
- Use `loom` for systematic concurrency testing of lock-free or shared-state code (rare, but the right tool when needed).
- Use explicit synchronization (channels, `Notify`, `Barrier`), never `thread::sleep`, for correctness.
- Test cancellation, timeout, shutdown, and drain behavior when code depends on them.
- For background tasks, verify `JoinHandle::await` returns on shutdown, not just that the task "stops eventually".

## Feature-Matrix Testing
- Tests inside `#[cfg(feature = "x")]` only run when the feature is enabled. CI must run a feature combination that actually enables them.
- Use `cargo hack test --feature-powerset --no-dev-deps` (or a curated subset) when feature interactions are non-trivial.
- Mark tests that require a non-default feature with `#[cfg(feature = "x")]` rather than `#[ignore]`; ignored tests rot.

## Coverage And Verification
- Use `cargo llvm-cov` (preferred over `tarpaulin` on modern Rust) for coverage.
- Coverage quality matters more than percentage.
- Prioritize branches, failure paths, boundary cases, and cancellation behavior.
- Use integration coverage where the risk lies in wiring or external interaction.
- Treat a missing regression test after a bug fix as a gap unless there is a concrete reason it cannot be added.
- Doc tests count toward coverage when run via `cargo llvm-cov --doctests` (requires nightly for full support).

## Test Runner Choice
- **`cargo test`** is the default; sufficient for most projects.
- **`cargo nextest run`** is a faster, friendlier runner: parallel by default, JUnit output for CI, retries-on-failure for known-flaky tests, smaller output. Adopt for medium and large workspaces.
- Doc tests still run via `cargo test --doc` even when nextest is the primary runner.

## Typical Commands
When the repository does not define a stricter workflow, prefer:
- `cargo test --workspace` (or `cargo nextest run --workspace`)
- `cargo test --workspace --all-features` for a maximalist run
- `cargo test --doc --workspace` for doc tests
- `cargo +nightly miri test -p <crate>` for `unsafe` paths
- `cargo llvm-cov --workspace --lcov --output-path lcov.info` for coverage
- `cargo bench --bench <name>` for criterion benchmarks
- `cargo fuzz run <target> -- -max_total_time=300` for fuzz runs
- `cargo hack test --feature-powerset --no-dev-deps` for feature matrices
- `cargo insta review` for snapshot updates

## Anti-Patterns To Reject
- `thread::sleep` for synchronization in tests
- `tokio::time::sleep` with real durations instead of paused time
- tests coupled to implementation call order without real need
- broad mocks replacing simple fakes; mock libraries are usually overkill in Rust
- no regression coverage after bug fixes
- snapshot files updated without reviewing the diff
- treating high line coverage as proof of meaningful test quality
- ignoring `unsafe` blocks under `miri` because "the test passes normally"
- relying on `#[ignore]` to deal with flaky tests instead of fixing the flake
- detached `tokio::spawn` in tests with no `JoinHandle::await`
- per-test global mutable state (`OnceLock` or `lazy_static`) without isolation

## Invocation Template
Use this skill with a prompt that supplies repository-specific context. Example:

```text
Use Rust Test Mode with Rust Engineering Policy and Rust Engineering Workflow.
Add regression coverage for the queue retry logic in /path/to/repo/crates/worker.
Prefer integration tests under tests/ for the public retry behavior; use unit tests inside the module for the private backoff helper.
Cover cancellation (drop the CancellationToken), invalid payloads, and retry exhaustion.
Use #[tokio::test(start_paused = true)] for the timing-sensitive cases.
Verify with cargo nextest run -p worker, cargo clippy, and cargo fmt --check.
```
