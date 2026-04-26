---
name: python-audit
description: Use for a phased, evidence-based deep audit of a Python codebase. The user must invoke this explicitly and supply repo path + phase. Skip for small patch reviews, narrow bug hunts, or ordinary implementation work.
---

<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 2.2.0 -->
# Python Audit Mode

## Purpose
Use this skill to run a phased, evidence-based deep audit of a Python codebase.

Apply this skill with:
- `policy.skill.md` for evaluation standards
- `workflow.skill.md` for tool-first execution discipline

This mode exists because audit work has materially different output constraints, evidence rules, and sequencing from normal implementation work.

## Skill Use
- Load this skill only when the user explicitly wants a deep Python repository audit or a clearly similar phased review.
- Treat this skill as the governing audit contract for the turn or session.
- Keep repository-specific instructions in the invoking prompt.
- Use this skill phase by phase. Do not treat it as permission to compress the whole audit into one response.

## Tool Use
This skill is tool-agnostic and works with Claude Code, Codex, OpenCode, and similar assistants. Map its guidance to whatever file-reading, editing, search, and shell-execution tools your environment exposes.

- Every factual claim in an audit must come from a tool invocation, not inference. Read the file, search the symbol, or run the command before writing the finding.
- Prefer structural or type-aware tooling (LSP, `pyright`, `mypy`) for references and call graphs; fall back to text search only when it is not available.
- Issue independent tool calls in parallel: inventory scans, symbol lookups, and multi-file reads should be batched.
- If evidence cannot be gathered, record it under `UNREVIEWED/INACCESSIBLE` rather than guessing.

## When To Use
Use this skill when the user asks for:
- a deep Python repository audit
- phased architecture review
- function, class, or module accounting
- observability or security review as part of a broader Python audit
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
- Python version or framework if known
- previous phase artifacts or `STATE_SNAPSHOT` when continuing

If scope or phase is missing, stop and ask.

## Operating Stance
- Prefer evidence over intuition.
- Describe the system as implemented, not as intended.
- Stay phase-disciplined.
- Treat tests, scripts, CI, infra, dependency files, and docs as first-class evidence.
- Account for dynamic Python behavior without overstating certainty.
- Read enough surrounding module and framework context to avoid symbol-level misinterpretation.
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
- key public symbols for Python files
- entrypoints such as `__main__.py`, CLI entry points, ASGI/WSGI apps, workers, scheduled tasks, and test harnesses
- startup, shutdown, config, secret-source, and dependency overview where evidenced
- totals and `UNREVIEWED/INACCESSIBLE`

### PHASE 2 - Function Accounting
Produce exactly:
- `symbol_index.csv`
- `module_index.csv`

Rules:
- include functions, methods, classes, properties, fixtures, tests, `__init__`, and `__main__`
- chunk outputs to 500 rows max per file part
- leave caller or callee fields blank when precision is not supportable and note `INFERENCE`

### PHASE 3 - Architecture + Data Boundaries
Using phase 1 and 2 evidence:
- describe architecture as implemented
- identify framework structure
- map ingress and egress
- identify validation points and missing validation points
- identify leakage between transport, domain, persistence, and framework code
- assess transaction boundaries, async boundaries, and dependency direction

### PHASE 4 - Observability + Security Audit
Review:
- logging structure and correlation
- metrics, tracing, health checks, shutdown, drain behavior
- trust boundaries, authn/authz, input validation, unsafe deserialization, subprocess use, template risks, path handling, secret handling

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
Use Python Audit Mode with Python Engineering Policy and Python Engineering Workflow.
Audit /path/to/repo.
Execute PHASE 3 - Architecture + Data Boundaries.
Focus on framework boundaries, async behavior, and worker entrypoints.
Summarize generated code instead of expanding it.
```
