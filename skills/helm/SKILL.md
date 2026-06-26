---
name: helm
description: Helm chart work — author or review charts, templates, `values.yaml` and `values.schema.json`; lint/template regression testing; chart docs; deep audits; and chart/API-version migrations. Triggers on `Chart.yaml`, `templates/`, `values.yaml`, or `helm` commands. Root skill that routes to its mode files.
---

<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.1.0 -->
# Helm

Root skill for Helm work. This `SKILL.md` is the only file loaded up front; it routes to
the mode files in this directory, which you read **on demand** with your file tool. Do
not guess a mode file's contents — read it.

## Modes — load the one matching the task
- `dev.md` — chart, template, and packaging implementation work.
- `test.md` — `helm lint`, `helm template` regression checks, `values.schema.json`, chart-testing.
- `audit.md` — explicit, phased, evidence-based audit of chart structure, security, operability.
- `docs.md` — chart READMEs, values references, deployment guides, runbooks.
- `migrate.md` — chart breaking changes, API versions, controller upgrades, rollback planning.

Cross-pack pointers use path form, e.g. `kubernetes/dev.md`, `cicd/gitops.md`. Stricter
repository-local conventions win when they are explicit and defensible.
