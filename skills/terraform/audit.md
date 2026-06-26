<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.0.0 -->
# Terraform Audit Deep Dive

## Purpose
Use this skill to run a phased, evidence-based deep-dive audit of a Terraform codebase. This skill is for enterprise-grade infrastructure-as-code audits that require broad codebase coverage, resource accounting, architecture analysis, security review, compliance review, and implementation-grade recommendations.

This skill defines the stable audit contract. Per-run inputs such as the repository path, the requested phase, and any special focus areas must be supplied in the task prompt.

## Skill Use
- Load this skill only when the user explicitly wants a deep Terraform repository audit or a clearly similar phased infrastructure-as-code review.
- Treat this skill as the governing audit contract for the session or turn.
- Keep repository-specific instructions in the task prompt, not in this file.
- When this skill conflicts with generic review behavior, follow this skill.

## Knowledge-Graph Discovery (When Available)
If a graphify knowledge graph (`graphify-out/`) is present, seed inventory and architecture mapping from `graphify-out/GRAPH_REPORT.md` and `graphify query`/`graphify path` instead of starting cold — then verify every graph-derived claim against repository evidence before recording it. `INFERRED` and `AMBIGUOUS` edges are leads, not findings; an unverified graph edge is `INFERENCE`, not evidence.

## Tool Use
This skill is tool-agnostic and works with Claude Code, Codex, OpenCode, and similar assistants. Map its guidance to whatever file-reading, editing, search, and shell-execution tools your environment exposes.

- Every factual claim in an audit must come from a tool invocation, not inference. Read the file, search for the resource, or run the command before writing the finding.
- Issue independent tool calls (directory listings, resource scans, provider pins, backend config reads) in parallel.
- When plan or state evidence is needed, inspect `terraform plan` output or state directly rather than describing expected behavior.
- If evidence cannot be gathered (no plan access, missing state visibility), record it under `UNREVIEWED/INACCESSIBLE` rather than guessing.

## When To Use
Use this skill when the user asks for any of the following:
- A deep audit of a Terraform repository
- A phased architecture review of infrastructure-as-code
- Full resource or module accounting for Terraform code
- A security or compliance audit as part of a broader infrastructure review
- Grading, refactor planning, or modernization guidance grounded in code evidence

Do not use this skill for:
- Small targeted code reviews on one change set
- Narrow bug hunts or single-resource debugging
- Pure security-only reviews when a dedicated security skill is more appropriate
- Repositories that are not primarily Terraform unless the user explicitly wants the Terraform-centric audit protocol

## Required Inputs
The invoking prompt must provide:
- Repository path or scope to audit
- Exact phase to execute

The invoking prompt may also provide:
- Areas to emphasize
- Areas to deprioritize
- Time or depth constraints
- Whether generated, vendored, or third-party module code should be summarized instead of exhaustively expanded

If the requested scope or phase is missing, stop and ask for it.

## Recommended Per-Run Inputs
Provide these when available:
- Repository path
- Requested phase name exactly as written in this skill
- Whether vendored and generated code should be fully reviewed or summarized
- Any explicit focus areas such as state boundaries, RBAC policies, network security, module design, CI/CD integration, compliance requirements, multi-cloud patterns, or TFC workspace organization
- Any explicit exclusions

## Operating Stance
- Act as a Principal Infrastructure Engineer performing an enterprise-grade audit.
- Prefer evidence over intuition.
- Describe the infrastructure as implemented, not as intended.
- Read broadly across Terraform code, variable files, backend configurations, CI/CD pipelines, scripts, and documentation within the accessible repository scope.
- Stay phase-disciplined. Do not leak later-phase conclusions into earlier phases.
- Optimize for defensible conclusions, not stylistic commentary.

## Hard Requirements
- Inspect every accessible file relevant to infrastructure behavior within the requested scope: `.tf` files, `.tfvars` files, variable definitions, output definitions, backend configs, CI/CD definitions, scripts, and docs.
- Account for every accessible resource and data source when executing Phase 2, including all resource types, data sources, module calls, locals, and provider configurations.
- Do not produce recommendations, grades, or refactor plans before the dedicated findings and synthesis phases.
- Never attempt the full audit in one response.
- Work only on the requested phase and stop at the phase boundary.

