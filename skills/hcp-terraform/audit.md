<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.3.0 -->
# HCP Terraform Audit Layer

Apply this overlay only when repository evidence or the task establishes HCP
Terraform usage. `terraform/audit.md` remains the complete contract: execute the
same requested phase, obey its gates, evidence and uncertainty rules, chunking,
failure behavior, footer, and stop. This file adds only TFC-specific deliverables.

Repository configuration may not expose control-plane state such as current team
access, workspace settings, sensitive flags, variable-set attachments, run history,
or policy-set assignments. Use accessible TFC API/UI evidence when supplied; otherwise
record each unavailable area under the base `UNREVIEWED/INACCESSIBLE` rules instead
of inferring it from HCL.

## PHASE 1 - Inventory + Entrypoints

Add inventories for:

- workspaces: name, mapped directory, organization, execution mode, and tags;
- every `data "tfe_outputs"` consumer, source organization/workspace, and count;
- visible workspace variables and variable sets, attachments, purpose, and
  `sensitive` status;
- `app.terraform.io/...` module sources and version constraints;
- Sentinel policy sets, enforcement levels, and attachments;
- VCS- or CLI-driven runs, run triggers, scheduled runs, and approval gates.

Report TFC totals. Inventory and describe only.

## PHASE 2 - Resource Accounting

Include every accessible `tfe_*` resource and `data.tfe_outputs` block in the base
CSV artifacts. Preserve the base schemas and coverage rules; do not create a third
artifact family.

## PHASE 3 - Architecture + State Boundaries

Map directory-to-workspace boundaries, environment and cluster isolation, naming
consistency, and per-workspace blast radius. Trace all `tfe_outputs` producer-to-
consumer edges, cycles and fan-out, and compare run triggers with that dependency
graph. Assess oversized or over-fragmented workspaces and overly broad or conflicting
variable-set attachments. Cite both the producer workspace/output and consuming
`data "tfe_outputs"` block for cross-workspace claims.

## PHASE 4 - Security + Compliance Audit

Review evidenced team permissions, plan/apply separation, state/output access,
static versus dynamic provider credentials, sensitive variables, Sentinel coverage
and enforcement, production approvals, scheduled drift runs, trigger correctness,
and private-module trust and pinning. Fold findings into the base `P0`/`P1`/`P2`
format; do not grade.

## PHASE 5 - Synthesis

Add a `TFC Integration` `A`–`F` subgrade covering workspace naming and mapping,
`tfe_outputs`, run triggers, Sentinel, and variable-set hygiene. Fold TFC evidence
into State Management, Security, and CI/CD where relevant, and add TFC remediation
to the base roadmap. Tie grades and recommendations to earlier-phase evidence.

For private-registry evidence, cite source and version together. For Sentinel, cite
the policy plus its evidenced enforcement level and attachment.
