<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

# LLM Coding Skills

A curated set of skills for LLM-powered coding assistants. Each skill is a structured prompt that gives your AI assistant deep, opinionated guidance for a specific task type — turning a general-purpose LLM into a focused specialist.

Skills are tool-agnostic and work with **Claude Code**, **Codex**, and **OpenCode**.

## Skills Matrix

### Language And Platform Skills

| Skill | Go | Rust | Python | Terraform | Helm | Kubernetes | Operator |
|-------|----|------|--------|-----------|------|------------|----------|
| **policy** | Standards layer for Go design, boundaries, errors, context, concurrency, config, observability, security, and review priorities | Standards layer for Rust design, ownership and borrowing, error handling, `unsafe` policy, async lifecycle, traits and generics, observability, security, and review priorities | Standards layer for Python design, boundaries, exceptions, typing, dependency injection, concurrency, config, observability, security, and review priorities | n/a | n/a | n/a | n/a |
| **workflow** | Execution layer for Go work: tool-first discovery, truth hierarchy, verification depth, and task-mode workflow | Execution layer for Rust work: rust-analyzer-first discovery, compiler-as-truth, parallel tool calls, verification depth proportional to risk | Execution layer for Python work: tool-first discovery, truth hierarchy, verification depth, and task-mode workflow | n/a | n/a | n/a | n/a |
| **dev** | Thin implementation mode layered on top of Go policy and workflow | Thin implementation mode layered on top of Rust policy and workflow | Thin implementation mode layered on top of Python policy and workflow | Write, modify, and review HCL with IaC best practices | Write, modify, and review Helm charts and release packaging | Write, modify, and review Kubernetes manifests with PSS-Restricted baseline, topology spread, and Kustomize best practices | Build controllers, APIs, reconcilers, webhooks, and operator wiring with kubebuilder/controller-runtime |
| **audit** | Resumable 5-stage deep-dive audit with evidence, chunking, and phase gates | Resumable 5-stage deep-dive audit including `unsafe` soundness, async lifecycle, public API stability, and supply-chain coverage | Resumable 5-stage deep-dive audit adapted for Python codebases and tooling | Resumable 5-stage infrastructure audit: resources, state, security, compliance | Resumable 5-stage chart audit: values surface, templates, release safety, upgrade risk | Resumable 5-stage cluster-manifest audit: resources, boundaries, security, operability | Resumable 5-stage operator audit: APIs, reconciliation, status, safety, and lifecycle |
| **docs** | Documentation mode for godoc, README, ADR, API, runbook, changelog, and onboarding work | Documentation mode for `///` and `//!` doc comments, doctests, `cargo doc`, mdbook guides, READMEs, ADRs, runbooks, changelogs | Generate docstrings (PEP 257), Sphinx/mkdocs, ADRs, runbooks, changelogs | Generate module READMEs, terraform-docs, ADRs, runbooks | Generate chart READMEs, values references, upgrade notes, runbooks | Generate workload docs, manifests guides, runbooks, platform ADRs | Generate CRD, controller, operational, and upgrade documentation |
| **test** | Test mode for strategy, unit/integration/e2e coverage, fuzzing, benchmarks, and regression work | Test mode for unit, integration (`tests/`), doctests, `cargo nextest`, `proptest`/`quickcheck`, `cargo-fuzz`, `criterion`, `insta` snapshots, `miri`, and coverage with `cargo llvm-cov` | Design test strategies, write tests (pytest, hypothesis, parametrize, fixtures) | Design test strategies (terraform test, Terratest, OPA/Rego, checkov) | Design chart test strategies (helm lint, template assertions, chart-testing) | Design manifest validation and conformance checks (kustomize, kubectl, policy) | Design envtest, reconciler, finalizer, status, and webhook test coverage |
| **migrate** | Migration mode for Go versions, dependency swaps, framework changes, architecture shifts, and rollback planning | Migration mode for edition upgrades, MSRV bumps, async-runtime swaps, framework migrations (e.g., actix→axum), removing `unsafe`, workspace restructuring, dependency replacement | Plan version upgrades, framework migrations (Flask/Django/FastAPI), packaging modernization | Plan version upgrades, state migrations, module refactoring, backend changes | Plan chart API, values, dependency, and release workflow migrations | Plan API version, controller, workload, and platform migrations | Plan CRD versioning, controller refactors, dependency upgrades, and conversion work |
| **ci** | Go CI job set grounded in Task + Ginkgo/stdlib + golangci-lint + staticcheck + govulncheck + goreleaser + svu + DCO + REUSE, OS matrix, per-package coverage, perf baselines | Rust CI job set: `rust-toolchain.toml`-pinned + `Swatinem/rust-cache` + rustfmt + clippy `-D warnings` + `cargo nextest` matrix + `cargo llvm-cov` coverage + `cargo doc` + `cargo deny` + `cargo audit` + MSRV verify + `miri` for unsafe + `release-plz`/`cargo-dist` | Python CI job set: ruff, mypy/pyright, pytest matrix across Python versions, pip-audit, bandit, uv build + twine check, OIDC trusted publishing | n/a | n/a | n/a | n/a |

