<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.0.0 -->
# Rust Engineering Workflow

## Purpose
Use this skill when the task is to understand, modify, debug, refactor, review, or validate Rust code and you need a reliable workflow for moving from request to verified result.

This skill is the execution layer for Rust work. It pairs with `policy.md`, which defines what good code looks like. This skill defines how to approach the work, which tools to prefer, how to verify results, and how to report uncertainty.

## Skill Use
- Load this skill when the task involves understanding, modifying, debugging, refactoring, reviewing, or validating Rust code.
- Use this skill together with Rust policy and any task-specific Rust mode when available.
- Treat this skill as the default execution contract unless the repository has stricter local workflow requirements.
- Prefer repository-specific commands, feature flags, and tooling conventions when they are explicit.

## Tool Use
This skill is tool-agnostic and works with Claude Code, Codex, OpenCode, and similar assistants. Map its guidance to whatever file-reading, editing, search, and shell-execution tools your environment exposes.

- Invoke tools to gather evidence; do not narrate or infer what a tool would show.
- Prefer structural or type-aware tooling (LSP, `analyzer.md`) when available; fall back to text search only when it is not.
- Issue independent tool calls in parallel rather than sequentially.
- If a claim can be verified by running a command, reading a file, or inspecting a symbol, verify it before asserting it.

## Operating Stance
- Use tools before explanation when the tool can resolve the question.
- Read repository evidence before proposing changes.
- Prefer structural and executable signals over textual interpretation.
- Treat reasoning as a fallback layer, not a primary source of truth.
- Do not summarize behavior that has not been traced through code, tool output, or both.

## Core Principle
- Do not trust inference when verification is possible. The compiler is the cheapest and strongest verifier in this language.

## Truth Hierarchy
When evaluating correctness, use this order of authority:
1. compiler and borrow checker (`cargo check`, `cargo build`)
2. test results (`cargo test`, `cargo nextest run`)
3. static analysis (`cargo clippy`, `cargo deny`, `cargo audit`)
4. runtime signals such as logs, traces, metrics, and `miri` output for `unsafe` paths
5. repository policy and guidance
6. LLM reasoning

## Tool Precedence
When multiple approaches are possible, prefer them in this order:
1. structural tooling such as `analyzer.md` definitions, references, implementations, type-on-hover, expand-macro, and rename-safe operations
2. Cargo-native repository tooling such as `cargo check`, `cargo build`, `cargo test`, `cargo doc`, `cargo tree`, `cargo metadata`
3. repository-standard analysis tools such as `clippy`, `cargo deny`, `cargo audit`, `cargo machete`/`cargo udeps`, `cargo expand`
4. runtime evidence such as logs, traces, metrics, reproduced failures, `miri` for unsafe code paths, `cargo flamegraph` for hotspots
5. text search and file reads for supplemental discovery
6. reasoning when stronger evidence is unavailable

## Evidence Types
Use these categories when reasoning or reporting:
- structural evidence: definitions, references, trait impls, lifetimes, type signatures, module edges, feature gates
- executable evidence: compiler diagnostics, test results, clippy lints, `cargo deny`/`cargo audit` output
- runtime evidence: logs, traces, metrics, reproduced failures, `miri` reports, `valgrind`/`heaptrack` output for FFI / leaks
- repository evidence: `Cargo.toml`, `Cargo.lock`, `rust-toolchain.toml`, CI config, scripts, build scripts (`build.rs`), feature combinations
- inference: reasoning not yet confirmed by stronger evidence

## Repository Workflow Discovery
Before running default commands, inspect the repository for its canonical workflow:
- `Cargo.toml` — workspace shape, members, features, lints, MSRV
- `Cargo.lock` — committed for binaries, absent or gitignored for libraries
- `rust-toolchain.toml` or `toolchain.md` — pinned channel and components
- `.cargo/config.toml` — build flags, custom registries, target overrides
- `Makefile`, `Justfile`, `Taskfile.yml`, `cargo-make` config, or `xtask` crate
- CI definitions
- `clippy.toml`, `rustfmt.toml`, `deny.toml`
- `build.rs` files and their generated outputs
- feature combinations the project tests in CI

