<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

# Rillan Skills

A curated set of skills for LLM-powered coding assistants. Each skill is a structured prompt that gives your AI assistant deep, opinionated guidance for a specific task type — turning a general-purpose LLM into a focused specialist.

Skills are tool-agnostic and work with **Claude Code**, **Codex**, and **OpenCode**.

## Skills Matrix

### Language and platform skills

A `✓` means a skill of that mode exists for that stack. See the per-mode sections below for what each mode does.

| Mode | Go | Rust | Python | Terraform | Helm | Kubernetes | Operator |
|---|:-:|:-:|:-:|:-:|:-:|:-:|:-:|
| `policy`   | ✓ | ✓ | ✓ | — | — | — | — |
| `workflow` | ✓ | ✓ | ✓ | — | — | — | — |
| `dev`      | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| `audit`    | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| `docs`     | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| `test`     | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| `migrate`  | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| `ci`       | ✓ | ✓ | ✓ | — | — | — | — |

The Go, Rust, and Python stacks are layered: `workflow` is the default entrypoint, `policy` is the standards layer, and `dev` / `audit` / `docs` / `test` / `migrate` are task modes layered on top. Terraform, Helm, Kubernetes, and Operator skills do not currently use a separate `workflow` skill — their task-mode skills are the entrypoints.

### Cross-cutting skills

| Skill | What it does |
|---|---|
| `adr-write` | Write, review, and supersede Architecture Decision Records. MADR-compatible. |
| `planning-decompose` | Turn an ambiguous request into a verifiable plan: goal, milestones, risks, rollback. |
| `rfc-write` | Structured RFC proposals upstream of an ADR — motivation, alternatives, rollout. |
| `docker-image` | Dockerfile design: scratch → distroless → slim, multi-stage, non-root, digest pinning, SBOM. |
| `security-review` | Stack-agnostic focused security review with severity-ranked findings. |
| `cicd-core` | Platform-agnostic CI/CD principles — stage ordering, secrets, OIDC, retries. |
| `cicd-github-actions` | GitHub Actions specifics — SHA pinning, OIDC, fork-safe triggers, reusable workflows. |
| `cicd-azure-devops` | Azure DevOps YAML pipelines — templates, Library + Key Vault, workload identity, Environments. |
| `cicd-gitops` | Argo CD / Flux — Application design, sync waves, progressive delivery, secret management. |
| `cicd-supply-chain` | SLSA, SBOM, Sigstore signing, provenance, signed commits, admission-time verification. |

## Quick Start

### Install the binary

#### From a GitHub release (recommended)

