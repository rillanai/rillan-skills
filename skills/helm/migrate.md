<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.0.0 -->
# Helm Migration Planning

## Purpose
Use this skill when planning Helm chart migrations: chart API changes, values contract changes, dependency upgrades, repository moves, or release-process changes.

## Knowledge-Graph Discovery (When Available)
If a graphify knowledge graph (`graphify-out/`) is present, use `graphify query`/`graphify path` to enumerate affected sites and latent cross-module couplings before sequencing the migration — graph traversal surfaces indirect dependencies that text search misses. Confirm every graph-derived call site or dependency with structural tooling, and run `graphify update .` after each step so the map tracks the migration.

## Tool Use
This skill is tool-agnostic and works with Claude Code, Codex, OpenCode, and similar assistants. Map its guidance to whatever file-reading, editing, search, and shell-execution tools your environment exposes.

- Use tools to inventory charts, subcharts, values keys, and installed releases before proposing a plan — do not estimate blast radius from memory.
- Diff rendered output between old and new values with `helm template` to confirm behavior changes; inspect the diff directly.
- Issue independent tool calls (reading multiple environments, scanning overlays, checking dependencies) in parallel.
- When renaming keys or resources, grep for every caller and overlay before declaring the mapping complete.

## Migration Concerns
- Breaking value key renames and default changes
- Chart dependency version changes and subchart behavior shifts
- Resource renames that can orphan or recreate workloads
- Hook lifecycle changes
- CRD installation and upgrade ordering

## Planning Rules
- Separate compatibility fixes from refactors.
- Identify whether the change is install-safe, upgrade-safe, and rollback-safe.
- Provide an explicit mapping for renamed values and resources.
- Call out one-way changes such as schema tightening, immutable field replacements, or CRD version drops.

## Deliverables
- Current vs. target state
- Compatibility risks
- Ordered migration steps
- Verification steps
- Rollback constraints