## Evidence And Anti-Hallucination Rules
- Every factual claim must be anchored to a file path and, when applicable, a resource address or block name.
- When a claim is not fully provable from accessible source, label it `INFERENCE` and explain why.
- When files or directories are inaccessible, excluded, generated, or otherwise not fully reviewed, list them under `UNREVIEWED/INACCESSIBLE` with an impact note.
- Do not imply certainty about runtime behavior unless it is directly supported by code, config, CI definitions, scripts, or docs.
- Distinguish implemented behavior from inferred intent.

## Evidence Sufficiency Standard
- A claim is `CONFIRMED` only when supported by directly inspected repository material.
- Acceptable support includes `.tf` source files, `.tfvars` files, backend configurations, CI definitions, scripts, and docs located in the audited scope.
- A single weak anchor is not enough for broad architectural claims. Use multiple anchors when behavior spans modules or state boundaries.
- If a claim depends on naming convention, incomplete module graph inference, or missing environment context, mark it `INFERENCE`.
- If a requested conclusion cannot be supported to this standard, state the limitation instead of filling the gap with assumptions.

## Output Contract
- Output only Markdown.
- All machine-readable artifacts must be in fenced code blocks labeled `csv` or `json`.
- If a hard requirement cannot be met, output exactly:

```text
ERROR: <short reason>
BLOCKED_BY: <what is missing>
```

- Do not output anything else in that failure mode.

## Chunking Rules
- Never attempt the entire audit in one response.
- Complete only the requested phase.
- Stop at the end of the requested phase and wait for continuation.
- When a phase is too large for one response, emit the current chunk, preserve numbering or artifact part names, and keep the footer pointing to the exact remaining next step.
- End every response with this exact footer shape:

```text
STATE_SNAPSHOT: (max 8 bullets)
- <bullet>

NEXT: <exact next phase name>
```

- Never omit the footer.

## General Audit Method
1. Establish accessible scope and identify obvious exclusions.
2. Build a fast inventory of directories, files, modules, providers, backends, and operational assets.
3. Read the relevant files for the current phase before producing conclusions.
4. Build evidence tables or inventories before making evaluative claims.
5. Keep analysis anchored to files and resource addresses.
6. Preserve phase boundaries strictly.

## Preferred Execution Method
- Prefer fast repository discovery and search tools.
- Start with file and module inventory before deep reading.
- Read enough surrounding context to avoid resource-level misinterpretation.
- For large repositories, review by environment or module subsystem and keep notes on what has and has not been covered.
- Use best-effort static analysis for module references, resource dependencies, cross-state references, and data source usage, but never overstate precision.
- Treat CI/CD pipelines, scripts, backend configs, and documentation as first-class evidence for intended and operational behavior.

## Phase Gate Rules
- Phase 1 may inventory and describe, but must not recommend.
- Phase 2 may account and index, but must not evaluate architecture quality or recommend changes.
- Phase 3 may assess boundaries, state isolation, and module dependencies, but must not produce detailed remediation plans.
- Phase 4 may produce prioritized findings with concrete fixes, but must not assign overall grades.
- Phase 5 may synthesize, grade, prioritize, and plan.
- If the user asks for a later-phase artifact early, either refuse that part or mark it as deferred to the proper phase.

## Phase Rules

