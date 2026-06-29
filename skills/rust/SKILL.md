---
name: rust
description: Rust engineering — implement, review, refactor, debug, test, document, audit, migrate, or set up CI for Rust code. Triggers on `Cargo.toml`, `.rs` files, rust-analyzer, clippy, `cargo test`/`nextest`, ownership/borrowing/unsafe/async questions, editions (2021→2024), or MSRV. Root skill that routes to its mode files.
---

<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.2.0 -->
# Rust

Root skill for Rust work. This `SKILL.md` is the only file loaded up front; it routes to
the mode files in this directory, which you read **on demand** with your file tool.
Do not guess a mode file's contents — read it.

## Baseline — load for any non-trivial Rust task
- `policy.md` — crate/module structure, ownership and borrowing discipline, error handling, unsafe policy, async lifecycle, the review bar.
- `workflow.md` — tool-first discovery (rust-analyzer), the truth hierarchy, verification.

## Modes — load the one matching the task, on top of the baseline
- `dev.md` — implement or modify Rust: features, bug fixes, targeted refactors.
- `test.md` — unit/integration/doctests, `cargo nextest`, proptest, fuzz, criterion, insta, miri, coverage.
- `audit.md` — explicit, phased, evidence-based deep audit. User supplies repo path + phase.
- `docs.md` — `///`/`//!` docs, doctests, crate docs, `cargo doc`, mdbook, READMEs, runbooks.
- `migrate.md` — edition upgrades (2021→2024), MSRV bumps, async-runtime/framework swaps, workspace restructuring, removing `unsafe`.
- `ci.md` — CI for a Rust project (fmt, clippy `-D warnings`, nextest, coverage, cargo-deny/audit, MSRV, miri, release).

Cross-pack pointers use path form, e.g. `cicd/core.md`. Stricter repository-local
conventions win when they are explicit and defensible.
