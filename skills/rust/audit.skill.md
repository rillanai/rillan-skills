---
name: rust-audit
description: Use for a phased, evidence-based deep audit of a Rust codebase. The user must invoke this explicitly and supply repo path + phase. Skip for small patch reviews, narrow bug hunts, or ordinary implementation work.
---

<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 0.1.0 -->
# Rust Audit Mode

## Purpose
Use this skill to run a phased, evidence-based deep audit of a Rust codebase.

Apply this skill with:
- `policy.skill.md` for evaluation standards
- `workflow.skill.md` for tool-first execution discipline

This mode exists because audit work has materially different output constraints, evidence rules, and sequencing from normal implementation work — and Rust audits have specific concerns around `unsafe`, async lifecycle, public API stability, and supply chain.

## Skill Use
- Load this skill only when the user explicitly wants a deep Rust repository audit or a clearly similar phased review.
- Treat this skill as the governing audit contract for the turn or session.
- Keep repository-specific instructions in the invoking prompt.
- Use this skill phase by phase. Do not treat it as permission to compress the whole audit into one response.

## Tool Use
This skill is tool-agnostic and works with Claude Code, Codex, OpenCode, and similar assistants. Map its guidance to whatever file-reading, editing, search, and shell-execution tools your environment exposes.

- Every factual claim in an audit must come from a tool invocation, not inference. Read the file, search the symbol, or run the command before writing the finding.
- Prefer structural or type-aware tooling (LSP, `rust-analyzer`) for references, trait impls, and call graphs; fall back to `cargo expand`, `cargo metadata`, `cargo tree`, and text search when needed.
- Issue independent tool calls in parallel: inventory scans, symbol lookups, multi-file reads, and dependency-graph queries should be batched.
- For `unsafe` audits, run `cargo +nightly miri test` (or note it cannot be run and treat that as a gap).
- For supply-chain audits, run `cargo deny check`, `cargo audit`, and `cargo tree --duplicates`; do not infer from `Cargo.toml` alone.
- If evidence cannot be gathered, record it under `UNREVIEWED/INACCESSIBLE` rather than guessing.

## When To Use
Use this skill when the user asks for:
- a deep Rust repository audit
- phased architecture review
- crate, module, or item accounting
- `unsafe` and soundness review
- async lifecycle and cancellation review
- public API and semver-stability review
- supply-chain (dependencies, advisories, licenses) review
- observability or security review as part of a broader Rust audit
- evidence-backed grading, modernization, or refactor planning

Do not use this skill for:
- small patch reviews
- narrow bug hunts
- ordinary implementation work

## Required Inputs
The invoking prompt must provide:
- repository path or scope
- exact phase to execute

Recommended inputs:
- focus areas (e.g., specific crates, the `unsafe` surface, async hot paths)
- exclusions (vendored, generated, third-party)
- depth constraints
- how to treat generated, vendored, or `build.rs`-emitted code
- previous phase artifacts or `STATE_SNAPSHOT` when continuing

If scope or phase is missing, stop and ask.

## Operating Stance
- Prefer evidence over intuition.
- Describe the system as implemented, not as intended.
- Stay phase-disciplined.
- Treat tests, examples, benches, scripts, CI, infra, and docs as first-class evidence.
- Read enough surrounding crate context to avoid item-level misinterpretation.
- If continuation context is provided, use it to resume the requested phase or exact next step, not to skip evidence gathering.
- Do not collapse multiple phases into one response.

## Evidence Rules
- Every factual claim must be anchored to a file path and, when applicable, an item path (`crate::module::Item`).
- Mark any non-provable conclusion as `INFERENCE`.
- List inaccessible or unreviewed material under `UNREVIEWED/INACCESSIBLE` with impact notes.
- For `unsafe` claims, anchor to the file, the `unsafe` block, and the `// SAFETY:` comment (or its absence).
- For dependency claims, anchor to `Cargo.toml`, `Cargo.lock`, and `cargo tree` output.
- Do not imply runtime certainty without code, config, test, or runtime evidence.

## Output Contract
- Output only Markdown.
- Machine-readable artifacts must be fenced `csv` or `json`.
- If a hard requirement cannot be met, output exactly:

```text
ERROR: <short reason>
BLOCKED_BY: <what is missing>
```

## Chunking And Continuation Rules
- Work only on the requested phase.
- Stop at the end of the phase boundary.
- Chunk large artifacts rather than compressing them inaccurately.
- When a phase is too large for one response, emit the current chunk, preserve artifact part names, and set `NEXT` to the exact remaining step or artifact part.
- If required information is missing, stop and identify exactly what is missing instead of guessing.
- End every response with:

```text
STATE_SNAPSHOT: (max 8 bullets)
- <bullet>

NEXT: <exact next phase name>
```

