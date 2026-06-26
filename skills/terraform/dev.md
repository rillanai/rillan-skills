<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.1.0 -->
# Terraform Development Guidance

## Purpose
Use this skill when writing, modifying, or reviewing Terraform code in repositories that value infrastructure reliability, security, modularity, operational clarity, and reviewability over cleverness.

This skill defines the default implementation and review contract for Terraform work. It is intended for module development, resource additions, refactors, environment promotion, state management changes, and infrastructure code review.

This guidance is backend- and cloud-neutral. Treat **OpenTofu as a first-class peer to Terraform** throughout: the `tofu` CLI mirrors `terraform`, the HCL is the same, and version/feature notes apply to both unless stated otherwise. When the project uses HCP Terraform / Terraform Cloud, load the `hcp-terraform` skill in addition to this one for the TFC-specific depth (`cloud {}` backend, `data "tfe_outputs"`, workspaces, private registry, Sentinel, run workflow).

## Skill Use
- Load this skill when the task is to write, modify, or review Terraform code.
- Treat this skill as the governing contract for the session or turn unless the repository has stricter local conventions.
- Keep repository-specific requirements in the task prompt.
- Match established project conventions when they are clear and defensible.
- When this skill conflicts with casual convenience, follow this skill.

## Tool Use
This skill is tool-agnostic and works with Claude Code, Codex, OpenCode, and similar assistants. Map its guidance to whatever file-reading, editing, search, and shell-execution tools your environment exposes.

- Invoke tools to read HCL, search for resources, and run `terraform` commands; do not describe what you would do.
- Issue independent tool calls in parallel rather than sequentially.
- Run `terraform fmt`, `terraform validate`, and `terraform plan` yourself — do not claim a change is verified without tool output.
- Before proposing a change to a shared module, inspect its callers to scope the blast radius.

## Knowledge-Graph Discovery (When Available)
If the repository carries a graphify knowledge graph (a `graphify-out/` directory), use it as a map to consult before broad text search — never as ground truth.
- Orient first from `graphify-out/GRAPH_REPORT.md` (or `graphify-out/wiki/index.md` when present): god nodes, communities, and cross-file relationships show module composition and resource dependencies before you open a file.
- For "what consumes this module/output", "what depends on this resource", and blast-radius questions, prefer `graphify query "<question>"`, `graphify path "<A>" "<B>"`, and `graphify explain "<name>"` over grep — they traverse extracted and inferred edges across module and resource boundaries that text search misses.
- Every edge is tagged `EXTRACTED`, `INFERRED`, or `AMBIGUOUS`. Treat `EXTRACTED` as structural evidence; treat `INFERRED` and `AMBIGUOUS` as leads to confirm with `terraform validate` and `terraform plan`. The graph never outranks `terraform plan` and validation output.
- After changing configuration, run `graphify update .` (AST-only, no API cost) to keep the graph current.

If no `graphify-out/` directory exists, ignore this section.

## Core Principles

### DRY
- Extract repeated resource patterns into reusable modules. If the same resource block with minor variations appears twice, refactor into a module with variables.
- Use shared `locals` for computed values, naming conventions, and tag maps that appear across resources.
- Prefer module composition over copy-paste of resource blocks across environments.
- A small amount of duplication is acceptable when the alternative is a brittle abstraction. Prefer clear duplication over a module that hides important resource differences behind obscure variable combinations.
- If removing duplication makes the infrastructure harder to reason about during an incident, prefer the more obvious code.

### KISS
- Choose the simplest resource configuration that meets the actual requirement. Avoid clever dynamic block gymnastics when a straightforward resource block is clear enough.
- If a `locals` expression needs a comment explaining how it works instead of why it exists, simplify it.
- Prefer straightforward conditional patterns over deeply nested ternaries or chained `try()` calls when the cases are few and clear.
- One module should manage one logical infrastructure concern. Split modules that combine unrelated resource groups.

### YAGNI
- Do not add variables, conditional resource creation, or module abstractions until a second consumer or environment exists or is imminent.
- Do not write speculative generic modules that try to handle every possible cloud configuration.
- Do not add feature toggles or provider-switching patterns for a single deployment target.
- A variable justified by environment differentiation is acceptable when there is a real second environment that needs it.

## Design Priorities
- Favor immutable infrastructure over in-place mutation. Prefer replacing resources over modifying them when the resource supports it and the blast radius is acceptable.
- Declarative over imperative. Let Terraform manage resource lifecycle. Avoid `local-exec` and `remote-exec` provisioners unless there is genuinely no resource or data source alternative.
- Minimize blast radius. Isolate state per environment and per logical boundary. A single `terraform apply` should not be able to destroy production and staging simultaneously.
- Make the plan readable. Structure code so that `terraform plan` output is understandable to a reviewer who did not write the change.
- Prefer designs that are simple to review, apply, and roll back.
- Keep provider concerns, module logic, environment configuration, and backend setup in separate layers.
- Choose clarity and operational safety over abstraction density.

