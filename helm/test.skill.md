---
name: helm-test
description: Use when designing validation or test strategy for Helm charts — `helm lint`, `helm template` regression checks, `values.schema.json`, chart-testing/snapshot/policy assertions, install tests for hooks. Treat CRDs, hooks, and stateful resources as upgrade-risk areas needing explicit coverage. Skip for non-chart Kubernetes manifests.
---

<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 0.2.0 -->
# Helm Test Strategy And Generation

## Purpose
Use this skill when designing validation and test strategy for Helm charts. Focus on linting, rendering, schema validation, manifest assertions, and release regression coverage.

## Tool Use
This skill is tool-agnostic and works with Claude Code, Codex, OpenCode, and similar assistants. Map its guidance to whatever file-reading, editing, search, and shell-execution tools your environment exposes.

- Run `helm lint` and `helm template` directly. A new test or values file is not done until the commands have executed and the output is known.
- When working with chart-testing, snapshot, or policy tooling (kubeconform, conftest), invoke the tools and report their output rather than describing expected behavior.
- Issue independent tool calls (multiple render targets, different values profiles) in parallel.
- Report rendering or lint failures with the exact command and output that produced them.

## Test Layers
- `helm lint` for static chart validation
- `helm template` with representative values for render-time regression checks
- Schema validation for values when `values.schema.json` is present or needed
- Manifest assertions with chart-testing, snapshot tests, or policy checks
- Optional install tests for hooks and live-cluster behavior when static checks are insufficient

## Guidance
- Test the values surface that users actually exercise: defaults, minimum valid config, and one or two realistic environment overrides.
- Prioritize selectors, names, labels, ports, probes, resources, affinity, tolerations, RBAC, and secret wiring.
- Treat CRDs, hooks, and stateful resources as upgrade-risk areas requiring explicit coverage.
- Prefer deterministic rendered-manifest checks before live-cluster tests.

## Completion Criteria
- The changed chart has at least one concrete verification path.
- Breaking values or rendered-manifest changes are called out explicitly.
