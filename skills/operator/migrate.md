<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.0.0 -->
# Kubernetes Operator Migration Planning

## Purpose
Use this skill when planning operator migrations: CRD version changes, controller refactors, `controller-runtime` upgrades, `achilles-sdk` adoption or updates, webhook introduction, or reconciliation-model changes.

## Knowledge-Graph Discovery (When Available)
If a graphify knowledge graph (`graphify-out/`) is present, use `graphify query`/`graphify path` to enumerate affected sites and latent cross-module couplings before sequencing the migration — graph traversal surfaces indirect dependencies that text search misses. Confirm every graph-derived call site or dependency with structural tooling, and run `graphify update .` after each step so the map tracks the migration.

If no `graphify-out/` directory exists, ignore this section.

## Tool Use
This skill is tool-agnostic and works with Claude Code, Codex, OpenCode, and similar assistants. Map its guidance to whatever file-reading, editing, search, and shell-execution tools your environment exposes.

- Use tools to inventory API versions, stored CRs, owner references, finalizers, and webhook registrations before proposing a plan.
- Run `make generate`, `make manifests`, and `go test`/`envtest` after each migration step and report actual output.
- Issue independent tool calls (reading API types across versions, scanning for conversion webhooks, checking RBAC) in parallel.
- For CRD version flips and storage version changes, confirm the conversion path against code and tests rather than asserting correctness.

## Migration Concerns
- CRD versioning, stored versions, conversion, and compatibility windows
- status and condition contract changes
- finalizer behavior changes
- owner reference or naming changes that can orphan managed resources
- dependency upgrades that alter controller-runtime behavior or generated artifacts

## Planning Rules
- Separate API compatibility work from reconciler cleanup.
- Identify whether existing custom resources continue to reconcile safely after the change.
- Provide ordered steps for generators, manifests, rollout, and rollback.
- Call out one-way changes such as storage version flips, defaulting changes, or conversion assumptions.

## Deliverables
- current vs. target operator model
- compatibility and rollout risks
- ordered migration plan
- verification steps
- rollback constraints
