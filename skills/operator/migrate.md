<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.1.0 -->
# Kubernetes Operator Migration Planning

## Purpose
Use this skill when planning operator migrations: CRD version changes, controller refactors, `controller-runtime` upgrades, `achilles-sdk` adoption or updates, webhook introduction or replacement (including validating-webhook → `ValidatingAdmissionPolicy`), or reconciliation-model changes.

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
- admission-control modernization: moving CEL-expressible validating webhooks to `ValidatingAdmissionPolicy` (GA 1.30) or CRD `x-kubernetes-validations`, and minimum cluster-version requirements that gate it

## Validating Webhook → ValidatingAdmissionPolicy Migration
When replacing a validating webhook with a VAP (the preferred direction for CEL-expressible validation — see `dev.md`):
- Confirm the **minimum supported cluster version is ≥ 1.30** (VAP GA). If older clusters must be supported, keep the webhook until the floor moves, or run both during the window.
- Translate each webhook check into CEL; identify any check that is **not** CEL-expressible (external calls, stateful Go logic) and leave those in a webhook. A partial migration is normal.
- Roll out **observe-only first**: ship the `ValidatingAdmissionPolicy` with `validationActions: [Audit, Warn]` (or `failurePolicy: Ignore`) and compare its decisions against the live webhook before switching to `Deny`.
- Sequence the cutover so validation is never dropped: policy enforcing **before** the webhook is removed, not after. Ship the `ValidatingAdmissionPolicyBinding` and any `paramRef` resources with the policy, and grant the API server read access to params.
- Rollback constraint: re-enabling the webhook must restore identical enforcement; keep the webhook manifests until the VAP has been enforcing cleanly across a full upgrade/restore cycle.

## Common controller-runtime Breaking Changes
When bumping `controller-runtime`, expect and check for:
- **Untyped → typed builder/sources:** `source.Kind(cache, &T{})` → `source.Kind[client.Object](cache, &T{}, handler, preds...)`; `handler.EnqueueRequestForObject{}` → `handler.TypedEnqueueRequestForObject[T]()`; reconcilers to `reconcile.TypedReconciler[T]`.
- **`component-config` removal:** the `ComponentConfig`/`config.Controller` file-based manager config is gone — move those settings into `ctrl.Options` in code.
- **Manager `Options` restructuring:** `MetricsBindAddress`/`Port`/`Host`/`CertDir` moved into `Metrics metricsserver.Options` and `WebhookServer webhook.NewServer(webhook.Options{...})`; update both call sites.
- **`cluster.Cluster` changes:** cache/client accessor and `NewClientFunc` signatures shifted; re-check any custom cluster or multi-cluster wiring against the new interfaces.

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

> Image/bundle/manifest packaging migrations (distroless, multi-arch, SBOM, OLM bundle, kustomize/PROJECT-v4 layout) belong to the `docker`, `cicd`, and `kubernetes` skills, not this one.
