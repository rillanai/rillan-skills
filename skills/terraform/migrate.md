<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.1.0 -->
# Terraform Migration Planning

## Purpose
Use this skill when planning or executing migrations for Terraform codebases. This includes Terraform version upgrades (1.x to 1.y), provider version upgrades, state migrations, module refactoring, backend changes, tool migrations, and state splitting operations.

This skill defines the migration planning and execution contract for Terraform work. It is intended for version upgrades, state operations, module restructuring, backend migrations, and tool transitions.

This guidance is backend- and cloud-neutral and applies to Terraform or OpenTofu. When the migration targets or originates from HCP Terraform / Terraform Cloud (the `cloud {}` backend, workspace renames/reorgs, cross-org moves, variable sets, or an `azurerm`/`s3`/`gcs` → TFC switch), load the `hcp-terraform` skill — it owns those TFC-specific procedures. The backend-change mechanics here apply to any backend pairing.

## Skill Use
- Load this skill when the task involves any kind of Terraform migration, upgrade, or structural change to state, modules, or backends.
- Treat this skill as the governing contract for migration safety and execution discipline unless the repository has stricter local conventions.
- Keep repository-specific migration requirements in the task prompt.
- Match established project migration conventions when they are clear and defensible.
- When this skill conflicts with casual convenience, follow this skill.

## Knowledge-Graph Discovery (When Available)
If a graphify knowledge graph (`graphify-out/`) is present, use `graphify query`/`graphify path` to enumerate affected sites and latent cross-module couplings before sequencing the migration — graph traversal surfaces indirect dependencies that text search misses. Confirm every graph-derived call site or dependency with structural tooling, and run `graphify update .` after each step so the map tracks the migration.

If no `graphify-out/` directory exists, ignore this section.

## Tool Use
This skill is tool-agnostic and works with Claude Code, Codex, OpenCode, and similar assistants. Map its guidance to whatever file-reading, editing, search, and shell-execution tools your environment exposes.

- Use tools to inventory modules, provider pins, state backends, and workspace layout before proposing a migration plan — do not estimate blast radius from memory.
- Issue independent tool calls (listing workspaces, reading lockfiles, checking CI) in parallel.
- Run `terraform plan` after each migration step and report actual output, not expected output.
- For state operations (`moved`, `import`, backend changes), inspect the plan and state listings rather than asserting correctness.

## When To Use
Use this skill when the user asks for any of the following:
- Terraform version upgrades (1.x to 1.y)
- Provider version upgrades with breaking changes
- Backend migrations (changing backends — any backend to any backend)
- State migrations (backend changes, state restructuring, cross-state moves)
- Module refactoring (extracting modules, splitting modules, reorganizing module boundaries)
- Tool migrations (Terraform to OpenTofu and back)
- State splitting (monolith state to multiple state units)
- Import of existing infrastructure into Terraform management
- Moving resources between state files or isolation units

For HCP Terraform–specific migrations (workspace restructuring, cross-org moves, variable sets, migrating an existing backend to the `cloud {}` block), use the `hcp-terraform` skill alongside this one.

Do not use this skill for:
- Writing new Terraform code without a migration context (use a Terraform development skill)
- Documentation tasks (use the Terraform documentation skill)
- Test strategy or generation (use the Terraform test skill)
- Routine `terraform apply` operations that do not involve migration

## Operating Stance
- Act as a Senior Infrastructure Engineer who has performed dozens of production state migrations and version upgrades.
- Prefer safety over speed. Every migration step must be independently verifiable.
- Treat state as the most critical artifact. State loss or corruption can cause infrastructure orphaning, duplicate resource creation, or accidental destruction.
- Describe the migration as it must be executed, not as it would work in theory.
- Read the actual Terraform code, state structure, and provider configurations before recommending migration steps.
- When uncertainty exists about the impact of a migration step, recommend testing in non-production first and document the uncertainty explicitly.

## Migration Types