Prefer repository-standard commands over generic defaults when they are explicit and relevant.

## Knowledge-Graph Discovery (When Available)
If the repository carries a graphify knowledge graph (a `graphify-out/` directory), use it as a map to consult before broad text search — never as ground truth.
- Orient first from `graphify-out/GRAPH_REPORT.md` (or `graphify-out/wiki/index.md` when present): god nodes, communities, and cross-file relationships give you the architecture before you open a file.
- For "how does X relate to Y", consumer discovery, and blast-radius questions, prefer `graphify query "<question>"`, `graphify path "<A>" "<B>"`, and `graphify explain "<symbol>"` over grep — they traverse extracted and inferred edges across crate and module boundaries that text search misses.
- Every edge is tagged `EXTRACTED`, `INFERRED`, or `AMBIGUOUS`. Treat `EXTRACTED` as structural evidence; treat `INFERRED` and `AMBIGUOUS` as leads to confirm with rust-analyzer, the compiler, or tests. The graph never outranks executable evidence in the Truth Hierarchy.
- After changing code, run `graphify update .` (AST-only, no API cost) to keep the graph current.

If no `graphify-out/` directory exists, ignore this section.

## Task Classification
Before acting, classify the task into one primary mode:
- understand code
- modify code
- add new functionality
- debug failure
- refactor
- review
- validate or verify

If a task spans multiple modes, choose the dominant mode first and sequence the others explicitly.

## Global Workflow Rules

### Identify The Source Of Truth
- Use `analyzer.md` for code understanding when available.
- Prefer explicit types, trait bounds, lifetimes, and tool output over inferred intent.
- Use `cargo expand` to see post-macro source when behavior depends on a macro.
- Confirm assumptions against code and tool output before acting.

### Minimize Blast Radius
- Identify affected items, modules, crates, and consumers before changing code.
- Distinguish public API changes (semver-relevant) from internal implementation changes.
- Use `cargo public-api` (or read the diff) when public API impact is unclear.
- When a graphify graph is present, use `graphify path`/`graphify query` to surface cross-crate consumers and latent couplings before estimating blast radius (see Knowledge-Graph Discovery).
- Prefer the smallest viable change that solves the task.

### Require Verification
- Do not consider code correct until it has been verified by the relevant tools.
- The minimum acceptable verification for most code changes is:
  - `cargo check` (or `cargo build`) — the compiler accepts it
  - relevant tests passing (`cargo test` for the affected crate or `cargo nextest run -p <crate>`)
  - `cargo clippy --all-targets -- -D warnings` when part of the repository workflow
- Increase verification depth with risk.

### State Uncertainty Explicitly
- If evidence is incomplete, say what is known, what is unknown, and what must be verified next.
- Do not present guesses as conclusions.
- If evidence is insufficient, stop and identify the missing information or validation step.

## Minimum Required Tool Use By Mode

### Understand Code
Must use:
- structural lookup for definitions, references, implementations, trait impls, or item boundaries when available
- `cargo metadata` / `cargo tree` for crate and dependency boundaries
- `cargo expand` when macros are load-bearing for the question

May use:
- text search for comments, attribute keys, and supplemental discovery

### Modify Code
Must use:
- structural lookup before editing
- `cargo check` for the affected crate after the change
- relevant tests after the change (`cargo test -p <crate>` or full run)
- `cargo clippy` when part of the repository workflow

### Add New Functionality
Must use:
- existing test discovery
- failing or coverage-driving tests first when practical
- compile and test verification after implementation
- public API impact check when the change touches `pub` items

