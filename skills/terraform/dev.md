<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.3.0 -->
# Terraform Development

Use this mode to author or review Terraform or OpenTofu configuration. Follow explicit repository conventions first. Keep changes small enough that a reviewer can connect each configuration change to its plan effect.

## Discover the execution contract

Before editing, identify the root module and its callers. Read nearby HCL, module sources, CI workflows, wrappers, version managers, lock files, and contributor instructions. Determine:

- whether the repository runs `terraform`, `tofu`, or a wrapper;
- the required CLI and provider versions from `required_version`, `required_providers`, `.terraform.lock.hcl`, CI, and version files;
- the backend, workspace/state boundary, working directory, variable sources, and provider aliases;
- all callers and output consumers of a shared module; and
- the repository's formatting, validation, planning, security, and documentation commands.

Do not introduce syntax or test features newer than the discovered minimum version. Do not silently change providers, credentials, backend settings, state addresses, or lock selections.

## Design for predictable plans

- Keep root modules focused on provider/backend configuration, environment values, and composition. Extract a child module when it represents a stable reusable concern; do not abstract a one-off block merely to remove repetition.
- Use typed variables with descriptions and validation for constraints known before provider calls. Prefer explicit object shapes over `map(any)`.
- Expose only outputs with known consumers. Mark sensitive outputs, but remember that state still contains their values.
- Put repeated naming, tags, and derived values in clear locals. Avoid nested conditionals and `try`/`can` chains that conceal malformed input.
- Prefer `for_each` with stable business keys when instances have identities. Changing `count` order or `for_each` keys changes resource addresses and can cause replacements.
- Use explicit dependencies only when Terraform cannot infer the relationship from references. Broad `depends_on` can make values unknown during planning and obscures the real dependency.
- Use data sources for authoritative external facts. Avoid using them as hidden configuration when an input would make the dependency clearer and testable.
- Avoid provisioners. If no provider resource or external workflow can perform the action, document retry behavior, idempotency, credentials, and failure recovery.

## Providers, modules, and dependencies

Every module declares each external provider source it uses; Terraform's built-in provider is implicit. Match the repository's constraint policy: reusable modules usually declare compatible minimums, while root modules and the committed dependency lock file select tested versions. Configure credentials and environment-specific provider settings in root modules; pass aliases explicitly to child modules through their declared provider configuration aliases.

Pin registry module versions and immutable git references according to repository policy. Treat a module source or provider upgrade as a dependency change: inspect release notes, update configuration deliberately, refresh the lock file only with the repository's command, and review the resulting plan. Never hand-edit `.terraform.lock.hcl`.

Commit `.terraform.lock.hcl` for root modules unless the repository documents another policy. Do not commit `.terraform/`, plan files, state, credentials, or generated secret-bearing artifacts.

References: [provider requirements](https://developer.hashicorp.com/terraform/language/providers/requirements), [dependency lock file](https://developer.hashicorp.com/terraform/language/files/dependency-lock).

## State and lifecycle safety

Treat state addresses and backend configuration as durable interfaces.

- Preserve the repository's separation between environments and operational boundaries. A state unit should have a comprehensible blast radius, locking where supported, restricted access, encryption, and recovery/versioning appropriate to its importance.
- Never hardcode credentials or expose them through committed HCL, variable defaults, tfvars, outputs, provisioner commands, or persisted backend arguments. When an interface must carry a secret, mark it sensitive, minimize its propagation, and use the established secret or workload-identity mechanism.
- Use `moved` blocks for reviewed address changes when the supported CLI version permits them. For imports, removals, backend migrations, or manual state commands, provide exact addresses, backups/recovery steps, and a post-change plan; never improvise against production state.
- Add `prevent_destroy`, `create_before_destroy`, or `ignore_changes` only for a specific lifecycle requirement. `ignore_changes` creates an ownership boundary: name the external owner and explain how drift is detected. Do not use lifecycle settings to suppress an unexplained plan.
- Check replacement triggers, uniqueness constraints, quotas, ordering, and downtime before accepting a replace action. Separate high-risk state or lifecycle work from unrelated refactors.

## Security and operational behavior

Grant the provider identity and provisioned identities the narrowest practical permissions. Keep public exposure, firewall rules, encryption, logging, deletion protection, backups, and retention explicit when relevant. Confirm provider defaults rather than assuming they are safe or stable.

Avoid leaking secrets through outputs, command arguments, logs, generated plans, or error messages. A `sensitive` mark redacts normal CLI display but does not remove data from state. Review plans and machine-readable plan JSON as sensitive artifacts.

Make failure behavior visible. Validate incompatible options early; use resource preconditions/postconditions for configuration invariants at the stage their values become known. Do not encode policy in comments alone.

## Verify the change

Use repository commands and the discovered CLI. A proportionate local sequence is:

1. Format only the affected scope and inspect the formatting diff.
2. Initialize the actual root module using the repository's normal setup. `init -backend=false` avoids backend initialization, but provider/module installation may still need network access.
3. Run `validate`. It checks syntax and internal consistency; it does not validate remote state or provider APIs.
4. Run focused tests required by the repository.
5. Run `plan` with the representative workspace and variable inputs. A plan includes validation and evaluates configuration with state and provider behavior; it may refresh remote objects and require credentials/network access.

Read the whole plan. Account for every create, update, replacement, and destroy; inspect sensitive or unknown values, provider/version changes, and output changes. Save a plan only when the established workflow protects the artifact and applies that exact reviewed plan.

If credentials, network access, providers, backend access, or representative variables are unavailable, run the safe checks that remain and report precisely what was not verified. Never describe `validate`, a speculative plan, or a mocked test as proof that infrastructure will apply successfully.

References: [`terraform validate`](https://developer.hashicorp.com/terraform/cli/commands/validate), [`terraform plan`](https://developer.hashicorp.com/terraform/cli/commands/plan).