### Terraform Version Upgrades (1.x to 1.y)
Minor version upgrades within the 1.x series. Review the upgrade guide for the target version and check for new features that replace workarounds in current code.

**Upgrade process**:
1. Review the upgrade guide for the target version.
2. Check for new features that replace workarounds in current code.
3. Verify provider compatibility with the new Terraform version.
4. Run `terraform init -upgrade` and `terraform plan` to verify no unexpected changes.
5. Test in non-production environments first.

**Key feature introductions by version**:
- **1.1**: `moved` blocks for refactoring without state surgery.
- **1.2**: Preconditions and postconditions in lifecycle blocks.
- **1.3**: Optional object attributes in variable type constraints.
- **1.4**: `null_resource` replaced by `terraform_data`.
- **1.5**: `import` blocks (declarative import) and `check` blocks (continuous validation). `terraform plan -generate-config-out` for import.
- **1.6**: `terraform test` with `.tftest.hcl` files.
- **1.7**: `removed` blocks (declarative resource removal from state), mock providers in tests.
- **1.8**: Provider-defined functions, `templatestring` function.
- **1.9+**: Check upgrade guides for the specific version.

**Feature adoption during upgrades**:
When upgrading, look for opportunities to adopt new features that improve the codebase:
- Replace `terraform state mv` workflows with `moved` blocks (1.1+).
- Add preconditions and postconditions to critical resources (1.2+).
- Replace `terraform import` CLI usage with `import` blocks (1.5+).
- Add `terraform test` files for untested modules (1.6+).
- Replace `terraform state rm` workflows with `removed` blocks (1.7+).
- Adopt mock providers in tests to reduce cloud costs (1.7+).

### Provider Version Upgrades
Provider upgrades can introduce breaking changes to resource schemas, behavior, and defaults.

- **Review the changelog** for every major version between the current and target versions. Do not skip intermediate versions.
- **Identify breaking changes**: Removed attributes, renamed attributes, changed defaults, removed resources, renamed resources.
- **Check resource schema changes**: Attributes that changed from optional to required, type changes, new required nested blocks.
- **Handle deprecated resources**: Replace deprecated resources with their replacements. Use `moved` blocks when the replacement has a different resource type.
- **Plan before apply**: After updating the provider version constraint, run `terraform init -upgrade` and `terraform plan`. A clean upgrade produces a no-op plan. Any planned changes must be reviewed and understood.
- **Pin provider versions**: Use `~>` constraints for minor version flexibility or exact pins for critical infrastructure. Never use `>=` without an upper bound on production infrastructure.

**Common provider upgrade patterns**:

azurerm major version upgrades often involve:
- Resource renames (e.g., resources moving to new API versions).
- Attribute restructuring (nested blocks becoming separate resources).
- Default value changes for security-related attributes.
- Required attributes that were previously optional.

google/google-beta provider upgrades often involve:
- Resources moving between `google` and `google-beta`.
- Field deprecations in favor of new field names.
- Changes to default values for IAM-related resources.

### State Migrations
Moving resources within state, between state files, or between backends.

**terraform state mv**:
```bash
# Move a resource to a new address within the same state
terraform state mv azurerm_virtual_network.old azurerm_virtual_network.new

# Move a resource into a module
terraform state mv azurerm_virtual_network.main module.network.azurerm_virtual_network.main

# Move a module to a new name
terraform state mv module.old_name module.new_name
```

**terraform import**:
```bash
# Import an existing resource into state
terraform import azurerm_resource_group.main /subscriptions/{sub-id}/resourceGroups/rg-example

# Import into a module
terraform import module.network.azurerm_virtual_network.main /subscriptions/{sub-id}/resourceGroups/rg-example/providers/Microsoft.Network/virtualNetworks/vnet-example
```

**moved blocks (Terraform 1.1+)**:
```hcl
moved {
  from = azurerm_virtual_network.old
  to   = azurerm_virtual_network.new
}

moved {
  from = azurerm_network_security_group.app
  to   = module.security.azurerm_network_security_group.app
}

moved {
  from = module.old_name
  to   = module.new_name
}
```

