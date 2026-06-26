---
name: kubernetes
description: Kubernetes manifests and Kustomize work — author or review Deployments/Services/Ingress/RBAC/NetworkPolicy/HPA/PDB, kustomize bases and overlays; validation and testing (kubeconform, `kubectl apply --dry-run=server`, conftest/kyverno); deployment docs and runbooks; deep audits; and API-version/controller migrations. Triggers on `kustomization.yaml` or k8s-shaped manifests. Root skill that routes to its mode files.
---

<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.1.0 -->
# Kubernetes

Root skill for Kubernetes manifest work. This `SKILL.md` is the only file loaded up
front; it routes to the mode files in this directory, which you read **on demand** with
your file tool. Do not guess a mode file's contents — read it.

## Modes — load the one matching the task
- `dev.md` — write or modify manifests and Kustomize overlays: selectors, probes, resources, security context, NetworkPolicy, PDBs, rollout strategy.
- `test.md` — kubeconform, `kubectl apply --dry-run=server`, conftest/kyverno, render diffs, deprecated-API checks.
- `audit.md` — explicit, phased, evidence-based audit of workloads, security/RBAC, operability, policy conformance.
- `docs.md` — resource docs, deployment flows, overlay usage, rollout/rollback runbooks.
- `migrate.md` — deprecated API-version transitions, controller swaps, storage/PVC/StatefulSet changes, security-policy adoption.

For chart packaging use the `helm` root; for controllers use the `operator` root.
Cross-pack pointers use path form, e.g. `helm/dev.md`. Stricter repository-local
conventions win when they are explicit and defensible.
