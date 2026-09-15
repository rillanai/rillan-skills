<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.3.0 -->
# Terraform Migration Planning

## Scope
Use this mode for Terraform or OpenTofu version and provider upgrades, backend changes,
resource-address and module refactors, imports, state splits, and migrations between
the two engines. For HCP Terraform workspace, organization, variable-set, VCS, or
`cloud {}` concerns, also load the `hcp-terraform` skill.

Treat state as the ownership record for real infrastructure. A migration is complete
only when configuration, state, provider schemas, backends, and downstream consumers
agree about that ownership.

## Discover the Actual Contract
Before designing commands, inspect:

- the pinned Terraform or OpenTofu version in configuration, CI, wrappers, and version
  managers; do not infer support from the newest documentation;
- `required_version`, provider source and version constraints, `.terraform.lock.hcl`,
  module versions and sources, and the upgrade guides for every crossed major version;
- every root module, backend, CLI workspace or remote state unit, and the credentials
  and locking behavior used by the execution environment;
- `state list` and `state show` output for affected addresses, plus provider import IDs;
- remote-state/output consumers, provider aliases, dependencies, `prevent_destroy`, and
  stateful resources whose replacement could lose data or cause downtime;
- CI, automation, and human sessions that could plan or apply concurrently.

When indirect module or output relationships matter and `graphify-out/` exists, use
graph traversal to investigate them, then confirm each result in code. Otherwise use
structural search and state inspection.

Record current engine/provider versions, state locations and workspaces, resource
counts, serial and lineage where exposed, affected addresses and IDs, expected plan
actions, owners, maintenance window, and rollback boundaries. Never print state or
commit backups: state commonly contains secrets.

## Choose Mechanisms by the Pinned Engine
Prefer configuration-driven changes because they are reviewable and appear in plans.
Use imperative state commands only when the pinned engine lacks the needed construct or
the move cannot be expressed declaratively.

For Terraform, `moved` requires 1.1+, `import` blocks require 1.5+, and `removed`
blocks require 1.7+. A non-destructive Terraform removal must say `destroy = false`:

```hcl
# Alternative A: preserve an object while changing its address.
moved {
  from = aws_instance.old
  to   = module.app.aws_instance.main
}

# Alternative B: stop managing an object without destroying it.
# Do not combine these alternatives for the same old address.
removed {
  from = aws_instance.retired
  lifecycle { destroy = false }
}
```