Prefer `moved` blocks over `terraform state mv` when possible. `moved` blocks are declarative, reviewable in code review, and execute as part of the normal plan/apply cycle. `terraform state mv` is an imperative operation that modifies state directly and is harder to review and audit.

**import blocks (Terraform 1.5+)**:
```hcl
import {
  to = azurerm_resource_group.main
  id = "/subscriptions/{sub-id}/resourceGroups/rg-example"
}

import {
  to = google_compute_network.main
  id = "projects/my-project/global/networks/vpc-main"
}
```

Prefer `import` blocks over `terraform import` CLI when possible. `import` blocks are declarative, can be code-reviewed, and generate configuration with `terraform plan -generate-config-out=generated.tf`.

**removed blocks (Terraform 1.7+)**:
```hcl
removed {
  from = azurerm_network_security_rule.legacy

  lifecycle {
    destroy = false
  }
}
```

Use `removed` blocks to remove resources from Terraform management without destroying them. This replaces the `terraform state rm` workflow with a declarative, reviewable approach.

**Cross-backend migration**:
When moving resources between separate state files managed by different backends:
```bash
# In the source configuration
terraform state mv -state-out=../destination/terraform.tfstate \
  azurerm_virtual_network.main azurerm_virtual_network.main

# Or pull state, manipulate, push
terraform state pull > source-state.json
# Edit as needed
terraform state push destination-state.json
```

### Changing Backends (Any Backend To Any Backend)
Moving state from one backend to another (e.g., `local` → `s3`, `azurerm` → `gcs`, `s3` → `cloud`). The mechanics are the same regardless of the backend pairing.

1. Back up the current state: `terraform state pull > state-backup-$(date +%Y%m%d%H%M%S).json`.
2. If the destination requires a pre-existing state container (bucket, prefix, workspace), create it first.
3. Update the `backend` (or `cloud`) block in the `terraform` block to the new configuration. Example, `azurerm` → `s3`:

Before:
```hcl
terraform {
  backend "azurerm" {
    resource_group_name  = "rg-terraform-state"
    storage_account_name = "stterraformstate"
    container_name       = "tfstate"
    key                  = "production/networking.tfstate"
  }
}
```

After:
```hcl
terraform {
  backend "s3" {
    bucket       = "my-tf-state"
    key          = "production/networking.tfstate"
    region       = "us-east-1"
    use_lockfile = true
  }
}
```

4. Run `terraform init -migrate-state` and confirm the migration when prompted.
5. Verify with `terraform plan` showing no changes.
6. Do not delete the source state until the migration is fully verified across all environments.
7. Update documentation and runbooks to reflect the new backend.
8. Update cross-state references: consumers using `data "terraform_remote_state"` must point at the new backend config.

For migrations into HCP Terraform (the `cloud {}` block, workspace renames/reorgs, cross-org moves, and variable-set handling), see the `hcp-terraform` skill — those procedures layer on top of this one.

### Module Refactoring
Restructuring module boundaries without affecting deployed infrastructure.

**Extracting inline resources to modules**:
1. Create the new module with the resources.
2. Add `moved` blocks mapping old addresses to new module addresses.
3. Run `terraform plan` to verify no changes (the `moved` blocks handle the address translation).
4. Apply to update state.
5. After successful apply, `moved` blocks can be kept for history or removed in a subsequent change.

**Splitting large modules**:
1. Identify cohesive resource groups that should be separate modules.
2. Create the new modules.
3. Add `moved` blocks for every resource being relocated.
4. Update the calling module to invoke both new modules.
5. Run `terraform plan` to verify no changes.
6. Apply, verify, then optionally clean up `moved` blocks.

**Module versioning strategy**:
- Use semantic versioning for shared modules.
- When using a registry (public, or a private registry such as HCP Terraform / Spacelift / Scalr), tag releases in the module repository; the registry picks up tagged versions.
- Pin consumers to specific versions or version ranges.
- Document breaking changes in a changelog.
- Provide migration guides when releasing major versions.

