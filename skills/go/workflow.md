<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.3.0 -->
# Go Engineering Workflow

## Purpose
Use this execution contract to understand, modify, debug, refactor, review, or verify Go code. Pair it with `policy.md` and a matching task mode when needed. User decisions and explicit repository conventions take precedence.

## Repository Workflow Discovery
Before choosing commands, inspect the relevant repository signals:
- `go.mod`
- `go.work` and module boundaries; scope `./...` correctly in multi-module repositories
- `Makefile`, `Taskfile.yml`, `magefiles`, or scripts
- CI, lint, and static-analysis configuration
- generators, build tags, and integration-test conventions

Use repository commands when present. Do not silently bypass required tags, generation steps, or package boundaries.

## Evidence And Discovery
Use evidence suited to the claim. User intent defines the requested behavior; repository policy defines local constraints; code and configuration show implementation; compiler, tests, analyzers, and runtime signals show different aspects of correctness. Resolve conflicts explicitly rather than treating one source as universally authoritative.

- Start from the relevant entrypoint or symbol and trace definitions, references, interfaces, registrations, and callers with `gopls` or an equivalent structural tool.
- Use `go list` for package and module boundaries. If structural tooling is unavailable, fall back to `go list`, `go doc`, `rg`, and careful file tracing.
- Read tests and configuration that define the behavior. Run commands instead of predicting their results.
- Batch independent discovery and verification when the environment supports it.
- Treat inference as a lead. State what remains uncertain when evidence is incomplete.

If a maintained `graphify-out/` graph exists and the task asks about relationships or blast radius, it can provide leads through `graphify query`, `path`, or `explain`. Confirm graph-derived claims against source or executable evidence. Its presence alone does not require a query or update.

## Planning Depth
Follow an established repository pattern for routine changes. For work with unresolved design choices, identify constraints and compare the viable options before implementation. Reuse decisions the user or repository has already made.

Ask for input only when missing information would materially change scope or correctness, or when a consequential decision remains unresolved; continue independent work that does not depend on the answer.

Use `planning` when decomposition or rollback boundaries are genuinely unclear. Use `adr` for a consequential decision that should be recorded, `rfc` when broader discussion is needed, and `socratic` when the remaining tradeoffs need structured challenge. Apply only the tools warranted by the decision; a new symbol, package, or dependency does not by itself require a ceremony.

## Execute By Intent

- **Understand:** trace the relevant path across package boundaries and explain it from evidence.
- **Modify or add:** identify consumers and tests, make the smallest coherent change, and preserve public contracts unless the request changes them.
- **Debug:** reproduce the failure when practical, trace the earliest divergence, fix the cause, and add regression coverage when it adds durable value.
- **Refactor:** establish behavior coverage, make reviewable steps, and stop when the stated structural problem is solved.
- **Review:** inspect the diff, affected callers, failure paths, API impact, and available verification; report actionable findings first.
- **Verify:** choose checks from the claimed behavior and risk, run them, and report the result and any gap.

## Verification
Prefer the repository's own commands. Otherwise start with the affected package: `go test ./path/to/package` and `go build` where compilation or binary wiring matters. Broaden verification when shared packages, public APIs, persistence, transport, dependency wiring, build tags, or multiple modules increase blast radius.

Add checks for the risk involved:

- `go test -race` for concurrency, shared state, lifecycle, cancellation, or shutdown changes
- integration or runtime-path checks for external boundaries and behavior unit tests cannot establish
- repository-standard `go vet`, `staticcheck`, or `golangci-lint` where configured
- `govulncheck` for security- or dependency-relevant changes
- required generators followed by a check of generated diffs

Do not claim a check passed without its output. If a command cannot run, give the concrete reason and pending gap. Finish with the implemented or observed result, checks run, and material uncertainty; omit process narration that does not help the user assess the result.
