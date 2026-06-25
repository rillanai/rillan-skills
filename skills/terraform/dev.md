<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.0.0 -->
# Terraform Development Guidance

## Purpose
Use this skill when writing, modifying, or reviewing Terraform code in repositories that value infrastructure reliability, security, modularity, operational clarity, and reviewability over cleverness.

This skill defines the default implementation and review contract for Terraform work. It is intended for module development, resource additions, refactors, environment promotion, state management changes, and infrastructure code review.

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
modules/          # Reusable child modules published to TFC private registry
  networking/
    main.tf                 # Core resources
    variables.tf            # Input variables with types, descriptions, and validations
    outputs.tf              # Output values
    versions.tf             # Required providers and Terraform version constraints
    README.md               # Module purpose, usage example, and input/output reference
  compute/
  database/
  identity/
environments/               # Root modules per environment and deployable unit; each maps to a TFC workspace
  prod/
    shared/                 # workspace: prod-shared (shared environment resources)
      main.tf               # Module calls and environment-specific resources
      variables.tf          # Environment-specific variable definitions
      outputs.tf            # Environment-specific outputs
      terraform.tfvars      # Variable values for this environment
      backend.tf            # TFC backend configuration for this workspace
      versions.tf           # Provider and Terraform version pins
      providers.tf          # Provider configuration and aliases
    cluster-a/              # workspace: prod-cluster-a
      main.tf
      variables.tf
      outputs.tf
      terraform.tfvars
      backend.tf
      versions.tf
      providers.tf
  nonprod/
    dev/
      shared/               # workspace: dev-shared
      cluster-b/            # workspace: dev-cluster-b
    staging/
      shared/               # workspace: staging-shared
.terraform/                 # Local provider cache and module cache (gitignored)
.terraform.lock.hcl         # Dependency lock file (committed)
```

- Keep root modules thin: compose child modules, set environment-specific variables, configure backends and providers.
- Put reusable infrastructure logic in `modules/` and publish to the TFC private registry.
- Group by infrastructure concern rather than by resource type when the project grows beyond a few modules.
- Separate environment configuration from module logic. Root modules in `environments/` compose modules; they do not own complex resource definitions directly.
- Each directory under `environments/` maps to exactly one TFC workspace. The workspace name should match the directory purpose (e.g., `prod-shared`, `cluster-a`).
- Keep `terraform.tfvars` environment-specific. Never share a single tfvars file across environments.
- Commit `.terraform.lock.hcl` to version control. Add `.terraform/` to `.gitignore`.
- Keep `backend.tf` separate from other configuration for clarity and to simplify backend migration.

## Module Design
- Define clear inputs (`variables.tf`), outputs (`outputs.tf`), and locals (`locals` blocks or `locals.tf`) for every module.
- Keep the module surface area minimal. Expose only the variables and outputs that consumers actually need. Do not expose every possible configuration knob.
- Source shared modules from the TFC private registry using the registry address format:

```hcl
module "networking" {
  source  = "app.terraform.io/my-org/networking/azurerm"
  version = "~> 2.0"

  # Module inputs...
}
```

- Use semantic versioning for shared modules consumed across repositories or teams. Pin module source versions in root modules.
- Prefer composition over inheritance. Build complex infrastructure by composing small, focused modules rather than creating deeply nested module hierarchies.
- Limit module nesting to two levels. A root module calling a child module calling a grandchild module is the practical maximum before debugging and state management become painful.
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
  name_prefix = "${var.name_prefix}-${var.environment}-${var.location}"
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
- Apply the `standard_tags` tagging strategy to every resource that supports tags. Use a shared `locals` block or a dedicated `tags.tf` for the default tag map, including dynamic timestamps with `lifecycle { ignore_changes }`:

```hcl
variable "standard_tags" {
  type        = map(string)
  description = "Standard tags applied to all resources."
  default     = {}
}

locals {
  common_tags = merge(var.standard_tags, {
    ApplicationName = var.application_name
    ProductName     = var.product_name
    Team            = var.team
    Contact         = var.contact
    Environment     = var.environment
    ManagedBy       = "terraform"
    CreationDate    = timestamp()
    Expiration      = timeadd(timestamp(), "8760h") # 1 year
  })
}

# Azure example
resource "azurerm_resource_group" "main" {
  name     = "${local.name_prefix}-rg"
  location = var.location
  tags     = local.common_tags

  lifecycle {
    # Dynamic tags (CreationDate, Expiration) are set once at creation.
    # Do not let Terraform revert them on subsequent applies.
    ignore_changes = [tags]
  }
}

