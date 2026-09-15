<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.3.0 -->
# Terraform Audit Deep Dive

Use this contract only for an explicitly requested deep, phased audit of Terraform
or OpenTofu. The task must supply the repository path or scope and one exact phase
below; ask when either is missing. It may also define focus, exclusions, depth, and
treatment of generated, vendored, or external modules. Add
`hcp-terraform/audit.md` only when HCP Terraform concerns are present.

## Audit contract

- Execute only the requested phase. Stop at its boundary and wait for the user;
  never auto-advance or combine phases.
- Inspect every accessible file relevant to the phase, including HCL, variable
  files, lock and backend configuration, CI, scripts, tests, and documentation.
  Treat third-party modules as dependency surface unless their source is locally
  accessible or the user requests exhaustive review.
- Anchor factual claims to tool-observed evidence using a path and, where applicable,
  a block address. State whether evidence is configured in source, proposed by a
  plan, recorded in state, or observed at runtime; none proves a stronger category
  by itself. Documentation proves stated intent, not deployed truth.
- Label a conclusion that is not fully proved `INFERENCE` and explain its basis.
  Record excluded, inaccessible, generated, external, or unreviewed material under
  `UNREVIEWED/INACCESSIBLE` with its effect on confidence. Missing plan, state,
  credentials, or tools is a limitation, not permission to guess.
- An existing `graphify-out/` graph may seed discovery. Treat inferred or ambiguous
  edges only as leads and verify them against repository evidence. Do not require or
  rebuild a graph.
- Build inventories before evaluation. Use the smallest set of evidence anchors that
  proves a claim; cite both producer and consumer for cross-state dependencies.
- Output Markdown and fence machine-readable artifacts as `csv` or `json`. Split
  outputs over 500 rows into stable `.partNN` names. A chunk is not a phase pause:
  its `NEXT` names the exact remaining part or step in the current phase.

A hard failure is the sole exception to all footer requirements. Emit only:

```text
ERROR: <short reason>
BLOCKED_BY: <missing input, evidence, tool, or access>
```

Every successful Phase 1–4 response, and every unfinished Phase 5 chunk, ends with:

```text
STATE_SNAPSHOT: (max 8 bullets)
- <coverage completed, evidence limits, and artifact parts>

NEXT: <exact next phase or remaining artifact part>
```

A completed Phase 5 instead ends with `FINAL STATE_SNAPSHOT` (at most eight
bullets) and `DONE`; it has no `NEXT`.

## Phase gates

- Phase 1 inventories and describes; it does not recommend or grade.
- Phase 2 accounts and indexes; it does not evaluate, recommend, or grade.
- Phase 3 assesses architecture, dependencies, state boundaries, and blast radius;
  it defers detailed remediation.
- Phase 4 reports evidence-backed, prioritized findings and fixes; it does not grade.
- Phase 5 synthesizes, grades, prioritizes, and plans.

## PHASE 1 - Inventory + Entrypoints

Produce a concise, directory-grouped inventory:

- every file, with one-line purpose; every directory, with purpose and
  `Critical`, `Important`, `Supporting`, `Legacy`, or `Generated`;
- root and child modules; provider requirements, versions, aliases, and cloud
  distribution; public/private registry, local, and Git module sources and pins;
- every backend and state/isolation unit, its mapped directory, identifier, locking
  and encryption evidence, and environment mapping. Classify local, S3, GCS,
  AzureRM, Consul, HCP/cloud, Kubernetes, HTTP, or other backends without assuming
  that an omitted setting means the deployed backend lacks the control;
- every `data` block, especially `terraform_remote_state` and `tfe_outputs`,
  with source and consumer;
- CI plan/apply entrypoints, approval gates, runner triggers and drift checks;
- variable definitions, `.tfvars`, evidenced precedence, outputs, and operational
  scripts.

Report totals for files, `.tf`, `.tfvars`, roots, child modules, state units,
cross-state references, resources and data sources by provider/cloud, plus
`UNREVIEWED/INACCESSIBLE`. For HCP Terraform, identify workspaces here and leave
deeper TFC accounting to the overlay.

Treat a directory as a root only when backend configuration, invocation wiring, or
other inspected evidence shows it is applied directly. Treat module-source code that
is not accessible as external, record the source and version constraint, and do not
invent its internal resource count.

## PHASE 2 - Resource Accounting

Account for every accessible `resource`, `data`, and `module` block, plus locals
and provider configurations in the coverage summary. Produce exactly two artifact
families:

`resource_index.csv`

```csv
resource_type,logical_name,module_path,provider,count_or_for_each,lifecycle_rules,dependencies_top
```

- One row per resource or data block. Prefix data types with `data.`.
- Set `module_path` to `root` for a root-module block or its full module address
  (for example, `module.networking.module.dns`) for a child-module block.
- Preserve full module paths and exact dynamic `count`/`for_each` expressions;
  record neither when absent. Note list-based `count` for later evaluation.
- Record active `create_before_destroy`, `prevent_destroy`, `ignore_changes`, and
  `replace_triggered_by`. In `dependencies_top`, record resolved resource or data
  addresses only when the evidence supports them; do not mislabel `local`, `var`,
  or `path` expressions as resource dependencies. Leave unknown dependencies blank
  and document the limitation.

`module_index.csv`

