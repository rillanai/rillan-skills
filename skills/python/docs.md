<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.0.0 -->
# Python Documentation Mode

## Purpose
Use this skill when generating, reviewing, or maintaining documentation for a Python codebase.

Apply this skill with:
- `policy.md` for standards that affect public APIs and maintainability
- `workflow.md` for code-grounded, tool-first execution

This mode is for documentation work that materially changes developer understanding: docstrings, package docs, README files, ADRs, API docs, runbooks, changelogs, and onboarding materials.

## Tool Use
This skill is tool-agnostic and works with Claude Code, Codex, OpenCode, and similar assistants. Map its guidance to whatever file-reading, editing, search, and shell-execution tools your environment exposes.

- Read the source before documenting it; do not write documentation from memory or assumption.
- Verify examples by running them when the repository treats them as testable (doctest, pytest examples, Sphinx `doctest-block`).
- Issue independent tool calls (reading multiple files, checking multiple symbols) in parallel.
- When documenting behavior, cite the file and symbol you read — do not paraphrase without grounding.

## When To Use
Use this skill for:
- module and package docstrings
- public class and function docstrings
- README updates
- ADRs
- API docs
- runbooks
- changelogs
- onboarding documentation
- documentation gap analysis

Do not use this skill for:
- logic review
- security audit
- broad architecture grading
- test design work

## Operating Stance
- Read the code before documenting it.
- Document actual behavior, not intent or aspiration.
- Prefer concise, scannable docs over commentary-heavy prose.
- Match the repository's existing docstring and docs style when it is coherent.

## Documentation Rules

### Module And Package Documentation
- Public packages and modules should have module-level documentation.
- `__init__.py` docstrings should describe the package purpose and align with `__all__` when present.
- Describe what the module provides, not implementation detail.

### Public Classes And Functions
- Public classes, functions, methods, and modules should have docstrings.
- Use one docstring convention per project and apply it consistently.
- Default to Google style when no convention exists, unless the project already uses NumPy or Sphinx/reST style.
- Document arguments, returns, yields, raises, side effects, invariants, and examples when they materially help the reader.
- Skip noise docstrings that only restate the signature.

### README Files
A Python project README should usually include:
1. project title and one-line description
2. overview
3. quickstart
4. installation
5. configuration
6. usage
7. API or package reference link
8. development setup
9. architecture overview
10. contributing
11. license

Rules:
- keep examples runnable or clearly marked as pseudocode
- do not document features that do not exist
- update README content when install, config, or project structure changes

### ADRs
Use:
- `Status`
- `Context`
- `Decision`
- `Consequences`
- `Alternatives Considered`

Rules:
- keep ADRs concise
- ground the context in actual project constraints
- supersede accepted ADRs rather than rewriting history

### API Docs
- Keep API docs aligned with real handlers, serializers, schemas, and auth rules.
- Prefer generated or code-near canonical sources when the project uses Sphinx, mkdocs, FastAPI OpenAPI, or similar tooling.
- Document method, path, request, response, status codes, authentication, and authorization.

### Runbooks
Each runbook should cover:
1. purpose
2. prerequisites
3. symptoms
4. diagnosis
5. resolution
6. escalation
7. post-incident checks

Rules:
- use numbered operational steps
- include actual commands or queries
- include expected outputs when practical

### Changelogs
- Follow the repository's changelog format; `Keep a Changelog` is a strong default.
- Group entries under user-visible categories such as `Added`, `Changed`, `Fixed`, and `Security`.
- Do not treat internal cleanup as a public changelog item unless it changes user-visible behavior.

## Evidence Rules
- Verify docs against source, config, tests, and framework definitions.
- Verify defaults from settings or config-loading code, not stale documentation.
- Verify API behavior from code or generated schema, not assumptions.
- Label planned work explicitly as planned or proposed.

## Quality Checklist
Before considering documentation work complete, verify:
- names, paths, imports, and symbols are current
- examples are executable when the repository treats them as testable
- configuration docs match actual defaults
- API docs match real handlers and types
- docs do not describe removed modules, commands, settings, or behavior

## Anti-Patterns To Reject
- stale documentation
- aspirational documentation presented as implemented behavior
- vacuous docstrings
- undocumented public APIs
- orphaned docs with no navigation path
- copy-pasted docstrings that drift from real behavior

## Invocation Template
Use this skill with a prompt that supplies repository-specific context. Example:

```text
Use Python Documentation Mode with Python Engineering Policy and Python Engineering Workflow.
Update package documentation and README content for /path/to/repo/src/myapp/auth.
Ground every statement in code or config.
Verify examples and defaults before finalizing the docs.
```