See HashiCorp's [refactoring](https://developer.hashicorp.com/terraform/language/modules/develop/refactoring),
[import](https://developer.hashicorp.com/terraform/language/import), and
[removed block](https://developer.hashicorp.com/terraform/language/block/removed)
references. Retain published-module `moved` history: removing it can break upgrades
from older module versions.

OpenTofu has its own release history and syntax. Confirm features against documentation
for the pinned OpenTofu version. In particular, [OpenTofu 1.7 introduced
`removed`](https://opentofu.org/docs/v1.7/intro/whats-new/), its initial semantics
differed from Terraform's, and later releases expanded lifecycle behavior. Do not copy
a Terraform `removed` block into OpenTofu without checking the target release's
[resource removal documentation](https://opentofu.org/docs/language/resources/syntax/).
OpenTofu documents `moved` under
[refactoring](https://opentofu.org/docs/language/modules/develop/refactoring/) and
configuration-driven [import](https://opentofu.org/docs/language/import/).

An import must have exactly one intended state address for each remote object, a
matching resource configuration, the provider-documented import identity, and a plan
that says import rather than create. Generated configuration is a starting point;
review it against the provider schema before applying.

`state mv`, `state rm`, and CLI import change state immediately. Enumerate exact
addresses first, use the backend lock, capture the command and result, and re-plan.
Never edit state JSON casually. `state push` is a recovery or exceptional transfer
tool, not a routine migration technique: it can overwrite remote state, and its
lineage and serial checks protect against pushing unrelated or stale snapshots. Do not
use `-force` to bypass those checks. See the Terraform
[state command](https://developer.hashicorp.com/terraform/cli/commands/state) and
[state push](https://developer.hashicorp.com/terraform/cli/commands/state/push)
references.

## Execution Contract
Separate state or address migration from provider behavior changes and infrastructure
changes. Use small batches with an independently reviewable saved plan.

1. Freeze automated and human applies for every affected state and its consumers.
   Confirm no active run holds or should hold the lock. Never disable locking; use
   force-unlock only after proving the recorded holder is stale.
2. Pull a fresh independent backup from every affected state and retain backend-native
   versions. Confirm the backup corresponds to the selected workspace and expected
   address inventory. Test the documented restore process in a safe environment.
3. Prepare configuration and migration declarations. Run formatting and validation,
   then refresh-aware plans with the actual target environment's variables, identity,
   workspace, engine, and provider lock file.
4. Review every action and replacement reason. Address-only moves should normally show
   moves with no remote changes. Imports should show imports, not creates. Non-destructive
   removals should explicitly say the object will remain. Any unexplained create,
   update, destroy, or replacement is a stop condition.
5. Execute only the saved plan or exact platform run whose plan was reviewed. If
   execution has not already been authorized, obtain explicit approval from the
   responsible owner; do not ask again when authorization already covers the step. Do
   not use automatic approval for interactive migration work.
6. Re-plan the changed state and every affected consumer. Verify ownership, outputs,
   resource counts, service health, and that no concurrent write occurred. Preserve
   the command log, plan evidence, and backups until the stability window closes.

A no-op final plan is the usual goal, but it is not proof by itself. It can be clean in
the wrong workspace or with missing configuration. Match backend identity, workspace,
addresses, IDs, and consumers as well.

## Module and Address Refactors
For a rename, module extraction, module split, or `count`/`for_each` conversion, map
every old instance address to exactly one new address. Put `moved` blocks in the module
that owns the relevant objects and inspect plans for callers using each supported
upgrade path. Resource-type changes are valid only when the provider supports migration
between those schemas; otherwise design an import or an explicit replacement and state
its infrastructure impact. Keep required provider configurations available until no
state object refers to them. When changing a module source, compare the resolved module
content and provider requirements, reinitialize, and do not assume matching input and
output names imply identical resource addresses or behavior.

## Version and Provider Upgrades
Read the engine upgrade guide and every provider migration guide/changelog across the
current-to-target range. Check supported engine versions, provider and module
constraints, protocol compatibility, renamed or removed resources, schema type
changes, new required fields, default changes, and provider-specific state upgraders.

Change one compatibility boundary at a time. Update constraints deliberately, run
`init -upgrade`, review lock-file changes, validate, and plan in non-production before
production. A version upgrade is not expected to change remote objects unless an
identified migration requirement explains each action. Keep the previous binary,
configuration commit, and lock file available; state written by a newer engine or
provider may not be readable by the old one, so rollback may require restoring the
pre-upgrade state and reconciling any remote changes made since it.

For Terraform-to-OpenTofu or reverse migrations, consult both engines' guides for the
exact source and target versions. Inventory registry sources, backend support, lock
files, encryption, and `terraform_remote_state` dependencies. OpenTofu documents that
Terraform may not reliably read state after OpenTofu-specific features write it; its
[interdependent configuration guide](https://opentofu.org/docs/intro/migration/multiple-configurations/)
therefore migrates consumers before producers. Avoid engine-specific state features
during a mixed-engine interval and verify consumers after each producer changes.

## Backend Migration
Provision and secure the destination backend and locking mechanism first. Back up all
source workspaces, update the backend configuration, then use `init -migrate-state` to
copy state. Terraform documents that `-migrate-state` can prompt for each workspace;
`-reconfigure` deliberately does not migrate state. Do not replace reviewed prompts
with `-force-copy` casually. See the
[init command](https://developer.hashicorp.com/terraform/cli/commands/init) and
[backend initialization](https://developer.hashicorp.com/terraform/language/backend)
references.

Verify each destination workspace independently: correct backend identity, expected
lineage/address inventory, no unexpected plan actions, and readable outputs for every
consumer. Keep the source immutable and recoverable until verification and the
stability window finish. Update consumers and automation before decommissioning it.
Rollback is not simply changing the backend block after new writes: freeze both sides,
identify the authoritative newest state, and prevent loss of changes made after the
copy.

## Splitting State Without Changing Infrastructure
A split transfers exclusive management; it must never intentionally destroy or create
the transferred objects. Map dependencies first and move cohesive batches in dependency
order. Establish outputs or a narrower data-sharing interface before consumers lose
their old references.

Prefer the documented declarative handoff when both pinned engines support it. For each
batch:

1. Freeze both roots and consumers; back up source and destination.
2. In the source, replace resource declarations with the engine-correct non-destructive
   `removed` form. Its reviewed plan must say the objects will be forgotten and show
   zero destroys. Apply it. The objects are temporarily unmanaged, so keep the gap
   short and make no remote changes.
3. In the destination, add matching resource and `import` blocks. Its reviewed plan
   must show the exact objects imported with zero remote creates, updates, or destroys.
   Apply it.
4. Confirm the source no longer contains those addresses, the destination contains
   them exactly once, and both roots and all consumers have no unexpected actions.

HashiCorp documents this remove-then-import sequence in
[Refactor Terraform state](https://developer.hashicorp.com/terraform/language/state/refactor).
Do not import first and leave two active states believing they own one object. Do not
accept a source plan that says transferred resources “will be destroyed.” If the
destination proposes creation, the import identity or configuration is wrong; stop
before applying.

For older versions, design an explicit CLI transfer using the exact version's official
`state mv`/`state rm`/import behavior. Rehearse it on disposable state copies, preserve
locking and backups, and verify after every batch. Manual pull/edit/push recipes and
forced pushes are too easy to apply to stale, wrong-lineage, or incomplete state and
must not be the default procedure.

## Rollback and Reporting
Define rollback per phase, not as “restore the backup.” State restoration can discard
legitimate writes made after the snapshot and does not reverse remote infrastructure
changes. Record the last safe point, authoritative state, code and lock-file revisions,
who must stop writes, exact restore mechanism, consumer rollback order, and how to
reconcile drift before resuming.

Report actual evidence: versions and pins inspected, state/workspace inventory,
backups captured, plan action counts and replacement reasons, approvals when required,
apply results, post-migration plans, consumer checks, unresolved risks, and the earliest
safe rollback point. Label any unrehearsed path as untested and run it in
non-production before production.
