<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.0.0 -->
# Python Development Mode

## Purpose
Use this skill when the task is to implement or modify Python code: feature work, bug fixes, targeted cleanup, or small refactors coupled to a concrete behavior change.

This skill is intentionally thin. It is the default implementation mode for Python work and is meant to be used with:
- `policy.md` for what good Python code looks like
- `workflow.md` for how to approach and verify the work

## Skill Use
- Load this skill when the primary task is to change Python code.
- Use this skill together with Python policy and Python workflow when available.
- Treat this skill as the implementation-mode overlay, not as the full Python contract by itself.
- Follow repository-specific commands, framework conventions, test markers, and tooling requirements when they are explicit.

## Tool Use
This skill is tool-agnostic and works with Claude Code, Codex, OpenCode, and similar assistants. Map its guidance to whatever file-reading, editing, search, and shell-execution tools your environment exposes.

- Invoke tools to read code, find references, and run tests; do not describe what you would do.
- Issue independent tool calls in parallel rather than sequentially.
- Run the relevant `pytest`, type-check, or lint commands yourself — do not claim a change is verified without tool output.

## Mode Focus
- Deliver the requested behavior with the smallest viable change.
- Preserve module and package boundaries unless the task explicitly requires changing them.
- Prefer regression tests for bug fixes and failing tests first for new behavior when practical.
- Keep unrelated cleanup out of the patch unless it is necessary for correctness or to avoid making the design worse.

## Implementation Workflow
1. Classify the task and identify the affected entrypoints, modules, packages, and symbols.
2. Use structural tooling and repository evidence to find definitions, references, and boundary crossings before editing.
3. Identify existing tests that cover the behavior and extend them when practical.
4. Make the smallest viable change.
5. Re-run verification proportional to the risk.
6. Re-check blast radius before considering the task complete.

## Default Verification
- At minimum, verify the touched module or relevant test target with the repository-standard test command.
- Use broader test runs for shared modules, framework wiring, persistence, async boundaries, or public APIs.
- Run repository-standard type checking, linting, or static analysis when it is part of normal workflow.
- Reproduce the runtime path when debugging behavior that tests do not cover well enough.

## Completion Criteria
Do not consider an implementation task complete until all applicable items are true:
- the requested behavior was implemented or corrected
- affected tests were added or updated when behavior changed
- verification was run or a concrete verification gap was reported
- no unnecessary abstraction, protocol, or framework surface was introduced
- remaining uncertainty and blast radius were stated explicitly

## Invocation Template
Use this skill with a prompt that supplies repository-specific context. Example:

```text
Use Python Development Mode with Python Engineering Policy and Python Engineering Workflow.
Implement the retry fix in /path/to/repo.
Keep the change scoped to the worker package.
Add or update regression coverage.
Verify the affected test target and any impacted async or integration path.
```
