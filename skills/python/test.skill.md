---
name: python-test
description: Use when designing, writing, reviewing, or refactoring Python tests — pytest unit/integration/e2e, async, fixtures, coverage analysis, regression tests, property-based, doctest, snapshot/golden-file. Run the tests you write before claiming done. Pair with `python-policy` and `python-workflow`. Skip for production-code review or documentation.
---

<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 2.1.0 -->
# Python Test Mode

## Purpose
Use this skill when designing, writing, reviewing, or refactoring tests for a Python codebase.

Apply this skill with:
- `policy.skill.md` for standards that define maintainable Python code
- `workflow.skill.md` for tool-first execution and verification discipline

This mode exists because test work has its own decision rules, fixtures, async patterns, mocking tradeoffs, and failure modes.

## Tool Use
This skill is tool-agnostic and works with Claude Code, Codex, OpenCode, and similar assistants. Map its guidance to whatever file-reading, editing, search, and shell-execution tools your environment exposes.

- Run the tests you write. A new or updated test is not done until `pytest` (or the repo's runner) has executed it and the result is known.
- Use structural tools to locate the code under test and its callers; do not guess at signatures or coverage.
- Issue independent tool calls (test discovery, coverage reads, symbol lookups) in parallel.
- Report flaky or async failures with the command and output that produced them, not paraphrased.

## When To Use
Use this skill for:
- writing new tests
- designing a test strategy
- analyzing coverage gaps
- refactoring flaky or brittle tests
- writing regression tests for bugs
- adding integration, e2e, property-based, doctest, snapshot, or golden-file coverage
- reviewing test quality

Do not use this skill for:
- broad production-code review outside test concerns
- documentation work
- security audit work

## Test Philosophy
- Test behavior, not implementation details.
- Test at the right level for the risk.
- Treat tests as production code.
- Prefer deterministic evidence over incidental timing or environment luck.
- A failing test is useful only when it fails for the right reason.

## Choosing Test Type

### Unit Tests
Use for:
- branching logic
- validation
- parsing and transformation
- boundary cases
- error-path coverage

### Parametrized Tests
Use for:
- many input-output combinations
- validation matrices
- parser and formatter coverage
- systematic boundary expansion

### Integration Tests
Use for:
- database behavior
- HTTP or external-service integration
- queue, cache, or file-system interaction
- multi-module wiring

### End-To-End Tests
Use sparingly for:
- critical request flows
- smoke coverage
- deploy-time or contract confirmation

### Property-Based Tests
Use for:
- invariants
- parsers and serializers
- validators
- edge spaces that are hard to enumerate manually

### Doctest, Snapshot, And Golden Files
Use for:
- executable documentation examples
- complex rendered output
- report formatting
- contract-like textual output

## pytest Patterns
- Prefer pytest as the default testing framework unless the repository clearly uses another one.
- Use fixtures for setup and teardown, not sprawling helper state.
- Keep `conftest.py` focused and local to the scope that needs it.
- Use markers to separate slow, integration, e2e, or environment-dependent tests.
- Prefer fakes and stubs over over-mocking.

## Mocking And Async Rules
- Patch where the symbol is looked up, not where it was originally defined.
- Do not mock what you do not need to control or observe.
- Use async-aware mocks and async test tooling for async code.
- Test async cancellation, timeout, and lifecycle behavior when code depends on it.
- Do not use time-based sleeps as the primary synchronization mechanism.

## Coverage And Verification
- Coverage quality matters more than percentage.
- Prioritize branches, failure paths, boundary cases, async cancellation, and framework edge behavior.
- Use integration coverage where the risk lies in wiring or external interaction.
- Treat a missing regression test after a bug fix as a gap unless there is a concrete reason it cannot be added.

## Typical Commands
When the repository does not define a stricter workflow, prefer:
- `pytest`
- `pytest path/to/tests_or_module`
- `pytest -m "not slow"`
- `pytest --cov`
- `pytest --maxfail=1 -q` for tight debug loops

Add repository-standard tools such as:
- `coverage`
- `hypothesis`
- `pytest-asyncio`
- `pytest-xdist`

when the project already uses them.

## Anti-Patterns To Reject
- flaky time-based synchronization
- excessive mocking
- tests coupled tightly to private implementation details
- no regression coverage after bug fixes
- snapshot or golden updates without reviewing the diff
- treating high coverage as proof of meaningful quality
- tests that leak global state or environment assumptions without cleanup

## Invocation Template
Use this skill with a prompt that supplies repository-specific context. Example:

```text
Use Python Test Mode with Python Engineering Policy and Python Engineering Workflow.
Add regression coverage for the retry logic in /path/to/repo/src/myapp/worker.py.
Prefer pytest and parametrized cases.
Cover cancellation, invalid payloads, and retry exhaustion.
Verify with the relevant test target and any async tooling already used by the repo.
```
