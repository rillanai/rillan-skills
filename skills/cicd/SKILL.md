---
name: cicd
description: CI/CD pipeline design and review across GitHub Actions, Azure DevOps, GitLab, Jenkins, Buildkite, and Argo CD/Flux — stage ordering, fast feedback, secret hygiene, OIDC/least-privilege credentials, caching, immutable artifacts, SBOM/SLSA/Sigstore signing and provenance, and GitOps delivery. Triggers on `.github/workflows/`, `azure-pipelines.yml`, `.gitlab-ci.yml`, or Argo/Flux manifests. Root skill that routes to its mode files.
---

<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.0.0 -->
# CI/CD

Root skill for CI/CD work. This `SKILL.md` is the only file loaded up front; it routes to
the mode files in this directory, which you read **on demand** with your file tool. Do
not guess a mode file's contents — read it.

## Modes — start with `core.md`, then add what the task needs
- `core.md` — platform-agnostic principles (load first): stage ordering cheapest-first, reproducibility, secret discipline, artifact immutability, OIDC-preferred credentials, caching, retries, the deploy boundary.
- `github-actions.md` — GitHub Actions: `pull_request` vs `pull_request_target` safety, SHA-pinned actions, `permissions: {}` default, OIDC to AWS/GCP/Azure, reusable workflows, matrix/caching, self-hosted runner trust.
- `azure-devops.md` — Azure DevOps YAML: step/job/stage/`extends` templates, Library + Key Vault, workload identity federation, Environments with checks, resources.
- `gitops.md` — Argo CD and Flux: Application/Kustomization design, ApplicationSet, sync waves, progressive delivery, secret management, RBAC, drift detection, PR promotion.
- `supply-chain.md` — SLSA, SBOM (SPDX/CycloneDX), in-toto/SLSA provenance, Sigstore keyless signing + Rekor, dependency pinning, vuln scanning, signed commits, admission-time verification.

Load the platform file(s) that match the target plus `supply-chain.md` for release
integrity. For per-language CI job sets, see `go/ci.md`, `rust/ci.md`, `python/ci.md`.
