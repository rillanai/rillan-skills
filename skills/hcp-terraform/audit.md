<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.1.0 -->
# HCP Terraform Audit Layer

## Purpose
Use this overlay when auditing a Terraform codebase that runs on HCP Terraform / Terraform Cloud. It adds a TFC-specific layer to the phased, evidence-based audit defined in `terraform/audit.md`. Run the base audit first; this file specifies the **additional** inventory, analysis, findings, and grading that TFC introduces. It does not replace any base phase.

The base contract still governs: same phase gates, same evidence/anti-hallucination rules, same chunking and footer, same `INFERENCE`/`UNREVIEWED/INACCESSIBLE` discipline. Every TFC claim here must be anchored to repository evidence (`cloud {}` blocks, `tfe_outputs` data sources, `tfe` provider resources, policy sets, CI/run config) the same way.

## When This Overlay Applies
Apply the TFC layer only when the audited repo shows evidence of HCP Terraform: `cloud {}` backend blocks, `data "tfe_outputs"` references, `app.terraform.io` module sources, `tfe_*` provider resources, or Sentinel policy sets. If none are present, the base audit is sufficient — do not fabricate a TFC layer.

## Phase 1 Additions — TFC Inventory
In addition to the base Phase 1 inventory, produce:
- **TFC workspace inventory**: workspace name, mapped directory, organization, execution mode (remote/local/agent), and any workspace tags. Cite the `cloud {}` block or `tfe_workspace` resource per entry.
- **`tfe_outputs` accounting**: every `data "tfe_outputs"` block — source organization + workspace and the consuming module/file. Count them.
- **Variable / variable-set inventory** (when visible in code or `tfe` resources): workspace variables, variable sets and their attachments, which carry provider credentials vs. tags vs. environment values, and `sensitive` status.
- **Private-registry module inventory**: modules sourced from `app.terraform.io/...` with version pins.
- **Sentinel policy inventory**: policy sets, enforcement levels, and workspace/org attachment.
- **Run configuration**: VCS-driven vs CLI-driven workspaces, run triggers between workspaces, scheduled runs, Confirm & Apply / approval gates.
- **TFC totals**: total workspaces, total `tfe_outputs` references, total variable sets, total private-registry modules, total Sentinel policy sets.

Constraints: inventory and describe only — no recommendations or grades (base Phase 1 rule).

## Phase 3 Additions — Workspace Boundaries & Blast Radius
On top of the base Phase 3 state-boundary analysis:
- **Workspace boundary map**: how many workspaces, what is in each, isolation level per environment and per cluster, and workspace-naming-convention consistency against the directory→workspace mapping.
- **Cross-workspace coupling**: all `data "tfe_outputs"` references — direction of dependency flow, which workspaces are producers vs. consumers, whether the graph has cycles or excessive fan-out, and whether run triggers match the data-dependency graph.
- **Workspace boundary quality**: are boundaries at the right granularity? Flag mega-workspaces (high blast radius) and over-fragmented workspaces (excessive coupling).
- **Blast radius per workspace**: maximum damage a single apply can cause in each workspace.
- **Variable-set sprawl**: variable sets attached too broadly (credential over-distribution) or precedence conflicts between workspace and variable-set values.

## Phase 4 Additions — TFC Security & Policy
On top of the base Phase 4 security audit:
- **Workspace access controls**: team permissions, separation of plan vs apply permissions, who can read state/outputs, run permissions. Flag over-broad access.
- **Secrets handling on TFC**: provider credentials and secrets in workspace variables / variable sets (marked `sensitive`?), use of dynamic provider credentials / OIDC vs static keys.
- **Sentinel coverage**: which policies are enforced, at what level (`advisory`/`soft-mandatory`/`hard-mandatory`), and gaps where org policy is not codified.
- **Run workflow safety**: production workspaces require Confirm & Apply (no `-auto-approve`), drift detection via scheduled runs, run-trigger correctness.
- **Supply chain**: trust and pinning of private-registry modules.

Findings follow the base format (`P0`/`P1`/`P2`, file path + resource address + evidence + concrete fix). No grades in this phase.

## Phase 5 Additions — TFC Integration Grade
Add one subgrade to the base Phase 5 synthesis:
- **TFC Integration** (`A–F`) — workspace naming consistency, cross-workspace reference patterns (`tfe_outputs` hygiene), run-trigger correctness, Sentinel policy coverage and enforcement levels, variable-set hygiene, and workspace-to-directory mapping quality.

Fold TFC-specific items into the base subgrades where they belong (State Management ← workspace isolation/coupling; Security ← workspace access controls + Sentinel; CI/CD ← run workflow), and report **TFC Integration** as the dedicated extra axis. Tie every grade to anchored evidence from the phases above. TFC-specific remediation items join the base roadmap, prioritized by severity — do not invent fixed calendar dates.

## Evidence Formatting (Additions)
- For cross-workspace references, cite both the producing workspace/output and the consuming `data "tfe_outputs"` block.
- For private-registry modules, cite the `source = "app.terraform.io/..."` line and its version constraint.
- For Sentinel, cite the policy file and its enforcement level / attachment.

## Invocation Template
```text
Use the terraform audit skill plus the hcp-terraform audit overlay on /path/to/repo.
Execute PHASE 1 — Inventory + Entrypoints, including the TFC workspace inventory and tfe_outputs accounting.
Emphasize workspace boundaries, cross-workspace coupling, variable-set hygiene, and Sentinel coverage.
```