Each release attaches prebuilt binaries for Linux, macOS, and Windows (amd64 + arm64)
plus a `checksums.txt`. Download the archive for your platform from the
[Releases page](https://github.com/rillanai/rillan-skills/releases/latest), verify it
against the checksums, extract, and put `rillan-skills` on your `PATH`:

```bash
# Linux x86_64 example — adjust VERSION and the os/arch suffix to taste.
VERSION=0.1.0
base="https://github.com/rillanai/rillan-skills/releases/download/v${VERSION}"
curl -fsSLO "${base}/rillan-skills_${VERSION}_linux_amd64.tar.gz"
curl -fsSLO "${base}/checksums.txt"
sha256sum --ignore-missing -c checksums.txt
tar -xzf "rillan-skills_${VERSION}_linux_amd64.tar.gz"
install -m 0755 rillan-skills ~/.local/bin/rillan-skills
rillan-skills version
```

(No Homebrew tap yet — the GitHub release archives are the supported download.)

#### From source

The installer is a single Go binary built from this repo:

```bash
task build
task install      # copies bin/rillan-skills to ~/.local/bin/
```

Or build and use directly:

```bash
task build
./bin/rillan-skills --help
```

### Install skills into a project

The installer detects which language and platform packs are relevant to a target repository, then writes skills into that repo's tool-specific directory:

```bash
# Detect packs for the current repo (no writes)
rillan-skills detect --target .

# Install into the current repo, for Claude Code
rillan-skills install --target . --tool claude

# Install for Codex and OpenCode at the same time
rillan-skills install --target . --tool claude,codex,opencode

# Override detection — install only specific packs
rillan-skills install --target . --packs go,security --tool claude

# Preview without writing
rillan-skills install --target . --tool claude --dry-run

# Force overwrite even when versions match
rillan-skills install --target . --tool claude --force
```

### Install skills globally (user-level, all tools)

Use `--global` to install into each tool's **user-level** config directory instead
of a repository, so the skills are available across every project. By default
`--global` installs **all** bundled packs (detection is meaningless against your home
dir); it composes with `--tool` and you can still pass `--packs go,security` to narrow
the selection. Pass `--tool all` to install for every supported tool at once:

```bash
# Install all packs globally for Claude Code only
rillan-skills install --global --tool claude

# Install all packs globally for every supported tool
rillan-skills install --global --tool all

# Install only a subset globally (--packs still narrows under --global)
rillan-skills install --global --tool all --packs go,security

# Preview a global install without writing
rillan-skills install --global --tool all --dry-run

# Remove everything installed globally (uses the manifest; never touches files it didn't install)
rillan-skills uninstall --global --tool all
```

Per-tool global destinations:

| Tool | Global destination |
|---|---|
| Claude Code | `~/.claude/skills/<pack>/SKILL.md` + `<pack>/<mode>.md` |
| Codex | `~/.codex/skills/<pack>/SKILL.md` + `<pack>/<mode>.md` |
| OpenCode | `~/.config/opencode/agent/<pack>.md` (single file: router + all modes inlined) |
| Grok | **No standard skills directory** — skipped with a warning. Copy the desired packs from `skills/<pack>/` into your Grok prompt/config manually. |

Grok is recognized (so `--tool grok` and `--tool all` report it) but has no
well-defined skills location, so it is skipped rather than failing the run.

### Install manifest

Every install (per-repo or `--global`) records what it wrote to a manifest so
re-installs can decide upgrade-vs-skip and uninstall can remove exactly what was
installed:

- **Per-repo:** `<target>/.rillan-skills/manifest.json`
- **Global:** `~/.rillan-skills/manifest.json`

The manifest records, per installed pack, the `tool`, `pack`, installed `version`,
and the list of `files` written (relative to the scope root). On (re)install the
bundled pack version (`<!-- version: X.Y.Z -->`) is compared against the manifest
(falling back to the installed router's marker), and each pack is reported as
`installed` / `upgraded X.Y.Z->A.B.C` / `up-to-date` / `reinstalled` /
`downgraded A.B.C->X.Y.Z`. A downgrade (installed version newer than bundled) is
**skipped by default** — pass `--force` to overwrite it.

`uninstall` removes only the files the manifest recorded — and only those that
resolve inside the scope root — so it never deletes unrelated skills sharing a
directory. For installs that predate the manifest (no recorded entry for a tool),
`uninstall` falls back to removing the known pack paths (`<base>/<pack>/` or
`<base>/<pack>.md`) that exist on disk.

### Other commands

```bash
rillan-skills list                    # list all skills bundled in the binary
rillan-skills uninstall --target .    # remove installed skill files
rillan-skills uninstall --target . --tool codex --dry-run
rillan-skills uninstall --global --tool all --dry-run
```

### Architecture: one root skill per concern

Each pack installs as a **single root skill** (`SKILL.md`) whose frontmatter is the only
thing your assistant loads up front. The root is a router: its body points at bundled
`<mode>.md` files (`policy`, `workflow`, `dev`, `test`, `audit`, `docs`, `migrate`, `ci`,
…) that the assistant reads **on demand** for the task at hand. This is progressive
disclosure — installing the Go, cicd, and kubernetes packs adds **three** descriptions to
the always-loaded set, not the ~20 it would take to register every mode as its own skill.

### Where skills land

| Tool | Project-scoped destination |
|---|---|
| Claude Code | `<repo>/.claude/skills/<pack>/SKILL.md` + `<pack>/<mode>.md` |
| Codex | `<repo>/.codex/skills/<pack>/SKILL.md` + `<pack>/<mode>.md` |
| OpenCode | `<repo>/.opencode/agents/<pack>.md` (single file: router + all modes inlined) |

Claude Code and Codex support bundled, progressively-disclosed files, so they get the
real router + mode-file directory. OpenCode uses single-file agents, so the installer
assembles one self-contained agent per pack with every mode inlined.

## Using Skills

Once installed, invoke the **root** for a concern; the router loads the right mode(s).

### Claude Code

```
> /go
> /python
> /terraform
```

For the layered language stacks the root loads `policy.md` + `workflow.md` as the
baseline, then the mode file matching the task:

```
> /go Implement the new queue consumer in this repo. Follow TDD.
> /go Add regression coverage for retry exhaustion.   # router pulls in test.md
```

### Codex

```
$go
$python
$terraform
```

Codex can also auto-select a root based on the task description when its trigger
conditions match.

### OpenCode

```
@go
@python
@terraform
```

## Skill modes

### `policy` + `workflow` (Go, Rust, Python)

The Go, Rust, and Python stacks use a layered model:

- **`policy`** defines what good engineering looks like in that language: package or crate boundaries, errors and panics, ownership and lifetimes (Rust), typing (Python), context (Go), concurrency, configuration, observability, security, and review priorities.
- **`workflow`** defines how work should be executed: structural-tool-first discovery (`gopls`, `rust-analyzer`, LSP/pyright/mypy), the compiler/type-checker as the cheapest verifier, parallel tool calls, truth hierarchy, verification depth proportional to risk.
- **`dev`**, **`audit`**, **`docs`**, **`test`**, and **`migrate`** are task modes layered on top.

In normal use, start with `workflow` and add a task mode only when the work is clearly specialized.

### `dev`

For Go, Rust, and Python, `dev` is a thin implementation mode layered on top of language `policy` and `workflow`. For Terraform, Helm, Kubernetes, and Operator, `dev` is the core implementation skill (HCL authoring, chart authoring, manifest authoring, controller wiring).

### `audit`

A structured, resumable 5-phase audit protocol for enterprise-grade codebase review:

1. **Inventory + Entrypoints** — files, packages, entry points, startup/shutdown, configuration sources.
2. **Accounting** — index every function/class/resource with CSV artifacts.
3. **Architecture + Boundaries** — data flow, boundary violations, isolation.
4. **Security + Observability** — prioritized findings (P0/P1/P2) with evidence and concrete fixes.
5. **Synthesis** — letter grades, prioritized refactor plan, 90-day roadmap.

Each audit skill uses explicit continuation rules: execute one phase at a time, stop at the phase boundary, preserve chunked artifacts when needed, and end with a `STATE_SNAPSHOT` plus exact `NEXT` step. Each phase has strict gate rules — no recommendations before findings, no grades before synthesis.

### `docs`

Covers all documentation types: inline (godoc, doc comments, docstrings), project (README, ADR, runbook, changelog), and API (OpenAPI, Sphinx, mkdocs, terraform-docs, `cargo doc`). Enforces language-specific conventions and requires every claim to be grounded in actual code.

### `test`

Covers the full testing spectrum for each stack:

- **Go** — table-driven tests, Ginkgo/Gomega, fuzz, benchmarks, golden files, `httptest`, race detection.
- **Rust** — unit, integration (`tests/`), doctests, `cargo nextest`, `proptest`/`quickcheck`, `cargo-fuzz`, `criterion`, `insta` snapshots, `miri`, coverage with `cargo llvm-cov`.
- **Python** — pytest fixtures and parametrize, hypothesis property-based testing, async testing, mocking, coverage analysis.
- **Terraform** — `terraform test`, Terratest, OPA/Rego, checkov, plan-time assertions.
- **Helm** — `helm lint`, `helm template` regression checks, `values.schema.json`, chart-testing.
- **Kubernetes** — kubeconform, `kubectl apply --dry-run=server`, conftest, kyverno test.
- **Operator** — envtest, reconciler tests, finalizer/status/webhook coverage.

### `migrate`

Structured migration planning with impact analysis, risk assessment, phased execution, and rollback strategies. Core rule: never migrate and refactor simultaneously.

### `ci`

Per-language CI skills with concrete job sets, toolchain pinning strategy, and publish workflows:

- **`go-ci`** — Task runner, golangci-lint, staticcheck, govulncheck, goreleaser, svu, DCO + REUSE, OS matrix, per-package coverage.
- **`rust-ci`** — `rust-toolchain.toml` pinning, `Swatinem/rust-cache`, rustfmt, clippy `-D warnings`, `cargo nextest`, `cargo llvm-cov`, `cargo deny`, `cargo audit`, MSRV verify, `miri`, `release-plz`, `cargo-dist`.
- **`python-ci`** — uv (or poetry) with lockfile integrity check, ruff, mypy or pyright, pytest matrix, pip-audit, bandit, OIDC trusted publishing.

### Helm + Kubernetes

The Helm and Kubernetes stacks follow the same task-mode pattern as Terraform:

- **`dev`** — chart, manifest, and packaging implementation work.
- **`audit`** — phase-by-phase review of chart structure, workload boundaries, security, operability.
- **`docs`** — chart READMEs, values references, deployment guides, runbooks.
- **`test`** — linting, rendering, schema validation, policy checks, deployment safety.
- **`migrate`** — API versions, chart breaking changes, controller upgrades, rollback planning.

### Operator skills and their relationship to Go

The operator stack is a specialization of Go development rather than a replacement for it:

- **`go-policy`** is the source of truth for general Go engineering quality.
- **`go-workflow`** is the source of truth for execution discipline and verification flow.
- **`operator-*`** skills add controller-specific rules for CRDs, reconciliation, status, finalizers, watches, webhooks, and lifecycle safety.

When operator-specific guidance conflicts with generic Go guidance, follow the operator skill for controller concerns and the Go skills for general engineering concerns.

The operator skills are intentionally centered on **kubebuilder**, **controller-runtime**, and **achilles-sdk** patterns. `operator-sdk` is not treated as the primary workflow; if a repository already uses it, preserve existing conventions rather than forcing a migration.

### `adr-write`

Language-agnostic skill for writing, reviewing, and superseding ADRs. Uses a MADR-compatible template, enforces supersession (new ADR rather than in-place edit), requires honest alternatives and real negative consequences, and grounds every claim in concrete constraints (code, config, incidents, prior ADRs). Pairs with any language or platform skill: plan with ADR mode, then execute with the relevant `dev` mode.

### `planning-decompose`

Turn an ambiguous or large request into a verifiable plan: goal, acceptance criteria, assumptions and unknowns, current vs. target state, milestones with per-step verification, risks with detection and rollback, and explicit sizing. Stops planning when it hits a hidden decision and escalates to `adr-write` or a clarifying question rather than absorbing the decision silently. Hand off each milestone to the appropriate execution-mode skill.

### `rfc-write`

Structured proposals for non-trivial changes to architecture, protocols, tooling, or process. RFC sits upstream of ADR — the RFC is the proposal under discussion, the ADR records the decision. Template covers motivation, goals/non-goals, current state, proposal, real alternatives including "do nothing", impact, rollout plan with reversibility, open questions, and review plan with dates and decision mechanism.

### `docker-image`

Container image design with a strong preference order for base images: **scratch → distroless → slim (Alpine, Debian slim) → full distro**. Covers multi-stage builds, BuildKit secrets and cache mounts, non-root UID, read-only root filesystem, digest pinning, OCI metadata labels, SBOM and cosign signing, multi-arch, and size budgets by tier.

### `security-review`

Stack-agnostic focused security review. Walks ten threat surfaces — authentication, authorization, input validation, secrets, transport/storage, dependencies and supply chain, configuration and deployment, logging and monitoring, cryptography, and business-logic abuse. Produces findings grouped by severity (Critical → High → Medium → Low → Info) with location, evidence, impact, exploitability, and concrete fix. Pairs with `cicd-supply-chain` for release integrity and `*-audit` for deeper stack-specific review.

### CI/CD skill stack

Five layered CI/CD skills. Start with `cicd-core` (platform-agnostic) and add one or more specialized skills:

- **`cicd-core`** — stage ordering cheapest-first, reproducibility, secret discipline, artifact immutability, OIDC-preferred credentials, caching rules, retry hygiene, deploy boundary.
- **`cicd-github-actions`** — `pull_request` vs. `pull_request_target` safety, SHA-pinned actions, `permissions: {}` default, OIDC to AWS/GCP/Azure, reusable workflows vs. composite actions, matrix and caching patterns, self-hosted runner trust model.
- **`cicd-azure-devops`** — YAML pipelines, step/job/stage/`extends` templates, Library variable groups + Key Vault, workload identity federation for service connections, ADO Environments with checks, resources (`repositories:`, `pipelines:`, `containers:`).
- **`cicd-gitops`** — Argo CD and Flux. Repository layout, Application / Kustomization design, ApplicationSet generators, sync waves, progressive delivery with Argo Rollouts / Flagger, SOPS / External Secrets / Sealed Secrets, RBAC and drift detection, promotion via PR.
- **`cicd-supply-chain`** — SLSA levels, SBOM (SPDX/CycloneDX), in-toto and SLSA provenance, Sigstore keyless signing + Rekor verification, dependency pinning, vulnerability scanning with failure criteria, signed commits (gitsign), admission-time verification with Kyverno / Sigstore policy-controller.

## Versioning

Each pack is versioned by a `<!-- version: X.Y.Z -->` comment in its router `SKILL.md`. The installer keys idempotency on that version (the installed `SKILL.md` for Claude/Codex, or the assembled agent file for OpenCode):

- **Same version** — skipped (use `--force` to overwrite).
- **Different version** — overwrites the whole pack (router + mode files) with the bundled version.
- **No prior version** — fresh install.

Check what's bundled with `rillan-skills list`.

This skill version is independent of the released **binary** version (see Releasing).

## Releasing

Releases are automated with [release-please](https://github.com/googleapis/release-please)
and [GoReleaser](https://goreleaser.com), wired together in
`.github/workflows/release.yml`:

1. Every push to `main` updates a **release PR** that release-please maintains from
   [Conventional Commits](https://www.conventionalcommits.org) (`feat:` → minor,
   `fix:` → patch, `feat!:`/`BREAKING CHANGE:` → major). Use conventional-commit
   subjects (or a conventional squash-merge title) so changes land in that PR.
2. Merging the release PR makes release-please cut the tag (`vX.Y.Z`) and the GitHub
   release, and emit `release_created=true`.
3. In the **same workflow run**, GoReleaser checks out that tag and uploads the
   cross-platform binaries + `checksums.txt` to the release it created.

Running GoReleaser inside the same run (gated on `release_created`) is deliberate: a
separate `on: push tags` workflow would not fire, because tags pushed with the default
`GITHUB_TOKEN` do not trigger further workflows. GoReleaser uses `release.mode:
keep-existing` and never pushes, moves, or deletes the tag — so the pipeline is safe
under tag-protection rules that forbid mutating a tag after it is cut.

**First release.** The manifest (`.release-please-manifest.json`) starts at `0.0.0`. To
cut a specific first version regardless of commit history, land a commit with a
`Release-As: 0.1.0` footer (or edit the version in the release PR before merging).
Validate the release config locally with `task release:check`; dry-run the artifacts
with `task release:snapshot`.

## Directory Structure

```
.
├── cmd/rillan-skills/        # installer CLI entrypoint
├── internal/
│   ├── detect/               # filesystem-based pack detection
│   └── install/              # skill writers per tool
├── embed.go                  # embeds skills/ into the binary
├── skills/                   # one dir per pack; SKILL.md is the router
│   ├── adr/      └─ SKILL.md
│   ├── cicd/     └─ SKILL.md + core, github-actions, azure-devops, gitops, supply-chain (.md)
│   ├── docker/   └─ SKILL.md
│   ├── go/       └─ SKILL.md + policy, workflow, dev, audit, docs, test, migrate, ci (.md)
│   ├── helm/     └─ SKILL.md + dev, audit, docs, test, migrate (.md)
│   ├── kubernetes/ └─ SKILL.md + dev, audit, docs, test, migrate (.md)
│   ├── operator/ └─ SKILL.md + dev, audit, docs, test, migrate (.md)
│   ├── planning/ └─ SKILL.md
│   ├── python/   └─ SKILL.md + policy, workflow, dev, audit, docs, test, migrate, ci (.md)
│   ├── rfc/      └─ SKILL.md
│   ├── rust/     └─ SKILL.md + policy, workflow, dev, audit, docs, test, migrate, ci (.md)
│   ├── security/ └─ SKILL.md
│   └── terraform/ └─ SKILL.md + dev, audit, docs, test, migrate (.md)
├── .goreleaser.yaml          # cross-platform release builds
├── release-please-config.json
├── .release-please-manifest.json
├── .github/workflows/        # ci.yml, release.yml
├── Taskfile.yml
├── AGENTS.md
└── README.md
```

## License

This repository is licensed under the SPDX license expression `Apache-2.0`.

Copyright attribution for the repository and included skills is `Rillan AI LLC`.

This project was originally developed under the Skaphos name (MIT licensed) at `github.com/skaphos/agent_resources` and relicensed when transferred to Rillan AI LLC.
