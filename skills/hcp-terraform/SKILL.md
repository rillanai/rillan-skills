---
name: hcp-terraform
description: HCP Terraform / Terraform Cloud (TFC) overlay for tasks involving the `cloud {}` backend, `data "tfe_outputs"`, TFC workspaces or variable sets, the private registry (`app.terraform.io`), Sentinel policy-as-code, TFC runs, or `tfe` provider resources. Load with the base `terraform` skill only for those concerns.
---

<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.3.0 -->
# HCP Terraform

TFC-specific overlay for the base `terraform` skill. Load it only when the task
involves `cloud {}`, `tfe_outputs`, TFC workspaces or variable sets, private-registry
modules, Sentinel, or TFC run behavior. Read the base mode and one matching overlay:

- `dev.md` — configure TFC-backed Terraform.
- `audit.md` — add TFC evidence and deliverables to the same requested audit phase.
- `docs.md` — document workspaces, variables, triggers, and runbooks.
- `migrate.md` — move backends, workspaces, organizations, or variable sets.

The base skill governs general Terraform concerns; this pack adds only TFC-specific
requirements.
