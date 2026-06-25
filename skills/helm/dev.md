<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.0.0 -->
# Helm Development Guidance

## Purpose
Use this skill when writing, modifying, or reviewing Helm charts. It is the default implementation contract for chart structure, templating, values design, release safety, and operability.

## Skill Use
- Load this skill when the task is to change Helm chart code, chart packaging, or Helm-based deployment workflow.
- Prefer stable chart interfaces over clever templates.
- Match repository conventions for chart layout, release tooling, and values structure when they are explicit.

## Tool Use
This skill is tool-agnostic and works with Claude Code, Codex, OpenCode, and similar assistants. Map its guidance to whatever file-reading, editing, search, and shell-execution tools your environment exposes.

- Invoke tools to read templates, values files, and `Chart.yaml`; do not describe what you would do.
- Run `helm lint` and `helm template` with representative values files to verify rendered output — do not claim a change is safe without executing these.
- Issue independent tool calls in parallel rather than sequentially.
- Before changing helper templates or label selectors, inspect callers across the chart and any subcharts.

## Core Principles
- Keep templates readable. Favor explicit manifests over dense helper indirection.
- Make `values.yaml` the public API. Keep keys predictable, documented, and backwards-compatible unless the task explicitly changes them.
- Fail fast with `required`, schema validation, and defensive defaults when omission would create broken releases.
- Keep Kubernetes concerns visible. Do not hide important probes, resources, security context, or affinity behind confusing helper stacks.

## Default Workflow
1. Inspect `Chart.yaml`, `values.yaml`, `templates/`, and any environment overlays before editing.
2. Identify the public values surface and any compatibility risk.
3. Make the smallest safe change to templates, helpers, or chart metadata.
4. Re-render output and inspect the generated manifests before considering the task done.

## Default Verification
- Run `helm lint` when Helm is available.
- Render with `helm template` using representative values files.
- Check that labels, selectors, names, hooks, and upgrade-sensitive resources remain stable unless intentionally changed.

## Completion Criteria
- Chart output matches the requested behavior.
- Values changes are intentional and documented.
- Upgrade and rollback risk is understood for stateful or identity-bearing resources.