**Migrating module sources**:
When moving modules between sources (local → git, git → registry, etc.):
```hcl
# Before: Git source
module "network" {
  source = "git::https://github.com/org/terraform-modules.git//network?ref=v1.0.0"
}

# After: registry source (public registry shown; a private registry uses app.terraform.io/{org}/{module}/{provider})
module "network" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "5.1.0"
}
```

After changing the module source, run `terraform init` to download from the new location. The module content must be identical. Run `terraform plan` to verify no changes.

### Backend Migrations
Moving state between storage backends.

**terraform init -migrate-state**:
```bash
# Update the backend configuration in the terraform block
# Then run:
terraform init -migrate-state
```

This interactively migrates state from the old backend to the new backend. For automation, use `-input=false` with appropriate backend configuration.

**Restructuring isolation units (e.g., CLI-workspace-per-env → directory-per-env)**:
When migrating from a single state with multiple CLI workspaces to a directory-per-environment layout:
1. For each existing state, pull it: `terraform state pull > <name>.tfstate`.
2. Create separate directory structures for each environment.
3. Create the new state unit for each environment directory (bucket key, prefix, or workspace).
4. Push state to each new unit: `terraform state push <name>.tfstate`.
5. Configure VCS/runner connections for each new unit.
6. Verify each environment with `terraform plan` showing no changes.
7. Decommission the old state units.

For HCP Terraform workspace restructuring specifically (renames, reorgs, cross-org moves, variable sets), see the `hcp-terraform` skill.

### Tool Migrations

