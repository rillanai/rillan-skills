<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.0.0 -->
# Terraform Test Strategy And Generation

## Purpose
Use this skill when designing test strategies, generating tests, analyzing test coverage, or reviewing tests for Terraform codebases. This covers variable validation and plan-time assertions, native `terraform test` with mock providers, contract testing, and compliance scanning recommendations.

This skill defines the test strategy and generation contract for Terraform work. It is intended for writing terraform tests, designing test strategies, implementing validation rules, module contract testing, and compliance testing.

## Skill Use
- Load this skill when the task is to write, design, review, or improve tests for Terraform code.
- Treat this skill as the governing contract for test design and implementation unless the repository has stricter local conventions.
- Keep repository-specific test requirements in the task prompt.
- Match established project test conventions when they are clear and defensible.
- When this skill conflicts with casual convenience, follow this skill.

## Tool Use
This skill is tool-agnostic and works with Claude Code, Codex, OpenCode, and similar assistants. Map its guidance to whatever file-reading, editing, search, and shell-execution tools your environment exposes.

- Run the tests you write. A new `.tftest.hcl` file is not done until `terraform test` has executed it and the result is known.
- Invoke `terraform fmt`, `terraform validate`, and `terraform plan` directly; do not describe what they would show.
- Issue independent tool calls (reading multiple files, scanning modules, checking CI config) in parallel.
- Report actual tool output for failures, not paraphrased summaries.

## When To Use
Use this skill when the user asks for any of the following:
- Writing terraform test files (`.tftest.hcl`)
- Designing a test strategy for a Terraform module or root module
- Adding variable validation rules, preconditions, or postconditions
- Module contract testing to verify interface stability
- Using mock providers in terraform test for cloud-agnostic testing
- Testing multi-cloud modules across Azure, GCP, and vSphere providers
- Compliance and policy scanning recommendations (tfsec, checkov, trivy, OPA)
- Test coverage analysis for infrastructure code
- Reviewing existing tests for gaps or anti-patterns

Do not use this skill for:
- Writing or modifying Terraform infrastructure code itself (use a Terraform development skill)
- Documentation tasks (use the Terraform documentation skill)
- Migration planning (use the Terraform migration skill)
- General Go test writing that is not Terraform-related (use a Go development skill)

## Test Philosophy
- **Test infrastructure contracts**: Verify that your inputs produce the expected resources with the expected configurations. You are testing your configuration, not the provider or the cloud API.
- **Test at the right level**: Use the cheapest test that catches the bug. Validation blocks catch type errors. Plan-time tests catch configuration errors. Integration tests catch wiring errors. Do not use integration tests for what a validation block would catch.
- **Minimize cloud costs**: Real resources cost real money. Prefer plan-only tests with mock providers. When apply is necessary, use the smallest possible resource sizes, shortest lifetimes, and dedicated test subscriptions or projects. Always destroy after testing.
- **Test what matters**: Focus test effort on module interfaces (inputs produce correct outputs), security-critical configuration (encryption, access controls, network rules), and complex conditional logic (count, for_each, dynamic blocks).
- **Keep tests deterministic**: Avoid tests that pass or fail depending on cloud API timing, region availability, or quota limits. Use mock providers for deterministic testing. When real cloud interaction is needed, use retries with backoff.

## Test Types And When To Use Each

### Terraform Native Tests (terraform test) — Primary Approach
Available in Terraform 1.6+. Mock providers available in Terraform 1.7+. Use `.tftest.hcl` files for module contract validation. This is the primary testing approach.

**When to use**:
- Validating that a module produces expected resource configurations from given inputs.
- Testing variable validation rules catch invalid input.
- Plan-time assertions that do not require real resource creation.
- Verifying output values and resource attributes against expected values.
- Quick feedback during module development.
- Testing module logic in isolation using mock providers (no cloud credentials needed).
- Testing multi-cloud modules by mocking each provider independently.

**When not to use**:
- When you need to verify actual cloud resource behavior after creation (runtime behavior, connectivity, API responses).
- When test logic requires complex programmatic assertions beyond HCL expressions.

