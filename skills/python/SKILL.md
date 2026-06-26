---
name: python
description: Python engineering — implement, review, refactor, debug, test, document, audit, migrate, or set up CI for Python code. Triggers on `pyproject.toml`, `requirements*.txt`, `.py` files, ruff, mypy/pyright, pytest, hypothesis, uv/poetry, or Python typing/packaging questions. Root skill that routes to its mode files.
---

<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.1.0 -->
# Python

Root skill for Python work. This `SKILL.md` is the only file loaded up front; it routes to
the mode files in this directory, which you read **on demand** with your file tool.
Do not guess a mode file's contents — read it.

## Baseline — load for any non-trivial Python task
- `policy.md` — package/module structure, typing discipline, error handling, configuration, the review bar.
- `workflow.md` — tool-first discovery (LSP, pyright/mypy), the truth hierarchy, verification.

## Modes — load the one matching the task, on top of the baseline
- `dev.md` — implement or modify Python: features, bug fixes, targeted refactors.
- `test.md` — pytest fixtures/parametrize, hypothesis, async testing, mocking, coverage.
- `audit.md` — explicit, phased, evidence-based deep audit. User supplies repo path + phase.
- `docs.md` — docstrings, Sphinx/mkdocs, READMEs, ADRs, runbooks, changelogs.
- `migrate.md` — Python version upgrades, dependency replacement, framework swaps, packaging transitions.
- `ci.md` — CI for a Python project (uv/poetry lockfile integrity, ruff, mypy/pyright, pytest matrix, pip-audit, bandit, OIDC publishing).

Cross-pack pointers use path form, e.g. `cicd/core.md`. Stricter repository-local
conventions win when they are explicit and defensible.