```csv
module,path,source,source_type,version,inputs,outputs,resources_managed,purpose
```

- One row per module call. `module` is the full call address; `path` is the
  resolved source filesystem path when accessible, otherwise blank; and `source`
  is the literal `source` argument. Classify `source_type` as `public_registry`,
  `private_registry`, `local`, or `git`; `version` is the literal constraint
  and is blank for local modules.
- `inputs`, `outputs`, and `resources_managed` are numeric counts.
  `resources_managed` counts resource plus data blocks in the resolved module.
  Leave any unknown count blank and document the limitation.

Use `resource_index.part01.csv` and `module_index.part01.csv` naming when chunked.
Add a provider/cloud count summary without creating another artifact family.

## PHASE 3 - Architecture + State Boundaries

Using Phases 1–2 evidence:

- describe implemented infrastructure and each cloud/on-prem role without assuming
  primacy;
- map root modules, child-module direction and nesting, reuse, registry/local
  sourcing, providers and aliases, account/subscription/project patterns, and
  authentication strategy;
- map each state unit to environments and clusters; assess naming, locking,
  encryption, backend consistency, isolation quality, and maximum single-apply
  blast radius;
- trace every `terraform_remote_state` and `tfe_outputs` producer/consumer edge,
  cycles, fan-out, and hardcoded coupling;
- identify oversized or fragmented state units, checked-in state, legacy syntax or
  deprecated resources, and the evidenced environment promotion/versioning flow;
- trace values across variables, locals, module inputs/outputs, data sources, and
  resources where that flow affects ownership, coupling, or blast radius.

State the mechanism behind each assessment, but defer implementation plans.

## PHASE 4 - Security + Compliance Audit

Review each applicable provider and resource class rather than assuming one cloud:

- identity and authorization: role/binding scope, custom roles, cross-account or
  cross-project trust, workload identity/OIDC/managed identity, service-account
  keys, shared identities, local admin access, and backend/state/run permissions;
- encryption and secrets: storage/database encryption, TLS enforcement, KMS/key
  vault access and rotation, hardcoded values, sensitive variables/outputs, secret
  managers, runner secrets, and state exposure;
- network and public exposure: broad ingress, firewall/NSG associations, public IPs
  and endpoints, private clusters/endpoints, segmentation, peering, DNS, WAF, and
  DDoS controls where applicable;
- compliance and operations: provider audit logs, diagnostic settings, security
  services, organization policy, centralized alerts, drift detection, state refresh,
  and policy-as-code;
- supply chain and governance: provider locks, module source trust and pinning,
  third-party provenance, required project-defined tags/labels, and narrowly scoped
  lifecycle ignores rather than blanket tag ignores.

Apply provider-specific checks where the resources exist: AWS IAM, KMS, security
groups, public endpoints, CloudTrail, Config, and organization controls; Azure RBAC,
managed identity, Key Vault, NSGs, Private Link, Activity Log, diagnostic settings,
Defender, and Policy; GCP IAM, service accounts, KMS, firewall rules, Private Service
Connect, Audit Logs, organization policy, and Security Command Center; and vSphere
roles, network policy, and audit logging. Absence from this list does not exclude
another provider's equivalent controls.

Use applicable static analysis and plan/state evidence when available; cite tool rule
IDs and locations. Do not assert absence of a cloud-managed control from source
alone. Recommend a scanner or control only after showing the uncovered mechanism.

Group findings by `P0`, `P1`, and `P2`. Every finding includes file path, resource
address, evidence, impact, and concrete fix. Mark missing operational visibility
under `UNREVIEWED/INACCESSIBLE`; do not turn it into a confirmed vulnerability.

## PHASE 5 - Synthesis

Produce an overall `A`–`F` grade and anchored subgrades for:

- Clean Code: formatting, naming, duplication, variables, and collection identity;
- Architecture: composition, separation, and directory/state alignment;
- Security: identity, encryption, network, secrets, exposure, and isolation;
- Compliance: auditability, monitoring, and policy enforcement;
- State Management: boundaries, coupling, locking, encryption, and blast radius;
- Modularity: interfaces, reuse, sources, versions, and nesting;
- CI/CD: plan review, apply approval, drift detection, tools, and runners;
- Docs/DX: accurate module and operational documentation.

Clean Code specifically includes Terraform identifier consistency, stable
`for_each` keys versus index-sensitive `count`, variable hygiene, and justified
lifecycle exceptions. Docs/DX distinguishes accurate hand-maintained intent and
runbooks from generated input/output reference material.

The HCP overlay alone adds `TFC Integration`. For each recommendation, cite prior
evidence, assign `P0`/`P1`/`P2`, effort `S`/`M`/`L`, and risk
`Low`/`Med`/`High`; give implementation, validation, rollout, and state-migration
steps. Sequence the roadmap by dependencies and risk rather than invented dates,
covering immediate stabilization, structural state/module work, then ongoing
hardening and drift controls. Include code cleanup, module isolation, future
enhancements, and the exact tests or commands that would validate each change.

## Recommendation quality

Tie recommendations to repository evidence and explain the failure mechanism and
expected effect. For consequential architectural or state changes, compare viable
alternatives, tradeoffs, reversibility, and the cost of no change; use a fuller
Socratic decision process only when ambiguity warrants it or the user requests it.
The host controls the output format. Remove generic advice that could be pasted into
an unrelated repository.