### Unit-Style Validation
Built into Terraform configuration. No test files required.

**Variable validation blocks**:
```hcl
variable "environment" {
  description = "Deployment environment. Must be one of: dev, staging, production."
  type        = string

  validation {
    condition     = contains(["dev", "staging", "production"], var.environment)
    error_message = "Environment must be one of: dev, staging, production."
  }
}

variable "address_space" {
  description = "Virtual network address space. Must be a valid CIDR range in RFC1918 space."
  type        = string

  validation {
    condition     = can(cidrhost(var.address_space, 0))
    error_message = "Address space must be a valid CIDR notation."
  }
}
```

**Preconditions** (verify assumptions before resource creation):
```hcl
resource "azurerm_linux_virtual_machine" "app" {
  name                = "${var.name_prefix}-vm"
  resource_group_name = var.resource_group_name
  location            = var.location
  size                = var.vm_size

  lifecycle {
    precondition {
      condition     = var.enable_accelerated_networking ? can(regex("Standard_D[0-9]+", var.vm_size)) : true
      error_message = "Accelerated networking requires a VM size that supports it (Standard_D series or higher)."
    }
  }
}
```

**Postconditions** (verify results after resource creation):
```hcl
resource "azurerm_postgresql_flexible_server" "main" {
  name                = "${var.name_prefix}-psql"
  resource_group_name = var.resource_group_name
  location            = var.location

  lifecycle {
    postcondition {
      condition     = self.fqdn != null
      error_message = "PostgreSQL server must have a valid FQDN after creation."
    }
  }
}
```

**Check blocks** (continuous validation, non-blocking):
```hcl
check "app_health" {
  data "http" "api" {
    url = "https://${azurerm_linux_web_app.main.default_hostname}/health"
  }

  assert {
    condition     = data.http.api.status_code == 200
    error_message = "Application health check failed after deployment."
  }
}
```

### Plan-Time Tests
Analyze `terraform plan` output without creating any real resources.

**When to use**:
- Verifying resource counts and types in the plan.
- Checking that conditional logic (count, for_each) produces the expected resources.
- Validating resource configurations before incurring cloud costs.
- CI pipelines where cost must be zero.

**How to use**:
- Run `terraform plan -out=tfplan` followed by `terraform show -json tfplan` to produce a machine-readable plan.
- Assert against the JSON plan structure for resource changes, output changes, and configuration values.
- Use OPA/Rego or custom scripts for plan JSON assertions.
- Use `command = plan` in terraform native test `run` blocks (preferred approach).

### Integration Tests
Create, verify, and destroy real cloud resources.

**When to use**:
- Verifying that resources are actually created and functional, not just planned.
- Testing network connectivity, RBAC permissions, firewall rules, and other runtime behaviors.
- Validating that resource dependencies and ordering work correctly.
- End-to-end module verification before release.

**When not to use**:
- For every PR. Integration tests are expensive. Run them on merge to main, on release, or on a schedule.
- When a plan-time test or mock provider test would catch the same issue.
- In shared cloud subscriptions or projects without test isolation.

### Compliance Tests (Recommended)
Policy enforcement against Terraform plans and configurations. These tools are recommended for adoption but may not be currently in use.

**When to consider**:
- Enforcing organizational security policies (encryption required, public access blocked, approved VM sizes).
- Meeting compliance frameworks (CIS benchmarks, SOC2, HIPAA, PCI-DSS).
- Preventing misconfigurations before they reach production.
- Continuous compliance monitoring in CI/CD.

**Tools to evaluate**:
- **tfsec/trivy**: Static analysis for security misconfigurations. Fast, broad rule sets, good CI integration. Supports Azure, GCP, and general Terraform patterns.
- **checkov**: Compliance framework scanning with built-in rules for CIS, SOC2, HIPAA. Supports custom policies. Good multi-cloud support.
- **OPA/Rego**: Flexible policy language for analyzing plan JSON. Good for custom organizational policies that do not map to existing tool rules.

