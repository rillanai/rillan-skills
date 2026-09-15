<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.3.0 -->
# Terraform Testing

Use this mode to design, write, or review tests for Terraform or OpenTofu configuration. Discover the repository's CLI, minimum version, module roots, provider constraints, test directories, fixtures, CI commands, and credential model before choosing a test feature or running a test.

## Choose the cheapest test that proves the claim

| Technique | What it establishes | What it does not establish |
|---|---|---|
| `fmt` and `validate` | canonical formatting, syntax, types, references, and internal configuration consistency | input-specific plan behavior or remote API acceptance |
| variable validation and preconditions | invalid inputs or known configuration invariants fail with useful diagnostics | provider or runtime behavior |
| native test with `command = plan` | assertions over a planned configuration | successful creation or runtime behavior |
| native test with a mock provider | module logic against the provider schema and controlled fake computed values | authentication, provider API behavior, eventual consistency, or real defaults |
| native apply test with real providers, or an external integration test | real provider interaction and post-apply assertions | broad production suitability without representative conditions |
| policy/static analysis | configured patterns match selected rules | runtime correctness or complete security coverage |

Prefer a narrow validation or plan assertion for module logic. Use real infrastructure only for claims that require provider or service behavior, such as permissions, connectivity, lifecycle, or API normalization.

## Native test semantics

Native test files conventionally live in `tests/` and use `.tftest.hcl` or `.tftest.json`. Confirm support in the repository's minimum version: Terraform introduced the test framework in 1.6 and provider mocking in 1.7. OpenTofu support and details vary by release, so use its documentation for the discovered version.

Each `run` is a plan or apply operation. **The default is `command = apply`**, which can create real infrastructure when real providers are configured. State the command explicitly so cost and side effects are obvious.

For a module whose `name` validation rejects empty strings:

```hcl
run "rejects_empty_name" {
  command = plan

  variables { name = "" }

  expect_failures = [var.name]
}
```

Use assertions that describe a public contract or a regression-prone transformation. Avoid duplicating HCL expressions verbatim or testing Terraform's own type checker. Give failures an actionable message with expected behavior.

Use `expect_failures` only for custom conditions on the named checkable object. Unrelated diagnostics still fail the test. Prefer it with `command = plan`; an apply run can fail during its planning phase before reaching a failure expected only during apply. Do not weaken production validation to make a negative test easier.

Runs that select the same module configuration or `state_key` can share test state; references to earlier run outputs also impose ordering. Keep that sequence understandable. Use distinct state keys, directories, accounts/projects, and unique names when tests can run concurrently.

Reference: [Terraform test language](https://developer.hashicorp.com/terraform/language/tests).

## Plans, mocks, and remote access

`command = plan` prevents resource creation for that run, but it is not automatically offline. A real provider may authenticate, read data sources, inspect remote state, or refresh existing objects. Use explicit mock providers or hermetic fixtures when the intended claim does not require those calls.

Mocks use real provider schemas while substituting generated or supplied values for provider operations. Provide deterministic defaults or targeted overrides only for computed values needed by the assertion. Assert your module's configuration and branching, not the fake value itself. A passing mock test proves Terraform-side wiring; claims about provider or remote behavior require real-provider verification.

Mocking still requires the provider schema to be available, so initialization/provider installation may require a configured mirror, cache, or network. Never claim an offline test unless the complete run was demonstrated without network and remote credentials.

Reference: [Terraform provider mocking](https://developer.hashicorp.com/terraform/language/tests/mocking).

## Real-resource tests

Before running an apply test, resolve its account/project, region, identity, permissions, naming isolation, expected cost, concurrency model, and cleanup path. Do not run it with ambient production credentials.

Terraform attempts to destroy resources created during native tests, including after many failures, but cleanup is not guaranteed. Provider errors, interrupts, crashes, timeouts, dependency bugs, or resources that resist deletion can leave infrastructure behind. Capture enough identifiers to find leftovers, set timeouts where supported, and report cleanup failures prominently. Use a scheduled sweeper or expiry tags only when the project already has that control.

Keep destructive scenarios isolated and explicitly authorized. Test deletion protection and replacement behavior with plans unless the real delete is the behavior under test.

## Contract and policy coverage

For a shared module, cover the minimum valid configuration, meaningful optional branches, boundary inputs, expected validation failures, stable outputs, and resource-address changes that would affect consumers. Add a regression test for a bug with a stable externally visible contract.

Use the repository's selected policy tool for organization rules such as public access, encryption, allowed regions, or required metadata. Pin rules/tool versions and document focused suppressions with an owner and reason. Policy scans complement Terraform tests; they do not replace plans or provider-backed verification.

## Execute and report

Run the exact repository command from the correct module root. Format test files, initialize as required, then run the narrow test selection before the broader suite. Never run apply tests merely because `terraform test` appears in a generic checklist.

Report:

- CLI and provider versions used;
- which runs used validate, plan, mocks, or real apply;
- whether network, backend, credentials, and remote APIs were involved;
- created resources and confirmed cleanup; and
- skipped checks or gaps caused by missing credentials, network, backend, provider binaries, quotas, or representative inputs.

Do not treat `validate`, a mock, or an unreviewed plan as evidence for a stronger claim than it checks.