# GCP example — use labels (GCP equivalent of tags)
resource "google_storage_bucket" "data" {
  name     = "${local.name_prefix}-data"
  location = var.location
  labels   = { for k, v in local.common_tags : lower(replace(k, "/[^a-z0-9_-]/", "_")) => lower(v) }

  lifecycle {
    ignore_changes = [labels]
  }
}
```

- Never hardcode names that should vary by environment.

## State Management
- Use Terraform Cloud (TFC) as the backend for all environments. Configure one TFC workspace per deployable unit (per directory under `environments/`):

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

- Each directory maps to exactly one TFC workspace. The workspace name should reflect the directory structure and purpose.
- Enable state locking. TFC handles locking automatically; never disable it.
- Isolate state per environment and per cluster. Each environment and cluster directory has its own workspace with its own state file. Never share state between dev, staging, and production.
- Consider isolating state per logical boundary within an environment when the blast radius of a single state file becomes too large (for example, separating networking state from compute state).
- Use `data "tfe_outputs"` for cross-workspace references. This is the standard pattern for referencing outputs from one TFC workspace in another:

```hcl
data "tfe_outputs" "networking" {
  organization = "my-org"
  workspace    = "prod-shared"
}

resource "azurerm_subnet" "app" {
  name                 = "${local.name_prefix}-app-subnet"
  resource_group_name  = data.tfe_outputs.networking.values.resource_group_name
  virtual_network_name = data.tfe_outputs.networking.values.vnet_name
  address_prefixes     = ["10.0.2.0/24"]
}
```

- Prefer `data "tfe_outputs"` over `terraform_remote_state` for cross-workspace references. The `tfe_outputs` data source is the TFC-native approach and supports workspace-level access controls.
- Some legacy configurations may use Azure Blob Storage backend for global configs. When encountered, plan migration to TFC but do not break existing workflows.
- Treat state files as sensitive. They contain resource metadata, attribute values, and potentially sensitive outputs. TFC encrypts state at rest; restrict workspace access via team permissions.
- Never commit state files to version control. Add `*.tfstate` and `*.tfstate.backup` to `.gitignore`.
- Use `terraform state` commands carefully. State manipulation (`mv`, `rm`, `import`) should be reviewed and documented like any infrastructure change.
- Use `moved` blocks for resource address refactoring instead of manual `terraform state mv` commands:

```hcl
moved {
  from = azurerm_resource_group.old_name
  to   = azurerm_resource_group.new_name
}
```

- Use `import` blocks (Terraform 1.5+) for importing existing resources declaratively:

```hcl
import {
  to = azurerm_resource_group.main
  id = "/subscriptions/.../resourceGroups/my-rg"
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
- Every output must have a `description`. This is especially critical for outputs consumed cross-workspace via `data "tfe_outputs"`, since the consumer has no other way to understand what the value represents.
- Organize outputs logically. Group them by resource or concern.
- Do not use `nullable = true` without a clear reason. Prefer required variables with validation over nullable variables with fallback logic.
- Keep `terraform.tfvars` files per environment. Do not use `terraform.tfvars` for shared defaults; use variable `default` values for that.

## Provider Configuration
- Pin provider versions to a specific minor version range in `versions.tf`. Support multi-cloud configurations with Azure as primary, GCP as secondary, and vSphere for on-prem:

```hcl
terraform {
  required_version = ">= 1.5.0"

  required_providers {
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
- Only include providers that the root module actually uses. Not every root module needs all three providers.
- Use provider aliases when managing resources across multiple subscriptions, projects, or regions:

```hcl
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

# GCP: multiple projects
provider "google" {
  project = var.primary_project_id
  region  = var.gcp_region
}

provider "google" {
  alias   = "shared_services"
  project = var.shared_services_project_id
  region  = var.gcp_region
}
```

- Keep provider configuration in a dedicated `providers.tf` file in root modules.
- **Bind provider identity at the root module, never in child modules.** Subscription IDs, project IDs, tenant IDs, and account identifiers must be variables in the root module's `providers.tf`, injected via TFC workspace variables or `terraform.tfvars`. Child modules must never hardcode or accept provider identity parameters — they inherit the provider from the calling root module. Even when a module is tightly bound to a specific organization or tenant, the root module is where that binding lives:

```hcl
# root module providers.tf — provider identity bound here
provider "azurerm" {
  features {}
  subscription_id = var.subscription_id  # Set via TFC workspace variable
  tenant_id       = var.tenant_id        # Set via TFC workspace variable
}

# child module — NO provider config, inherits from root
module "network" {
  source  = "app.terraform.io/my-org/network/azurerm"
  version = "~> 2.0"
  # ...
}
```

- Do not hardcode credentials, subscription IDs, project IDs, or tenant IDs in provider blocks or anywhere in `.tf` files. Use TFC workspace variables, workload identity federation, managed identities, environment variables, or credential helpers. Never use static service account keys or client secrets in code.
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

- When referencing resources across state boundaries, prefer `data "tfe_outputs"` for TFC-managed workspaces. Fall back to provider-specific data sources when querying resources not managed by Terraform.

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
- Restrict access to TFC workspaces via team permissions. Separate plan and apply permissions where possible.
- Audit provider permissions. The credentials Terraform uses should have only the permissions required for the resources it manages.
- Never store `.tfvars` files containing secrets in version control. Use TFC workspace variables, environment variables, or secret manager references instead.
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

## Backend And Workspace Strategy
- Use the directory + workspace hybrid strategy. Each directory under `environments/` maps to a TFC workspace. This provides clear separation by environment and cluster while leveraging TFC for state management, locking, and access control.
- **Directory-based per environment and per cluster**: Each environment (dev, staging, prod) has its own directory tree. Within each environment, separate directories exist for the environment-level resources and for each cluster. Each directory is a root module with its own TFC workspace.
- **Workspace naming convention**: Workspace names should match the directory purpose. For example, `environments/prod/environment/` maps to workspace `prod-shared`, and `environments/prod/cluster-a/` maps to workspace `cluster-a`.
- Configure the TFC backend in a dedicated `backend.tf` file:

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

- Cross-workspace references use `data "tfe_outputs"` to read outputs from other workspaces. This allows cluster workspaces to reference environment-level resources (e.g., networking, identity) without tight state coupling:

```hcl
data "tfe_outputs" "environment" {
  organization = "my-org"
  workspace    = "prod-shared"
}

# Use outputs from the environment workspace
resource "azurerm_kubernetes_cluster" "main" {
  name                = "${local.name_prefix}-aks"
  location            = data.tfe_outputs.environment.values.location
  resource_group_name = data.tfe_outputs.environment.values.resource_group_name
  dns_prefix          = local.name_prefix

  default_node_pool {
    name           = "default"
    node_count     = var.node_count
    vm_size        = var.vm_size
    vnet_subnet_id = data.tfe_outputs.environment.values.cluster_subnet_id
  }

  identity {
    type = "UserAssigned"
    identity_ids = [data.tfe_outputs.environment.values.cluster_identity_id]
  }
}
```

- Some legacy configurations may use Azure Blob Storage backend for global configs. Document these and plan TFC migration when practical.
- If the workspace name must vary per deployment, use workspace tags or variables rather than hardcoding:

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
- Run `terraform plan` in CI on every pull request. TFC can run speculative plans automatically on PRs.
- Never run `terraform apply -auto-approve` in production pipelines. Require explicit human approval for production applies. TFC supports approval gates via Sentinel policies or manual confirmation.
- Save the plan as an artifact and apply the saved plan, not a fresh plan. This prevents drift between the reviewed plan and the applied changes:

```bash
terraform plan -out=tfplan
# After approval:
terraform apply tfplan
```

- Run `terraform fmt -check` and `terraform validate` as early CI checks.
- Consider adding `tflint`, `tfsec`/`trivy`, and `checkov` as part of the CI pipeline. These are not currently in use but are recommended for catching issues early.
- Implement drift detection by scheduling regular `terraform plan` runs and alerting on unexpected changes. TFC supports scheduled runs for this purpose.
- Use separate CI/CD credentials per environment. Production apply credentials should not be available to development pipelines.
- Cache the `.terraform` directory in CI to speed up `terraform init`, but invalidate the cache when `.terraform.lock.hcl` changes.
- Pin the Terraform version in CI. Use `tfenv`, `asdf`, or a container image with a specific version. Never rely on "latest."
- Store plan artifacts securely. Plans can contain sensitive data.

## Anti-Patterns To Reject
- Monolith root modules that manage hundreds of resources in a single state file with enormous blast radius
- Hardcoded provider identity (subscription IDs, project IDs, tenant IDs) in `.tf` files — these must be variables injected via TFC workspace variables
- Hardcoded values that should be variables, especially resource IDs, region names, or CIDR blocks
- Overuse of provisioners (`local-exec`, `remote-exec`, `file`) for tasks that should be handled by proper resources, data sources, or configuration management tools
- Module nesting deeper than two levels, making debugging and state inspection painful
- `terraform apply -auto-approve` in production environments or shared environments
- Using `local-exec` to run scripts that create or modify infrastructure that should be a Terraform resource
- Storing secrets in `terraform.tfvars` or variable defaults committed to version control
- Using `count` with lists where `for_each` with a map or set would prevent index-based cascading changes
- Ignoring `.terraform.lock.hcl` or not committing it, leading to inconsistent provider versions across machines
- Blanket `ignore_changes = [tags]` instead of ignoring only specific computed tag attributes
- Using `ignore_changes` on core resource attributes (`name`, `address_space`, `resource_group_name`) as a permanent fixture instead of fixing configuration-state mismatches
- Overusing `ignore_changes` to suppress legitimate configuration drift instead of fixing the root cause
- Kebab-case output names (`my-output`) instead of snake_case (`my_output`)
- Outputs without `description` fields, especially cross-workspace outputs consumed via `tfe_outputs`
- Committing plan files or state files to git — these contain sensitive data and belong in CI artifacts or TFC
- Storing state files in version control instead of a remote backend
- Using `terraform_remote_state` as the primary integration mechanism between loosely coupled stacks — prefer `data "tfe_outputs"` for TFC workspaces
- Creating resources with public access (public IPs, `0.0.0.0/0` security group rules) without explicit, documented justification
- Writing modules that expose every possible variable as an input, creating an unmanageable surface area
- Using `terraform taint` (deprecated) instead of `terraform apply -replace`
- Mixing workspace-based and directory-based environment strategies in the same project without clear justification
- Running `terraform apply` without first reviewing the plan output
- Using the `default` workspace in production when workspaces are the environment strategy
- Using local accounts or static credentials when RBAC and workload identity are available
- Sourcing modules from git URLs when TFC private registry is available

## Code Quality Checklist
Before considering a Terraform task complete, verify:
- `terraform fmt` has been run and the code is formatted.
- `terraform validate` passes without errors.
- Consider running `tflint` — passes without errors or warnings, or warnings are acknowledged and documented.
- Consider running `tfsec` or `trivy` — passes, or findings are reviewed and accepted risks are documented.
- Consider running `checkov` for compliance-relevant resources — passes, or exceptions are documented.
- All variables have `type`, `description`, and `validation` where appropriate.
- All outputs have `description`.
- All resources that support tags have the `standard_tags` tag set applied. If dynamic tags are used, only specific computed tag attributes are ignored — not all tags.
- All output names use snake_case consistently. No kebab-case output names.
- No plan files (`*.tfplan`, `plan`) or state files are committed to git.
- No secrets are hardcoded in any `.tf` or `.tfvars` file.
- Sensitive variables and outputs are marked `sensitive = true`.
- TFC backend is configured with the correct workspace name.
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
3. Blast radius — How much can this change break? Is state properly isolated across TFC workspaces?
4. Modularity — Are modules focused, minimal, composable, and sourced from the TFC private registry?
5. Readability — Can a reviewer understand the plan output and the code?
6. Naming and tagging — Are `standard_tags` conventions consistent? Is naming computed via locals?
7. State safety — Is state isolated per workspace, locked, encrypted, and backed up?

Do not lead with style preferences. Flag security, correctness, and blast radius concerns first; address formatting and naming only after operational concerns.

## What Good Output Looks Like
- Code is straightforward to review by reading `terraform plan` output.
- Module boundaries are obvious and follow logical infrastructure concerns.
- State boundaries match operational and security boundaries, with one TFC workspace per deployable unit.
- Failure modes are explicit: `prevent_destroy` on critical resources, validation on inputs, clear error messages.
- Security posture is visible: RBAC-only, workload identity, per-service identities, encryption enabled, network rules restrictive by default.
- Operational concerns such as `standard_tags` tagging, naming via locals, and TFC workspace management are addressed during implementation, not bolted on afterward.
- Cross-workspace references use `data "tfe_outputs"` with clear workspace naming.
- The code can pass a serious infrastructure audit without needing the reader to infer design intent.

## Invocation Template
Use this skill with a prompt that supplies repository-specific context. Example:

```text
Use Terraform Development Guidance.
Add a new AKS cluster module in /path/to/repo/modules/compute.
Keep the module focused on a single managed Kubernetes cluster.
Follow the existing standard_tags tagging and naming conventions in the repository.
Source the module from the TFC private registry.
Ensure RBAC-only, private cluster, and workload identity are enabled.
Add validation blocks on all input variables.
```
