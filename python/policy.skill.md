---
name: python-policy
description: Use as the Python standards layer (DRY/KISS/YAGNI, project structure, design priorities, review bar) for any Python implementation, review, refactor, migration, documentation, or audit work. Pair with `python-workflow` for execution discipline. Repo-specific stricter rules win when explicit and defensible.
---

<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 2.1.0 -->
# Python Engineering Policy

## Purpose
Use this skill when Python work needs a clear definition of what good, correct, maintainable engineering looks like.

This skill is the standards layer for Python tasks. It defines coding rules, architectural expectations, review priorities, and quality bars. It does not define the execution workflow; pair it with `workflow.skill.md` for that.

## Skill Use
- Load this skill for Python implementation, review, refactor, migration, documentation, and testing work when standards matter.
- Treat this skill as the default Python quality contract unless the repository has stricter local rules.
- Repository-specific conventions may override this skill when they are explicit, coherent, and defensible.
- When this skill conflicts with casual convenience, follow this skill.

## Tool Use
This skill is tool-agnostic and works with Claude Code, Codex, OpenCode, and similar assistants. Map its guidance to whatever file-reading, editing, search, and shell-execution tools your environment exposes.

- Invoke tools to check standards against real code; do not apply rules from memory to code you have not read.
- Issue independent tool calls in parallel rather than sequentially.
- Verify claims about API shape, typing, or boundaries against the current source before asserting them.

## Core Principles

### DRY
- Extract repeated logic into named functions, methods, or reusable helpers when repetition is real and local.
- Prefer small, clear duplication over brittle abstraction.
- Use shared constants or aliases for meaningful repeated values and strings.

### KISS
- Choose the simplest solution that solves the actual problem.
- Prefer straightforward control flow over indirection.
- Split functions, methods, or modules that mix unrelated responsibilities.

### YAGNI
- Do not add protocols, ABCs, extension points, or config knobs without a real consumer.
- Do not add speculative utility modules.
- Do not introduce plugin or provider patterns for a single implementation.

## Design Priorities
- Make runtime behavior easy to trace from entrypoint to side effect.
- Keep module and package boundaries explicit and directional.
- Separate transport, domain, persistence, framework, and infrastructure concerns when the repository size justifies it.
- Prefer designs that are easy to test, observe, and operate.
- Choose operational clarity over abstraction density.

## Project Structure
- Keep `__main__.py`, application factories, and framework entrypoints thin.
- Put business logic in domain and service modules by default.
- Use `src` layout for applications when import discipline matters.
- Use flat layout for libraries only when the repository intentionally prefers that simplicity.

## Module And Boundary Rules
- Keep modules cohesive with one primary reason to change.
- Prevent inward layers from importing outward layers.
- Do not import transport or framework modules into core business logic without deliberate reason.
- Make boundary crossings explicit in names and types.
- Avoid circular imports by extracting shared concepts or inverting dependencies.

## Protocols, ABCs, And Types
- Define protocols or ABCs where they are consumed, not where they are implemented.
- Keep protocols and ABCs small.
- Accept abstract capabilities and return concrete types when practical.
- Do not introduce protocols or ABCs only to appease mocking.
- Add type hints to public interfaces and important internal boundaries.
- Avoid `Any` in public APIs without a strong reason.

## Data Handling
- Start with types and invariants before implementation detail.
- Validate untrusted input at the boundary closest to ingress.
- Normalize once, then operate on validated forms.
- Be explicit about defaults, optional values, and partial updates.
- Avoid leaking persistence or framework models across unrelated layers unless the tradeoff is deliberate.

## Error Handling
- Raise exceptions; do not hide them.
- Prefer specific exception types.
- Preserve causality with `raise ... from err` when wrapping.
- Handle exceptions at boundaries and avoid duplicate logging.
- Do not use exceptions as ordinary control flow.

## Dependency Injection And Context
- Pass dependencies explicitly through constructors, factories, or function arguments.
- Do not rely on module-level mutable singletons for stateful dependencies.
- Use `contextvars` only for cross-cutting concerns such as request or trace identifiers.
- Keep wiring in the entrypoint or application factory.

## Concurrency And Lifecycle
- Choose the concurrency model deliberately: async, threading, or multiprocessing based on the work.
- Every background task, thread, or worker must have an owner, shutdown path, and error handling strategy.
- Handle cancellation and timeout behavior explicitly.
- Protect shared mutable state deliberately.
- Make retry, idempotency, drain, and shutdown behavior explicit.

## Configuration
- Load configuration into typed, validated structures at startup.
- Keep precedence explicit: CLI args, env vars, config file, defaults.
- Do not scatter `os.environ` reads across the codebase.
- Keep secrets out of logs and out of casual configuration dumping.

## Logging And Observability
- Use structured or consistently machine-readable logging in production code.
- Log at meaningful operation boundaries.
- Include correlation identifiers when available.
- Do not log secrets, tokens, or sensitive payloads by default.
- Write code so metrics and tracing can be added at meaningful boundaries.

## Security
- Treat all external input as untrusted.
- Validate inputs before database, file, template, shell, deserialization, or outbound-network use.
- Do not use `eval`, `exec`, unsafe YAML loaders, or unsafe deserialization on untrusted input.
- Use safe subprocess invocation and explicit authorization checks.
- Be cautious with file paths, redirects, remote fetches, and proxy-like behavior.

## Handlers And Entrypoints
- Keep handlers in a consistent flow: parse, validate, call service, return response.
- Inject dependencies through framework mechanisms or explicit constructors.
- Do not leak raw domain exceptions directly to clients.
- Use middleware or centralized hooks for cross-cutting concerns.

## Graceful Shutdown
- Handle process shutdown explicitly.
- Drain in-flight work before exit when the runtime model allows it.
- Make background work participate in cancellation.
- Log shutdown paths clearly enough for operators to understand them.

## Refactoring Rules
- Refactor toward simpler control flow and clearer boundaries.
- Keep refactors incremental unless a larger redesign is explicitly requested.
- Make the smallest structural change that solves the real problem.
- Leave touched code more testable and observable than you found it.

## Review Standard
Evaluate Python changes in this order:
1. correctness
2. exception and error handling
3. concurrency or async safety
4. API and boundary design
5. testability
6. simplicity
7. performance where it is obvious and material

## Shared Quality Checklist
Before considering Python work complete, verify all applicable items:
- behavior changes have relevant tests or an explicit gap explanation
- concurrency- or async-sensitive changes are verified appropriately
- exceptions are handled at the right level
- no dead code or stray TODOs without tracking context were introduced
- naming follows repository and Python conventions
- exported API surface has the needed documentation
- repository-standard type-checking, linting, or static-analysis issues were not ignored silently

## Anti-Patterns To Reject
- god modules or god classes
- oversized protocols or ABCs
- public APIs built around `Any` or loose dictionaries without strong reason
- global mutable state used as dependency injection
- hidden background tasks with no lifecycle management
- low-level logging that forces duplicate logs upstream
- `eval`, `exec`, or unsafe deserialization on untrusted data
- clever abstractions that obscure control flow

## Invocation Template
Use this skill with a prompt that supplies repository-specific context. Example:

```text
Use Python Engineering Policy with Python Engineering Workflow.
Apply Python standards while reviewing and updating /path/to/repo.
Keep boundaries explicit, avoid speculative abstractions, and prioritize correctness over convenience.
```