```bash
# tfsec/trivy static analysis (recommended for CI adoption)
tfsec /path/to/terraform --format json --out results.json
trivy config /path/to/terraform --severity HIGH,CRITICAL

# checkov compliance scanning
checkov --directory /path/to/terraform
checkov --directory /path/to/terraform --framework terraform --output junitxml > test-results.xml

# OPA for custom policies against plan JSON
terraform plan -out=tfplan
terraform show -json tfplan > tfplan.json
opa eval --input tfplan.json --data policies/ 'data.terraform.policies.deny'
```

### Contract Tests
Verify that module interfaces remain stable across changes.

**When to use**:
- Before releasing a new version of a shared module.
- Verifying that outputs still have the expected names, types, and non-null values.
- Ensuring that required variables have not changed in breaking ways.
- Detecting unintentional interface changes during refactoring.

**How to implement**:
- Maintain a set of terraform native test files that call the module with known inputs and assert specific output values.
- Test with minimum valid configuration (only required variables) to catch accidental required variable additions.
- Test with full configuration to catch output regressions.

## Terraform Test Framework (.tftest.hcl)

### File Structure
Test files use the `.tftest.hcl` extension and are placed in a `tests/` directory by convention.

```
modules/
  network/
    main.tf
    variables.tf
    outputs.tf
    tests/
      basic.tftest.hcl
      validation.tftest.hcl
      full_config.tftest.hcl
      multi_cloud.tftest.hcl
```

### Run Blocks
Each `run` block is a test step. Steps execute sequentially and can share state.

```hcl
run "creates_network_with_correct_address_space" {
  command = plan

  variables {
    address_space       = "10.0.0.0/16"
    environment         = "test"
    location            = "eastus2"
    resource_group_name = "rg-test"
  }

  assert {
    condition     = azurerm_virtual_network.main.address_space[0] == "10.0.0.0/16"
    error_message = "Virtual network address space does not match input variable."
  }

  assert {
    condition     = length(azurerm_subnet.main) == 2
    error_message = "Expected 2 subnets based on default subnet configuration."
  }
}
```

### Variables Blocks
Supply test inputs. Each `run` block can have its own `variables` block.

```hcl
run "minimal_configuration" {
  command = plan

  variables {
    address_space       = "10.0.0.0/16"
    environment         = "test"
    location            = "eastus2"
    resource_group_name = "rg-test"
  }

  assert {
    condition     = azurerm_virtual_network.main.address_space[0] == "10.0.0.0/16"
    error_message = "Virtual network should be created with the specified address space."
  }
}

run "full_configuration" {
  command = plan

  variables {
    address_space       = "10.0.0.0/16"
    environment         = "production"
    location            = "eastus2"
    resource_group_name = "rg-prod"
    enable_nat_gateway  = true
    subnets = {
      app  = { address_prefix = "10.0.1.0/24" }
      data = { address_prefix = "10.0.2.0/24" }
      mgmt = { address_prefix = "10.0.3.0/24" }
    }
  }

  assert {
    condition     = length(azurerm_subnet.main) == 3
    error_message = "Expected 3 subnets when 3 subnet definitions are provided."
  }

  assert {
    condition     = azurerm_nat_gateway.main[0].name != ""
    error_message = "NAT gateway should be created when enable_nat_gateway is true."
  }
}
```

### Assert Blocks
Each assert has a `condition` (HCL expression evaluating to bool) and an `error_message`.

```hcl
assert {
  condition     = output.vnet_id != ""
  error_message = "Virtual network ID output must not be empty."
}

assert {
  condition     = azurerm_virtual_network.main.dns_servers == null || length(azurerm_virtual_network.main.dns_servers) > 0
  error_message = "DNS servers must either be null (Azure default) or explicitly configured."
}

assert {
  condition     = alltrue([for nsg in azurerm_network_security_group.main : length(nsg.security_rule) > 0])
  error_message = "All network security groups must have at least one security rule."
}
```

### Plan vs Apply Commands
- `command = plan`: Validates the plan without creating resources. Fast and free. Use this by default.
- `command = apply`: Creates real resources, runs assertions, and destroys on completion. Use only when plan-time assertions are insufficient.

