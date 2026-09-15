<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.3.0 -->
# HCP Terraform Migration Planning

Load `terraform/migrate.md` first. This overlay covers the HCP Terraform details for
`cloud {}`, workspace identity and settings, remote runs, variables, output sharing,
and organization boundaries. Its state backup, locking, plan review, authorization,
verification, and rollback rules still apply.

## Inventory and Freeze
Record workspace IDs as well as names, organization and project, Terraform version,
execution mode and agent pool, VCS connection, working directory, variables and
variable sets, dynamic credentials, teams, policies, run triggers, notifications,
state-sharing permissions, and `tfe_outputs` consumers. Sensitive variable values are
write-only and may require the secrets owner to recreate them.

Before moving state, stop or finish pending and active runs, disable new VCS/run-trigger
activity, and lock every affected workspace. HashiCorp's migration guidance requires
[pausing operations](https://developer.hashicorp.com/terraform/enterprise/migrate);
a state lock alone does not account for queued automation. Keep automation disabled,
hold locks between operations, unlock only the workspace needed for the exact reviewed
migration run, and relock it immediately afterward. Restore the intended trigger and
lock state only after verification.

## Existing Backend to `cloud {}`
Create and configure the destination workspace first. Match the source execution inputs
and grant only the output/state access consumers need. Back up every source workspace,
then replace the backend configuration with the intended `cloud {}` organization and
workspace mapping.

Run plain `terraform init` and review its migration prompts. HashiCorp's
[`cloud` connection documentation](https://developer.hashicorp.com/terraform/cli/cloud/settings)
specifically directs this workflow and may prompt to map CLI workspaces to HCP
workspaces. Do not substitute generic backend flags without checking the pinned CLI's
documented `cloud` behavior.

Verify the destination workspace ID, state inventory, variables, execution mode, and a
plan with no unexplained remote changes. Verify all consumers before leaving the source
immutable for the stability window. Decommission it only after the rollback boundary
has passed.

## Rename or Reorganize Workspaces
A rename changes name-based API URLs and references even though the workspace ID stays
fixed. Update every `cloud {}` mapping, `tfe_outputs` reference, API/CI integration,
run trigger, and document that embeds the old name. Re-plan the renamed workspace and
its consumers. See HCP Terraform
[workspace settings](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings).

For a state split, create and fully configure each destination workspace, including
variables, VCS/working directory, permissions, policies, and run triggers. Then follow
the base mode's declarative remove-then-import batches. A destination plan does not
“adopt” state: it must explicitly show imports, followed by exclusive ownership and
zero remote creates, updates, or destroys, followed by clean source, destination, and
consumer plans. Execute only the exact HCP platform run whose plan was reviewed. Do not
use casual `state mv -state-out` or `state push` recipes against remote workspaces.

Prefer `tfe_outputs` for intentional cross-workspace interfaces. Publish only required
outputs and set explicit state-sharing permissions; sequence run triggers according to
the actual producer/consumer graph.

## Organization or Platform Boundaries
Do not assume a rename, copy, or state upload transfers workspace metadata. Determine
the supported operation for the exact platform and release. Terraform Enterprise has a
documented [workspace transfer](https://developer.hashicorp.com/terraform/enterprise/workspaces/transfer)
workflow that preserves specified history and metadata; confirm its prerequisites and
review stage. If no supported transfer exists, treat the move as a new workspace plus
the base mode's controlled ownership transfer, and inventory/recreate variables,
variable sets, teams, policies, VCS, triggers, notifications, and output permissions.

When leaving HCP Terraform, first migrate ownership to a backend supported by the
target engine using HashiCorp's documented source workflow. Replace `tfe_outputs` and
workspace-managed inputs, then verify producers and consumers. For OpenTofu, check the
exact version pair and avoid writing OpenTofu-specific state until all Terraform
consumers have migrated.

Rollback must identify which state is authoritative after the cutover. Restoring an
older HCP state version can discard later writes and does not reverse remote changes;
freeze runs, reconcile drift, and verify consumers before resuming automation.
