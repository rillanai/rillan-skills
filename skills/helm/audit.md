<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.0.0 -->
# Helm Audit Deep Dive

## Purpose
Use this skill for a phased, evidence-based audit of Helm charts and chart repositories.

This skill is the audit contract for chart structure, values surface, template safety, release ergonomics, upgrade risk, and operational clarity.

## Skill Use
- Load this skill when the user explicitly wants a deep Helm or chart audit.
- Treat this skill as the governing audit contract for the turn or session.
- Keep repository-specific scope, focus areas, and exclusions in the invoking prompt.
- Execute only the requested phase and stop at the phase boundary.

## Knowledge-Graph Discovery (When Available)
If a graphify knowledge graph (`graphify-out/`) is present, seed inventory and architecture mapping from `graphify-out/GRAPH_REPORT.md` and `graphify query`/`graphify path` instead of starting cold — then verify every graph-derived claim against repository evidence before recording it. `INFERRED` and `AMBIGUOUS` edges are leads, not findings; an unverified graph edge is `INFERENCE`, not evidence.

## Tool Use
This skill is tool-agnostic and works with Claude Code, Codex, OpenCode, and similar assistants. Map its guidance to whatever file-reading, editing, search, and shell-execution tools your environment exposes.

- Every factual claim in an audit must come from a tool invocation, not inference. Read the template, values file, or helper before writing the finding.
- Render charts with `helm template` against representative values files when behavior needs verification.
- Issue independent tool calls (inventory scans, multi-file reads, subchart walks) in parallel.
- If evidence cannot be gathered, record it under `UNREVIEWED/INACCESSIBLE` rather than guessing.

## When To Use
Use this skill for:
- deep chart audits
- release safety reviews
- values sprawl or helper-template complexity reviews
- security and operability assessment for Helm-packaged workloads
- evidence-backed upgrade planning or chart simplification work

Do not use this skill for:
- ordinary template edits
- narrow rendering bug hunts
- one-change patch review without broader audit intent

## Required Inputs
The invoking prompt must provide:
- repository path or chart scope
- exact phase to execute

Recommended inputs:
- focus charts or subcharts
- focus values files or environments
- exclusions
- whether generated manifests should be summarized or expanded
- previous phase artifacts or `STATE_SNAPSHOT` when continuing

If scope or phase is missing, stop and ask.

## Operating Stance
- Prefer evidence over intuition.
- Describe the chart behavior as implemented, not as intended.
- Stay phase-disciplined.
- Separate templating correctness from workload correctness.
- Treat `Chart.yaml`, `values.yaml`, schema files, templates, hooks, helpers, docs, and release tooling as first-class evidence.

## Evidence Rules
- Every factual claim must be anchored to a chart path, template name, helper name, value key, hook, or rendered object when applicable.
- Mark any non-provable conclusion as `INFERENCE`.
- List inaccessible or unreviewed material under `UNREVIEWED/INACCESSIBLE` with impact notes.
- Do not imply runtime certainty unless it is supported by chart code, rendered output, tests, release tooling, or docs.

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
1. Establish accessible chart scope and obvious exclusions.
2. Build a fast inventory of charts, subcharts, values files, helpers, hooks, CRDs, and release assets.
3. Read the files relevant to the current phase before making conclusions.
4. Build inventories or rendered evidence before evaluative claims.
5. Preserve phase boundaries strictly.

## Phase Gate Rules
- Phase 1 may inventory and describe, but must not recommend.
- Phase 2 may account and index, but must not recommend or grade.
- Phase 3 may assess chart boundaries, upgrade risk, and release architecture, but must not produce detailed remediation plans.
- Phase 4 may produce prioritized findings with fixes, but must not assign overall grades.
- Phase 5 may synthesize, grade, prioritize, and plan.

## Phase Rules

### PHASE 1 - Inventory + Chart Surface
Produce:
- chart inventory grouped by chart or subchart
- one-line purpose for each chart, dependency, helper file, hook, and values file
- values surface summary: main values files, schema coverage, and environment overlays where evidenced
- release asset summary: CRDs, hooks, helper stacks, chart dependencies, and testing or packaging files
- totals and `UNREVIEWED/INACCESSIBLE`

### PHASE 2 - Template + Values Accounting
Produce exactly:
- `template_index.csv`
- `values_index.csv`

Rules:
- include one row per rendered template file, helper file, hook template, and CRD manifest
- include one row per meaningful top-level value key and notable nested compatibility boundary
- chunk outputs to 500 rows max per file part
- leave rendered object counts or ownership fields blank when precision is not supportable and note `INFERENCE`

### PHASE 3 - Release Architecture + Boundaries
Using phase 1 and 2 evidence:
- describe chart architecture as implemented
- map chart, subchart, helper, and values ownership boundaries
- identify unstable selectors, names, labels, immutable field risk, and hook side effects
- assess values hygiene: duplicate knobs, undocumented keys, weak defaults, and compatibility risk
- identify where a problem belongs: chart contract, templating logic, or workload manifest behavior

### PHASE 4 - Security + Operability Findings
Review:
- RBAC, service accounts, secrets handling, pod security context, and network exposure
- upgrade and rollback safety, dependency pinning, CRD handling, and install-vs-upgrade behavior
- operability signals such as probes, resource defaults, and workload-facing hooks where chart evidence exists

Output findings grouped by `P0`, `P1`, and `P2`, each with:
- chart path
- template, helper, or value key
- evidence
- concrete fix

### PHASE 5 - Synthesis
Produce:
- overall grade `A-F`
- subgrades for values design, template clarity, security, operability, upgrade safety, and docs/DX
- anchored justification
- prioritized recommendations with `P0`, `P1`, and `P2`
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
Use Helm Audit Deep Dive.
Audit /path/to/charts.
Execute PHASE 3 - Release Architecture + Boundaries.
Focus on values sprawl, hook side effects, and upgrade-sensitive resources.
Summarize generated manifests instead of expanding them.
```
