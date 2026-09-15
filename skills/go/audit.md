<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.3.0 -->
# Go Audit Mode

Use with `policy.md` and `workflow.md` only for an explicitly requested deep,
phased Go audit. The task must name the repository path or scope and one exact
phase below; ask for either if missing. Repository instructions define focus,
exclusions, depth, and treatment of generated or vendored code.

## Audit contract

- Execute only the requested phase, stop at its boundary, and wait. Never advance
  automatically or combine phases.
- Inspect relevant source, tests, scripts, CI, configuration, and documentation.
  Anchor factual claims to tool-observed evidence: a path and, where applicable,
  a symbol. Documentation can establish stated intent, not deployed behavior.
- Label conclusions not proven by accessible evidence `INFERENCE`. Record excluded,
  inaccessible, or unreviewed material under `UNREVIEWED/INACCESSIBLE`, including
  its effect on confidence. Do not guess when tooling or access is missing.
- Prefer structural or type-aware evidence such as `gopls`; use text search when
  necessary. Treat an existing `graphify-out/` graph as optional discovery leads,
  and verify every graph-derived claim in the repository. Do not build a graph as
  an audit prerequisite.
- Build inventories before evaluation. Read enough package context to avoid
  interpreting a symbol in isolation. Reuse prior artifacts, but verify new claims.
- Output Markdown; fence machine-readable artifacts as `csv` or `json`. Split an
  oversized artifact into stable, named parts and set `NEXT` to the exact remaining
  part or step.

A hard failure is the sole exception to the footer requirement. Emit only:

```text
ERROR: <short reason>
BLOCKED_BY: <missing input, evidence, tool, or access>
```

Every successful Phase 1–4 response, and every unfinished Phase 5 chunk, ends with:

```text
STATE_SNAPSHOT: (max 8 bullets)
- <completed coverage, evidence limits, and artifact parts>

NEXT: <exact next phase or remaining artifact part>
```

## Phase gates

- Phase 1 inventories and describes; it does not recommend.
- Phase 2 accounts and indexes; it does not recommend or grade.
- Phase 3 assesses architecture and boundary violations; it does not provide
  detailed remediation plans.
- Phase 4 reports prioritized findings and fixes; it does not assign overall grades.
- Phase 5 synthesizes, grades, prioritizes, and plans.

## PHASE 1 - Inventory + Entrypoints

Produce a directory-grouped repository inventory with a one-line purpose and
importance tag for each directory, a one-line purpose for every file, and key
exported symbols for Go files. Identify evidenced entrypoints, startup and shutdown
paths, configuration, and secret sources. Report totals and
`UNREVIEWED/INACCESSIBLE`.

## PHASE 2 - Function Accounting

Produce exactly `function_index.csv` and `package_index.csv`, using stable part names
and at most 500 rows per part. Include one row per function or method, including
`init`, `main`, and tests. Leave caller or callee fields blank when precise analysis
is unavailable and document the limitation as `INFERENCE`. Run
`golang.org/x/tools/cmd/deadcode` when available; otherwise record that evidence gap.

## PHASE 3 - Architecture + Data Boundaries

Using Phases 1–2 evidence, describe the implemented architecture; map ingress,
egress, validation, and missing validation; identify transport/domain/persistence
leakage; and assess dependency direction, transaction boundaries, and idempotency.

## PHASE 4 - Observability + Security Audit

Review structured logging and correlation; metrics, tracing, health checks,
shutdown, and drain behavior; trust boundaries, authentication, authorization,
validation, injection, path handling, and secrets. Use available tool evidence such
as `govulncheck ./...` for reachable vulnerabilities and `gosec`/`staticcheck` for
source findings; cite rule IDs and locations. If a tool cannot run, record the gap.

Group findings by `P0`, `P1`, and `P2`. Each finding includes path, symbol, evidence,
impact, and a concrete fix.

## PHASE 5 - Synthesis

Produce an overall `A`–`F` grade and evidence-backed subgrades for code,
architecture, observability, security, testing, performance, modularity, and
docs/DX. Prioritize recommendations as `P0`/`P1`/`P2`, size each `S`/`M`/`L`, and
provide verification and rollout considerations. A completed Phase 5 ends with
`FINAL STATE_SNAPSHOT` (at most eight bullets) and `DONE`; an unfinished chunk
uses the resumable footer above.

When modernization is in scope, use available `gopls` modernization analysis and
`deadcode` evidence to assess supported Go/toolchain directives, structured logging,
error wrapping/inspection, current iteration features and built-ins, and appropriate
use of generics. Record unavailable analyzers rather than substituting intuition.

## Recommendation quality

Tie every recommendation to evidence and explain its mechanism. For consequential
architectural changes, compare viable alternatives, tradeoffs, and the cost of no
change; use a fuller Socratic decision process only when ambiguity warrants it or the
user requests it. The host controls presentation format. Reject generic advice that
could apply unchanged to an unrelated repository.
