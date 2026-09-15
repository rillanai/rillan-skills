---
name: go
description: Go engineering — implement, review, refactor, debug, test, document, audit, migrate, or set up CI for Go code. Triggers on `.go` files, `go.mod`, gopls, `go test`/`vet`/`build`, table-driven tests, govulncheck, golangci-lint, or Go project structure. Root skill that routes to its mode files.
---

<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.3.0 -->
# Go

Root skill for Go work. This `SKILL.md` is the only file loaded up front; it routes to
the mode files in this directory, which you read **on demand** with your file tool.
Do not guess a mode file's contents — read it.

## Baseline — load for any non-trivial Go task
- `policy.md` — what good Go looks like: structure, errors, concurrency, the review bar.
- `workflow.md` — how to discover, execute, verify, and report with evidence proportional to risk.

## Modes — load a matching mode when the task needs one, on top of the baseline
- `dev.md` — implement or modify Go: features, bug fixes, targeted refactors, cleanup tied to behavior.
- `test.md` — design, write, review, or refactor Go tests (table-driven, race, fuzz, benchmarks, golden files).
- `audit.md` — explicit, phased, evidence-based deep audit.
- `docs.md` — godoc, package docs, READMEs, ADRs, runbooks, changelogs.
- `migrate.md` — Go version upgrades, dependency/framework swaps, package or service extraction.
- `ci.md` — CI for a Go project (lint, vet, race-test, parallel/shuffled tests, fuzz, coverage, vuln, static analysis, build-time discipline, release).

Load another mode only when the work actually crosses that boundary; a development
task does not automatically require the test, docs, migration, audit, or CI mode.
The baseline alone covers understanding, investigation, ordinary review, and
verification. Use `dev.md` when debugging requires code changes. Reserve `audit.md`
for an explicitly requested deep audit.
Cross-pack pointers use path form, e.g. `cicd/core.md`. Explicit repository-local
conventions and user decisions govern the work.
