<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

# Skill Improvement Plan

## Goal

Reduce avoidable skill context and duplicated instructions while preserving pack
paths, installer interfaces, safety rules, evidence requirements, and phase
boundaries. This pass prioritizes Socratic reasoning as a reusable audit of
judgment and Terraform as the main stack-specific pilot. Python and Rust remain
deferred until these results are reviewed.

## Acceptance criteria

- Socratic Mode keeps the human's decisions visible, asks factual questions
  before a defensible position when needed, honors explicit `off`, marks checks
  `PASS`, `FLAG`, or `N/A`, and reports a concise verdict without exposing an
  internal reasoning transcript.
- Terraform reductions preserve evidence rules, safety gates, useful
  distinctions, and explicit phase pauses.
- Routers select the smallest applicable mode, and host skills own output
  formats and phase control.
- Metadata, references, packaging, `task vet`, `task test`, and representative
  dry-run installs pass before this pass is called complete.
- Behavioral evidence is described accurately: static review and controlled
  scenario exercises check contracts and discipline; they are not broad model
  efficacy benchmarks.

## Baseline and current measurements

The pre-refactor source baseline supplies the baseline values; current values are
`wc -w` counts from the checked-out files. Word counts proxy context load and are
not exact token counts. Terraform rows include their routers; HCP rows include
the base and overlay.

| Workflow | Baseline | Current | Change |
|---|---:|---:|---:|
| Terraform dev | 6,305 | 1,168 | -81.5% |
| Terraform test | 4,374 | 1,142 | -73.9% |
| Terraform docs | 3,750 | 800 | -78.7% |
| Terraform audit | 4,405 | 1,743 | -60.4% |
| Terraform migrate | 4,534 | 1,878 | -58.6% |
| Terraform pack total | 22,376 | 5,979 | -73.3% |
| HCP audit (base + overlay) | 5,597 | 2,296 | -59.0% |
| HCP migrate (base + overlay) | 5,901 | 2,660 | -54.9% |

The Socratic router is now 1.1.0 and fell from 3,249 to 1,341 words
(-58.7%). This is a content-load metric, not a token measurement. The current
pass aligns selected-level and Socratic-only hooks in planning, ADR, RFC,
security, and CI, while retaining host formats and phase boundaries. The
integrated hooks fell from 2,140 to 1,014 words (-52.6%).

## Milestones

### 1 — Go pilot (S, complete)

Go routine dev fell from 4,426 to 2,532 words (42.8%); policy remained intact.
Static scenarios corrected baseline-only routing for investigate, review, and
verify. Vet, tests, packaging, and project/global dry-run checks
passed. It remains a comparison point, not proof of model behavior.

### 2 — Socratic reasoning and Terraform pass (L, implementation complete)

Socratic Mode now supports four explicit levels, synthesis before preference
questions, visible consequential decisions, applicable `N/A`, scoped memory
with consent, and a concise `PASS`/`FLAG` verdict. Integration hooks are aligned
across the selected host skills; old audit appendices and the Python/Rust
workflow step 3 are updated to use Socratic checks where applicable.

Terraform and HCP Terraform routers and modes are compacted at version 3.3.0
where applicable. The five audit phases still pause at each boundary: execute
only the requested phase, stop, and wait for the next explicit phase. A static
review and one controlled Phase 2 scenario checked the exact CSV contract,
restored definitions for path, count, and address semantics, and confirmed the
correct `NEXT` for Phase 3 without auto-advancing. The scenario exposed and
recorded inaccessible upstream state; it is a contract and phase-discipline
check, not a broad benchmark.

The current pass is verified: all 68 Markdown files pass metadata checks, audit
phase/router/workflow preservation checks pass, and `git diff --check` is clean.
Byte-exact installs across 15 packs for Claude, Codex, and OpenCode pass, as do
repository detection and project/global install/uninstall dry-runs. The full
`go tool task ci` run passes with the compatible Go toolchain; the Go 1.27.1
linter analyzer crash is toolchain-specific. The known Taskfile glob mismatch
(`skills/*/*.skill.md` versus embedded `skills/*/*.md`) remains deferred.

### 3 — Shared composition and examples (deferred, M)

Measure common and subset-install loads before moving guidance or examples, and
decide installer support from those measurements.

### 4 — Python/Rust general rollout (deferred, M)

Apply only patterns validated by the Socratic/Terraform pass after review.

## Rollback and done criteria

Keep milestones in separate commits and revert an affected pack independently
while preserving paths and metadata. This pass is verified; shared composition,
examples, and later Python/Rust rollout remain deferred.