## Project Structure

```text
modules/          # Reusable child modules (local path, git, or a module registry)
  networking/
    main.tf                 # Core resources
    variables.tf            # Input variables with types, descriptions, and validations
    outputs.tf              # Output values
    versions.tf             # Required providers and Terraform version constraints
    README.md               # Module purpose, usage example, and input/output reference
  compute/
  database/
  identity/
environments/               # Root modules per environment and deployable unit; each maps to one state/isolation unit
  prod/
    shared/                 # shared environment resources
      main.tf               # Module calls and environment-specific resources
      variables.tf          # Environment-specific variable definitions
      outputs.tf            # Environment-specific outputs
      terraform.tfvars      # Variable values for this environment
      backend.tf            # Backend configuration for this state/isolation unit
      versions.tf           # Provider and Terraform version pins
      providers.tf          # Provider configuration and aliases
    cluster-a/
      main.tf
      variables.tf
      outputs.tf
      terraform.tfvars
      backend.tf
      versions.tf
      providers.tf
  nonprod/
    dev/
      shared/
      cluster-b/
    staging/
      shared/
.terraform/                 # Local provider cache and module cache (gitignored)
.terraform.lock.hcl         # Dependency lock file (committed)
```

- Keep root modules thin: compose child modules, set environment-specific variables, configure backends and providers.
- Put reusable infrastructure logic in `modules/`. Source it locally during development; publish to a module registry (public registry, a private registry such as HCP Terraform / Spacelift / Scalr, or git) when sharing across repos or teams.
- Group by infrastructure concern rather than by resource type when the project grows beyond a few modules.
- Separate environment configuration from module logic. Root modules in `environments/` compose modules; they do not own complex resource definitions directly.
- Each directory under `environments/` maps to exactly one state/isolation unit — a TFC workspace, an S3 key + lock table, a GCS prefix, a `consul` path, etc. Name it to match the directory purpose (e.g., `prod-shared`, `cluster-a`).
- Keep `terraform.tfvars` environment-specific. Never share a single tfvars file across environments.
- Commit `.terraform.lock.hcl` to version control. Add `.terraform/` to `.gitignore`.
- Keep `backend.tf` separate from other configuration for clarity and to simplify backend migration.

## Module Design
- Define clear inputs (`variables.tf`), outputs (`outputs.tf`), and locals (`locals` blocks or `locals.tf`) for every module.
- Keep the module surface area minimal. Expose only the variables and outputs that consumers actually need. Do not expose every possible configuration knob.
- Source shared modules from wherever the project publishes them. Common options, each with a versioned `source`:

```hcl
# Public registry
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 5.0"
}

# Private registry (HCP Terraform, Spacelift, Scalr, …)
module "networking" {
  source  = "app.terraform.io/my-org/networking/aws"
  version = "~> 2.0"
}

# Git
module "networking_git" {
  source = "git::https://github.com/my-org/terraform-modules.git//networking?ref=v2.0.0"
}

# Local path
module "networking_local" {
  source = "../../modules/networking"
}
```

- Use semantic versioning for shared modules consumed across repositories or teams. Pin module source versions in root modules. (Local-path sources are not versioned; pin git sources with `?ref=`.)
- Prefer composition over inheritance. Build complex infrastructure by composing small, focused modules rather than creating deeply nested module hierarchies.
- Keep module nesting shallow — roughly two levels (a root module calling a child calling a grandchild) before debugging and state inspection get painful. This is a soft guideline, configurable per project, not a hard universal cap.
- Every module must have a `versions.tf` specifying `required_providers` and the minimum `required_version` of Terraform.
- Provide sensible defaults for variables where a reasonable default exists. Require explicit input where no safe default is possible.
- Write a `README.md` for every shared module with purpose, usage example, inputs table, and outputs table. Maintain these by hand; do not rely solely on auto-generated documentation.
- Each module should focus on one infrastructure concern. Do not combine unrelated resource groups (e.g., networking and identity) into a single module.
- Test modules in isolation before composing them into root modules.

## Resource Naming
- Use consistent, predictable naming conventions for all resources. Derive names from a combination of prefix, environment, and location using `locals`.
- Prefer computing names in `locals` blocks rather than constructing them inline in resource arguments.
- Use the standard naming pattern:

```hcl
locals {
  name_prefix = "${var.name_prefix}-${var.environment}-${var.region}"
}

# AWS example
resource "aws_s3_bucket" "data" {
  bucket = "${local.name_prefix}-data"
}

# Azure example
resource "azurerm_resource_group" "main" {
  name     = "${local.name_prefix}-rg"
  location = var.location
}

# GCP example
resource "google_storage_bucket" "data" {
  name     = "${local.name_prefix}-data"
  location = var.location
}
```