**Terraform to OpenTofu (and back)**:
OpenTofu is a first-class peer to Terraform; the migration is mechanical for compatible versions.
- OpenTofu 1.6.x is compatible with Terraform 1.6.x state and configuration.
- Replace the `terraform` CLI with `tofu` CLI (or vice versa). Commands are identical.
- Update CI/CD pipelines to use the chosen binary.
- Backend configuration: most backends (`local`, `s3`, `gcs`, `azurerm`, `consul`, `kubernetes`, `http`) work with both tools. The HCP Terraform `cloud {}` backend is Terraform-only — if migrating off it to OpenTofu, switch to another supported backend first (see the `hcp-terraform` skill for that backend's specifics).
- Update provider lock files: run `tofu init -upgrade` (or `terraform init -upgrade`).
- Verify with `tofu plan` (or `terraform plan`) showing no changes.
- Update documentation references to the chosen tool.
- Key differences to address:
  - Registry: OpenTofu uses its own registry. Most providers are available.
  - Licensing: OpenTofu is MPL 2.0; Terraform is BSL.
  - Features: the two may diverge over time. Check compatibility for the specific version.
  - State encryption: OpenTofu supports native state encryption, which Terraform does not.

### State Splitting
Breaking a monolith state file into multiple state files or isolation units.

**When to split state**:
- State file is large enough to cause slow plans (hundreds of resources).
- Different teams own different parts of the infrastructure.
- Different change cadences (networking changes rarely, application infra changes frequently).
- Blast radius reduction: a bad apply should not risk unrelated infrastructure.

**Splitting into separate state units**:
1. Back up the monolith state.
2. Identify cohesive resource groups for each new state unit.
3. Create the new state unit for each group (bucket key, prefix, or workspace).
4. Write the Terraform configuration for each new root module.
5. Move resources from the monolith state to each new unit:
```bash
# Pull monolith state
cd monolith/
terraform state pull > ../monolith-backup.json

# Move resources to new state files
terraform state mv -state-out=../networking/terraform.tfstate \
  aws_vpc.main aws_vpc.main
terraform state mv -state-out=../networking/terraform.tfstate \
  aws_subnet.app aws_subnet.app
terraform state mv -state-out=../networking/terraform.tfstate \
  aws_subnet.data aws_subnet.data

# Push to the new state unit's backend
cd ../networking/
terraform state push terraform.tfstate
```
6. Verify each new state unit with `terraform plan` showing no changes.
7. Verify the monolith `terraform plan` shows only the removed resources as "will be destroyed" (because they are no longer in its configuration).

**Dependency management between split units**:
Use `data "terraform_remote_state"` (or the backend's equivalent) to share values between units:
```hcl
data "terraform_remote_state" "networking" {
  backend = "s3"
  config = {
    bucket = "my-tf-state"
    key    = "platform/prod/networking.tfstate"
    region = "us-east-1"
  }
}

resource "aws_instance" "app" {
  subnet_id = data.terraform_remote_state.networking.outputs.app_subnet_id
  ami       = var.ami_id
  # ...
}
```

`terraform_remote_state` is the backend-agnostic default. On HCP Terraform, `data "tfe_outputs"` is a purpose-built alternative that integrates with TFC access controls — see the `hcp-terraform` skill.

## Migration Planning Process

### State Backup
**ALWAYS back up state before any migration operation.** No exceptions.

```bash
# For remote-backend state
terraform state pull > state-backup-$(date +%Y%m%d%H%M%S).json

# For local state
cp terraform.tfstate terraform.tfstate.backup.$(date +%Y%m%d%H%M%S)
```

Some backends maintain their own state version history usable for rollback (versioned S3 buckets, HCP Terraform state history, etc.). Regardless, always keep an independent backup before migration operations.

Keep backups until the migration is fully verified and stable in all environments.

### Impact Analysis
Before executing any migration:
- **List all resources in state**: `terraform state list` to understand the scope.
- **Identify affected resources**: Which resources will be moved, imported, or re-addressed.
- **Determine blast radius**: What is the worst case if this migration goes wrong? Data loss? Downtime? Resource recreation?
- **Check for stateful resources**: Databases, storage accounts, DNS records, and certificates are high-risk. Recreation means data loss or downtime.
- **Check for resources with prevent_destroy**: These will block accidental destruction but should still be migrated carefully.
- **Check cross-state dependencies**: Which other state units consume outputs from the unit being migrated? Those consumers will need updates.

### Dependency Analysis
- **What depends on migrated resources?** Other state units reading outputs via `terraform_remote_state` (or the backend's equivalent), applications using resource IDs, monitoring pointing at resource names.
- **What do migrated resources depend on?** Resources in other state units, external systems, service principals, network connectivity.
- **Are there circular dependencies?** State splitting can reveal circular dependencies that must be broken.

### Risk Assessment
Rate each migration step on:
- **Data loss potential**: Can this step cause data loss? (Database migrations, storage changes, encryption key rotations.)
- **Downtime potential**: Can this step cause service interruption? (Resource recreation, DNS changes, network changes.)
- **Rollback difficulty**: How hard is it to undo? (Restoring a previous state version is straightforward. Recreated databases are not.)
- **Blast radius**: How many services or users are affected if this goes wrong?

### Plan/Apply Diff Analysis
After every migration step:
- Run `terraform plan` and verify the output.
- The ideal result is "No changes. Your infrastructure matches the configuration."
- Any planned changes must be reviewed and explicitly approved. Unexpected changes indicate a migration error.
- Pay special attention to "must be replaced" actions. These mean resource destruction and recreation.
- If the plan shows unexpected changes, stop. Diagnose before proceeding.

### Phased Execution Plan
Structure every migration as a series of small, independently verifiable steps:
1. Each step produces a clean `terraform plan` with no unexpected changes.
2. Each step can be paused and resumed without leaving state in an inconsistent state.
3. Each step has a documented rollback procedure.
4. Steps are ordered to minimize blast radius: start with the lowest-risk resources.

### Rollback Plan
For every migration, document:
- **State restore procedure**: How to restore the backed-up state file — `terraform state push`, or the backend's own state-version rollback (versioned bucket, HCP Terraform state history).
- **Version pin rollback**: How to revert Terraform or provider version constraints.
- **Code rollback**: Which git commit to revert to.
- **Backend config rollback**: How to revert backend/state-unit configuration changes. (For HCP Terraform workspace settings — variables, VCS connections, run triggers — see the `hcp-terraform` skill.)
- **Timing**: How long the rollback takes and what happens to changes made between migration and rollback.

## Migration Execution Rules
These rules are non-negotiable for any migration operation:

- **ALWAYS backup state before migrating.** Run `terraform state pull` and save the output before any state-modifying operation. This is not optional.
- **Verify with terraform plan showing no changes after each step.** Every migration step must produce a clean plan. If the plan shows unexpected changes, stop and diagnose.
- **Never migrate state and change resources simultaneously.** Migration steps should be pure state operations. Infrastructure changes should be separate commits and separate applies.
- **Keep each migration step small and independently verifiable.** Do not batch dozens of state moves into one operation. Move a few resources, verify, then continue.
- **Use moved blocks over state surgery when possible (Terraform 1.1+).** `moved` blocks are declarative, reviewable, and reversible. `terraform state mv` is imperative and harder to audit.
- **Use import blocks over terraform import CLI when possible (Terraform 1.5+).** `import` blocks are declarative and can be code-reviewed. `terraform import` CLI modifies state directly.
- **Use removed blocks over terraform state rm when possible (Terraform 1.7+).** `removed` blocks are declarative and reviewable.
- **Lock state during migration.** Use the backend's native locking (S3 lockfile/DynamoDB, Azure Blob lease, GCS, or HCP Terraform's automatic locking). If manual locking is needed, use `terraform force-unlock` only as a last resort.
- **Test migration in non-production first.** Always migrate dev or staging before production. Verify the process works and the plan is clean before touching production state.
- **Document every migration step as it is executed.** Keep a log of commands run, plan output reviewed, and verification results. This log is critical for debugging if something goes wrong.
- **Never run terraform apply -auto-approve during migration.** Always review the plan manually during migration operations. Auto-approve is for routine CI/CD, not for state surgery.

## Terraform-Specific Migration Tools

### terraform state Commands
```bash
terraform state list                    # List all resources in state
terraform state show <address>          # Show details of a specific resource
terraform state mv <src> <dst>          # Move a resource to a new address
terraform state rm <address>            # Remove a resource from state (does not destroy it)
terraform state pull                    # Download remote state to stdout
terraform state push <file>             # Upload state to the configured backend
terraform state replace-provider <old> <new>  # Replace a provider in state
terraform force-unlock <lock-id>        # Manually unlock state (last resort)
```

### moved Blocks (Terraform 1.1+)
Declarative resource address changes. Applied during `terraform plan` and `terraform apply`.

```hcl
# Rename a resource
moved {
  from = azurerm_network_security_group.web_nsg
  to   = azurerm_network_security_group.application
}

# Move into a module
moved {
  from = azurerm_network_security_group.app
  to   = module.security.azurerm_network_security_group.app
}

# Move between modules
moved {
  from = module.old.azurerm_storage_account.data
  to   = module.new.azurerm_storage_account.data
}

# Rename a module
moved {
  from = module.legacy_network
  to   = module.networking
}

# Move indexed resources (count to for_each)
moved {
  from = azurerm_subnet.app[0]
  to   = azurerm_subnet.app["app"]
}
```

### import Blocks (Terraform 1.5+)
Declarative resource import. Applied during `terraform plan` and `terraform apply`.

```hcl
import {
  to = azurerm_resource_group.main
  id = "/subscriptions/{sub-id}/resourceGroups/rg-example"
}

import {
  to = google_compute_network.main
  id = "projects/my-project/global/networks/vpc-main"
}

# Generate configuration for imported resources
# terraform plan -generate-config-out=generated.tf
```

### removed Blocks (Terraform 1.7+)
Declarative resource removal from state without destroying the resource.

```hcl
removed {
  from = azurerm_network_security_rule.legacy_allow_all

  lifecycle {
    destroy = false
  }
}
```

### terraform init -migrate-state
Migrates state between backends. Run after changing the `backend` or `cloud` block in the `terraform` block.

```bash
terraform init -migrate-state
```

For automated pipelines:
```bash
terraform init -migrate-state -input=false
```

### tfmigrate
Third-party tool for declarative state migrations:
```hcl
# tfmigrate configuration
migration "state" "rename_resource" {
  dir = "."
  actions = [
    "mv azurerm_virtual_network.old azurerm_virtual_network.new",
  ]
}

migration "state" "split_to_module" {
  dir = "."
  actions = [
    "mv azurerm_virtual_network.main module.networking.azurerm_virtual_network.main",
    "mv azurerm_subnet.app module.networking.azurerm_subnet.app",
  ]
}
```

```bash
tfmigrate plan migration.hcl   # Dry run
tfmigrate apply migration.hcl  # Execute
```

## Evidence Rules
- Ground all migration recommendations in actual state analysis, plan output, and code analysis.
- Do not recommend migration steps without first reading the relevant Terraform configuration and understanding the current state structure.
- When recommending version upgrades, cite the specific changelog entries or upgrade guide sections that apply.
- When recommending state operations, verify the resource addresses exist in state.
- When estimating blast radius, base it on actual resource dependencies, not assumptions.
- When recommending backend/state-unit changes, verify current backend configuration and cross-state dependencies. (For HCP Terraform workspace/variable-set changes, see the `hcp-terraform` skill.)
- If a migration path has not been tested in the current codebase, label it as untested and recommend non-production verification first.

## Anti-Patterns To Reject
- **Migrating without state backup**: Every state-modifying operation must be preceded by a state backup. No exceptions. No shortcuts.
- **terraform apply -auto-approve during migration**: Never auto-approve during migration. Always review the plan manually.
- **Migrating production first**: Always migrate non-production environments first. Verify the process and the resulting plan before touching production.
- **State surgery when moved blocks would work**: Prefer `moved` blocks over `terraform state mv`. `moved` blocks are reviewable, declarative, and part of the normal plan/apply workflow.
- **terraform state rm when removed blocks would work**: Prefer `removed` blocks (1.7+) over `terraform state rm`. `removed` blocks are reviewable and declarative.
- **Mixing migration with infrastructure changes**: A migration commit should contain only migration operations (moved blocks, import blocks, state moves). Infrastructure changes (new resources, modified configurations) should be separate commits and separate applies.
- **Skipping the no-op plan verification**: After every migration step, run `terraform plan` and verify it shows no unexpected changes. Skipping this step means you do not know if the migration succeeded.
- **Big-bang migrations**: Moving all resources in one operation. If it fails, everything is in an unknown state. Break migrations into small, verifiable steps.
- **Migrating state across environment boundaries without isolation**: Using the same state operations on resources from different environments in a single operation.
- **Ignoring provider version compatibility during Terraform upgrades**: Upgrading Terraform without verifying that all providers are compatible with the new version.
- **Deleting backup state before verifying the migration**: After migrating to a new backend or state unit, always verify with `terraform plan` before deleting backup state files or decommissioning old backends.
- **Using terraform state push without understanding the contents**: `terraform state push` overwrites remote state. Only use it when you are certain the local state is correct and complete.
- **Force-unlocking state without understanding who holds the lock**: `terraform force-unlock` should be a last resort. Verify that no other process is actively modifying state before forcing an unlock.
- **Forgetting to update cross-state references**: When renaming or moving a state unit, all `data "terraform_remote_state"` (or `tfe_outputs`) references in consumers must be updated.

## Invocation Template
Use this skill with a prompt that supplies repository-specific context. Example:

```text
Use Terraform Migration Planning.
Plan the upgrade from Terraform 1.5 to 1.8 for the infrastructure at /path/to/repo.
Identify breaking changes in providers, recommend moved blocks for the module refactoring,
and produce a phased execution plan with rollback procedures.
Split the networking and compute state into separate isolation units.
Test the migration path in the dev environment first.
```