```hcl
run "plan_only_validation" {
  command = plan

  variables {
    environment = "test"
    name_prefix = "test"
  }

  assert {
    condition     = azurerm_storage_account.main.name == "testsa"
    error_message = "Storage account name must follow naming convention."
  }
}

run "apply_and_verify" {
  command = apply

  variables {
    environment = "test"
    name_prefix = "test"
  }

  assert {
    condition     = output.storage_account_id != ""
    error_message = "Storage account ID must be populated after creation."
  }
}
```

### Mock Providers
Use mock providers (Terraform 1.7+) to test module logic without any cloud interaction. Mock providers return synthetic values for all resource attributes, allowing you to test naming conventions, conditional logic, and output wiring without credentials or cloud API calls.

**Basic mock provider**:
```hcl
mock_provider "azurerm" {}

run "test_naming_convention" {
  command = plan

  variables {
    environment = "staging"
    app_name    = "myapp"
    location    = "eastus2"
  }

  assert {
    condition     = azurerm_resource_group.main.name == "rg-staging-myapp"
    error_message = "Resource group name must follow the naming convention: rg-{environment}-{app_name}."
  }
}
```

**Mock provider with alias**:
```hcl
mock_provider "azurerm" {
  alias = "hub"
}

mock_provider "azurerm" {
  alias = "spoke"
}

run "test_hub_spoke_peering" {
  command = plan

  variables {
    hub_vnet_name   = "vnet-hub"
    spoke_vnet_name = "vnet-spoke"
  }

  assert {
    condition     = azurerm_virtual_network_peering.hub_to_spoke.name == "peer-hub-to-spoke"
    error_message = "Peering name must follow naming convention."
  }
}
```

**Mock provider with mock data sources**:
When your module uses data sources to look up existing resources, you can provide mock data to control what those data sources return:

```hcl
mock_provider "azurerm" {
  mock_data "azurerm_client_config" {
    defaults = {
      tenant_id       = "00000000-0000-0000-0000-000000000000"
      subscription_id = "00000000-0000-0000-0000-000000000001"
      object_id       = "00000000-0000-0000-0000-000000000002"
    }
  }

  mock_data "azurerm_resource_group" {
    defaults = {
      name     = "rg-existing"
      location = "eastus2"
      id       = "/subscriptions/00000000-0000-0000-0000-000000000001/resourceGroups/rg-existing"
    }
  }
}

run "test_role_assignment_uses_current_identity" {
  command = plan

  variables {
    resource_group_name = "rg-existing"
  }

  assert {
    condition     = azurerm_role_assignment.main.principal_id == "00000000-0000-0000-0000-000000000002"
    error_message = "Role assignment must use the current client's object ID."
  }
}
```

**Mock provider with mock resources**:
Control what values mock resources return when referenced by other resources:

```hcl
mock_provider "azurerm" {
  mock_resource "azurerm_key_vault" {
    defaults = {
      id        = "/subscriptions/00000000/resourceGroups/rg-test/providers/Microsoft.KeyVault/vaults/kv-test"
      vault_uri = "https://kv-test.vault.azure.net/"
    }
  }
}

run "test_secret_references_correct_vault" {
  command = plan

  variables {
    key_vault_name = "kv-test"
  }

  assert {
    condition     = azurerm_key_vault_secret.main.key_vault_id == "/subscriptions/00000000/resourceGroups/rg-test/providers/Microsoft.KeyVault/vaults/kv-test"
    error_message = "Secret must be created in the correct Key Vault."
  }
}
```

### Expect Failures (Negative Testing)
Test that invalid inputs are correctly rejected by validation rules.

```hcl
run "rejects_invalid_environment" {
  command = plan

  variables {
    environment         = "invalid"
    address_space       = "10.0.0.0/16"
    location            = "eastus2"
    resource_group_name = "rg-test"
  }

  expect_failures = [
    var.environment,
  ]
}

run "rejects_invalid_cidr" {
  command = plan

  variables {
    environment         = "dev"
    address_space       = "not-a-cidr"
    location            = "eastus2"
    resource_group_name = "rg-test"
  }

  expect_failures = [
    var.address_space,
  ]
}

run "rejects_empty_name_prefix" {
  command = plan

  variables {
    environment         = "dev"
    address_space       = "10.0.0.0/16"
    location            = "eastus2"
    resource_group_name = "rg-test"
    name_prefix         = ""
  }

  expect_failures = [
    var.name_prefix,
  ]
}
```

