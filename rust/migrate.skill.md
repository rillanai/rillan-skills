---
name: rust-migrate
description: Use when planning or executing Rust migrations — Rust edition upgrades (2021→2024), MSRV bumps, async runtime swaps, framework migrations (e.g., actix→axum), dependency replacement, workspace restructuring, removing `unsafe`, schema/API transitions, legacy modernization. Prefer reversibility over speed. Skip for greenfield code or isolated bug fixes without a migration dimension.
---

<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 0.1.0 -->
# Rust Migration Mode

## Purpose
Use this skill when planning or executing migrations in a Rust codebase.

Apply this skill with:
- `policy.skill.md` for design and boundary standards
- `workflow.skill.md` for tool-first discovery and verification

This mode exists for migrations that materially change toolchain, edition, MSRV, async runtime, dependencies, framework, workspace shape, schema, or public API and therefore require explicit impact analysis, sequencing, and rollback planning.

## Tool Use
This skill is tool-agnostic and works with Claude Code, Codex, OpenCode, and similar assistants. Map its guidance to whatever file-reading, editing, search, and shell-execution tools your environment exposes.

- Use tools to inventory call sites, trait impls, feature flags, and schema references before proposing a migration plan — do not estimate blast radius from memory.
- Use `cargo tree`, `cargo metadata`, and `rust-analyzer` references to map dependency and consumer surface; `rg` for textual sweeps.
- Issue independent tool calls (listing affected crates, reading `Cargo.toml`, checking CI, scanning features) in parallel.
- Run `cargo check`, `cargo test`, `cargo clippy`, and `cargo fmt --check` after each migration step and report actual output, not expected output.
- For dependency upgrades, read the changelog of every intermediate version, not just the target.
- For edition upgrades, run `cargo fix --edition` and read every change before committing it.
- If the migration touches persistence, run schema or data verification against the target state rather than inferring correctness.
- For public-API migrations, use `cargo public-api` (or `cargo semver-checks`) to surface the diff.

## When To Use
Use this skill for:
- Rust edition upgrades (e.g., 2021 → 2024)
- toolchain channel changes (stable → MSRV-pinned, or stable → nightly for a specific feature)
- MSRV bumps (and the release-note + CI implications)
- major dependency upgrades with breaking changes (`tokio 0.2 → 1.x`, `hyper 0.14 → 1.x`, `axum 0.6 → 0.7`)
- async runtime migrations (`async-std → tokio`, `actix → tokio + axum`)
- framework migrations (`actix-web → axum`, `rocket → axum`, `warp → axum`)
- workspace restructuring (single-crate → workspace, splitting a crate, extracting a library)
- removing `unsafe` (replacing with safe alternatives or wrapping in safer abstractions)
- removing `Box<dyn Error>` in favor of typed errors (or vice versa)
- API version transitions
- database / schema migrations (often paired with `sqlx`/`diesel` migration tooling)
- modernizing legacy code with compatibility constraints

Do not use this skill for:
- greenfield implementation
- ordinary code review
- isolated bug fixes without a migration dimension

## Operating Stance
- Prefer reversibility over speed.
- Describe the current system as implemented before proposing the target state.
- Do not migrate and opportunistically redesign at the same time.
- Every migration plan must answer: what can fail, how do we detect it, and how do we roll it back?
- For library crates, every migration step that changes the public API has a semver impact — call it out in the plan.

## Migration Planning Process
1. Inventory the current surface area: crates, modules, public items, trait impls, feature flags, dependencies, MSRV, edition, async runtime.
2. Define the target state and what must remain compatible during transition.
3. Identify blast radius, sequencing constraints, and rollback boundaries.
4. Choose incremental steps that are independently verifiable.
5. Define verification for each step before making changes.
6. Remove temporary compatibility layers, deprecated re-exports, and feature shims once migration is complete.

## Migration Types

