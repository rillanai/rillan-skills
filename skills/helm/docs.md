<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.2.0 -->
# Helm Documentation Guidance

## Purpose
Use this skill when writing or updating Helm chart documentation: chart READMEs, values references, upgrade notes, operator runbooks, and release guidance.

## Tool Use
This skill is tool-agnostic and works with Claude Code, Codex, OpenCode, and similar assistants. Map its guidance to whatever file-reading, editing, search, and shell-execution tools your environment exposes.

- Read `Chart.yaml`, `values.yaml`, schemas, and templates before documenting behavior — do not write from memory.
- Verify example commands by running `helm template` or `helm lint` where practical.
- Issue independent tool calls (reading multiple values files, scanning subcharts) in parallel.
- Cite the exact values key or template path when describing behavior.

## Scope
- `README.md` for chart purpose, install flow, and examples
- `values.yaml` reference tables and override examples
- Upgrade and rollback notes for breaking values or resource changes
- Runbooks for rendering, linting, packaging, and release troubleshooting

## Rules
- Treat `values.yaml` as a public contract; document keys that users are expected to set.
- Keep examples realistic and minimal.
- Document defaults, required values, and compatibility caveats explicitly.
- Ground every statement in actual chart behavior. Do not invent flags, hooks, or values.

## Minimum Deliverables
- What the chart deploys
- Required values and important optional values
- Example `helm install` or `helm upgrade` commands
- Upgrade considerations for users carrying existing releases
