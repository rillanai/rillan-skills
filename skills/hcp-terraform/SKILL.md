---
name: hcp-terraform
description: HCP Terraform / Terraform Cloud (TFC) overlay — the `cloud {}` backend, `data "tfe_outputs"` cross-workspace references, TFC workspaces and the directory+workspace hybrid, the private registry (`app.terraform.io`), Sentinel policy-as-code, variable sets, and the TFC run workflow (speculative plans, Confirm & Apply, run triggers, scheduled runs). Load IN ADDITION TO the base `terraform` skill whenever a project uses HCP Terraform / Terraform Cloud. Triggers on `cloud {}` blocks, `tfe_outputs`, `app.terraform.io`, Sentinel, TFC workspaces, or `tfe` provider resources.
---

<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.1.0 -->
# HCP Terraform

Overlay skill for HCP Terraform / Terraform Cloud (TFC) work. This `SKILL.md` is the only
file loaded up front; it routes to the mode files in this directory, which you read **on
demand** with your file tool. Do not guess a mode file's contents — read it.

**This is an overlay, not a standalone skill.** It is a specialization of the base
`terraform` skill, not a replacement. Load the base `terraform` skill (its `SKILL.md` and
the matching mode file) as the baseline for all general Terraform/OpenTofu quality,
state, module, and review discipline; the mode files here add only the TFC-specific
depth. When TFC guidance conflicts with generic Terraform guidance, follow the TFC file
for HCP Terraform concerns. When the project is NOT on HCP Terraform, do not load this
pack — the base `terraform` skill is backend-neutral and already covers it.

Load this pack when the project uses the `cloud {}` backend, references
`data "tfe_outputs"`, sources modules from `app.terraform.io`, runs Sentinel policies,
or manages TFC workspaces / variable sets.

## Modes — load the one matching the task, on top of `terraform/<same mode>`
- `dev.md` — the `cloud {}` backend, TFC workspace organization and the directory+workspace hybrid, private-registry module sourcing, `data "tfe_outputs"`, variable sets, Sentinel, and the TFC run workflow.
- `audit.md` — the TFC audit layer: workspace inventory, `tfe_outputs` accounting, workspace-boundary/blast-radius analysis, and the **TFC Integration** grade dimension, framed as additions to the base terraform audit phases.
- `docs.md` — TFC workspace documentation, variable sets, run triggers, and the TFC-UI plan/apply/rollback runbook.
- `migrate.md` — migrating a backend to the `cloud {}` block, TFC workspace renames/reorgs, cross-org moves, and variable-set handling during migration.

Cross-pack pointers use path form, e.g. `terraform/dev.md`. Sentinel is HCP Terraform–only;
OPA/conftest and Checkov work on any runner and stay in the base `terraform` skill.