### Testing Preconditions and Postconditions
Test that lifecycle preconditions and postconditions trigger correctly:

```hcl
run "rejects_incompatible_vm_size_for_accelerated_networking" {
  command = plan

  variables {
    vm_size                      = "Standard_B1s"
    enable_accelerated_networking = true
  }

  expect_failures = [
    azurerm_linux_virtual_machine.app,
  ]
}
```

### Testing with Override Files
Use override files in test configurations to substitute provider configurations or backend settings for testing:

```hcl
run "test_with_overridden_provider" {
  command = plan

  override_resource {
    target = azurerm_resource_group.existing
    values = {
      name     = "rg-override"
      location = "eastus2"
    }
  }

  assert {
    condition     = azurerm_virtual_network.main.resource_group_name == "rg-override"
    error_message = "Virtual network must be created in the overridden resource group."
  }
}
```

## Testing Multi-Cloud Modules

When modules support multiple cloud providers (Azure, GCP, vSphere), test each provider path independently using mock providers.

### Azure Module Testing
```hcl
mock_provider "azurerm" {
  mock_data "azurerm_client_config" {
    defaults = {
      tenant_id       = "00000000-0000-0000-0000-000000000000"
      subscription_id = "00000000-0000-0000-0000-000000000001"
    }
  }
}

run "azure_compute_naming" {
  command = plan

  variables {
    cloud       = "azure"
    environment = "production"
    app_name    = "myapp"
  }

  assert {
    condition     = azurerm_linux_virtual_machine.main[0].name == "vm-production-myapp"
    error_message = "Azure VM must follow naming convention: vm-{environment}-{app_name}."
  }
}
```

### GCP Module Testing
```hcl
mock_provider "google" {}
mock_provider "google-beta" {}

run "gcp_compute_naming" {
  command = plan

  variables {
    cloud       = "gcp"
    environment = "production"
    app_name    = "myapp"
    project_id  = "my-project-123"
  }

  assert {
    condition     = google_compute_instance.main[0].name == "production-myapp"
    error_message = "GCP instance must follow naming convention: {environment}-{app_name}."
  }
}
```

### vSphere Module Testing
```hcl
mock_provider "vsphere" {
  mock_data "vsphere_datacenter" {
    defaults = {
      id   = "datacenter-1"
      name = "dc-onprem"
    }
  }

  mock_data "vsphere_compute_cluster" {
    defaults = {
      id            = "cluster-1"
      name          = "cluster-prod"
      resource_pool_id = "resgroup-1"
    }
  }

  mock_data "vsphere_datastore" {
    defaults = {
      id   = "datastore-1"
      name = "ds-ssd-01"
    }
  }
}

run "vsphere_vm_configuration" {
  command = plan

  variables {
    datacenter_name = "dc-onprem"
    cluster_name    = "cluster-prod"
    vm_name         = "app-server-01"
    num_cpus        = 4
    memory          = 8192
  }

  assert {
    condition     = vsphere_virtual_machine.main.num_cpus == 4
    error_message = "vSphere VM must be provisioned with the requested CPU count."
  }

  assert {
    condition     = vsphere_virtual_machine.main.memory == 8192
    error_message = "vSphere VM must be provisioned with the requested memory."
  }
}
```

### Cross-Provider Utility Module Testing
For modules that use utility providers alongside cloud providers (e.g., infoblox for DNS):

```hcl
mock_provider "azurerm" {}
mock_provider "infoblox" {
  mock_data "infoblox_ipv4_network" {
    defaults = {
      cidr = "10.0.0.0/24"
    }
  }
}

run "ip_allocation_from_infoblox" {
  command = plan

  variables {
    network_name = "app-network"
    dns_zone     = "internal.example.com"
  }

  assert {
    condition     = infoblox_ip_allocation.main.cidr == "10.0.0.0/24"
    error_message = "IP allocation must reference the Infoblox network."
  }
}
```