### Debug Failure
Must use:
- reproduction or concrete failing evidence when practical (`cargo test --no-run` then targeted `cargo test <name>`)
- code-path tracing
- post-fix regression verification

### Refactor
Must use:
- structural lookup before broad edits
- tests before and after meaningful steps
- compile and test confirmation that intended behavior is preserved
- `cargo public-api` (or equivalent) when refactor crosses a public boundary

### Review
Must use:
- diff inspection
- affected item, module, or crate tracing
- verification evidence review when available
- semver-impact check on public crates

### Validate Or Verify
Must use:
- the relevant compiler, test, and analysis commands, not code inspection alone, when executable verification is possible

## Workflow By Mode

### Understand Code
1. Identify the relevant entrypoint, boundary, or starting symbol.
2. Trace definitions, references, and implementations via `analyzer.md`.
3. Follow call flow across module and crate boundaries; respect feature gates.
4. Identify core types, trait bounds, and lifetimes before summarizing behavior.
5. If macros are load-bearing, expand them and trace the generated code.
6. Summarize runtime path from input to side effect.

### Modify Code
1. Identify affected items and crates.
2. Determine whether the change affects public API, private implementation, or both.
3. Identify existing tests and update or extend them when practical.
4. Make the smallest viable change.
5. Re-run verification (`cargo check`, then `cargo test`, then `cargo clippy`).
6. Re-check blast radius — features, optional deps, downstream consumers in the workspace.

### Add New Functionality
1. Identify the correct module or crate boundary for the new behavior.
2. Reuse existing module structure unless there is a real reason not to.
3. Start from types and trait bounds — encode invariants in the signature.
4. Add a failing test first when practical.
5. Implement the minimum viable code.
6. Verify behavior, placement, and observability implications.

### Debug Failure
1. Reproduce the failure deterministically when possible (`RUST_BACKTRACE=1`, `RUST_LOG=trace`, `cargo test <name> -- --nocapture`).
2. Identify the failing entrypoint, test, request path, or runtime boundary.
3. Trace inputs, transformations, and outputs.
4. For `unsafe`, run `miri` to expose UB. For panics, use the backtrace and `--nocapture`.
5. Find the earliest point where reality diverges from expectation.
6. Fix the root cause.
7. Add or update regression coverage.

### Refactor
1. Confirm current behavior is covered or add coverage first.
2. Identify the structural problem explicitly (deep generics, leaking abstractions, lifetime gymnastics).
3. Make small, incremental changes; let `cargo check` guide you.
4. Re-run tests after meaningful steps.
5. Stop when the structural problem is solved.

### Review
1. Identify what changed and why.
2. Check whether the change matches the stated task.
3. Check correctness first (compiler-level, then test-level).
4. Check edge cases, failure paths, cancellation, drop behavior, and blast radius where relevant.
5. Check `unsafe` blocks for `// SAFETY:` comments and soundness.
6. Check public API impact, feature interaction, and verification coverage.
7. Check style and consistency last.

### Validate Or Verify
1. Identify the risks introduced by the change.
2. Choose verification depth proportional to those risks.
3. Run the relevant commands and checks.
4. Confirm success criteria explicitly.
5. Report remaining uncertainty.

## Default Verification Profile
When repository-specific commands are not provided, prefer:
- affected-crate verification first:
  - `cargo check -p <crate>` then `cargo test -p <crate>` (or `cargo nextest run -p <crate>`)
- broader confirmation when blast radius is non-trivial:
  - `cargo build --workspace --all-targets`
  - `cargo test --workspace`
- lints when part of repository workflow:
  - `cargo clippy --workspace --all-targets -- -D warnings`
  - `cargo fmt --all -- --check`
- feature-matrix coverage when features are non-trivial:
  - `cargo hack check --feature-powerset --no-dev-deps`
- `miri` for code that exercises `unsafe`:
  - `cargo +nightly miri test -p <crate>`
- supply-chain checks when relevant:
  - `cargo deny check`
  - `cargo audit`

