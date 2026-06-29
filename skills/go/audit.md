<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.2.0 -->
# Go Audit Mode

## Purpose
Use this skill to run a phased, evidence-based deep audit of a Go codebase.

Apply this skill with:
- `policy.md` for evaluation standards
- `workflow.md` for tool-first execution discipline

This mode exists because audit work has materially different output constraints, evidence rules, and sequencing from normal implementation work.

## Skill Use
- Load this skill only when the user explicitly wants a deep Go repository audit or a clearly similar phased review.
- Treat this skill as the governing audit contract for the turn or session.
- Keep repository-specific instructions in the invoking prompt.
- Use this skill phase by phase. Do not treat it as permission to compress the whole audit into one response.

## Tool Use
This skill is tool-agnostic and works with Claude Code, Codex, OpenCode, and similar assistants. Map its guidance to whatever file-reading, editing, search, and shell-execution tools your environment exposes.

- Every factual claim in an audit must come from a tool invocation, not inference. Read the file, search the symbol, or run the command before writing the finding.
- Prefer structural or type-aware tooling (LSP, `gopls`) for references and call graphs; fall back to text search only when it is not available.
- Issue independent tool calls in parallel: inventory scans, symbol lookups, and multi-file reads should be batched.
- If evidence cannot be gathered, record it under `UNREVIEWED/INACCESSIBLE` rather than guessing.

## When To Use
Use this skill when the user asks for:
- a deep Go repository audit
- phased architecture review
- function or package accounting
- observability or security review as part of a broader Go audit
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
- focus areas
- exclusions
- depth constraints
- how to treat generated or vendored code
- previous phase artifacts or `STATE_SNAPSHOT` when continuing

If scope or phase is missing, stop and ask.

## Operating Stance
- Prefer evidence over intuition.
- Describe the system as implemented, not as intended.
- Stay phase-disciplined.
- Treat tests, scripts, CI, infra, and docs as first-class evidence.
- Read enough surrounding package context to avoid symbol-level misinterpretation.
- If continuation context is provided, use it to resume the requested phase or exact next step, not to skip evidence gathering.
- Do not collapse multiple phases into one response.

## Evidence Rules
- Every factual claim must be anchored to a file path and, when applicable, a symbol.
- Mark any non-provable conclusion as `INFERENCE`.
- List inaccessible or unreviewed material under `UNREVIEWED/INACCESSIBLE` with impact notes.
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
- repository inventory grouped by directory
- one-line purpose and importance tag for each directory
- one-line purpose for each file
- key exported symbols for Go files
- entrypoints, startup, shutdown, config, and secret source summary where evidenced
- totals and `UNREVIEWED/INACCESSIBLE`

### PHASE 2 - Function Accounting
Produce exactly:
- `function_index.csv`
- `package_index.csv`

Rules:
- one row per function or method, including `init`, `main`, and tests
- chunk outputs to 500 rows max per file part
- leave caller or callee fields blank when precision is not supportable and note `INFERENCE`
- run `golang.org/x/tools/cmd/deadcode` to identify unreachable functions as harder evidence than reference-counting by hand

### PHASE 3 - Architecture + Data Boundaries
Using phase 1 and 2 evidence:
- describe architecture as implemented
- map ingress and egress
- identify validation points and missing validation points
- identify leakage between transport, domain, and persistence
- assess transaction boundaries, idempotency, and dependency direction

### PHASE 4 - Observability + Security Audit
Review:
- logging structure and correlation
- metrics, tracing, health checks, shutdown, drain behavior
- trust boundaries, authn/authz, input validation, injection risks, path handling, secret handling

Anchor security findings to tool evidence, consistent with the "every claim from a tool" rule:
- `govulncheck ./...` for reachability-based CVEs (a vulnerable dependency that is actually called)
- `gosec` and `staticcheck` output for injection, path, and unsafe-use findings — cite the rule ID and line, not a paraphrase

Output findings grouped by `P0`, `P1`, and `P2`, each with:
- file path
- symbol
- evidence
- concrete fix

### PHASE 5 - Synthesis
Produce:
- overall grade `A-F`
- subgrades for code, architecture, observability, security, testing, performance, modularity, docs/DX
- anchored justification
- prioritized refactor recommendations with `P0`, `P1`, and `P2`
- effort sizing `S`, `M`, `L`

## Modernization Evidence
When the audit covers modernization, gather evidence rather than eyeballing. Run the `modernize` analyzer (`gopls`) and `golang.org/x/tools/cmd/deadcode`, and report against this checklist:
- `log/slog` adoption versus ad-hoc or third-party logging without an `slog.Handler` bridge
- error wrapping with `%w` and aggregation with `errors.Join`, branching via `errors.Is`/`errors.As` rather than string matching
- range-over-func iterators (`iter.Seq`/`iter.Seq2`) where channels or slice-builders are used for custom sequences
- `min`/`max`/`clear` builtins and `for range n` loops versus hand-rolled equivalents
- generics replacing `interface{}` containers and utility code
- `go` and `toolchain` directive currency against the supported Go release window

Cite the analyzer or `deadcode` output as the evidence, not a hand inspection.

## Completion Rule
An audit response is incomplete if it:
- mixes phases
- makes unsupported claims
- omits required artifacts
- grades before synthesis
- recommends fixes before the proper phase
- omits the continuation footer

## Invocation Template
Use this skill with a prompt that supplies repository-specific context. Example:

```text
Use Go Audit Mode with Go Engineering Policy and Go Engineering Workflow.
Audit /path/to/repo.
Execute PHASE 3 - Architecture + Data Boundaries.
Focus on package direction, shutdown behavior, and queue consumers.
Summarize generated code instead of expanding it.
```

## Recommendations Are Architecture: Full Socratic Intent
Findings are evidence-driven, but a recommendation is a design act — and architectural recommendations are exactly where buried decisions and generic patterns do the most damage. Scale the rigor to the recommendation, not to a fixed template:

- **Architectural output** — a modernization or refactor plan, a new boundary, package split, or dependency, a structural remediation, anything that forecloses future options — gets the **full** `socratic` protocol: the **Synthesis-First Gate** (commit to your own recommendation and red-team it before you present it), the **Decision Ledger** (surface each consequential fork as a weighted choice the user decides, not a pick buried in the report), and the **Depth Audit**.
- **Localized, evidence-obvious fixes** — a missing nil check, a pinned version, a tightened permission — get the **Depth Audit** alone.

Run the Depth Audit on every recommendation regardless. A recommendation FLAGs when it could be pasted into a different repo unchanged, reaches for a reflex default (add a cache/queue/retry/index/abstraction, split the service, "add rate limiting") without tying it to the evidence you gathered, leans on buzzwords ("scalable", "robust", "idiomatic", "best-practice") in place of a mechanism, or offers no alternative and no "do nothing" cost. Fix the FLAG by anchoring to the evidence, or surface it as an open question — never ship generic advice as a finding.
