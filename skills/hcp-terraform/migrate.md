<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.1.0 -->
# HCP Terraform Migration Planning

## Purpose
Use this overlay when a migration targets or originates from HCP Terraform / Terraform Cloud. It adds the TFC-specific procedures to the migration contract in `terraform/migrate.md`. Load `terraform/migrate.md` first as the baseline — all of its safety discipline applies unchanged here:

- **ALWAYS back up state before any migration.** `terraform state pull > backup-$(date +%Y%m%d%H%M%S).json`.
- **Verify `terraform plan` shows no changes after each step.**
- **Keep steps small and independently verifiable; test in non-production first.**
- **Never `-auto-approve` during migration.**
- **Use `moved`/`import`/`removed` blocks over state surgery where possible.**

This file covers only the TFC-target/source specifics: migrating a backend into the `cloud {}` block, workspace renames/reorgs, cross-org moves, and variable handling.

## Migrating An Existing Backend → `cloud {}`
When moving state from any backend (`azurerm`, `s3`, `gcs`, `consul`, local, …) to a TFC workspace:

1. Ensure the target TFC workspace exists and is configured with the needed variables, variable sets, and VCS connection.
2. Back up the current state (`terraform state pull > backup.json`).
3. Change the `terraform` block from the old backend to `cloud {}`:

Before:
```hcl
terraform {
  backend "s3" {
    bucket = "my-tf-state"
    key    = "production/networking.tfstate"
    region = "us-east-1"
  }
}
```

After:
```hcl
terraform {
  cloud {
    organization = "my-org"

    workspaces {
      name = "platform-prod-networking"
    }
  }
}
```

4. Run `terraform init -migrate-state` and confirm when prompted.
5. Verify with `terraform plan` showing no changes.
6. Do not delete the source state until the TFC migration is verified across all environments.
7. Update documentation and runbooks to reflect the new TFC backend.
8. Update cross-state consumers: references that used `terraform_remote_state` against the old backend may move to `data "tfe_outputs"` once both sides are on TFC.

(The general backend-change mechanics — pre-creating the destination, `init -migrate-state`, no-op-plan verification — are in `terraform/migrate.md`; this section adds only the TFC target.)

## Workspace Renaming
1. Update the workspace name in TFC (UI or `tfe_workspace` resource / API).
2. Update the `name` in every affected `cloud {}` block.
3. Update all `data "tfe_outputs"` references in consuming workspaces that point at the old name.
4. Update VCS trigger configurations if the workspace name appears in branch/path filters.
5. Update any CI/CD scripts that reference the workspace name.
6. Run `terraform plan` in all affected workspaces to verify no unexpected changes.
7. Update documentation (READMEs, AGENTS.md, runbooks) with the new name.

## Workspace Reorganization (Splitting / Restructuring)
When splitting a monolith workspace into component workspaces (the state-move mechanics are the base skill's; the TFC wiring is here):
1. Back up the source workspace state via `terraform state pull` or the TFC API.
2. Create the new workspaces in TFC with appropriate naming, VCS connections, and variable configuration.
3. Move resources from the source workspace state to destination workspace states with `terraform state mv -state-out=...` (see `terraform/migrate.md` → State Splitting).
4. Push state into each new workspace (`terraform state push`), or let the new workspace's first plan adopt the migrated state.
5. Wire consumers with `data "tfe_outputs"`:
```hcl
data "tfe_outputs" "networking" {
  organization = "my-org"
  workspace    = "platform-prod-networking"
}
```
6. Run `terraform plan` in each new workspace to verify no changes.
7. Add or update run triggers between the new workspaces to match the data-dependency graph.

## Moving Workspaces Between Organizations
1. Document all workspace configuration: variables, variable sets, team access, VCS connections, run triggers, notification configs.
2. Pull state from the source workspace: `terraform state pull > migration-backup.json`.
3. Create the new workspace in the destination organization with matching configuration.
4. Push state to the new workspace: `terraform state push migration-backup.json`.
5. Update VCS connections to point to the correct repository.
6. **Recreate workspace variables and variable sets — they do NOT transfer between organizations.** Sensitive values must be re-entered manually.
7. Update all `data "tfe_outputs"` references in other workspaces to the new organization + workspace name.
8. Run `terraform plan` in the migrated workspace to verify no changes.
9. Verify consuming workspaces can read outputs from the new location.
10. Decommission the old workspace only after full verification.

## Managing TFC Variables During Migration
- **Terraform variables**: export non-sensitive values and recreate in the destination. For sensitive variables, coordinate with the secrets owner to re-enter.
- **Environment variables** (provider credentials): document and recreate; sensitive values re-entered manually. Prefer switching to dynamic provider credentials / OIDC during the migration if not already in use.
- **Variable sets**: verify which sets are attached and ensure equivalent sets exist in the destination org/workspace.
- **Variable precedence**: workspace-specific variables override variable-set values — document the intended precedence so the migration does not silently change resolved values.

## Migrating Off TFC (→ another backend / OpenTofu)
- The `cloud {}` backend is Terraform-only and not OpenTofu-compatible. To move to OpenTofu, first migrate state to a supported backend (`s3`, `gcs`, `azurerm`, `consul`, …) using `terraform init -migrate-state`, then follow the base skill's Terraform→OpenTofu steps.
- Replace `data "tfe_outputs"` references with `terraform_remote_state` against the new backend.
- Recreate workspace variables/variable sets as backend/runner variables (env vars, CI secrets, tfvars, Vault).

## Anti-Patterns To Reject (TFC-Specific)
- Forgetting to update `data "tfe_outputs"` references when renaming/moving workspaces.
- Assuming workspace variables and variable sets transfer between organizations — they do not; recreate them.
- Deleting the source backend's state before verifying the TFC migration with a clean plan.
- Migrating production workspaces first instead of validating in non-production.

## Invocation Template
```text
Use the terraform migration skill plus the hcp-terraform migration overlay.
Migrate the networking root module at /path/to/repo from the s3 backend to the cloud {} block
(org "my-org", workspace platform-prod-networking). Back up state, init -migrate-state, verify a no-op plan,
then update consumers from terraform_remote_state to data "tfe_outputs". Validate in dev first.
```
