<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.2.0 -->
# Kubernetes Operator Documentation Guidance

## Purpose
Use this skill when documenting Kubernetes operator behavior, CRD contracts, reconciliation semantics, operational procedures, or upgrade guidance.

## Tool Use
This skill is tool-agnostic and works with Claude Code, Codex, OpenCode, and similar assistants. Map its guidance to whatever file-reading, editing, search, and shell-execution tools your environment exposes.

- Read API types, CRD manifests, reconciler code, and status builders before documenting behavior — do not write from memory.
- Verify CRD field defaults and validation by reading the generated `config/crd/bases/*.yaml`, not by paraphrasing code comments.
- Issue independent tool calls (reading types, conditions, webhooks, sample CRs) in parallel.
- Cite the Go type and field path when documenting behavior; stale documentation drifts fastest from generated API surface.

## Scope
- CRD and API documentation
- controller behavior and ownership documentation
- install, upgrade, and rollback guides
- troubleshooting and operational runbooks
- webhook and status semantics documentation

## Rules
- Document the API as users experience it: required fields, defaults, status, conditions, and lifecycle semantics.
- Explain what the controller owns, what it watches, and what side effects it produces.
- Keep examples grounded in actual CRD fields and real reconciliation behavior.
- Treat generated CRDs and manifests as artifacts to explain, not as the only source of documentation.
- Generate the CRD/API reference from source markers (`crd-ref-docs` or `gen-crd-api-reference-docs`) rather than hand-writing field tables — generated reference stays in sync with `*_types.go` and CRD bases. Hand-write only the prose around it.
- Document the **metrics surface**: which custom collectors the controller exposes, the authenticated endpoint (no kube-rbac-proxy), and the RBAC a scraper needs.
- Provide a **conditions reference table**: each condition `type`, its `reason` values, what `True`/`False`/`Unknown` mean, and how `observedGeneration` signals freshness.
- Document **RBAC**: the ClusterRole/Role the controller needs and why, derived from `+kubebuilder:rbac` markers.

> Image/bundle/manifest packaging (distroless base, multi-arch, SBOM, OLM bundle, kustomize/PROJECT-v4 layout) is covered by the `docker`, `cicd`, and `kubernetes` skills — document it there, not here.