### Cross-Cutting Skills

| Skill | Description |
|-------|-------------|
| **adr-write** | Language-agnostic skill for writing, reviewing, and superseding Architecture Decision Records. MADR-compatible format, supersession rules, grounded in constraints rather than memory. |
| **planning-decompose** | Turn an ambiguous request into a verifiable plan: goal, acceptance criteria, milestones, per-step verification, risks, and rollback. Hands off to execution-mode skills. |
| **rfc-write** | Structured RFC proposals with motivation, goals/non-goals, alternatives, impact, rollout, and review plan. Upstream of ADRs: the RFC is the proposal under debate, the ADR records the decision. |
| **docker-image** | Dockerfile and container image design: scratch → distroless → slim tier preference, multi-stage builds, BuildKit secrets, non-root, digest pinning, SBOM, signing, size budgets. |
| **security-review** | Stack-agnostic focused security review: threat surfaces, auth, input validation, secrets, transport, supply chain, findings with severity and concrete fix. |
| **cicd-core** | Platform-agnostic CI/CD principles: stage ordering, reproducibility, secret discipline, artifact immutability, least-privilege credentials, caching, retries, observability. |
| **cicd-github-actions** | GitHub Actions specifics: reusable workflows, SHA pinning, explicit permissions, OIDC cloud auth, fork-safe triggers, concurrency, matrix, environments, self-hosted runner trust. |
| **cicd-azure-devops** | ADO YAML pipelines: templates, `extends` for org standards, Library variable groups + Key Vault, workload identity federation for service connections, Environments with approvals. |
| **cicd-gitops** | Argo CD and Flux patterns: pull-based deploy, PR-driven promotion, sync waves, progressive delivery, SOPS/ESO secret management, RBAC, drift detection. |
| **cicd-supply-chain** | SLSA levels, SBOM, in-toto/SLSA provenance, Sigstore keyless signing + verification, dependency pinning, vuln scanning, signed commits, reproducible builds, admission-time enforcement. |

## Quick Start

### Install

The installer auto-detects which tools you have and installs skills at the user level for each:

```bash
./install.sh
```

### Selective Install

Install for specific tools or languages:

```bash
# Only Claude Code
./install.sh --tool=claude

# Only Claude Code and OpenCode
./install.sh --tool=claude,opencode

# Only Go and Python skills
./install.sh --lang=go,python

# Only Rust skills
./install.sh --lang=rust

# Combine filters
./install.sh --tool=claude --lang=terraform

# Only Helm and Kubernetes skills
./install.sh --tool=codex --lang=helm,kubernetes

# Only operator skills
./install.sh --tool=codex --lang=operator

# Preview without installing
./install.sh --dry-run
```

### Other Commands

```bash
# List detected tools and all available skills
./install.sh --list

# Remove all installed skills
./install.sh --uninstall

# Remove only Codex skills
./install.sh --uninstall --tool=codex
```

## Using Skills

Once installed, skills are invoked differently depending on your tool:

### Layered stacks vs. single-mode stacks

The Go, Rust, and Python stacks are layered:

- Start with `workflow` as the default entrypoint for most work.
- Add `policy` when you want the standards layer made explicit.
- Add a task mode such as `dev`, `audit`, `docs`, `test`, or `migrate` when the task benefits from a more specialized overlay.

Terraform, Helm, Kubernetes, and Operator skills do not currently use a separate `workflow` skill, so their task-mode skills remain the normal entrypoints.

### Claude Code

Skills are installed as custom slash commands in `~/.claude/commands/`:

```
> /go-workflow
> /python-workflow
> /terraform-test
```

For layered stacks, start with workflow and add a task mode only when needed:

```
> /go-workflow Implement the new queue consumer in this repo. Follow TDD.
> /go-workflow /go-test Add regression coverage for retry exhaustion.
> /python-workflow Refactor the API client and verify the existing pytest flow.
```

### Codex

Skills are installed in `~/.agents/skills/` and invoked with the `$` prefix:

```
$go-workflow
$python-workflow
$terraform-test
```