## Typical Risk Levels

### Low Risk
Examples:
- comments
- doc updates
- rename with strong `analyzer.md` support and no behavior change
- private helper extraction with covering tests

Typical checks:
- `cargo check` and focused tests if applicable

### Moderate Risk
Examples:
- behavior change in one crate
- handler change
- refactor with existing test coverage
- adding a feature flag

Typical checks:
- relevant crate tests or `cargo test --workspace`
- `cargo clippy`
- feature-matrix smoke when features were touched

### High Risk
Examples:
- async lifecycle changes (cancellation, drop, spawn)
- `unsafe` code added or modified
- public API changes (semver-relevant)
- shutdown behavior
- persistence or transport changes
- dependency changes (`Cargo.toml` / `Cargo.lock`)
- `build.rs` or proc-macro changes

Typical checks:
- full `cargo test --workspace`
- `cargo clippy --workspace --all-targets -- -D warnings`
- `miri` for `unsafe`
- `cargo deny check` / `cargo audit` for dependency changes
- public API diff (`cargo public-api`) when applicable
- integration tests when applicable

## Fallback Rules
- If `analyzer.md` or equivalent structural tooling is unavailable, fall back to `cargo check --message-format=json`, `cargo doc --no-deps`, `cargo expand`, `rg`, and careful file tracing.
- If tests cannot be run, state exactly why and identify the pending verification gap.
- If runtime reproduction is unavailable, anchor the analysis to the strongest available structural and executable evidence.
- If a build script (`build.rs`) is failing, read the script and its outputs (`OUT_DIR`) before assuming dependency or platform issues.

## Escalation Rules
- If structural understanding is uncertain, get compiler-visible or type-aware evidence (`cargo check`, type-on-hover, `cargo expand`).
- If behavior is uncertain, reproduce the behavior or run tests.
- If `unsafe` correctness is uncertain, run `miri` and re-derive the `// SAFETY:` invariants.
- If safety is uncertain in a broader sense, increase verification depth.
- If evidence conflicts, report the conflict and resolve it using the truth hierarchy.
- If required information is missing, stop and identify exactly what is missing.

## Output Contract
When responding to a Rust task, include the following when relevant:
- task classification
- current understanding
- plan
- affected crates, modules, items, or boundaries
- semver / API impact when public surface changed
- risks or blast radius (including feature interactions)
- verification steps
- remaining uncertainty

Prefer concise structure over narrative sprawl.

## Completion Criteria
Do not consider a Rust task complete until all applicable items are true:
- the task was correctly classified and scoped
- affected code paths were identified
- the requested change or analysis was completed
- relevant verification was run or explicitly identified as pending
- public API impact was characterized when applicable
- unresolved uncertainty was disclosed
- the result is supported by evidence rather than inference alone

## Anti-Patterns To Reject
- guessing when `cargo check` or `analyzer.md` can verify
- using grep as the primary structural tool when type-aware tooling is available
- making broad edits without understanding references, trait impls, and feature gates
- mixing refactor, feature work, and debugging implicitly
- claiming correctness without compile or test evidence when executable verification is possible
- silencing warnings with `#[allow(...)]` instead of fixing the underlying issue
- disabling clippy lints inline without justification
- adding `unsafe` without `miri` consideration
- presenting tentative analysis as established fact
- skipping regression coverage after fixing a bug
- leaving verification implicit
- ignoring `cargo deny` / `cargo audit` failures as "noise"

## Invocation Template
Use this skill with a prompt that supplies repository-specific context. Example:

```text
Use Rust Engineering Workflow and Rust Engineering Policy.
Understand the request pipeline in /path/to/repo.
Trace entrypoint to side effects, including async cancellation behavior.
Identify affected crates, modules, and feature gates.
Prefer actual tool output (cargo check, cargo test, rust-analyzer) over inference.
Report risks, missing information, and verification steps.
```