## General Audit Method
1. Establish accessible scope and obvious exclusions.
2. Read the files relevant to the requested phase before making conclusions.
3. Build inventories or evidence tables before evaluative claims.
4. Reuse prior phase artifacts when supplied, but verify any new claims against repository evidence.
5. Preserve phase boundaries strictly.

## Phase Gate Rules
- Phase 1 may inventory and describe, but must not recommend.
- Phase 2 may account and index, but must not recommend or grade.
- Phase 3 may assess architecture and boundary violations, but must not produce detailed remediation plans.
- Phase 4 may produce prioritized findings with fixes, but must not assign overall grades.
- Phase 5 may synthesize, grade, prioritize, and plan.

## Phase Rules

### PHASE 1 - Inventory + Entrypoints
Produce:
- workspace inventory (all crates with kind: bin / lib / proc-macro, edition, MSRV, default features)
- per-crate one-line purpose and importance tag
- one-line purpose for each non-trivial source file
- key exported items per crate (`pub` types, traits, functions, macros)
- entrypoints (`main` functions, `#[tokio::main]`, library re-exports), startup/shutdown, config sources, secret sources
- feature flag inventory (which features enable which behavior)
- `build.rs` inventory and what each generates
- totals and `UNREVIEWED/INACCESSIBLE`

### PHASE 2 - Item Accounting
Produce exactly:
- `item_index.csv` — one row per `pub` item with: crate, module path, kind (fn/struct/enum/trait/macro/const/static), name, signature summary, doc-status, feature-gate
- `crate_index.csv` — one row per workspace crate with: name, kind, edition, MSRV, dep count, public-item count, `unsafe`-block count, `unsafe`-fn count, test count, doctest count

Rules:
- chunk outputs to 500 rows max per file part
- count `unsafe` blocks textually if structural tooling is unavailable, and note `INFERENCE`
- include test items in counts but distinguish them in the kind field

### PHASE 3 - Architecture + Data Boundaries + Async + Unsafe
Using phase 1 and 2 evidence:
- describe architecture as implemented (crate dependency graph, module hierarchy, layered separation)
- map ingress and egress (HTTP/RPC handlers, queue consumers, FFI exposure)
- identify validation points and missing validation points
- identify leakage between transport, domain, and persistence
- assess transaction boundaries, idempotency, and dependency direction
- enumerate every `unsafe` block with: file, function, `// SAFETY:` justification (or its absence), apparent invariants, audit verdict (`SOUND` / `QUESTIONABLE` / `UNSOUND` / `UNVERIFIED`)
- enumerate every detached `tokio::spawn` (no join handle stored); flag cancellation behavior
- enumerate `Mutex`-across-`.await` instances and shared mutable state patterns
- assess feature-flag interactions (additive? non-additive? combinations tested in CI?)

### PHASE 4 - Observability + Security + Supply-Chain Audit
Review:
- logging structure (`tracing`?), correlation IDs, span discipline
- metrics, tracing exports, health checks, shutdown, drain behavior
- panic boundaries (panic-on-startup vs. propagated `Result`); `#[catch_unwind]` usage
- trust boundaries, authn/authz, input validation, injection risks (SQL, shell, path traversal), secret handling (`secrecy`?)
- dependency advisories (`cargo audit` output)
- license compliance (`cargo deny check licenses`)
- duplicate dependencies (`cargo tree --duplicates`)
- pinned vs. floating dependency versions
- `unsafe` soundness against `miri` output (or note as gap if `miri` cannot run)

Output findings grouped by `P0`, `P1`, and `P2`, each with:
- file path and item path
- evidence (tool output, file:line)
- concrete fix

### PHASE 5 - Synthesis
Produce:
- overall grade `A-F`
- subgrades for code, architecture, async lifecycle, unsafe, observability, security, supply-chain, testing, performance, modularity, docs/DX
- anchored justification
- prioritized refactor recommendations with `P0`, `P1`, and `P2`
- effort sizing `S`, `M`, `L`

## Completion Rule
An audit response is incomplete if it:
- mixes phases
- makes unsupported claims
- omits required artifacts
- grades before synthesis
- recommends fixes before the proper phase
- omits the continuation footer
- claims `unsafe` soundness without `miri` evidence (or without explicitly recording the gap)
- claims supply-chain hygiene without `cargo audit` / `cargo deny` evidence (or without recording the gap)

## Invocation Template
Use this skill with a prompt that supplies repository-specific context. Example:

```text
Use Rust Audit Mode with Rust Engineering Policy and Rust Engineering Workflow.
Audit /path/to/repo.
Execute PHASE 3 - Architecture + Data Boundaries + Async + Unsafe.
Focus on the worker crates, the FFI surface, and any spawn/cancel patterns in the request pipeline.
Run cargo +nightly miri test on crates with unsafe; record gaps if miri cannot run.
Summarize generated code instead of expanding it.
```
