<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.2.0 -->
# Python Migration Mode

## Purpose
Use this skill when planning or executing migrations in a Python codebase.

Apply this skill with:
- `policy.md` for design and boundary standards
- `workflow.md` for tool-first discovery and verification

This mode exists for migrations that materially change Python version, dependencies, frameworks, packaging, architecture, schema, or API shape and therefore require explicit impact analysis, sequencing, and rollback planning.

## Tool Use
This skill is tool-agnostic and works with Claude Code, Codex, OpenCode, and similar assistants. Map its guidance to whatever file-reading, editing, search, and shell-execution tools your environment exposes.

- Use tools to inventory call sites, imports, and schema references before proposing a migration plan — do not estimate blast radius from memory.
- Issue independent tool calls (listing affected files, reading config, checking CI) in parallel.
- Run `pytest`, type-check, and lint after each migration step and report actual output, not expected output.
- If the migration touches persistence, run schema or data verification against the target state rather than inferring correctness.

## When To Use
Use this skill for:
- Python version upgrades
- dependency replacement or major version upgrades
- framework migrations
- module or architecture restructuring
- sync-to-async transitions
- database migration planning
- API version transitions
- packaging or build-system modernization

Do not use this skill for:
- greenfield implementation
- ordinary code review
- isolated bug fixes without a migration dimension

## Operating Stance
- Prefer reversibility over speed.
- Describe the current system as implemented before proposing the target state.
- Do not migrate and opportunistically redesign at the same time.
- Every migration plan must answer: what can fail, how do we detect it, and how do we roll it back?

## Migration Planning Process
1. Inventory the current surface area: modules, imports, entrypoints, APIs, schemas, frameworks, and consumers.
2. Define the target state and what must remain compatible during transition.
3. Identify blast radius, sequencing constraints, and rollback boundaries.
4. Choose incremental steps that are independently verifiable.
5. Define verification for each step before making changes.
6. Remove temporary compatibility layers once migration is complete.

## Migration Types

### Python Version Upgrades
- inspect Python version declarations in project metadata, CI, Dockerfiles, and local version files
- review release notes and deprecations for each intermediate version
- identify syntax, typing, stdlib, and dependency compatibility changes
- adopt new syntax or language features only after the version upgrade is stable

### Dependency Migrations
- map imports, APIs, exception behavior, and call patterns before swapping
- prefer drop-in replacement only when behavior is truly compatible
- otherwise use adapters or compatibility layers at the consumption boundary
- do not leave dual-dependency coexistence in place indefinitely

### Framework Migrations
- inventory routing, middleware, dependency injection, request parsing, worker behavior, or ORM usage
- verify lifecycle, error handling, auth, and serialization semantics before and after
- keep business logic framework-independent where possible during the migration

### Architecture Migrations
- map current module relationships before moving code
- introduce clearer boundaries incrementally
- verify dependency direction and import health after each meaningful move
- keep temporary re-exports or shims only as long as needed for safe transition

### Database And API Migrations
- separate schema migration from data migration
- plan compatibility and rollout windows explicitly
- preserve backward compatibility unless a deliberate version break is being introduced
- document consumer-facing version and deprecation behavior

### Packaging Migrations
- keep install, build, and publish workflows explicit
- update metadata, lock files, CI, and developer setup together
- verify editable install and test workflow after the packaging change

## Execution Rules
- Keep each migration step small and independently verifiable.
- Run verification after each meaningful step.
- Preserve backward compatibility during the transition when consumers depend on it.
- Use temporary adapters, shims, or dual-path behavior only when the migration truly requires it.
- Remove temporary migration scaffolding after the cutover.

## Verification Expectations
- Verify the affected test target on every step.
- Use broader verification for framework, packaging, API, persistence, or architecture shifts.
- Run repository-standard type checking, linting, or static analysis when part of normal workflow.
- Reproduce runtime paths when migration risk lies in environment or framework behavior rather than pure module structure.

## Anti-Patterns To Reject
- migrating and refactoring simultaneously without need
- changing behavior during a dependency swap unless explicitly intended
- skipping rollback planning
- one-shot big bang migrations when incremental steps are possible
- leaving compatibility layers in place indefinitely
- claiming migration safety without verification evidence

## Invocation Template
Use this skill with a prompt that supplies repository-specific context. Example:

```text
Use Python Migration Mode with Python Engineering Policy and Python Engineering Workflow.
Plan and execute the requests-to-httpx migration in /path/to/repo.
Identify impacted modules, exception behavior, and async implications.
Keep the migration incremental and reversible.
Define verification and rollback for each step.
```
