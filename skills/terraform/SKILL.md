---
name: terraform
description: Terraform / HCL work — author or review `.tf` modules, resources, and providers; plan/apply safety; state moves; testing (terraform test, Terratest, OPA/checkov); module docs; deep audits; and provider/version migrations. Triggers on `*.tf`, `terraform.tfvars`, `.terraform/`, or HCL. Root skill that routes to its mode files.
---

<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.2.0 -->
# Terraform

Root skill for Terraform work. This `SKILL.md` is the only file loaded up front; it
routes to the mode files in this directory, which you read **on demand** with your file
tool. Do not guess a mode file's contents — read it.

This core is backend- and cloud-neutral: it applies to any backend (local, s3, azurerm,
gcs, consul, cloud/HCP, kubernetes, http), any cloud (AWS, Azure, GCP, vSphere, …), and
Terraform or OpenTofu. **Using HCP Terraform / Terraform Cloud? Also load the
`hcp-terraform` skill** — it overlays this one with TFC-specific depth (the `cloud {}`
block, `data "tfe_outputs"`, workspaces, private registry, Sentinel, run workflow).

## Modes — load the one matching the task
- `dev.md` — author or modify HCL: modules, resources, variables, outputs, provider config.
- `test.md` — `terraform test`, Terratest, OPA/Rego, checkov, plan-time assertions.
- `audit.md` — explicit, phased, evidence-based audit. User supplies scope + phase.
- `docs.md` — module READMEs, input/output references, `terraform-docs`.
- `migrate.md` — provider/version upgrades, state moves, module restructuring.

Cross-pack pointers use path form, e.g. `cicd/core.md`, `security/SKILL.md`. Stricter
repository-local conventions win when they are explicit and defensible.