- Use snake_case for Terraform resource logical names and variable names.
- Use kebab-case for actual cloud resource names.
- Apply a consistent tag (or label) map to every resource that supports tags. Keep a shared `locals` block or a dedicated `tags.tf` for the default map. The **specific tag keys are project policy** — pass the required key set (e.g. application, team, contact, environment) in via the invoking prompt rather than hardcoding one organization's governance schema here:

```hcl
variable "standard_tags" {
  type        = map(string)
  description = "Standard tags applied to all resources. Keys are defined by project policy."
  default     = {}
}

locals {
  # Merge the project's standard tags with a few stable, computed defaults.
  # Prefer stable values; avoid timestamp()-based tags that drift on every apply.
  common_tags = merge(var.standard_tags, {
    Environment = var.environment
    ManagedBy   = "terraform"
  })
}

# AWS example
resource "aws_s3_bucket" "data" {
  bucket = "${local.name_prefix}-data"
  tags   = local.common_tags
}

# Azure example
resource "azurerm_resource_group" "main" {
  name     = "${local.name_prefix}-rg"
  location = var.location
  tags     = local.common_tags
}

# GCP example — use labels (GCP equivalent of tags)
resource "google_storage_bucket" "data" {
  name     = "${local.name_prefix}-data"
  location = var.location
  labels   = { for k, v in local.common_tags : lower(replace(k, "/[^a-z0-9_-]/", "_")) => lower(v) }
}
```

- Never hardcode names that should vary by environment.

## State Management
- Use a remote backend with state locking for any shared or production state. Choose the backend the project standardizes on and let the project decide — common choices:

```hcl
# S3 (with native state locking, Terraform 1.10+ / use a DynamoDB table on older versions)
terraform {
  backend "s3" {
    bucket       = "my-tf-state"
    key          = "prod/shared/terraform.tfstate"
    region       = "us-east-1"
    use_lockfile = true
  }
}

# Other common backends: gcs, azurerm, consul, kubernetes, http, and cloud (HCP Terraform).
# Local state is acceptable only for throwaway or single-operator experiments.
```

