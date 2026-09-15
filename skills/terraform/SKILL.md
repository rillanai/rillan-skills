---
name: terraform
description: Terraform / HCL work — author or review `.tf` modules, resources, and providers; plan/apply safety; state moves; testing (terraform test, Terratest, OPA/checkov); module docs; deep audits; and provider/version migrations. Triggers on `*.tf`, `terraform.tfvars`, `.terraform/`, or HCL. Root skill that routes to its mode files.
---

<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.3.0 -->
# Terraform

Backend- and cloud-neutral router for Terraform and OpenTofu. Read only the mode
that matches the task; ordinary targeted review does not trigger `audit.md`.

- `dev.md` — author, debug, change, or perform a targeted review of HCL.
- `test.md` — test or validate Terraform behavior and policy.
- `audit.md` — run an explicitly requested phased deep audit; requires scope and phase.
- `docs.md` — write module and operational documentation.
- `migrate.md` — change providers, versions, state, or module structure.

For those modes that have an HCP overlay, also load the relevant
`hcp-terraform` mode only when the task involves `cloud {}`, `tfe_outputs`, TFC
workspaces, the private registry, variable sets, Sentinel, or TFC runs. Use this
pack's `test.md` for testing. Explicit repository instructions take precedence.