Typical usage:

```
$go-workflow
$go-test
$operator-dev
```

Codex can also auto-select skills based on task descriptions when the skill's trigger conditions match.

### OpenCode

Skills are installed as agents in `~/.config/opencode/agents/` and invoked with the `@` prefix:

```
@go-workflow
@python-workflow
@terraform-test
```

## Skill Details

### Go policy + workflow

The Go stack now uses a layered model:

- **policy** defines what good Go engineering looks like: package boundaries, interfaces, errors, context, concurrency, config, logging, security, shutdown, and review priorities.
- **workflow** defines how Go work should be executed: repository discovery, tool precedence, truth hierarchy, verification depth, task classification, and uncertainty reporting.
- **dev**, **audit**, **docs**, **test**, and **migrate** are task modes layered on top of policy and workflow.
- In normal use, start with **workflow** and add a task mode only when the work is clearly specialized.

This gives the Go stack a clearer precedence model and reduces duplication between skills.

### Python policy + workflow

The Python stack now uses the same layered model:

- **policy** defines what good Python engineering looks like: module boundaries, protocols and ABCs, exceptions, typing, dependency injection, concurrency, config, logging, security, shutdown, and review priorities.
- **workflow** defines how Python work should be executed: repository discovery, tool precedence, truth hierarchy, verification depth, task classification, and uncertainty reporting.
- **dev**, **audit**, **docs**, **test**, and **migrate** are task modes layered on top of policy and workflow.
- In normal use, start with **workflow** and add a task mode only when the work is clearly specialized.

This gives the Python stack the same explicit precedence model and reduces duplicated boilerplate across skills.

### Rust policy + workflow

The Rust stack uses the same layered model:

- **policy** defines what good Rust engineering looks like: crate and module boundaries, ownership and borrowing discipline, traits and generics, error handling (`thiserror` for libraries, `anyhow` for apps), `unsafe` policy with `// SAFETY:` comments, async lifecycle (cancellation, drop, blocking), Cargo features, edition and MSRV, observability with `tracing`, security, and review priorities.
- **workflow** defines how Rust work should be executed: `rust-analyzer`-first discovery, the compiler as the cheapest verifier, parallel tool calls, `cargo expand` for macro-heavy paths, `miri` for `unsafe`, verification depth proportional to risk.
- **dev**, **audit**, **docs**, **test**, and **migrate** are task modes layered on top of policy and workflow.
- In normal use, start with **workflow** and add a task mode only when the work is clearly specialized.

### dev (Development Guidance)

For Go, Rust, and Python, `dev` is now a thin implementation mode layered on top of language-specific policy and workflow. For Terraform, `dev` remains the core coding skill.

### audit (Audit Deep Dive)

A structured, resumable 5-phase audit protocol for enterprise-grade codebase review:

1. **Phase 1 — Inventory + Entrypoints**: Catalog files, packages, entry points, startup/shutdown behavior, configuration sources
2. **Phase 2 — Accounting**: Index every function/class/resource with CSV artifacts
3. **Phase 3 — Architecture + Boundaries**: Map data flow, identify boundary violations, assess isolation
4. **Phase 4 — Security + Observability**: Prioritized findings (P0/P1/P2) with evidence and concrete fixes
5. **Phase 5 — Synthesis**: Letter grades, prioritized refactor plan, 90-day roadmap

Each audit skill now uses explicit continuation rules: execute one phase at a time, stop at the phase boundary, preserve chunked artifacts when needed, and end with a `STATE_SNAPSHOT` plus exact `NEXT` phase or follow-up step. Each phase has strict gate rules — no recommendations before findings, no grades before synthesis.

### docs (Documentation Guidance)

Covers all documentation types: inline docs (godoc/docstrings/variable descriptions), project docs (READMEs, ADRs, runbooks, changelogs), and API docs (OpenAPI, Sphinx, mkdocs, terraform-docs). Enforces language-specific conventions (godoc format, PEP 257, terraform-docs markers) and requires all documentation to be grounded in actual code.

### Helm + Kubernetes

The Helm and Kubernetes stacks follow the same task-mode pattern as Terraform:

- **dev** covers chart, manifest, and packaging implementation work.
- **audit** covers resumable, phase-by-phase review of chart structure, workload boundaries, security, and operability.
- **docs** covers chart READMEs, values references, deployment guides, and runbooks.
- **test** covers linting, rendering, schema validation, policy checks, and deployment safety verification.
- **migrate** covers API version changes, chart breaking changes, controller upgrades, and rollback planning.

### Operator Skills And Their Relationship To Go

The operator stack is a specialization of Go development rather than a replacement for it:

- **go-policy** remains the source of truth for general Go engineering quality.
- **go-workflow** remains the source of truth for execution discipline and verification flow.
- **operator-dev**, **operator-test**, **operator-docs**, **operator-audit**, and **operator-migrate** add controller-specific rules for CRDs, reconciliation, status, finalizers, watches, webhooks, and lifecycle safety.
- When operator-specific guidance conflicts with generic Go guidance, follow the operator skill for controller concerns and the Go skills for general engineering concerns.

The operator skills are intentionally centered on **kubebuilder**, **controller-runtime**, and **achilles-sdk** patterns. `operator-sdk` is not treated as the primary workflow; if a repository already uses it, preserve existing conventions rather than forcing a migration.

### test (Test Strategy and Generation)

Covers the full testing spectrum for each language:

- **Go**: Table-driven tests, Ginkgo/Gomega BDD, fuzz testing, benchmarks, golden files, httptest, race detection
- **Python**: pytest fixtures/parametrize, hypothesis property-based testing, async testing, mocking patterns, coverage analysis
- **Terraform**: Native terraform test (.tftest.hcl), Terratest, OPA/Rego policy testing, checkov compliance, plan-time assertions

Includes guidance on test philosophy, coverage analysis, fixture design, and common anti-patterns.

### migrate (Migration Planning)

Structured migration planning with impact analysis, risk assessment, phased execution, and rollback strategies:

- **Go**: Version upgrades, dependency swaps, framework migrations, architecture transitions, database migrations
- **Python**: Version upgrades, framework migrations (Flask/Django/FastAPI), sync-to-async transitions, packaging modernization (setup.py to pyproject.toml)
- **Terraform**: Version upgrades, state migrations (moved/import blocks), module refactoring, backend changes, state splitting, tool migrations (Terraform to OpenTofu)

Core rule: never migrate and refactor simultaneously.

### adr-write (Architecture Decision Records)

Language-agnostic skill for writing, reviewing, and superseding ADRs. Uses a MADR-compatible template, enforces supersession (new ADR rather than in-place edit), requires honest alternatives and real negative consequences, and grounds every claim in concrete constraints (code, config, incidents, prior ADRs).

Pairs with any language or platform skill: plan the decision with ADR mode, then execute with the relevant dev mode.

### planning-decompose (Task Decomposition)

Turn an ambiguous or large request into a verifiable plan: goal, acceptance criteria, assumptions/unknowns, current vs. target state, milestones with per-step verification, risks with detection and rollback, and explicit sizing. Stops planning when it hits a hidden decision and escalates to `adr-write` or a clarifying question rather than absorbing the decision silently.

Hand off each milestone to the appropriate execution-mode skill; planning-decompose does not implement.

### rfc-write (Request for Comments)

Structured proposals for non-trivial changes to architecture, protocols, tooling, or process. RFC sits upstream of ADR: the RFC is the proposal under discussion, the ADR records the decision. Template covers motivation, goals/non-goals, current state, proposal, real alternatives including "do nothing", impact (users, operators, compatibility, security, performance, cost), rollout plan with reversibility, open questions, and review plan with dates and decision mechanism.

### docker-image (Dockerfile And Container Image)

Container image design with a strong preference order for base images: **scratch → distroless → slim (Alpine, Debian slim) → full distro**. Covers multi-stage builds, BuildKit secrets and cache mounts, non-root UID, read-only root filesystem, digest pinning, OCI metadata labels, SBOM and cosign signing, multi-arch, and size budgets by tier.

### security-review (Security Review)

Stack-agnostic focused security review. Walks ten threat surfaces — authentication, authorization, input validation, secrets, transport/storage, dependencies/supply chain, configuration/deployment, logging/monitoring, cryptography, business logic/abuse. Produces findings grouped by severity (Critical → High → Medium → Low → Info) with location, evidence, impact, exploitability, and concrete fix. Pairs with `cicd-supply-chain` for release integrity and `*-audit` for deeper stack-specific review.

### CI/CD Skill Stack

Five layered CI/CD skills. Start with `cicd-core` (platform-agnostic) and add one or more specialized skills:

- **cicd-core** — principles: stage ordering cheapest-first, reproducibility, secret discipline, artifact immutability, OIDC-preferred credentials, caching rules, retry hygiene, deploy boundary.
- **cicd-github-actions** — Actions specifics: `pull_request` vs. `pull_request_target` safety, SHA-pinned actions, `permissions: {}` default, OIDC to AWS/GCP/Azure, reusable workflows vs. composite actions, matrix and caching patterns, self-hosted runner trust model.
- **cicd-azure-devops** — YAML pipelines, step/job/stage/extends templates, Library variable groups + Key Vault, workload identity federation for service connections, ADO Environments with checks, resources (`repositories:`, `pipelines:`, `containers:`).
- **cicd-gitops** — Argo CD and Flux. Repository layout, Application / Kustomization design, ApplicationSet generators, sync waves, progressive delivery with Argo Rollouts / Flagger, SOPS / External Secrets / Sealed Secrets, RBAC and drift detection, promotion via PR.
- **cicd-supply-chain** — SLSA levels, SBOM (SPDX/CycloneDX), in-toto and SLSA provenance, Sigstore keyless signing + Rekor verification, dependency pinning, vulnerability scanning with failure criteria, signed commits (gitsign), admission-time verification with Kyverno / Sigstore policy-controller.

### Language-Specific CI

Per-language CI skills with concrete job sets, toolchain pinning strategy, and publish workflows:

- **go-ci** — Task runner, Ginkgo or stdlib testing, golangci-lint, staticcheck, govulncheck, goreleaser, svu, DCO and REUSE when required, cross-OS matrix, per-package coverage thresholds, benchmark history as artifact, CI summary.
- **rust-ci** — `rust-toolchain.toml` pinning, `Swatinem/rust-cache`, rustfmt, clippy with `-D warnings`, `cargo nextest` matrix, doctests, `cargo llvm-cov` coverage, `cargo doc` with broken-link checking, `cargo deny` + `cargo audit`, MSRV verify, `miri` for unsafe-bearing crates, `release-plz` for release PRs, `cargo-dist` for artifact pipelines.
- **python-ci** — uv (or poetry) with lockfile integrity check, ruff (lint + format), mypy or pyright, pytest matrix across supported Python versions, pip-audit, bandit, uv build + twine check, OIDC trusted publishing to PyPI.

## Versioning

All skills are versioned using a `<!-- version: X.Y.Z -->` comment at the top of each file. The installer detects existing installations and reports:

- **Upgrade**: when the source version is newer than the installed version
- **Downgrade**: when the installed version is newer (warns before overwriting)
- **No change**: when versions match (still overwrites to pick up non-version changes)

Check installed versions with `./install.sh --list`.

## License

This repository is licensed under the SPDX license expression `Apache-2.0`.

Copyright attribution for the repository and included skills is `Rillan AI LLC`.

This project was originally developed under the Skaphos name (MIT licensed) at
`github.com/skaphos/agent_resources` and relicensed when transferred to Rillan AI LLC.

## Directory Structure

```
skills/
├── go/
│   ├── audit.skill.md
│   ├── ci.skill.md
│   ├── dev.skill.md
│   ├── docs.skill.md
│   ├── migrate.skill.md
│   ├── policy.skill.md
│   ├── test.skill.md
│   └── workflow.skill.md
├── rust/
│   ├── audit.skill.md
│   ├── ci.skill.md
│   ├── dev.skill.md
│   ├── docs.skill.md
│   ├── migrate.skill.md
│   ├── policy.skill.md
│   ├── test.skill.md
│   └── workflow.skill.md
├── python/
│   ├── audit.skill.md
│   ├── ci.skill.md
│   ├── dev.skill.md
│   ├── docs.skill.md
│   ├── migrate.skill.md
│   ├── policy.skill.md
│   ├── test.skill.md
│   └── workflow.skill.md
├── terraform/
│   ├── audit.skill.md
│   ├── dev.skill.md
│   ├── docs.skill.md
│   ├── migrate.skill.md
│   └── test.skill.md
├── helm/
│   ├── audit.skill.md
│   ├── dev.skill.md
│   ├── docs.skill.md
│   ├── migrate.skill.md
│   └── test.skill.md
├── kubernetes/
│   ├── audit.skill.md
│   ├── dev.skill.md
│   ├── docs.skill.md
│   ├── migrate.skill.md
│   └── test.skill.md
├── operator/
│   ├── audit.skill.md
│   ├── dev.skill.md
│   ├── docs.skill.md
│   ├── migrate.skill.md
│   └── test.skill.md
├── adr/
│   └── write.skill.md
├── planning/
│   └── decompose.skill.md
├── rfc/
│   └── write.skill.md
├── docker/
│   └── image.skill.md
├── security/
│   └── review.skill.md
├── cicd/
│   ├── azure-devops.skill.md
│   ├── core.skill.md
│   ├── github-actions.skill.md
│   ├── gitops.skill.md
│   └── supply-chain.skill.md
├── install.sh
└── README.md
```