- For HCP Terraform / Terraform Cloud (the `cloud {}` block and TFC workspaces), see the `hcp-terraform` skill — it is the overlay that owns that backend's specifics.
- Each directory maps to exactly one state/isolation unit. Name it to reflect the directory structure and purpose.
- Enable state locking. Most remote backends support locking; never disable it where it is available.
- Isolate state per environment and per cluster. Each environment and cluster directory has its own state. Never share state between dev, staging, and production.
- Consider isolating state per logical boundary within an environment when the blast radius of a single state file becomes too large (for example, separating networking state from compute state).
- Use `terraform_remote_state` (or your backend's equivalent) for cross-state references. Publish the values a consumer needs as outputs in the producer, then read them downstream:

```hcl
data "terraform_remote_state" "networking" {
  backend = "s3"
  config = {
    bucket = "my-tf-state"
    key    = "prod/shared/terraform.tfstate"
    region = "us-east-1"
  }
}

resource "aws_subnet" "app" {
  vpc_id     = data.terraform_remote_state.networking.outputs.vpc_id
  cidr_block = "10.0.2.0/24"
}
```

- `terraform_remote_state` is backend-agnostic and is the default cross-state mechanism. (HCP Terraform offers `data "tfe_outputs"` as a TFC-native alternative with workspace-level access controls — see the `hcp-terraform` skill.)
- Treat state files as sensitive. They contain resource metadata, attribute values, and potentially sensitive outputs. Use a backend that encrypts state at rest and restrict access via the backend's access controls (bucket/IAM policies, workspace team permissions, etc.).
- Never commit state files to version control. Add `*.tfstate` and `*.tfstate.backup` to `.gitignore`.
- Use `terraform state` commands carefully. State manipulation (`mv`, `rm`, `import`) should be reviewed and documented like any infrastructure change.
- Use `moved` blocks for resource address refactoring instead of manual `terraform state mv` commands:

```hcl
moved {
  from = aws_security_group.old_name
  to   = aws_security_group.new_name
}
```

- Use `import` blocks (Terraform 1.5+) for importing existing resources declaratively:

```hcl
import {
  to = aws_s3_bucket.main
  id = "my-existing-bucket"
}
```

- Back up state files before migrations or major refactors.

## Variables And Outputs
- Type every variable. Use `string`, `number`, `bool`, `list()`, `map()`, `set()`, `object()`, or `tuple()` as appropriate.
- Use `optional()` for object attributes that have sensible defaults:

```hcl
variable "cluster_config" {
  type = object({
    name         = string
    node_count   = number
    machine_type = optional(string, "Standard_D4s_v3")
    auto_scale   = optional(bool, true)
  })
  description = "Cluster configuration with optional tuning parameters."
}
```

- Add a `description` to every variable and every output. Descriptions appear in documentation and `terraform plan` output; they are not optional.
- Use `validation` blocks for variables that have constraints beyond type. Use regex, uniqueness checks, and non-empty validations:

```hcl
variable "environment" {
  type        = string
  description = "Deployment environment name."

  validation {
    condition     = contains(["dev", "staging", "production"], var.environment)
    error_message = "Environment must be one of: dev, staging, production."
  }
}

variable "subnet_cidrs" {
  type        = list(string)
  description = "List of subnet CIDR ranges. Must be non-empty and contain valid CIDR notation."

  validation {
    condition     = length(var.subnet_cidrs) > 0
    error_message = "At least one subnet CIDR must be provided."
  }

  validation {
    condition     = alltrue([for cidr in var.subnet_cidrs : can(cidrhost(cidr, 0))])
    error_message = "All entries must be valid CIDR notation."
  }
}

variable "resource_names" {
  type        = list(string)
  description = "List of resource names. Must be unique."

  validation {
    condition     = length(var.resource_names) == length(toset(var.resource_names))
    error_message = "Resource names must be unique."
  }
}
```

- Mark variables and outputs that contain secrets as `sensitive = true`. This prevents Terraform from displaying their values in plan and apply output.
- Provide `default` values only when a safe, reasonable default exists. Require explicit input for values that must vary per deployment.
- **Use snake_case for all output names.** Never use kebab-case — it forces consumers to use bracket syntax (`module.foo["my-output"]`) instead of dot syntax (`module.foo.my_output`). Be consistent across all modules.
- Every output must have a `description`. This is especially critical for outputs consumed across state boundaries (via `terraform_remote_state` or your backend's equivalent), since the consumer has no other way to understand what the value represents.
- Organize outputs logically. Group them by resource or concern.
- Do not use `nullable = true` without a clear reason. Prefer required variables with validation over nullable variables with fallback logic.
- Keep `terraform.tfvars` files per environment. Do not use `terraform.tfvars` for shared defaults; use variable `default` values for that.

## Provider Configuration
- Pin provider versions to a specific minor version range in `versions.tf`. Declare only the providers a root module actually uses. The cloud roster and which provider is "primary" is project context — the example below shows several together, but most root modules use one:

```hcl
terraform {
  required_version = ">= 1.5.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
    }
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
    vsphere = {
      source  = "hashicorp/vsphere"
      version = "~> 2.0"
    }
  }
}
```

- Always specify the `source` in `required_providers`, even for HashiCorp providers.
- Only include providers that the root module actually uses. Not every root module needs every provider.
- Use provider aliases when managing resources across multiple accounts, subscriptions, projects, or regions:

```hcl
# AWS: multiple accounts/regions
provider "aws" {
  region = var.primary_region
}

provider "aws" {
  alias  = "network"
  region = var.primary_region
  assume_role {
    role_arn = var.network_account_role_arn
  }
}

# Azure: multiple subscriptions
provider "azurerm" {
  features {}
  subscription_id = var.primary_subscription_id
}

provider "azurerm" {
  alias           = "connectivity"
  features {}
  subscription_id = var.connectivity_subscription_id
}
```

- Keep provider configuration in a dedicated `providers.tf` file in root modules.
- **Bind provider identity at the root module, never in child modules.** Account IDs, subscription IDs, project IDs, tenant IDs, and similar identifiers must be variables in the root module's `providers.tf`, injected via backend/runner variables (env vars, `terraform.tfvars`, CI secrets, TFC workspace variables, Vault). Child modules must never hardcode or accept provider identity parameters — they inherit the provider from the calling root module. Even when a module is tightly bound to a specific organization or tenant, the root module is where that binding lives:

```hcl
# root module providers.tf — provider identity bound here
provider "aws" {
  region = var.region
  assume_role {
    role_arn = var.deploy_role_arn  # Set via runner/backend variable
  }
}

# child module — NO provider config, inherits from root
module "network" {
  source  = "../../modules/network"
  # ...
}
```

- Do not hardcode credentials, account IDs, subscription IDs, project IDs, or tenant IDs in provider blocks or anywhere in `.tf` files. Use backend/runner variables, workload identity federation, OIDC role assumption, managed identities, environment variables, or credential helpers. Never use static access keys, service account keys, or client secrets in code.
- Commit `.terraform.lock.hcl` to ensure reproducible builds across machines and CI.
- Review provider changelogs before major version upgrades. Provider upgrades can introduce breaking changes to resource behavior.

## Data Sources
- Use data sources to reference existing infrastructure that Terraform does not manage, such as shared virtual networks, VM images, DNS zones, or subscription/project metadata.
- Do not use data sources as a substitute for proper variable passing when Terraform manages both the producer and consumer resources.
- Be aware of staleness. Data source values are read at plan time. If the referenced resource changes between plan and apply, the plan may be stale.
- Prefer `data` sources over hardcoded resource IDs. Hardcoded IDs break across environments.
- Use `data` sources to look up dynamic values like latest VM images, availability zones, or subscription/project IDs:

```hcl
# Azure: look up latest image
data "azurerm_platform_image" "ubuntu" {
  location  = var.location
  publisher = "Canonical"
  offer     = "0001-com-ubuntu-server-jammy"
  sku       = "22_04-lts"
}

# GCP: look up latest image
data "google_compute_image" "ubuntu" {
  family  = "ubuntu-2204-lts"
  project = "ubuntu-os-cloud"
}
```

- When referencing resources across state boundaries, prefer `terraform_remote_state` (or your backend's equivalent; on HCP Terraform, `data "tfe_outputs"`). Fall back to provider-specific data sources when querying resources not managed by Terraform.

## Lifecycle Rules
- Use `create_before_destroy` when replacing a resource that must remain available during replacement, such as security groups referenced by active instances or DNS records.
- Use `prevent_destroy` on critical resources that should never be accidentally deleted, such as production databases, storage accounts with important data, or encryption keys:

```hcl
resource "azurerm_key_vault" "main" {
  # ...
  lifecycle {
    prevent_destroy = true
  }
}
```

- Use `ignore_changes` only when an external process legitimately modifies a resource attribute that Terraform should not revert. **Always list specific attributes, never blanket-ignore entire blocks.** Document the reason in a comment:

```hcl
resource "azurerm_kubernetes_cluster" "main" {
  # ...
  lifecycle {
    # Node count managed by cluster autoscaler, not Terraform.
    ignore_changes = [default_node_pool[0].node_count]
  }
}
```

- **Be surgical with tag ignores.** If you use dynamic timestamps (`timestamp()`, `timeadd()`) in tags, ignore only the specific computed tag attributes, not all tags. Blanket `ignore_changes = [tags]` masks legitimate tag corrections and forces tainting to fix them. Prefer computing stable tag values (e.g., deploy date set once and not updated) over `timestamp()` which changes on every apply:

```hcl
# Preferred: stable tags that don't drift
locals {
  common_tags = merge(var.standard_tags, {
    DeployedBy = "terraform"
  })
}

# If dynamic timestamps are unavoidable, ignore only the specific keys
# and document which external process sets them
```

- **Never use `ignore_changes` on core resource attributes** like `name`, `address_space`, or `resource_group_name`. If Terraform wants to change these, the configuration does not match deployed state. Fix the configuration or import the resource correctly — do not paper over the mismatch with `ignore_changes`. Temporary import fixups should be removed once state is aligned.
- Use `replace_triggered_by` when a resource must be recreated in response to changes in a related resource that Terraform would not otherwise detect as a dependency.
- Do not overuse `ignore_changes` to suppress legitimate configuration drift. If Terraform keeps wanting to change an attribute, understand why before ignoring it.

## Security
- Use RBAC exclusively. Disable local accounts on all services that support it (e.g., AKS clusters, databases). Never create local admin accounts when RBAC is available.
- Use workload identity federation for service-to-cloud authentication. Avoid static credentials, client secrets, and service account keys.
- Use per-service user-assigned managed identities (Azure) or dedicated service accounts (GCP) rather than shared credentials:

```hcl
# Azure: user-assigned identity per service
resource "azurerm_user_assigned_identity" "app" {
  name                = "${local.name_prefix}-app-identity"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  tags                = local.common_tags

  lifecycle {
    ignore_changes = [tags]
  }
}

# GCP: dedicated service account per service
resource "google_service_account" "app" {
  account_id   = "${var.name_prefix}-app"
  display_name = "App service account for ${var.environment}"
  project      = var.project_id
}
```

- Use Key Vault (Azure) or Secret Manager (GCP) for encryption keys and secrets. Never store secrets in Terraform code or tfvars files.
- Deploy private clusters. Kubernetes API servers and databases should not be publicly accessible.
- Apply NSGs (Azure) or firewall rules (GCP) per subnet. Do not allow `0.0.0.0/0` ingress on sensitive ports. Prefer specific CIDR ranges or service tags.
- Do not create public-facing resources without explicit intent. Review any `0.0.0.0/0` CIDR, `public = true`, or `publicly_accessible = true` carefully.
- Follow least privilege for all IAM roles and policies. Never use wildcard permissions in production unless there is a documented, reviewed justification.
- Mark sensitive variables and outputs with `sensitive = true`.
- Restrict access to remote state and the execution backend via its access controls (bucket/IAM policies, workspace team permissions, etc.). Separate plan and apply permissions where the backend or runner supports it.
- Audit provider permissions. The credentials Terraform uses should have only the permissions required for the resources it manages.
- Never store `.tfvars` files containing secrets in version control. Use backend/runner variables (env vars, CI secrets, TFC workspace vars), or secret manager references (Vault, cloud secret managers) instead.
- Validate that encryption is enabled on storage resources (Azure Storage, Cloud SQL, managed disks, Key Vault, etc.) by default.

## Conditional And Dynamic Patterns
- Use `count` only for simple conditional resource creation (create or do not create, 0 or 1):

```hcl
resource "azurerm_log_analytics_workspace" "main" {
  count               = var.enable_logging ? 1 : 0
  name                = "${local.name_prefix}-logs"
  location            = var.location
  resource_group_name = azurerm_resource_group.main.name
}
```

- Use `for_each` with maps or objects for creating multiple instances of a resource, especially when each instance has a meaningful identity:

```hcl
variable "subnets" {
  type = map(object({
    address_prefix = string
    service_endpoints = optional(list(string), [])
  }))
  description = "Map of subnet configurations keyed by subnet name."
}

resource "azurerm_subnet" "this" {
  for_each             = var.subnets
  name                 = "${local.name_prefix}-${each.key}"
  resource_group_name  = azurerm_resource_group.main.name
  virtual_network_name = azurerm_virtual_network.main.name
  address_prefixes     = [each.value.address_prefix]
  service_endpoints    = each.value.service_endpoints
}
```

- Prefer `for_each` over `count` when the collection may change, because `count` uses index-based addressing and causes cascading replacements when items are added or removed from the middle.
- Use `setproduct()` in locals for Cartesian product patterns when you need to create resources for every combination of two sets:

```hcl
locals {
  # Create a subnet in every combination of environment and zone
  subnet_combinations = { for pair in setproduct(var.environments, var.zones) :
    "${pair[0]}-${pair[1]}" => {
      environment = pair[0]
      zone        = pair[1]
    }
  }
}
```

- Use `for` expressions to transform lists to maps when `for_each` requires a map:

```hcl
locals {
  users_map = { for user in var.users : user.name => user }
}
```

- Use dynamic blocks to generate repeated nested blocks inside a resource, but do not overuse them. If a dynamic block makes the resource harder to read than writing out the blocks explicitly, prefer explicit blocks:

```hcl
resource "azurerm_network_security_group" "app" {
  name                = "${local.name_prefix}-app-nsg"
  location            = var.location
  resource_group_name = azurerm_resource_group.main.name

  dynamic "security_rule" {
    for_each = var.security_rules
    content {
      name                       = security_rule.value.name
      priority                   = security_rule.value.priority
      direction                  = security_rule.value.direction
      access                     = security_rule.value.access
      protocol                   = security_rule.value.protocol
      source_port_range          = security_rule.value.source_port_range
      destination_port_range     = security_rule.value.destination_port_range
      source_address_prefix      = security_rule.value.source_address_prefix
      destination_address_prefix = security_rule.value.destination_address_prefix
    }
  }
}
```

- Use `try()` to safely access nested attributes that may not exist, especially in complex data structures from data sources or module outputs.
- Use `coalesce()` to select the first non-empty value from a list of candidates. Useful for defaulting patterns.
- Use `one()` to convert a list with zero or one element to a single value or null.
- Prefer `lookup()` over direct map indexing when you need a default value for missing keys.

## Dependencies
- Prefer implicit dependencies through resource attribute references. Terraform automatically determines the correct ordering when one resource references another's attributes.
- Use `depends_on` sparingly. It is a blunt instrument that creates a hard ordering dependency on the entire resource, not just specific attributes. Use it only when implicit dependencies cannot express the real ordering requirement.
- Document why `depends_on` is needed whenever it is used. If you cannot explain the dependency, it probably does not need to be explicit.
- Module-level `depends_on` forces all resources in the target module to be created or destroyed in order. Use it only when the entire module truly depends on another resource or module completing first.
- Avoid circular dependencies. If two resources or modules depend on each other, restructure the module boundaries.

## Backend And State Isolation Strategy
- Use the directory-per-isolation-unit strategy. Each directory under `environments/` is a root module with its own backend configuration and its own state. This provides clear separation by environment and cluster while keeping blast radius bounded.
- **Directory-based per environment and per cluster**: Each environment (dev, staging, prod) has its own directory tree. Within each environment, separate directories exist for the environment-level resources and for each cluster. Each directory owns one state/isolation unit — a TFC workspace, an S3 key + lock table, a GCS prefix, etc.
- **Naming convention**: State unit names (workspace name, S3 key, GCS prefix) should match the directory purpose. For example, `environments/prod/shared/` → `prod-shared`, and `environments/prod/cluster-a/` → `cluster-a`.
- Configure the backend in a dedicated `backend.tf` file. Pick the backend the project uses (`s3`, `gcs`, `azurerm`, `consul`, `kubernetes`, `http`, `cloud`); keep one isolation unit per directory:

```hcl
# Example: S3 backend, one key per deployable unit
terraform {
  backend "s3" {
    bucket       = "my-tf-state"
    key          = "prod/shared/terraform.tfstate"
    region       = "us-east-1"
    use_lockfile = true
  }
}
```

- For HCP Terraform / Terraform Cloud (the `cloud {}` block, workspace tags, and the directory + workspace hybrid pattern), see the `hcp-terraform` skill.
- Cross-state references use `terraform_remote_state` (or your backend's equivalent) to read outputs from other isolation units. This lets cluster state reference environment-level resources (e.g., networking, identity) without tight coupling:

```hcl
data "terraform_remote_state" "environment" {
  backend = "s3"
  config = {
    bucket = "my-tf-state"
    key    = "prod/shared/terraform.tfstate"
    region = "us-east-1"
  }
}

# Use outputs from the environment state
resource "aws_eks_cluster" "main" {
  name     = "${local.name_prefix}-eks"
  role_arn = data.terraform_remote_state.environment.outputs.cluster_role_arn

  vpc_config {
    subnet_ids = data.terraform_remote_state.environment.outputs.private_subnet_ids
  }
}
```

## Terraform 1.x+ Features
- Use `moved` blocks for all resource address refactoring. Never use manual `terraform state mv` for changes that can be expressed declaratively.
- Use `import` blocks (Terraform 1.5+) for importing existing resources into state declaratively.
- Use `check` blocks for continuous validation of infrastructure assumptions:

```hcl
check "cluster_health" {
  data "http" "health" {
    url = "https://${azurerm_kubernetes_cluster.main.fqdn}/healthz"
  }

  assert {
    condition     = data.http.health.status_code == 200
    error_message = "Cluster health check failed."
  }
}
```

- Use `terraform test` for module testing. Write `.tftest.hcl` files to validate module behavior:

```hcl
# tests/basic.tftest.hcl
run "creates_resource_group" {
  command = plan

  assert {
    condition     = azurerm_resource_group.main.name == "test-dev-eastus-rg"
    error_message = "Resource group name does not match expected convention."
  }
}
```

- Do not use pre-1.0 syntax or maintain migration paths for legacy Terraform versions.

## CI/CD Integration
- Run `terraform plan` in CI on every pull request. Your runner (TFC, Atlantis, Spacelift, env0, or a plain CI job) can run speculative plans automatically on PRs.
- Never run `terraform apply -auto-approve` in production pipelines. Require explicit human approval for production applies, via the runner's approval gate (TFC Confirm & Apply, Atlantis `apply` comment, a protected CI stage, etc.).
- Enforce policy-as-code in the pipeline where it adds value: OPA/conftest and Checkov policies work on any runner; Sentinel is HCP Terraform–only (see the `hcp-terraform` skill).
- Save the plan as an artifact and apply the saved plan, not a fresh plan. This prevents drift between the reviewed plan and the applied changes:

```bash
terraform plan -out=tfplan
# After approval:
terraform apply tfplan
```

- Run `terraform fmt -check` and `terraform validate` as early CI checks.
- Add `tflint`, `tfsec`/`trivy`, and `checkov` to the CI pipeline to catch issues early.
- Implement drift detection by scheduling regular `terraform plan` runs and alerting on unexpected changes. Most runners support scheduled runs for this purpose.
- Use separate CI/CD credentials per environment. Production apply credentials should not be available to development pipelines.
- Cache the `.terraform` directory in CI to speed up `terraform init`, but invalidate the cache when `.terraform.lock.hcl` changes.
- Pin the Terraform version in CI. Use `tfenv`, `asdf`, or a container image with a specific version. Never rely on "latest."
- Store plan artifacts securely. Plans can contain sensitive data.

## Anti-Patterns To Reject
- Monolith root modules that manage hundreds of resources in a single state file with enormous blast radius
- Hardcoded provider identity (account IDs, subscription IDs, project IDs, tenant IDs) in `.tf` files — these must be variables injected via backend/runner variables
- Hardcoded values that should be variables, especially resource IDs, region names, or CIDR blocks
- Overuse of provisioners (`local-exec`, `remote-exec`, `file`) for tasks that should be handled by proper resources, data sources, or configuration management tools
- Excessive module nesting (beyond roughly two levels, as a soft guideline), making debugging and state inspection painful
- `terraform apply -auto-approve` in production environments or shared environments
- Using `local-exec` to run scripts that create or modify infrastructure that should be a Terraform resource
- Storing secrets in `terraform.tfvars` or variable defaults committed to version control
- Using `count` with lists where `for_each` with a map or set would prevent index-based cascading changes
- Ignoring `.terraform.lock.hcl` or not committing it, leading to inconsistent provider versions across machines
- Blanket `ignore_changes = [tags]` instead of ignoring only specific computed tag attributes
- Using `ignore_changes` on core resource attributes (`name`, `address_space`, `resource_group_name`) as a permanent fixture instead of fixing configuration-state mismatches
- Overusing `ignore_changes` to suppress legitimate configuration drift instead of fixing the root cause
- Kebab-case output names (`my-output`) instead of snake_case (`my_output`)
- Outputs without `description` fields, especially outputs consumed across state boundaries via `terraform_remote_state`
- Committing plan files or state files to git — these contain sensitive data and belong in CI artifacts or a remote backend
- Storing state files in version control instead of a remote backend
- Tight cross-stack coupling: reaching into another stack's raw state instead of consuming a published output via `terraform_remote_state` (or the backend's equivalent)
- Creating resources with public access (public IPs, `0.0.0.0/0` security group rules) without explicit, documented justification
- Writing modules that expose every possible variable as an input, creating an unmanageable surface area
- Using `terraform taint` (deprecated) instead of `terraform apply -replace`
- Mixing CLI-workspace-based and directory-based environment strategies in the same project without clear justification
- Running `terraform apply` without first reviewing the plan output
- Using the `default` CLI workspace in production when workspaces are the environment strategy
- Using local accounts or static credentials when RBAC and workload identity are available

## Code Quality Checklist
Before considering a Terraform task complete, verify:
- `terraform fmt` has been run and the code is formatted.
- `terraform validate` passes without errors.
- Consider running `tflint` — passes without errors or warnings, or warnings are acknowledged and documented.
- Consider running `tfsec` or `trivy` — passes, or findings are reviewed and accepted risks are documented.
- Consider running `checkov` for compliance-relevant resources — passes, or exceptions are documented.
- All variables have `type`, `description`, and `validation` where appropriate.
- All outputs have `description`.
- All resources that support tags have the project's standard tag map applied. If dynamic tags are used, only specific computed tag attributes are ignored — not all tags.
- All output names use snake_case consistently. No kebab-case output names.
- No plan files (`*.tfplan`, `plan`) or state files are committed to git.
- No secrets are hardcoded in any `.tf` or `.tfvars` file.
- Sensitive variables and outputs are marked `sensitive = true`.
- The backend is configured correctly for this isolation unit (workspace name, state key, prefix, etc.).
- The change has been reviewed via `terraform plan` output before apply.
- Module documentation (`README.md`) is updated if module inputs or outputs changed.
- `.terraform.lock.hcl` is committed if provider versions changed.
- No `TODO` was added without a tracking reference.
- Lifecycle rules are appropriate for the resource type and environment.
- No public exposure was introduced without explicit, documented intent.
- `for_each` is used instead of `count` for multi-instance resources; `count` is only used for 0-or-1 conditional creation.
- `moved` blocks are used for any resource address changes.

## Review Standard
When reviewing Terraform code, evaluate in this order:
1. Correctness — Does the plan produce the intended infrastructure?
2. Security — Are IAM/RBAC, encryption, network, secrets, and workload identity handled properly?
3. Blast radius — How much can this change break? Is state properly isolated across state/isolation units?
4. Modularity — Are modules focused, minimal, composable, and sourced from a versioned registry or pinned source?
5. Readability — Can a reviewer understand the plan output and the code?
6. Naming and tagging — Are tag-map conventions consistent? Is naming computed via locals?
7. State safety — Is state isolated per unit, locked, encrypted, and backed up?

Do not lead with style preferences. Flag security, correctness, and blast radius concerns first; address formatting and naming only after operational concerns.

## What Good Output Looks Like
- Code is straightforward to review by reading `terraform plan` output.
- Module boundaries are obvious and follow logical infrastructure concerns.
- State boundaries match operational and security boundaries, with one state/isolation unit per deployable unit.
- Failure modes are explicit: `prevent_destroy` on critical resources, validation on inputs, clear error messages.
- Security posture is visible: RBAC-only, workload identity, per-service identities, encryption enabled, network rules restrictive by default.
- Operational concerns such as consistent tagging, naming via locals, and state-boundary management are addressed during implementation, not bolted on afterward.
- Cross-state references use `terraform_remote_state` (or the backend's equivalent) with clear naming.
- The code can pass a serious infrastructure audit without needing the reader to infer design intent.

## Invocation Template
Use this skill with a prompt that supplies repository-specific context. Pass project-specific policy (the exact tag-key set, cloud roster, backend choice, module-nesting limit) in the prompt. Example:

```text
Use Terraform Development Guidance.
Add a new EKS cluster module in /path/to/repo/modules/compute.
Keep the module focused on a single managed Kubernetes cluster.
Apply the repository's standard tag map; required keys: ApplicationName, Team, Environment.
Backend is s3 with one state key per deployable unit; source the module from the local path.
Ensure RBAC-only, private cluster, and IRSA/workload identity are enabled.
Add validation blocks on all input variables.
```