### Edition Upgrades
- Read the [edition guide](https://doc.rust-lang.org/edition-guide/) for every intermediate edition (2018 → 2021 → 2024).
- Run `cargo fix --edition` per crate and review the diff carefully — the tool is conservative but not infallible.
- Update `edition = "..."` in each `Cargo.toml`. Mixed-edition workspaces compile fine but introduce unnecessary cognitive load.
- Re-run the full test suite (including `cargo test --doc`) after the edition flip; some changes affect macro hygiene and trait method resolution.
- Update `rust-toolchain.toml` if the new edition requires a newer minimum stable.

### Toolchain / MSRV Bumps
- Update `rust-toolchain.toml` (channel, components, profile) and `package.rust-version` in `Cargo.toml`.
- Run `cargo +<msrv> check --workspace` and `cargo +<msrv> test --workspace` to confirm the new floor still works.
- For libraries, document the MSRV change in the changelog and call out the semver implication (an MSRV bump is, by community convention, at least a minor bump).
- Update CI matrices to include the new MSRV cell.

### Dependency Migrations
- Map old import paths, call sites, types, and behavior contracts.
- Read the upstream changelog for *every* intermediate major version, not just current → target.
- Prefer drop-in replacement when behavior is truly compatible (re-export at the old name temporarily if needed).
- Otherwise use adapters or temporary trait impls at the consumption boundary; remove them after cutover.
- Run `cargo update -p <dep> --precise <version>` for surgical bumps; full `cargo update` is for end-of-cycle hygiene.
- Do not leave dual-dependency coexistence (`tokio = "0.2"` and `tokio = "1"`) in place indefinitely; every dependent crate eventually pays the bridging cost.
- Re-run `cargo deny check` and `cargo audit` after the swap.

### Async Runtime Migrations
- Inventory every `#[tokio::main]`, `#[async_std::main]`, `tokio::spawn`, `async_std::task::spawn`, `tokio::sync::*`, runtime-specific timers, and runtime-specific channels.
- Audit traits with `Send` bounds; switching runtimes often surfaces missing or extra `Send` requirements.
- Identify all `select!`, `join!`, `try_join!` macros — they are runtime-specific in subtle ways.
- Plan the cutover crate-by-crate; coexistence of two async runtimes in one process is technically possible but operationally messy.
- Test cancellation behavior carefully; cancellation semantics differ between runtimes.

### Framework Migrations
- Inventory handlers, middleware, route behavior, extractors, request/response types, and dependency wiring (`State`, `Data`, etc.).
- Verify middleware order and error-handling semantics — these almost always differ between frameworks.
- Compare route tables before and after; generate them from code if possible.
- Test the request lifecycle: timeouts, cancellation when the client drops, body size limits, multipart handling.
- Migrate one route group at a time when the framework supports mixed mounting; otherwise plan a single cutover with a tested rollback.

### Workspace Restructuring
- Map current crate relationships before moving code.
- Introduce new crate boundaries incrementally — extract a module to a new crate, verify the workspace builds, then move on.
- Use `[workspace.dependencies]` to unify versions across members; this also catches accidental drift.
- Use `[workspace.lints]` for shared lint configuration.
- Verify dependency direction after each meaningful move (`cargo depgraph` or `cargo modules` can help).
- Keep temporary re-exports (`pub use new_crate::*;`) only as long as needed for safe transition; remove and bump the major version.

### Removing `unsafe`
- Inventory every `unsafe` block and `unsafe fn`; categorize by reason (FFI, performance, transmute, raw pointers, `Send`/`Sync`).
- For each, evaluate the safe alternative: standard library types, `bytemuck`, `zerocopy`, `Pin`, the `crossbeam` family, `Arc<T>` patterns, or restructured ownership.
- Replace one block at a time; verify with `cargo test` and `cargo +nightly miri test` after each removal.
- Document the migration in the commit message; future reviewers benefit from the rationale.

### Database / Schema Migrations
- Separate schema migration from data migration.
- Plan expand/migrate/contract for zero-downtime constraints (add nullable column → backfill → make non-null → remove old column).
- Use the project's migration tooling (`sqlx-cli`, `diesel migration`, `refinery`); do not hand-write `ALTER` statements unless the tooling truly does not fit.
- Define rollback behavior explicitly. Some migrations (storage version flips, dropped columns) are intrinsically one-way — call this out.
- Never depend on production startup to run migrations automatically without a release-time gate.

### API Migrations
- Choose a versioning strategy and apply it consistently (`v1::Foo`, `v2::Foo`, or feature-gated alternates).
- Preserve backward compatibility unless a new version is being introduced.
- Document deprecation windows and consumer migration signals (`#[deprecated]` attribute with a message and `since` version).
- For library crates, characterize the semver impact precisely; `cargo semver-checks` catches many but not all cases.

## Execution Rules
- Keep each migration step small and independently verifiable.
- Run verification after each meaningful step.
- Preserve backward compatibility during the transition when consumers depend on it.
- Use `#[deprecated(since = "...", note = "...")]` on items being phased out so consumers see the warning before the removal.
- Use temporary adapters, shims, re-exports, or feature flags only when the migration truly requires it.
- Remove temporary migration scaffolding after the cutover and before the next release tag.

## Verification Expectations
- Verify the affected crate set on every step: `cargo check`, `cargo test`, `cargo clippy`.
- Use broader verification for public API, persistence, dependency, async runtime, or workspace shifts.
- Use `cargo +nightly miri test` for `unsafe`-touching migrations.
- Run `cargo deny check` and `cargo audit` for dependency changes.
- Use `cargo public-api` or `cargo semver-checks` for public-surface changes on library crates.
- Run feature-matrix verification (`cargo hack check --feature-powerset --no-dev-deps`) when feature interactions might be affected.
- Run repository-standard lint or static analysis when part of normal workflow.

## Anti-Patterns To Reject
- migrating and refactoring simultaneously without need
- changing behavior during a dependency swap unless explicitly intended
- skipping rollback planning
- one-shot big-bang migrations when incremental steps are possible
- leaving compatibility layers, deprecated re-exports, or feature shims in place indefinitely
- bumping a dependency without reading the changelog
- claiming migration safety without verification evidence
- silently ratcheting MSRV without a changelog entry
- removing `unsafe` by replacing one unsafe block with another in a different module
- adding a new async runtime without removing the old one in the same release

## Invocation Template
Use this skill with a prompt that supplies repository-specific context. Example:

```text
Use Rust Migration Mode with Rust Engineering Policy and Rust Engineering Workflow.
Plan and execute the actix-web → axum migration in /path/to/repo.
Identify impacted crates, handlers, middleware order, and extractor differences.
Keep the migration incremental: one route group per step with tests in between.
Define verification (cargo test, cargo clippy, integration tests) and rollback for each step.
Call out semver impact on the public crate surface.
```