### PHASE 1 - Inventory + Entrypoints
Produce:
- Repository inventory grouped by directory
- One-line purpose and tag for each directory: `Critical`, `Important`, `Supporting`, `Legacy`, or `Generated`
- One-line purpose for each file
- All identified root modules (directories containing backend configuration or intended to be applied directly)
- All identified child modules (directories used as `module` sources)
- TFC workspace inventory: workspace name, mapped directory, organization, and any workspace tags
- Cross-workspace reference inventory: all `data "tfe_outputs"` blocks with source workspace and consuming module
- Provider inventory: all providers, their versions, their aliases, and cloud distribution (Azure, GCP, vSphere, other)
- Backend configuration summary: backend type (TFC cloud block, Azure Blob Storage legacy, or other), workspace name, organization, locking mechanism, encryption status per root module
- Environment structure: directory + workspace hybrid pattern with evidence showing how directories map to TFC workspaces
- Data source inventory: all `data` blocks with their type and purpose, with special attention to `data "tfe_outputs"` and cross-workspace references
- CI/CD integration points: pipeline definitions, plan/apply steps, approval gates, drift detection, TFC run triggers
- Variable file inventory: all `.tfvars` files, `variables.tf` files, and variable precedence
- Module source inventory: TFC private registry modules (with `app.terraform.io` source addresses), local modules, git-sourced modules, and public registry modules
- Totals block with:
  - Total files
  - Total `.tf` files
  - Total `.tfvars` files
  - Total root modules
  - Total child modules
  - Total TFC workspaces identified
  - Total cross-workspace references (`data "tfe_outputs"` count)
  - Total resources by provider (count), broken down by cloud (Azure, GCP, vSphere, other)
  - Total data sources by provider (count)
  - `UNREVIEWED/INACCESSIBLE` list with impact notes

Constraints:
- Do not include architecture recommendations.
- Do not include grades or refactor suggestions.
- Keep descriptions concise enough to remain scannable for large repositories.

### PHASE 2 - Resource Accounting
Produce exactly two artifact families:
- `resource_index.csv`
- `module_index.csv`

`resource_index.csv` rules:
- One row per resource or data source block
- Columns:
  - `resource_type` — The Terraform resource type (e.g., `azurerm_resource_group`, `google_compute_instance`, `vsphere_virtual_machine`, `data.tfe_outputs`)
  - `logical_name` — The logical name in the Terraform configuration (e.g., `main`, `primary`, `environment`)
  - `module_path` — Full module path (e.g., `root`, `module.networking`, `module.compute.module.cluster`)
  - `provider` — The provider managing this resource (e.g., `azurerm`, `google`, `vsphere`, `tfe`, `kubernetes`). Note the cloud distribution explicitly.
  - `count_or_for_each` — Whether the resource uses `count`, `for_each`, or neither; include the expression if dynamic. Flag any use of `count` with lists (should be `for_each`).
  - `lifecycle_rules` — Active lifecycle rules: `create_before_destroy`, `prevent_destroy`, `ignore_changes`, `replace_triggered_by`; blank if none. Note `ignore_changes = [tags]` for dynamic tag patterns.
  - `dependencies_top` — Top explicit and notable implicit dependencies (resource addresses), including cross-workspace references via `data.tfe_outputs`

`module_index.csv` rules:
- One row per module call
- Columns:
  - `module` — The module call name as it appears in the configuration (e.g., `module.networking`)
  - `path` — File system path to the module source
  - `source` — The `source` argument value (local path, TFC private registry address `app.terraform.io/...`, or git URL)
  - `source_type` — Classification: `tfc_registry`, `local`, `git`, `public_registry`
  - `version` — The `version` constraint if specified; blank for local modules
  - `inputs` — Count of input variables passed to the module call
  - `outputs` — Count of outputs defined by the module
  - `resources_managed` — Count of resources and data sources defined in the module
  - `purpose` — One-line description of the module's purpose

Additional constraints:
- If dependencies cannot be computed precisely, leave the field blank and note `INFERENCE`.
- Chunk large outputs to 500 rows maximum per file part.
- Use filenames like `resource_index.part01.csv`.
- Do not include architecture analysis, recommendations, grades, or refactor advice in this phase.
- Include both `resource` and `data` blocks in `resource_index.csv`, prefixing data source types with `data.` for clarity.
- Record `count` and `for_each` expressions accurately; do not simplify dynamic expressions.
- Note multi-cloud provider distribution in a summary table: count of resources per provider (azurerm, google, vsphere, tfe, kubernetes, other).

