<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.1.0 -->
# Terraform Documentation Guidance

## Purpose
Use this skill when generating, reviewing, or maintaining documentation for Terraform codebases. This includes module READMEs, root module documentation, variable and output descriptions, architecture decision records, runbooks, changelogs, AGENTS.md contributor guidance, and environment documentation.

This skill defines the documentation contract for Terraform work. It is intended for README generation, module documentation, input/output documentation, ADRs for infrastructure decisions, operational runbooks, AGENTS.md files for AI-assisted contributor guidance, state-backend documentation, and changelog maintenance.

This guidance is backend- and cloud-neutral and applies to Terraform or OpenTofu. When the project uses HCP Terraform / Terraform Cloud, load the `hcp-terraform` skill for the TFC-specific documentation depth (workspace documentation, variable sets, run triggers, and the TFC-UI plan/apply runbook).

## Skill Use
- Load this skill when the task is to create, update, or review documentation for Terraform code.
- Treat this skill as the governing contract for documentation quality and completeness unless the repository has stricter local conventions.
- Keep repository-specific documentation requirements in the task prompt.
- Match established project documentation conventions when they are clear and defensible.
- When this skill conflicts with casual convenience, follow this skill.

## Tool Use
This skill is tool-agnostic and works with Claude Code, Codex, OpenCode, and similar assistants. Map its guidance to whatever file-reading, editing, search, and shell-execution tools your environment exposes.

- Read the HCL before documenting it; do not write documentation from memory or assumption.
- Verify variable/output descriptions, defaults, and validation blocks against the actual `.tf` files.
- Run `docs.md` (or equivalent) to confirm generated sections match the current module surface.
- Issue independent tool calls (reading multiple modules, scanning for resources) in parallel.
- Cite the file path when documenting behavior — do not paraphrase without grounding.

## When To Use
Use this skill when the user asks for any of the following:
- README generation for a Terraform module or root module
- Module documentation including usage examples, inputs, and outputs
- Variable or output description review and improvement
- AGENTS.md files for contributor and AI-assistant guidance
- State-backend documentation including naming conventions, cross-state references, and isolation-unit structure
- Architecture Decision Records for infrastructure choices
- Operational runbooks for Terraform / OpenTofu workflows
- Changelog generation or maintenance for infrastructure changes
- Environment documentation covering promotion flows and access requirements
- Documentation audits or gap analysis for Terraform repositories

Do not use this skill for:
- Writing or modifying Terraform code itself (use a Terraform development skill)
- Test strategy or test generation (use the Terraform test skill)
- Migration planning (use the Terraform migration skill)
- Pure infrastructure design without a documentation deliverable

## Operating Stance
- Act as a Senior Infrastructure Engineer who treats documentation as a first-class operational artifact.
- Prefer evidence over aspiration. Document what the code does, not what it should do.
- Read the actual Terraform source before writing documentation. Do not invent resources, variables, outputs, or provider configurations.
- Keep documentation accurate, scannable, and operationally useful.
- When documentation and code diverge, flag the divergence and correct the documentation to match the code.

## Documentation Types And Standards

### Module README
Every Terraform module must have a README.md. Module READMEs are hand-maintained by default. The README must include:

- **Purpose**: A concise statement of what the module provisions and why it exists as a module.
- **Usage Example**: A complete, valid HCL block showing how to call the module with realistic inputs. Use `module` blocks with actual variable values, not placeholder comments.
- **Requirements**: Minimum Terraform version, required provider versions, and any external dependencies.
- **Providers**: Which providers the module uses and their expected configuration.
- **Resources Managed**: A table of all resources and data sources the module creates, reads, or manages. Include the resource type, logical name, and a brief description. Keep this table updated when resources change.
- **Inputs Table**: All variables with columns for Name, Description, Type, Default, and Required.
- **Outputs Table**: All outputs with columns for Name and Description.
- **Identity and RBAC Overview** (when applicable): Document which managed identities, service principals, or service accounts the module creates, what roles or permissions are assigned, and to which scopes. This is especially important for modules that create identity resources or assign roles.

Example usage block format:

```hcl
module "network" {
  source = "../../modules/network"

  name_prefix        = "production"
  address_space      = ["10.0.0.0/16"]
  location           = var.location
  resource_group_name = azurerm_resource_group.main.name
  enable_nat_gateway = true

  subnets = {
    app = {
      address_prefix = "10.0.1.0/24"
    }
    data = {
      address_prefix = "10.0.2.0/24"
    }
  }

  tags = {
    Environment = "production"
    Team        = "platform"
  }
}
```

### terraform-docs Integration (Recommended for Input/Output Tables)
Hand-maintained READMEs are the default for prose documentation — purpose statements, usage examples, resource tables, identity overviews, and RBAC mappings. These sections require human judgment and cannot be generated.

However, **input/output tables should be automated with terraform-docs.** Hand-maintained input/output tables inevitably drift from the actual variable and output blocks as modules evolve. This is a known and common documentation failure mode.

**Recommended hybrid approach**:
- Use terraform-docs injection markers (`<!-- BEGIN_TF_DOCS -->` / `<!-- END_TF_DOCS -->`) for input/output tables only.
- Keep all other README sections (purpose, usage, resource tables, identity docs, RBAC mappings) as hand-maintained prose outside the markers.
- Run terraform-docs in CI as a drift check — if the generated tables don't match the committed README, the CI check fails.

```yaml
# .terraform-docs.yml
formatter: markdown table
output:
  file: README.md
  mode: inject

sections:
  show:
    - inputs
    - outputs
```

**When NOT to use terraform-docs**:
- When the generated output would overwrite valuable hand-maintained context within the injection markers.
- When a module's variable documentation requires rich prose that the `description` field alone cannot convey — in that case, hand-maintain the table but keep the `description` fields accurate as the source of truth.

### AGENTS.md (Contributor Guidance)
AGENTS.md files provide guidance for AI-assisted development tools and human contributors working in the repository. They complement READMEs by documenting conventions that are not obvious from the code alone.

An AGENTS.md file should include:

- **Repository Structure**: How the repository is organized — module directories, environment directories, shared configuration, test locations.
- **Naming Conventions**: Resource naming patterns, variable naming conventions, module naming standards, file naming expectations.
- **Coding Standards**: Required Terraform formatting, variable ordering conventions, when to use `count` vs `for_each`, when to create a module vs inline resources.
- **Provider Patterns**: Which providers are in use (e.g., azurerm, google, vsphere, infoblox, tfe), how provider aliases are structured, authentication patterns.
- **Testing Expectations**: Where tests live, what testing framework is used (e.g., native `terraform test`), what test coverage is expected for new modules.
- **Documentation Expectations**: What documentation must be updated when code changes, whether READMEs are hand-maintained or auto-generated.
- **Backend/Workflow Conventions**: State-unit naming patterns, how cross-state references work, where variables are supplied (tfvars, env vars, CI secrets, workspace variables). For HCP Terraform specifics, see the `hcp-terraform` skill.
- **Commit and PR Conventions**: Commit message format, PR description requirements, required reviewers.

Place AGENTS.md at the repository root for repository-wide guidance. Place additional AGENTS.md files in subdirectories when those directories have conventions that differ from the root.

### Root Module Documentation
Root modules (environments, stacks, deployments) require documentation that covers:

- **Environment Overview**: What this root module deploys and into which subscription, project, region, or datacenter.
- **Prerequisites**: What must exist before applying — subscriptions, resource groups, bootstrap resources, service principals, network connectivity.
- **Backend Configuration**: Where state is stored (S3 bucket + key, GCS bucket + prefix, azurerm container, HCP Terraform workspace, etc.), how to access it, locking mechanism.
- **Authentication**: How to authenticate to the provider — OIDC role assumption, service principals, managed identities, service accounts, environment variables, credential files.
- **State/Isolation Unit**: The state unit this root module owns (workspace name, state key, prefix), where its variables come from, and any cross-state dependencies via `terraform_remote_state` (or the backend's equivalent). For HCP Terraform workspace details, see the `hcp-terraform` skill.
- **Deployment Procedures**: Step-by-step instructions for plan, apply, and destroy workflows including CLI-driven runs, runner workflows, and any approval gates.
- **Variable Configuration**: How variables are supplied — tfvars files, environment variables, CI secrets, backend/runner variables.
- **Dependencies**: Other root modules, state units, or external systems this deployment depends on.

### State-Backend And Cross-State Documentation
Whatever backend the project uses, document:

- **Naming Convention**: The pattern used for state-unit names (e.g., `{project}-{environment}-{component}`) and how names map to infrastructure boundaries.
- **Access Structure**: Who can read state and run plan/apply, and how access is controlled (bucket/IAM policies, workspace team permissions, etc.).
- **Cross-State References**: Which units share data via `terraform_remote_state` (or the backend's equivalent), the direction of data flow, and what outputs are consumed.
- **Variable Supply**: Where variables and provider credentials come from (tfvars, env vars, CI secrets, secret managers).
- **VCS/Runner Integration**: How applies are triggered, which branches deploy where, and working-directory configuration.

For HCP Terraform–specific documentation (workspace organization, variable sets, run triggers), see the `hcp-terraform` skill.

Example cross-state reference documentation:

```
State Dependencies:
  platform-prod-network
    └─ Consumed by: platform-prod-compute (via remote_state: vpc_id, subnet_ids)
    └─ Consumed by: platform-prod-database (via remote_state: private_subnet_id)

  platform-prod-identity
    └─ Consumed by: platform-prod-compute (via remote_state: instance_role_arn)
```

### Variable Documentation
Every `variable` block must include a `description` field. Variable documentation rules:

- The description must explain what the variable controls and how it affects the infrastructure, not just restate the variable name.
- Include information about valid values, constraints, or formats when the type alone is insufficient.
- Document the implications of the default value when one is set.
- When validation rules exist, the description should summarize what is validated.
- For complex types (objects, maps of objects), document the structure and purpose of each attribute.

Bad:
```hcl
variable "vm_size" {
  description = "The VM size"
  type        = string
  default     = "Standard_D2s_v3"
}
```

Good:
```hcl
variable "vm_size" {
  description = "Azure VM size for the application servers. Use Standard_B2s for dev/staging, Standard_D2s_v3 or Standard_D4s_v3 for production workloads. Must be a size available in the target region."
  type        = string
  default     = "Standard_D2s_v3"

  validation {
    condition     = can(regex("^Standard_", var.vm_size))
    error_message = "VM size must be a Standard-tier Azure VM size."
  }
}
```

### Output Documentation
Every `output` block must include a `description` field. Output documentation rules:

- The description must explain what the output value represents and who or what consumes it.
- When an output is consumed across a state boundary (via `terraform_remote_state` or the backend's equivalent), document the consuming state unit.
- When an output contains sensitive data, the `sensitive = true` flag must be set and the description should note this.
- For outputs that expose resource identifiers, document what resource they reference.

Good:
```hcl
output "vpc_id" {
  description = "ID of the created VPC. Consumed by the compute and database state units via terraform_remote_state."
  value       = aws_vpc.main.id
}

output "subnet_ids" {
  description = "Map of subnet name to subnet ID for all subnets in the VPC. Used by the compute unit for instance placement and the database unit for private endpoints."
  value       = { for k, v in aws_subnet.main : k => v.id }
}
```

### Architecture Decision Records
ADRs for infrastructure document the reasoning behind structural decisions that are not obvious from the code alone. Infrastructure ADRs should cover:

- **Module Structure Decisions**: Why resources are grouped into these modules, why certain resources are in the root module instead of a child module, why a module boundary exists where it does.
- **Provider Decisions**: Why this cloud provider or this specific provider configuration — multi-cloud strategy, provider aliases, provider version constraints, why one cloud for this workload vs another.
- **State Strategy Decisions**: Why this backend (e.g., S3 vs GCS vs HCP Terraform), why state is split this way across isolation units, why this state structure.
- **Networking Decisions**: Why this address space, why this connectivity model (peering, VPN, Direct Connect, ExpressRoute, Cloud Interconnect), why these regions or zones.
- **Security Decisions**: Why this identity model (managed identity, service principal, service account), why these role assignments, why this secret management approach.
- **Cost Decisions**: Why this VM sizing strategy, why reserved vs on-demand, why this autoscaling configuration.

ADR format:
```markdown
# ADR-NNN: Title

## Status
Accepted | Superseded by ADR-NNN | Deprecated

## Context
What is the situation that requires a decision?

## Decision
What was decided and why?

## Consequences
What are the tradeoffs? What becomes easier? What becomes harder?
```

### Runbooks
Operational runbooks document procedures that operators need to execute. Terraform / OpenTofu runbooks must cover:

- **Plan and Apply Procedures**: How to trigger a plan (CLI, VCS push, runner), how to review plan output, how to approve and apply, how to cancel a run, the saved-plan workflow (`terraform plan -out=tfplan` then `terraform apply tfplan`).
- **State Recovery**: How to recover from corrupted state, how to download/inspect state from the backend, how to restore a previous state version, how to handle state lock conflicts (`terraform force-unlock` as a last resort).
- **Import Procedures**: How to import existing resources into state (using `import` blocks or `terraform import`), and post-import plan verification.
- **Drift Remediation**: How to detect drift (scheduled plan runs, manual plan), how to decide between reconciling in Terraform vs accepting the drift, how to handle manual changes made outside Terraform.
- **Incident Response for Infrastructure**: What to do when apply fails mid-way, how to handle provider outages, how to roll back a bad apply by restoring a previous state version.
- **Access and Permissions**: How to obtain access to state and the execution backend, required membership, how to authenticate to cloud providers for local CLI runs.

When the backend is HCP Terraform, the TFC-UI runbook steps (Confirm & Apply, Discard Run, state-history rollback) live in the `hcp-terraform` skill.

Each runbook procedure should include:
- Prerequisites
- Step-by-step commands or UI instructions
- Expected output at each step
- What to do if the step fails
- Who to escalate to

Example CLI-driven workflow runbook entry:

```markdown
### Applying Changes via CLI

1. Pull the latest code for the target state unit and run `terraform init`.
2. Run `terraform plan -out=tfplan` and review the plan. Check:
   - Resource additions, changes, and destructions.
   - No unexpected "must be replaced" actions.
   - Output changes match expectations.
3. If the plan is correct, get approval, then apply the saved plan: `terraform apply tfplan`.
4. If the plan shows unexpected changes, stop and investigate before applying.
5. After apply completes, verify outputs and check resource health.

**Rollback**: Restore the previous state version from the backend (versioned bucket object, backend state history, or a pre-change backup), then run a fresh `terraform plan` to reconcile.
```

### Changelogs
Infrastructure changelogs track changes to modules and root modules over time.

- Document breaking changes in module interfaces: removed variables, changed variable types, renamed outputs, changed resource addressing that forces replacement.
- Document new resources added, resources removed, and resources that will be replaced.
- Note provider version bumps and Terraform version requirements changes.
- For modules consumed by multiple teams, distinguish between internal changes (no interface impact) and breaking changes (consumers must update).

Format:
```markdown
## [1.2.0] - 2025-01-15

### Breaking Changes
- Removed `enable_legacy_mode` variable. All deployments now use the current architecture.
- Renamed output `db_endpoint` to `database_endpoint` for consistency.

### Added
- High-availability NAT gateway mode with one gateway per zone.
- Support for private endpoints on database resources.

### Changed
- Default VM size updated from Standard_D2s_v3 to Standard_D4s_v3.
- Minimum Terraform version bumped to 1.5.0.

### Fixed
- Role assignment ordering that caused intermittent plan diffs.
```

### Environment Documentation
Each environment (dev, staging, production, etc.) should have documentation covering:

- **What the environment contains**: Which state units are deployed, which services run there, what data it holds.
- **Promotion Flow**: How changes move from one environment to the next — VCS branch strategy, runner triggers, manual promotion, approval gates.
- **Access Requirements**: Who can access the environment, what membership/permissions are needed, what cloud-provider credentials are required, what approval is required.
- **Differences from Other Environments**: What is intentionally different — VM sizes, replica counts, feature flags, monitoring thresholds, network connectivity.
- **Cost Profile**: Approximate cost, reserved capacity, spot/preemptible usage.

## Terraform-Specific Documentation Rules
These rules apply to all documentation generated or reviewed under this skill:

- Every `variable` block must have a `description` argument. No exceptions.
- Every `output` block must have a `description` argument. No exceptions.
- Every module (child or root) must have a `README.md`.
- Hand-maintain READMEs by default. Include resource tables, identity overviews, and RBAC mappings that automated tools cannot generate. Use terraform-docs only as an optional supplementary automation layer when adopted by the team.
- Document required permissions, roles, or service account bindings needed to apply the Terraform configuration.
- Document provider authentication requirements including service principals, managed identities, service accounts, and required environment variables.
- Document state backend access requirements — S3/GCS bucket permissions, azurerm container access, HCP Terraform workspace access, or other backend permissions as applicable.
- Include example usage blocks showing real, valid HCL that can be copied and adapted. Do not use pseudo-code or incomplete examples.
- When modules are sourced from a registry, document the full source path and version constraint (e.g. `terraform-aws-modules/vpc/aws` for the public registry, `app.terraform.io/{org}/{module}/{provider}` for a private registry).

## Evidence Rules
- All documentation must reflect the actual infrastructure code as it currently exists.
- Do not document aspirational architecture that is not implemented.
- When code and documentation diverge, the code is the source of truth. Update the documentation.
- Verify that documented variable names, types, defaults, and descriptions match the actual variable blocks.
- Verify that documented outputs match the actual output blocks.
- Verify that documented resources match the actual resources in the module.
- When documenting provider versions, check the actual `required_providers` block.
- When documenting Terraform version requirements, check the actual `required_version` constraint.
- When documenting backend/state configuration, verify state-unit names and cross-state references against the actual backend blocks and `terraform_remote_state` data sources. (For HCP Terraform workspace/variable-set verification, see the `hcp-terraform` skill.)

## Anti-Patterns To Reject
- **Stale READMEs**: Documentation that describes a previous version of the module and has not been updated to reflect current resources, variables, or outputs.
- **Undescribed Variables**: Variable blocks without a `description` argument or with descriptions that merely restate the variable name.
- **Undescribed Outputs**: Output blocks without a `description` argument.
- **Documentation-Code Mismatch**: Docs that reference provider versions, resource names, or variable names that do not exist in the current code.
- **Aspirational Architecture Docs**: Documentation that describes intended future state as if it were current implementation.
- **Missing Usage Examples**: Module READMEs without a concrete HCL usage example.
- **Placeholder Descriptions**: Descriptions like "TODO", "TBD", or empty strings.
- **Copy-Pasted Provider Docs**: Restating upstream provider documentation instead of explaining how the module uses the resource.
- **Missing Authentication Docs**: Assuming operators know how to authenticate without documenting the specific requirements.
- **Missing Identity and RBAC Documentation**: Modules that create managed identities, service principals, or role assignments without documenting who gets what access and to which scope.
- **Missing State Context**: Root module documentation that does not mention which state unit manages it, how variables are supplied, or how cross-state dependencies work.
- **Over-reliance on terraform-docs**: Using terraform-docs as the sole documentation source when the module requires resource tables, identity overviews, or RBAC mappings that terraform-docs cannot generate.

## Quality Checklist
Before considering a Terraform documentation task complete, verify:
- All example usage blocks contain valid HCL that would pass `terraform validate` if provider and backend were configured.
- All links in documentation resolve to valid targets.
- Variable descriptions are meaningful — they explain what the variable controls and its implications, not just its name.
- Output descriptions explain what the value represents and who consumes it.
- The resources listed in the README match the actual resources in the module.
- Provider and Terraform version requirements documented match the actual constraints.
- Runbook procedures include failure handling, not just the happy path.
- Runbook procedures reflect the actual workflow (CLI saved-plan apply, runner approval, backend state operations); TFC-UI specifics belong in the `hcp-terraform` skill when TFC is the backend.
- ADRs include consequences, not just the decision.
- Changelog entries distinguish breaking changes from non-breaking changes.
- No sensitive values (subscription IDs, secrets, internal hostnames) appear in example blocks unless they are clearly fake.
- AGENTS.md files accurately reflect the repository's actual conventions and tooling.
- State-backend documentation covers naming, cross-state references, and variable configuration.
- If terraform-docs is in use, its output is consistent with the hand-maintained sections. If terraform-docs is not in use, hand-maintained tables are current.

## Invocation Template
Use this skill with a prompt that supplies repository-specific context. Example:

```text
Use Terraform Documentation Guidance.
Generate a README for the network module at /path/to/modules/network.
Include usage example, resource table, inputs table, outputs table, identity overview, and required permissions.
Document the backend/state configuration and cross-state dependencies.
```