## When To Consider Terratest

Terratest (Go-based integration testing) is a heavyweight tool for cases where `terraform test` is insufficient. Most modules should be tested with native `terraform test` and mock providers. Consider Terratest only when:

- You need to make HTTP requests to deployed resources, SSH into instances, or query cloud APIs directly to verify runtime behavior that cannot be asserted from plan output or Terraform state.
- You need complex programmatic retry logic with exponential backoff for eventually-consistent resources.
- You need to orchestrate multi-step workflows across multiple independent Terraform configurations in a specific sequence.
- The repository already has an established Terratest suite and the team has Go expertise.

If Terratest is adopted, follow these principles:
- Always use `defer terraform.Destroy(t, terraformOptions)` immediately after setting up options, before `InitAndApply`. Never place Destroy after assertions.
- Use `t.Parallel()` for independent tests. Ensure resource names include a random component to avoid collisions.
- Use unique naming prefixes per test run (e.g., include a short random string or timestamp).
- Use the smallest viable resource sizes. Do not test with large VM sizes when a small size validates the same behavior.
- Run Terratest on merge to main or on a schedule, not on every PR.

## Testing Patterns

### Fixture Modules
Create minimal fixture modules that wrap the module under test with provider configuration and test-specific defaults.

```
tests/
  fixtures/
    network_basic/
      main.tf       # Calls the real module with minimal config
      variables.tf  # Exposes only what tests need to vary
      outputs.tf    # Re-exports module outputs for test assertions
    network_full/
      main.tf       # Calls the real module with full config
      variables.tf
      outputs.tf
```

### Test Isolation
- Use unique naming prefixes or suffixes per test run to avoid resource name collisions. Include a random or timestamp-based component.
- Use dedicated test subscriptions, projects, or vSphere resource pools when possible.
- Never run integration tests against shared development or staging environments.
- Use separate state files per test run. Never share state between parallel tests.

### Cost Management
- Default to `command = plan` in terraform native tests.
- Use mock providers (Terraform 1.7+) to eliminate cloud interaction entirely for logic tests.
- Use the smallest viable resource sizes in integration tests.
- Set aggressive timeouts and destroy resources immediately after assertions.
- Run integration tests on a schedule (nightly, weekly) rather than on every commit.
- Prefer mock providers in terraform native tests when cloud interaction is not the point of the test.

### Parallel Test Execution
- Terraform native tests: `run` blocks within a file execute sequentially. Separate test files may execute in parallel depending on the runner.
- Be aware of cloud API rate limits when running many parallel tests. Use retry with backoff.
- Each parallel test must have its own state file. Never share Terraform state between parallel tests.

### State Isolation In Tests
- Each test run must use a separate state file or separate backend path.
- For terraform native tests, state is managed automatically per test run.
- Never point tests at production or shared state backends.

## Coverage Expectations

### Minimum Coverage Standard
Every module published to the TFC private registry should have at minimum:
- **Validation tests**: Every variable with a `validation` block must have a corresponding `expect_failures` test proving the validation rejects bad input.
- **Minimal configuration test**: A plan-only test with only required variables set, proving the module plans successfully with defaults.
- **Naming convention test**: A plan-only test asserting that computed resource names follow the expected pattern.
- **Security-critical assertion tests**: Plan-only tests asserting encryption is enabled, public access is disabled, RBAC is configured, and network rules are restrictive — for every resource that has these attributes.

Critical infrastructure modules (networking, identity, compute clusters, encryption) should additionally have:
- **Full configuration test**: All optional variables exercised, asserting correct resource counts and configurations.
- **Conditional logic tests**: Each `count` and `for_each` path tested (both the 0/false and N/true cases).
- **Output tests**: Assertions that outputs are non-empty and correctly shaped.

### Resource Coverage
Track which module resources have tests that assert their configuration:
- List all resources in the module.
- Map each resource to the test files and assertions that cover it.
- Identify resources with no test coverage.
- Prioritize coverage for security-critical resources (role assignments, network security groups, encryption settings, public access configurations).