### PHASE 3 - Architecture + State Boundaries
Using evidence from Phase 1 and Phase 2:
- Describe the infrastructure architecture as implemented, including multi-cloud distribution (Azure primary, GCP secondary, vSphere on-prem)
- Map TFC workspace boundaries: how many workspaces, what is in each, isolation level per environment and per cluster, workspace naming convention consistency
- Analyze cross-workspace coupling: all `data "tfe_outputs"` references, direction of dependency flow, which workspaces are producers vs. consumers, whether the dependency graph has cycles or excessive fan-out
- Assess workspace boundary quality: are boundaries drawn at the right granularity? Are there workspaces that are too large (high blast radius) or too small (excessive cross-workspace coupling)?
- Analyze module dependency graph: which root modules call which child modules, depth of nesting (max 2 levels expected), reuse patterns, TFC private registry vs. local sourcing
- Review provider configuration: aliases, multi-subscription (Azure) patterns, multi-project (GCP) patterns, vSphere integration, authentication strategy (workload identity, managed identity, service accounts)
- Identify environment promotion flow: how changes move from dev to staging to production, whether module versions are pinned differently per environment via TFC registry versions
- Assess blast radius: what is the maximum damage a single `terraform apply` can cause in each root module/workspace
- Map cross-workspace references: `data "tfe_outputs"` usage patterns, any legacy `terraform_remote_state` usage, hardcoded references between state boundaries
- Identify state management risks: overly broad workspaces, missing locking (should not occur with TFC), any legacy Azure Blob Storage backends not yet migrated, state files in version control
- Assess backend configuration consistency across environments: TFC cloud block consistency, organization and workspace naming patterns
- Identify any legacy patterns: Azure Blob Storage backends, `terraform_remote_state` usage that should be migrated to `data "tfe_outputs"`, pre-1.x syntax

Constraints:
- Anchor every claim to file paths and resource addresses.
- Keep detailed refactor plans deferred to Phase 5.
- Focus on actual module dependency direction, workspace boundaries, cross-workspace coupling, and blast radius, not desired architecture labels.

### PHASE 4 - Security + Compliance Audit
Security review areas:

**Identity and Access Management (Multi-Cloud)**:
- Azure RBAC: role assignments, custom role definitions, overly permissive roles, condition-based access, subscription-level vs. resource-group-level scoping
- GCP IAM: role bindings, custom roles, overly permissive roles, condition-based bindings, project-level vs. resource-level scoping
- vSphere permissions: role assignments, privilege escalation paths
- Workload identity federation: Azure managed identities (user-assigned vs. system-assigned), GCP workload identity, service account key usage (should be zero)
- Per-service identity isolation: whether each service has its own identity or shares a common one
- Local accounts: any resources with local admin accounts enabled when RBAC is available (e.g., AKS `local_account_disabled`, database admin accounts)
- TFC workspace access: team permissions, run permissions, state access controls
- Cross-account/cross-project trust: Azure service principals in foreign tenants, GCP cross-project IAM bindings

**Encryption**:
- At-rest encryption on storage: Azure Storage (encryption scope, customer-managed keys via Key Vault), GCP Cloud Storage (CMEK), managed disks, databases
- In-transit encryption: TLS configuration, HTTPS enforcement, minimum TLS version settings
- Key management: Azure Key Vault usage, GCP KMS usage, key rotation policies, access policies on key vaults
- Database encryption: Azure SQL TDE, Cloud SQL encryption, PostgreSQL/MySQL SSL enforcement

**Network Security**:
- NSG rules per subnet (Azure): overly permissive rules, `0.0.0.0/0` ingress, missing NSG associations
- GCP firewall rules: overly permissive rules, `0.0.0.0/0` source ranges, missing network tags
- vSphere network security: distributed firewall rules, port group security policies
- Public IP exposure: resources with public IPs, public load balancers, public endpoints
- Private clusters: whether Kubernetes clusters have private API endpoints, private node pools
- VNet/VPC design: subnet segmentation, peering configuration, service endpoints, private endpoints/Private Link (Azure), Private Service Connect (GCP)
- DNS security: private DNS zones, split-horizon DNS

