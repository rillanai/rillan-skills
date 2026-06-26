<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.1.0 -->
# HCP Terraform Development Guidance

## Purpose
Use this overlay when writing or reviewing Terraform on HCP Terraform / Terraform Cloud (TFC). It adds the TFC-specific backend, state, module-sourcing, secrets, policy, and run-workflow depth on top of `terraform/dev.md`. Load `terraform/dev.md` first as the baseline; everything here assumes its general rules (DRY/KISS/YAGNI, typing/validation, lifecycle, `for_each`, `moved`/`import`) already apply.

When this overlay and the base skill agree, the base wins by default; this file only overrides where TFC changes the mechanics (the backend, cross-state references, module registry, secrets injection, runner workflow).

## When This Overlay Applies
- The `terraform` block uses a `cloud {}` block (or the project is configured against `app.terraform.io` / a TFE instance).
- Modules are sourced from a TFC private registry (`app.terraform.io/{org}/{module}/{provider}`).
- Cross-state references use `data "tfe_outputs"`.
- Policy is enforced with Sentinel.
- Secrets and provider auth come from TFC workspace variables or variable sets.

If none of these hold, stay in the base `terraform` skill — it is backend-neutral.

## The `cloud {}` Backend
- Use the `cloud {}` block as the execution backend. TFC stores state, runs plan/apply remotely, locks automatically, and encrypts state at rest. Configure it in a dedicated `backend.tf`:

```hcl
terraform {
  cloud {
    organization = "my-org"

    workspaces {
      name = "prod-shared"
    }
  }
}
```

- Each directory under `environments/` maps to exactly one TFC workspace. The workspace name should reflect the directory structure and purpose (e.g., `prod-shared`, `prod-cluster-a`).
- TFC handles state locking automatically; never disable it.
- If the workspace must vary per deployment, select workspaces by tag instead of hardcoding a single name:

```hcl
terraform {
  cloud {
    organization = "my-org"

    workspaces {
      tags = ["platform", "prod"]
    }
  }
}
```

