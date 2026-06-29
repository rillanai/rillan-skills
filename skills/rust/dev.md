<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.2.0 -->
# Rust Development Mode

## Purpose
Use this skill when the task is to implement or modify Rust code: feature work, bug fixes, targeted cleanup, or small refactors coupled to a concrete behavior change.

This skill is intentionally thin. It is the default execution mode for writing Rust code and is meant to be used with:
- `policy.md` for what good Rust code looks like
- `workflow.md` for how to approach and verify the work

## Skill Use
- Load this skill when the primary task is to change Rust code.
- Use this skill together with Rust policy and Rust workflow when available.
- Treat this skill as the implementation-mode overlay, not as the full Rust contract by itself.
- Follow repository-specific commands, feature flags, MSRV constraints, and CI conventions when they are explicit.

## Tool Use
This skill is tool-agnostic and works with Claude Code, Codex, OpenCode, and similar assistants. Map its guidance to whatever file-reading, editing, search, and shell-execution tools your environment exposes.

- Invoke tools to read code, find references, and run tests; do not describe what you would do.
- Issue independent tool calls in parallel rather than sequentially.
- Run the relevant `cargo check` / `cargo test` / `cargo clippy` commands yourself — do not claim a change is verified without tool output.

## Mode Focus
- Deliver the requested behavior with the smallest viable change.
- Preserve module and crate boundaries unless the task explicitly requires changing them.
- Respect existing feature flags; do not silently flip defaults or add features that aren't required.
- Prefer regression tests for bug fixes and failing tests first for new behavior when practical.
- Keep unrelated cleanup out of the patch unless it is necessary for correctness or to avoid making the design worse.
- Do not add `unsafe`, new external dependencies, or new public API surface without a clear reason tied to the task.

## Implementation Workflow
1. Classify the task and identify the affected items, modules, crates, and feature gates.
2. Use structural tooling (`analyzer.md`) to find definitions, references, and trait impls before editing.
3. Identify existing tests that cover the behavior and extend them when practical.
4. Make the smallest viable change — let the type system and borrow checker confirm intent as you go.
5. Re-run verification proportional to the risk.
6. Re-check blast radius (features, optional deps, downstream workspace consumers) before considering the task complete.

## Default Verification
- At minimum, run `cargo check -p <affected-crate>` and `cargo test -p <affected-crate>` (or `cargo nextest run -p <crate>`).
- Run `cargo clippy --all-targets -- -D warnings` for the affected crate when clippy is part of the repository workflow.
- Run `cargo fmt --all -- --check` (or fix with `cargo fmt`) before claiming done.
- Use `cargo test --workspace` when the change affects shared crates, public API, persistence, transport, async lifecycle, or dependency wiring.
- Use `cargo +nightly miri test` for code that adds or modifies `unsafe`.
- For non-trivial feature interactions, run `cargo hack check --feature-powerset --no-dev-deps` (or the project's matrix subset).

## Public API Discipline
- If the change touches `pub` items, characterize the semver impact (patch / minor / major).
- If the crate uses `cargo public-api` or similar in CI, run it locally and include the diff in the report.
- Mark new public enums and structs `#[non_exhaustive]` if they may grow.
- Do not silently add new trait impls for foreign types unless intentional — coherence and downstream impact deserve a comment.

## Async Discipline
- New `tokio::spawn` calls need an owner: a `JoinHandle` stored somewhere with a join or cancel path, or an explicit comment if detached.
- Do not introduce a `std::sync::Mutex` (or `parking_lot::Mutex`) held across `.await`. Use `tokio::sync::Mutex` if the guard genuinely must span an await point, or restructure to release the lock first.
- Long-running loops should `select!` on a cancellation signal, not just busy-wait or sleep.
- CPU-bound work added to an async path goes through `spawn_blocking` (or `rayon`) — do not block the runtime.

## Completion Criteria
Do not consider an implementation task complete until all applicable items are true:
- the requested behavior was implemented or corrected
- affected tests were added or updated when behavior changed
- `cargo check`, `cargo test`, `cargo clippy`, and `cargo fmt --check` pass for the affected scope (or a concrete verification gap was reported)
- no new `unsafe` block lacks a `// SAFETY:` comment
- no new `unwrap()` / `expect()` was added in fallible production paths without justification
- public API impact was characterized when public items changed
- no unnecessary boundary widening, generic parameter, trait, or feature was introduced
- remaining uncertainty and blast radius were stated explicitly

## Invocation Template
Use this skill with a prompt that supplies repository-specific context. Example:

```text
Use Rust Development Mode with Rust Engineering Policy and Rust Engineering Workflow.
Implement the queue retry backoff fix in /path/to/repo.
Keep the change scoped to the worker crate.
Add or update regression coverage.
Verify the affected crate with cargo test, cargo clippy, and cargo fmt --check.
Run the workspace test suite if the change crosses crate boundaries.
```