**Secrets Management**:
- Hardcoded secrets in `.tf` or `.tfvars` files
- `sensitive` marking on variables and outputs
- Key Vault / Secret Manager usage for runtime secrets
- TFC workspace variable sensitivity settings
- State file exposure: sensitive values in outputs without `sensitive = true`

**Public Exposure**:
- Publicly accessible resources: storage accounts with public access, databases with public endpoints, VMs with public IPs, load balancers with public frontends
- WAF or DDoS protection presence where applicable
- Private endpoint usage for PaaS services

**Compliance and Monitoring**:
- Azure: Activity Log configuration, Diagnostic Settings on key resources, Microsoft Defender for Cloud enablement, Azure Policy assignments
- GCP: Cloud Audit Logs configuration, organization policy constraints, Security Command Center enablement
- vSphere: audit logging configuration
- Cross-cloud: centralized log aggregation strategy, alerting on security-relevant events
- Drift detection posture: whether regular plan runs are scheduled (TFC scheduled runs), alerting on unexpected changes, state refresh strategy
- Policy-as-code: Sentinel policies in TFC, Azure Policy, GCP organization policies. Note that `tfsec`/`checkov`/OPA are not currently in use but should be recommended.

**Supply Chain Security**:
- Module sources: TFC private registry (preferred), local, git, public registry. Flag any modules sourced from untrusted public registries.
- Provider version pinning and `.terraform.lock.hcl` presence and commitment
- Third-party module trust assessment

**Tagging and Resource Governance**:
- `standard_tags` compliance: whether all taggable resources have the standard `standard_tags` applied
- Required tag keys present: `ApplicationName`, `ProductName`, `Team`, `Contact`, `Environment`
- Dynamic tags (`CreationDate`, `Expiration`) with `lifecycle { ignore_changes = [tags] }` pattern
- GCP label compliance (equivalent of tags)

Output requirements:
- Findings must be grouped by `P0`, `P1`, and `P2`
- Every finding must include:
  - File path
  - Resource address
  - Evidence
  - Concrete fix
- Prefer findings that are actionable within the existing architecture unless the risk clearly requires structural change.

Constraints:
- No finding without evidence.
- Do not assign overall project grades in this phase.

### PHASE 5 - Synthesis
Produce:
- Overall letter grade `A-F`
- Subgrades for:
  - Clean Code — formatting, naming consistency (kebab-case resources, snake_case Terraform names), DRY, variable hygiene, `for_each` over `count`
  - Architecture — module design, composition, separation of concerns, directory + workspace hybrid consistency
  - Security — RBAC, workload identity, encryption, network, secrets, public exposure, per-service identities
  - Compliance — audit logging coverage (Azure Activity Log, GCP Cloud Audit Logs), monitoring, policy-as-code readiness
  - State Management — TFC workspace isolation, workspace boundaries, cross-workspace coupling via `data "tfe_outputs"`, blast radius, legacy backend migration status
  - Modularity — module boundaries, TFC private registry usage, reuse, versioning, surface area, nesting depth
  - TFC Integration — workspace naming consistency, cross-workspace reference patterns, run triggers, Sentinel policy usage, workspace-to-directory mapping quality
  - CI/CD — plan review, apply approval, drift detection, tooling, TFC integration
  - Docs/DX — module README files (hand-maintained), variable descriptions, output descriptions, onboarding clarity