## Workspace Organization (Directory + Workspace Hybrid)
- Use the directory + workspace hybrid strategy. Each directory under `environments/` is a root module with its own TFC workspace. This gives clear separation by environment and cluster while leveraging TFC for state, locking, and access control. (This is the TFC realization of the base skill's "one state/isolation unit per deployable unit" rule — it is one valid layout, not a universal mandate.)
- **Directory-based per environment and per cluster**: each environment (dev, staging, prod) has its own directory tree; within each, separate directories exist for environment-level resources and for each cluster. Each directory is a root module with its own workspace.
- **Workspace naming convention**: workspace names match the directory purpose. For example, `environments/prod/shared/` → workspace `prod-shared`, `environments/prod/cluster-a/` → workspace `prod-cluster-a`. Keep the convention consistent across the organization.
- One workspace per deployable unit. Avoid mega-workspaces (high blast radius) and over-fragmentation (excessive cross-workspace coupling).

## Cross-Workspace References (`data "tfe_outputs"`)
- On TFC, prefer `data "tfe_outputs"` over `terraform_remote_state` for cross-workspace references. It is the TFC-native mechanism and respects workspace-level access controls, so a consumer needs read access to the producing workspace's outputs rather than to raw state:

```hcl
data "tfe_outputs" "networking" {
  organization = "my-org"
  workspace    = "prod-shared"
}

resource "aws_instance" "app" {
  subnet_id = data.tfe_outputs.networking.values.app_subnet_id
  ami       = var.ami_id
  # ...
}
```

- The consuming reference reads `data.tfe_outputs.<name>.values.<output>`. Document every output consumed this way (the base `docs.md` rule still applies; the consumer has no other way to learn what the value means).
- Use `terraform_remote_state` only when referencing state in non-TFC backends.

## Module Sourcing — TFC Private Registry
- Publish shared modules to the TFC private registry and source them by registry address with a pinned version:

```hcl
module "networking" {
  source  = "app.terraform.io/my-org/networking/aws"
  version = "~> 2.0"

  # Module inputs...
}
```

- The address format is `app.terraform.io/{org}/{module}/{provider}`. Tag releases in the module repository; TFC picks up tagged versions automatically.
- Pin module versions in root modules. Use semantic version constraints (`~>`, exact pins for critical infrastructure).
- The private registry is one option among several (the base skill also covers public registry, git, and local sources). Do not treat git/https/local sources as anti-patterns — choose the source the project standardizes on.

## Secrets, Variables, And Provider Auth
- Inject provider identity and secrets via TFC workspace variables and variable sets, not committed `.tfvars`:
  - **Terraform variables**: non-secret inputs and provider identity (account/subscription/project IDs, region) as workspace variables; mark secrets `sensitive`.
  - **Environment variables**: provider credentials (e.g., `AWS_ACCESS_KEY_ID`, `ARM_CLIENT_ID`, dynamic provider credentials / OIDC) as environment-type workspace variables.
  - **Variable sets**: share common values (provider credentials, standard tags, org-wide settings) across workspaces via variable sets. Be aware that workspace-specific variables override variable-set values — document the intended precedence.
- Prefer TFC's dynamic provider credentials (workload identity / OIDC) over static keys stored as variables.
- Bind provider identity at the root module (the base skill's rule); on TFC the values arrive as workspace variables:

```hcl
# root module providers.tf — values set via TFC workspace variables
provider "aws" {
  region = var.region
  assume_role {
    role_arn = var.deploy_role_arn
  }
}
```

## Policy-As-Code — Sentinel
- Sentinel is HCP Terraform–only. Use it for org/workspace policy enforcement on runs (required tags, approved regions/sizes, encryption-required, public-access-blocked, cost limits via the cost-estimation data):

```hcl
# Sentinel policy: deny public S3 buckets (illustrative)
import "tfplan/v2" as tfplan

deny_public_buckets = rule {
  all tfplan.resource_changes as _, rc {
    rc.type is not "aws_s3_bucket_acl" or
      rc.change.after.acl is not "public-read"
  }
}

main = rule { deny_public_buckets }
```

- Attach policy sets to the organization or specific workspaces; set enforcement level (`advisory`, `soft-mandatory`, `hard-mandatory`).
- Sentinel complements, not replaces, the runner-agnostic policy tools in the base skill (OPA/conftest, Checkov). Use those too where they add value — they run in any pipeline and in local pre-commit.

## TFC Run Workflow
- **Speculative plans**: TFC runs a speculative (non-applyable) plan automatically on PRs against connected VCS branches. Treat it as the PR gate.
- **Confirm & Apply**: never `-auto-approve` production. Require explicit human confirmation via the TFC UI "Confirm & Apply" step (or an API-driven equivalent with an approval gate). Sentinel `soft-mandatory`/`hard-mandatory` policies and manual approval are the approval mechanisms.
- **Run triggers**: configure run triggers so that an apply in an upstream workspace (e.g., networking) queues a run in dependent workspaces. Document the trigger graph; avoid cycles.
- **Scheduled runs**: enable scheduled runs for drift detection — TFC plans on a cadence and surfaces unexpected changes. Alert on non-empty scheduled plans.
- **CLI-driven vs VCS-driven**: prefer VCS-driven workspaces for auditability. Use CLI-driven runs (`terraform plan`/`apply` against the `cloud {}` backend) for bootstrap and exceptional operations.

## Anti-Patterns To Reject (TFC-Specific)
- Hardcoding secrets or provider identity in `.tf`/`.tfvars` instead of TFC workspace variables / variable sets.
- Using `terraform_remote_state` for TFC-to-TFC references when `data "tfe_outputs"` (with workspace access controls) is available.
- `-auto-approve` on production TFC workspaces instead of Confirm & Apply.
- One mega-workspace with enormous blast radius, or so many tiny workspaces that cross-workspace coupling explodes.
- Unpinned private-registry module sources.
- Treating Sentinel as the only policy layer (it does not run in local pre-commit or non-TFC CI) — pair with OPA/Checkov.
- Inconsistent workspace naming that breaks the directory→workspace mapping.

## Review Standard (Additions)
On top of the base `terraform/dev.md` review order, also check:
- State boundaries map cleanly to TFC workspaces; blast radius per workspace is bounded.
- Cross-workspace references use `data "tfe_outputs"` with documented producing workspace/output.
- Modules sourced from the private registry are version-pinned.
- Secrets and provider auth come from workspace variables / variable sets / dynamic credentials, never code.
- Sentinel enforcement levels are appropriate; production applies require confirmation.

## Invocation Template
Use this overlay with the base `terraform` skill and a prompt that supplies project context. Example:

```text
Use the terraform skill plus the hcp-terraform overlay.
Add a new compute workspace under environments/prod/cluster-a backed by the cloud {} block, org "my-org".
Source the networking module from the TFC private registry (app.terraform.io/my-org/networking/aws), pinned.
Read the shared VPC via data "tfe_outputs" from workspace prod-shared.
Provider auth comes from a variable set with dynamic AWS credentials; no static keys.
```