### Scenario Coverage
Track which input scenarios are tested:
- **Happy path**: Default configuration produces expected resources.
- **Minimal configuration**: Only required variables produce a valid plan.
- **Full configuration**: All optional variables exercised.
- **Error cases**: Invalid inputs are rejected by validation rules.
- **Edge cases**: Zero counts, empty lists, maximum values, conditional resources (count = 0 vs count = 1).
- **Environment variations**: Different environments produce appropriately different configurations.
- **Multi-cloud variations**: Each supported provider path is tested with appropriate mock providers.

### Compliance Coverage
Track which compliance policies are enforced:
- Map compliance requirements (CIS benchmarks, organizational policies) to specific test assertions or policy rules.
- Identify compliance requirements without automated enforcement.
- When compliance scanning tools (tfsec, checkov) are adopted, track rule results over time to detect regressions.

## Test Commands
Standard test commands for Terraform codebases:

```bash
# Native terraform testing (primary approach)
terraform test                                   # Run all .tftest.hcl files
terraform test -filter=tests/basic.tftest.hcl    # Run a specific test file
terraform test -verbose                          # Verbose output with plan details

# Validation (no tests, but catches syntax and type errors)
terraform validate                               # Validate configuration syntax and types
terraform fmt -check -recursive                  # Check formatting
terraform fmt -recursive                         # Auto-format

# Linting
tflint                                           # Lint with default rules
tflint --init                                    # Initialize tflint plugins
tflint --module                                  # Lint as a module (check unused variables, etc.)

# Security scanning (recommended for adoption)
tfsec .                                          # Security scan
tfsec . --soft-fail                              # Security scan, non-blocking
trivy config .                                   # Trivy config scan
checkov --directory .                            # Compliance scan

# Plan-based testing
terraform plan -out=tfplan                       # Generate plan
terraform show -json tfplan > plan.json          # Export plan as JSON for policy testing
```

## Anti-Patterns To Reject
- **Tests that only check plan succeeded**: Running `terraform plan` and asserting it did not error is not a test. Assert specific resource attributes, counts, and configurations.
- **No cleanup or destroy**: Integration tests that create resources and do not destroy them. Always use `defer terraform.Destroy` or `command = apply` (which auto-destroys).
- **Hardcoded regions or subscriptions**: Tests that only work in a specific region or subscription. Parameterize region and subscription-specific values, or use mock providers.
- **Testing provider behavior**: Do not test that `azurerm_virtual_network` creates a virtual network. That is the provider's job. Test that your configuration produces the right network with the right settings.
- **Flaky tests from eventual consistency**: Tests that intermittently fail because a resource is not yet available. Use retries with backoff, not hope. Or use mock providers to eliminate flakiness.
- **Shared state between tests**: Tests that read or modify shared Terraform state. Each test must be independent.
- **Integration tests on every commit**: Running expensive apply-and-destroy tests on every PR. Use plan-only and mock-provider tests for PRs, integration tests on merge or on a schedule.
- **No negative tests**: Testing only that valid input works. Also test that invalid input is correctly rejected.
- **Enormous test modules**: Test fixtures that deploy entire environments to test one module. Keep fixtures minimal and focused.
- **Ignoring test output**: Running tests in CI without failing the pipeline on test failure. Tests must gate deployment.
- **Hardcoded resource names without uniqueness**: Test resources with static names that collide when tests run in parallel or are re-run before cleanup completes.
- **Tests that depend on execution order across files**: Assuming test files run in a specific order. Each test file must be independently executable.
- **Skipping mock providers for logic tests**: Using real cloud credentials to test naming conventions, conditional logic, or output wiring when mock providers would suffice. Mock providers are faster, cheaper, and more deterministic.

## Invocation Template
Use this skill with a prompt that supplies repository-specific context. Example:

```text
Use Terraform Test Strategy And Generation.
Write terraform native tests for the network module at /path/to/modules/network.
Cover minimal config, full config, and invalid CIDR rejection.
Use mock providers for all plan-only tests.
Add variable validation blocks for any unvalidated inputs.
Include mock data sources for azurerm_client_config and azurerm_resource_group.
```