- Justification for each grade with anchored evidence
- Prioritized refactor recommendations using `P0`, `P1`, and `P2`
- Effort estimate for each item: `S`, `M`, or `L`
- Risk estimate for each item: `Low`, `Med`, or `High`
- Step-by-step implementation plan per item
- Tests or validation steps required
- Rollout notes including state migration considerations and TFC workspace changes
- Code cleanup recommendations
- Module isolation and design recommendations
- Future enhancements and ideas
- 90-day roadmap:
  - Weeks 1-2: stabilize — fix P0 security issues (RBAC enforcement, workload identity gaps, public exposure), ensure all TFC workspaces have correct access controls, pin provider versions, add `terraform fmt` and `terraform validate` to CI
  - Weeks 3-6: restructure — decompose any oversized workspaces, establish module versioning in TFC private registry, implement proper workspace boundaries, migrate any legacy Azure Blob Storage backends to TFC, add `tflint` to CI, consider adding `tfsec`/`trivy`
  - Weeks 7-12: harden — implement drift detection via TFC scheduled runs, consider adding compliance scanning with `checkov`, establish environment promotion flow via TFC registry module versions, document all modules with hand-maintained READMEs, implement TFC team-based RBAC for workspace access, audit and clean up cross-workspace `data "tfe_outputs"` coupling

Constraints:
- End Phase 5 with `FINAL STATE_SNAPSHOT` and `DONE`.
- Tie every major recommendation to evidence established in earlier phases.

## Evidence Formatting Guidance
- Prefer compact evidence references inline, for example: `environments/prod/environment/main.tf::module.networking`, `modules/compute/main.tf::azurerm_kubernetes_cluster.main`.
- When citing module-level behavior without one clear resource, cite the most relevant file path and state that the claim is module-level.
- When behavior is distributed across several files, cite all primary anchors rather than collapsing to one.
- For multi-file claims, list the smallest set of anchors that fully supports the conclusion.
- For cross-workspace references, cite both the producing workspace/output and the consuming `data "tfe_outputs"` block.

## Handling Generated, Vendor, And External Material
- Generated code such as `.terraform.lock.hcl` may be summarized if it is clearly machine-generated, but its existence and impact on reproducibility must still be recorded.
- Vendored or third-party modules should usually be treated as dependency surface, not first-party architecture, unless the repository modifies or relies heavily on them operationally.
- TFC private registry modules referenced via `source = "app.terraform.io/..."` should be noted for version pinning, registry organization, and trust assessment. Their internal implementation is in scope only if the module source code is accessible in the audited repository.
- Public registry modules referenced via `source = "registry.terraform.io/..."` or short-form `source = "org/module/provider"` should be noted for version pinning and trust assessment, but their internal implementation is out of scope unless explicitly requested.
- If the user explicitly requires exhaustive treatment of vendored or third-party module code, follow that request.

## What Good Output Looks Like
- Dense, evidence-backed, and phase-correct
- Explicit about uncertainty
- Operationally useful to an engineering lead or infrastructure team
- Free of filler, motivational language, and generic best-practice padding
- Readable in chunks for large repositories
- Multi-cloud aware: findings and accounting distinguish between Azure, GCP, and vSphere resources
- TFC-aware: workspace boundaries, cross-workspace references, and registry sourcing are first-class audit concerns

## Practical Execution Notes
- Prefer fast repository discovery and indexing tools.
- Use chunked output aggressively to keep responses reviewable.
- If precise module dependency graph extraction is not available, provide best-effort accounting and mark uncertainty as `INFERENCE`.
- Favor completeness within the requested phase over speculative depth outside it.
- Avoid spending disproportionate effort on generated lock files or vendored modules when higher-signal root modules and security configurations remain unread.
- Surface inaccessible scope early so the user can correct it before later phases.
- When auditing multi-cloud codebases, track provider distribution throughout and surface any inconsistencies in security posture between clouds.

## Invocation Template
Use this skill with a prompt that supplies the missing run-specific inputs. Example:

```text
Use Terraform Audit Deep Dive on /path/to/repo.
Execute PHASE 1 - Inventory + Entrypoints only.
Emphasize TFC workspace boundaries, cross-workspace coupling, RBAC policies, and network security.
Treat vendored modules as summarized unless they are modified locally.
```

## Completion Rule
- Stop after the requested phase.
- Do not continue automatically into the next phase.
- Always end with the required state footer.
