---
name: operator
description: Kubernetes operators in Go (kubebuilder, controller-runtime, achilles-sdk) — reconcilers, CRD/API types, webhooks, finalizers, status/conditions, watch wiring, manager setup; plus operator testing (envtest, fake-client), docs, deep audits, and CRD/controller-runtime migrations. Triggers on `api/*/*_types.go`, controllers, or `sigs.k8s.io/controller-runtime`. Root skill that routes to its mode files.
---

<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.1.0 -->
# Operator

Root skill for Kubernetes operator work. This `SKILL.md` is the only file loaded up
front; it routes to the mode files in this directory, which you read **on demand** with
your file tool. Do not guess a mode file's contents — read it.

This stack is a **specialization of Go development**, not a replacement. For general Go
quality and execution discipline, load `go/policy.md` and `go/workflow.md` as the
baseline; the mode files below add controller-specific rules. When operator guidance
conflicts with generic Go guidance, follow the operator file for controller concerns.

## Modes — load the one matching the task
- `dev.md` — API types, reconcilers, watch wiring, finalizers, status/conditions, webhooks, manager setup.
- `test.md` — reconciler unit tests, envtest, fake-client, webhook validation/defaulting, CRD generation checks.
- `audit.md` — explicit, phased, evidence-based audit of reconciler correctness, CRD/webhook/status models, lifecycle safety.
- `docs.md` — CRDs/API contracts, reconciliation semantics, ownership, status/conditions, install/upgrade/rollback guides.
- `migrate.md` — CRD version flips, storage-version changes, conversion webhooks, controller-runtime upgrades, achilles-sdk adoption.

Centered on kubebuilder, controller-runtime, and achilles-sdk; `operator-sdk` is not the
primary workflow — preserve existing conventions if a repo already uses it.
