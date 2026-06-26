<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.1.0 -->
# HCP Terraform Documentation Guidance

## Purpose
Use this overlay when documenting a Terraform codebase that runs on HCP Terraform / Terraform Cloud. It adds the TFC-specific documentation surfaces to `terraform/docs.md`. Load that base file first — its rules (every variable/output has a `description`, every module has a hand-maintained README, document what the code does not what it should do, code is the source of truth) apply unchanged. This file covers only TFC-specific documentation: workspace docs, variable sets, run triggers, and the TFC-UI runbook.

## TFC Workspace Documentation
When HCP Terraform is the backend, document:

- **Workspace Naming Convention**: the pattern used for workspace names (e.g., `{project}-{environment}-{component}`) and how names map to infrastructure boundaries and to the `environments/` directory tree.
- **Organization Structure**: how workspaces are organized within the TFC organization, which teams have access, and how RBAC (plan vs apply permissions) is configured.
- **Cross-Workspace References**: which workspaces share data via `data "tfe_outputs"`, the direction of data flow, and what outputs are consumed.
- **Variable Sets**: which variable sets are attached to which workspaces and what they contain (provider credentials, common tags, environment-specific values). Note workspace-vs-variable-set precedence.
- **Run Triggers**: configured run triggers between workspaces and their purpose; keep the trigger graph documented and acyclic.
- **VCS Integration**: how workspaces connect to VCS repositories, which branches trigger which workspaces, and working-directory configuration.

Example cross-workspace reference documentation:

```
Workspace Dependencies:
  platform-prod-network
    └─ Consumed by: platform-prod-compute (via tfe_outputs: vpc_id, subnet_ids)
    └─ Consumed by: platform-prod-database (via tfe_outputs: private_subnet_id)

  platform-prod-identity
    └─ Consumed by: platform-prod-compute (via tfe_outputs: instance_role_arn)
```

## Output Documentation (TFC)
- When an output is consumed by another TFC workspace via `data "tfe_outputs"`, name the consuming workspace in the output's `description`:

```hcl
output "vpc_id" {
  description = "ID of the created VPC. Consumed by the compute and database workspaces via tfe_outputs."
  value       = aws_vpc.main.id
}
```

## Root Module / Workspace Context
For each root module on TFC, document the **TFC Workspace** alongside the base root-module docs: workspace name, organization, which variables and variable sets supply it, and cross-workspace dependencies via `data "tfe_outputs"`.

## TFC Runbook (UI Workflow)
The base `terraform/docs.md` runbook covers the CLI saved-plan workflow and generic state recovery. On TFC, document the UI-driven workflow:

```markdown
### Applying Changes via TFC

1. Push your changes to the VCS branch connected to the target workspace.
2. Navigate to the workspace in the TFC UI.
3. Review the auto-triggered (speculative) plan. Check:
   - Resource additions, changes, and destructions.
   - No unexpected "must be replaced" actions.
   - Output changes match expectations.
   - Sentinel policy checks pass (or have documented overrides).
4. If the plan is correct, click "Confirm & Apply" and add a comment describing the change.
5. If the plan shows unexpected changes, click "Discard Run" and investigate.
6. After apply completes, verify the workspace outputs and check resource health.

**Rollback**: In the TFC UI, open the workspace's state history, select the previous
state version, and click "Rollback to this state". Then trigger a new plan to reconcile.
```

Also document, as TFC additions to the base runbook:
- **State recovery on TFC**: downloading state from TFC, restoring a previous state version via the UI/API, handling state lock conflicts.
- **Drift remediation**: TFC scheduled runs / health assessments as the drift signal, and how to reconcile vs accept.
- **Access**: how to obtain TFC access, required team membership, and authenticating for CLI-driven runs against the `cloud {}` backend.

## Quality Checklist (Additions)
On top of the base docs checklist, verify:
- TFC workspace documentation covers naming, cross-workspace references, variable sets, and run triggers.
- Runbook procedures reflect the TFC UI workflow (Confirm & Apply, Discard Run, state-history rollback) and Sentinel checks.
- Documented workspace names, variable sets, and `tfe_outputs` references match the actual `cloud {}` blocks, `tfe_*` resources, and data sources.

## Invocation Template
```text
Use the terraform documentation skill plus the hcp-terraform documentation overlay.
Document the prod-compute root module: its TFC workspace, the variable sets that supply provider
credentials, and the cross-workspace dependencies it reads from prod-shared via tfe_outputs.
Include the TFC UI apply/rollback runbook.
```
