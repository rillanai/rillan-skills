---
name: python-ci
description: Use when designing or modifying CI for a Python project — `.github/workflows/*.yml`, `azure-pipelines.yml`, `tox.ini` / `noxfile.py`. Defines the lockfile-integrity / lint / type-check / test / coverage / security-scan / build / publish job set. Default toolchain is uv + ruff + mypy or pyright + pytest + nox + OIDC trusted publishing. Pair with `cicd-*` skills and `python-test` / `python-policy`.
---

<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 0.1.0 -->
# Python CI Mode

## Purpose
Use this skill when designing, writing, or refactoring CI for a Python project. It defines the job set, toolchain, and conventions that produce high-signal, reproducible Python pipelines.

This skill is the Python-specific CI layer. Pair it with `cicd-core` for platform-agnostic principles, a platform-specific skill (`cicd-github-actions` or `cicd-azure-devops`) for wiring, `cicd-supply-chain` for release integrity, and `python-test` / `python-policy` for test-shape and quality rules.

## Skill Use
- Load this skill when the task involves authoring or modifying CI for a Python codebase.
- Treat this skill as the governing contract for Python-specific CI jobs (lint, type check, test, coverage, security scan, build, publish).
- Keep project-specific toolchain preferences (uv vs. poetry vs. pip, ruff vs. black+isort, mypy vs. pyright, tox vs. nox) in the invoking prompt.

## Tool Use
This skill is tool-agnostic and works with Claude Code, Codex, OpenCode, and similar assistants. Map its guidance to whatever file-reading, editing, search, and shell-execution tools your environment exposes.

- Read `pyproject.toml`, `uv.lock` / `poetry.lock` / `requirements*.txt`, `tox.ini` / `noxfile.py`, and existing CI workflows before proposing changes. Python CI typically wraps project-defined tasks.
- Run jobs locally first: `uv run pytest`, `nox -s tests`, `tox`. A CI change that is never executed locally is a CI change that will fail on the runner.
- Use `actionlint`/`az pipelines validate` for the workflow, and run each job's tool directly to verify.
- Issue independent tool calls (reading multiple config files, checking lockfiles, verifying tool versions) in parallel.

## Opinionated Default Toolchain
This is the default toolchain for a modern Python project. Substitute only when the project has a clear reason.

- **Python version pinning**: `.python-version` (read by `uv`, `pyenv`, `mise`) plus `requires-python` in `pyproject.toml`. CI reads one of these — never hardcodes a version.
- **Package manager**: [`uv`](https://docs.astral.sh/uv/) as the default for new projects (fast, lockfile-native, venv-aware). `poetry` or `pip-tools` are fine substitutes when already in use.
- **Lint + format**: [`ruff`](https://docs.astral.sh/ruff/) for both. One tool replaces `flake8`, `pyflakes`, `pycodestyle`, `isort`, `black`. Config in `pyproject.toml` under `[tool.ruff]`.
- **Type check**: `mypy` or `pyright`. Pick one per project; both are fine, `pyright` is faster and stricter by default.
- **Tests**: `pytest` (use the python-test skill for the test-shape rules). Plugins: `pytest-cov`, `pytest-xdist` for parallel, `pytest-asyncio` when async.
- **Coverage**: `pytest --cov` plus `coverage.py`. Enforce per-module thresholds.
- **Security / vulnerabilities**: `pip-audit` for installed dependencies; `bandit` for static source analysis. Optional: `semgrep` for pattern-based rules.
- **Packaging**: PEP 517/518 build via `hatchling` / `setuptools` / `flit`. Build with `python -m build` or `uv build`.
- **Publishing**: OIDC-based "trusted publishing" to PyPI (no API token). `pypa/gh-action-pypi-publish` for Actions.
- **Matrix runner**: `nox` (prefer) or `tox` for running the full job set across Python versions.
- **Pre-commit**: `pre-commit` for local guardrails; run the same hooks in CI for consistency.

Pin every tool's version explicitly. Unpinned `pip install ruff` is non-deterministic.

## Pipeline Job Set
A full Python CI pipeline usually has these jobs. Skip a job only when it demonstrably doesn't apply, and say so.

### 1. Lockfile Integrity
Confirm the lockfile matches `pyproject.toml`:

```yaml
- name: Check lockfile
  run: uv lock --check       # or: poetry lock --check
```

This fails the pipeline if someone edited `pyproject.toml` without regenerating the lockfile — a very common bug.

### 2. Lint And Format (ruff)
One job; ruff handles both:

```yaml
- name: Ruff lint
  run: uv run ruff check .
- name: Ruff format check
  run: uv run ruff format --check .
```

Rules:
- Do not `ruff format --write` in CI. Format drift is a PR fix, not a CI auto-commit.
- Rule selection in `[tool.ruff.lint]` is the source of truth; do not override in the CI step.
- Consider enabling rule sets `E`, `F`, `W`, `I` (import sorting), `B` (bugbear), `SIM`, `UP` (pyupgrade), `RUF` as a baseline.

### 3. Type Check (mypy / pyright)
```yaml
# mypy
- name: Type check
  run: uv run mypy .

# or pyright
- name: Type check
  run: uv run --with pyright pyright
```

Rules:
- `strict = true` (mypy) or the equivalent pyright strictness where practical. Loose typing produces unhelpful CI signal.
- Add `# type: ignore[<code>]` with the specific code, never bare `# type: ignore`.
- For codebases migrating into typing, set a strict baseline and `mypy --strict` only on annotated packages via `[[tool.mypy.overrides]]`.

### 4. Tests (Python Matrix)
Run tests across supported Python versions:

```yaml
test:
  runs-on: ${{ matrix.os }}
  strategy:
    fail-fast: false
    matrix:
      os: [ubuntu-24.04]
      python: ["3.11", "3.12", "3.13"]
      include:
        - os: macos-latest
          python: "3.13"
        - os: windows-latest
          python: "3.13"
  steps:
    - uses: actions/checkout@<sha>
    - uses: astral-sh/setup-uv@<sha>
      with:
        enable-cache: true
    - name: Set up Python
      run: uv python install ${{ matrix.python }}
    - name: Install
      run: uv sync --frozen --all-extras
    - name: Tests
      run: uv run pytest --cov --cov-report=xml -n auto
    - name: Coverage check
      if: matrix.os == 'ubuntu-24.04' && matrix.python == '3.13'
      run: uv run coverage report --fail-under=<threshold>
```

Rules:
- Test on every supported Python version. Dropping an older version is a breaking change; dropping 3.13 support because you forgot to test is a bug.
- Matrix OS coverage is for cross-platform libraries; pure-server projects can stick to Linux + one macOS spot check.
- `fail-fast: false` — knowing which cell failed is the point.
- Enforce coverage thresholds on a single cell (usually Linux + latest Python). Cross-OS coverage variance isn't meaningful signal.
- `-n auto` (`pytest-xdist`) parallelizes on CPU count; drop for test suites with shared-state issues.

### 5. Async / Integration Tests
If the project has async tests or integration tests with real services, split them into a separate job:

```yaml
- name: Integration tests
  env:
    DATABASE_URL: postgres://postgres:postgres@localhost:5432/test
  run: uv run pytest tests/integration -v
```

Use `services:` (Actions) or service containers (ADO) for Postgres, Redis, etc. Run on Linux only unless the integration genuinely depends on OS-specific behavior.

### 6. Security Scans
Run every PR:

```yaml
- name: pip-audit
  run: uv tool run pip-audit -r <(uv export --frozen --format requirements-txt)

- name: bandit
  run: uv tool run bandit -r src/ -q
```

Rules:
- `pip-audit` reports CVEs for installed dependencies. Fail the pipeline on High/Critical with an available fix.
- `bandit` catches common source-level issues (hardcoded secrets, `shell=True`, `eval`). Tune `skips` in `[tool.bandit]` for known false positives.
- Add `trufflehog` or `gitleaks` as a secret scanner — catches credentials committed to the repo.

### 7. Build
Build every PR to catch packaging regressions before release:

```yaml
- name: Build
  run: uv build
- name: Validate distribution
  run: uv tool run twine check dist/*
```

`twine check` validates PyPI-render-ability (README, classifiers, metadata).

### 8. Pre-commit (optional, redundant-but-useful)
Run the project's pre-commit hooks in CI to catch anything the author skipped:

```yaml
- uses: pre-commit/action@<sha>
```

This overlaps with ruff/mypy jobs; run it only if the project uses hooks the other jobs don't cover.

### 9. CI Summary
Publish a summary of all job results on main-branch pushes. Same pattern as Go CI — see `go-ci`.

## Release Pipeline
A separate pipeline, triggered on tags, handles signed releases to PyPI. It adds:

- Full build
- SBOM generation (CycloneDX via `cyclonedx-py` or syft)
- Sigstore keyless signing of wheels + sdist
- PyPI upload via OIDC trusted publishing (no API token)
- GitHub Release creation with changelog

### PyPI OIDC Trusted Publishing
Configure in PyPI (project settings → trusted publishers) and the workflow:

```yaml
release-pypi:
  runs-on: ubuntu-24.04
  environment: pypi
  permissions:
    id-token: write    # required for OIDC
    contents: read
  steps:
    - uses: actions/checkout@<sha>
    - uses: astral-sh/setup-uv@<sha>
    - run: uv build
    - uses: pypa/gh-action-pypi-publish@<sha>
      # No token needed; OIDC-authenticated
```

This replaces long-lived PyPI API tokens entirely — the gold standard for Python publishing. See `cicd-supply-chain` for signing and provenance.

## Python Version Strategy
- Declare supported Python versions in `pyproject.toml`: `requires-python = ">=3.11"`.
- Test the declared range in CI. If you support 3.11+, run 3.11, 3.12, 3.13.
- Drop a version deliberately: update `requires-python`, bump major version, announce in CHANGELOG.
- For libraries, stay conservative: support ≥2 current Python versions. For internal services, pin to one.

## Dependency Declaration
- `pyproject.toml` is the source of truth for dependencies.
- Lockfile (`uv.lock`, `poetry.lock`, `requirements*.txt`) is committed and regenerated via CI or pre-commit hook.
- Separate optional dependency groups for dev, test, docs, etc.:

```toml
[project]
dependencies = ["httpx>=0.27", "pydantic>=2.0"]

[dependency-groups]
dev = ["ruff", "mypy", "pytest", "pytest-cov", "pytest-xdist"]
docs = ["sphinx", "myst-parser"]
```

(`[dependency-groups]` is PEP 735, supported by uv and picked up by most modern tools.)

## Caching
- `astral-sh/setup-uv` or `actions/setup-python` built-in cache is correct; don't roll your own.
- Cache key must include the lockfile hash.
- Do not cache the built virtualenv across unrelated runs; re-solve from lockfile each time.

## Dependencies (Dependabot / Renovate)
- Enable `pip`, `uv`, or `poetry` ecosystem depending on what the project uses.
- Group dev dependencies so they don't flood the PR queue.

```yaml
# .github/dependabot.yml (uv example)
version: 2
updates:
  - package-ecosystem: "uv"
    directory: "/"
    schedule:
      interval: "weekly"
    groups:
      dev-deps:
        dependency-type: "development"
  - package-ecosystem: "github-actions"
    directory: "/"
    schedule:
      interval: "weekly"
```

## Runner Pinning
- Pin Ubuntu (`ubuntu-24.04`) rather than `ubuntu-latest`.
- Pin the action/task SHAs (see `cicd-github-actions` / `cicd-azure-devops`).

## Anti-Patterns To Reject
- Hardcoded Python version in the workflow instead of reading `.python-version` or `requires-python`
- Unpinned tool installs (`pip install ruff` without a version)
- `pytest` job without coverage or vuln scan in the pipeline
- Auto-formatting commits pushed from CI (defeats signed-commit / DCO flow)
- Skipping type checking because "it's too strict" — fix the types, not the job
- Integration tests running on every OS / Python cell in the matrix
- Coverage thresholds applied as a blanket global percentage
- `fail-fast: true` on the test matrix (hides per-version failures)
- Using long-lived PyPI API tokens when OIDC trusted publishing is available
- `pip install --upgrade pip` without pinning — CI reproducibility fails silently
- `pyproject.toml` and lockfile drift (no lockfile integrity check)
- Separate lint jobs for `black`, `isort`, `flake8`, `pyupgrade` when ruff does all of them
- Tests that depend on network state without `pytest.mark.network` or similar gating
- CI summary absent — failure location not scannable from PR view

## Completion Criteria
Do not consider a Python CI task complete until all applicable items are true:
- Python version comes from `.python-version` / `requires-python`, not the workflow
- every tool is version-pinned (ruff, mypy/pyright, pytest, uv/poetry)
- lockfile integrity check runs before any job installs dependencies
- lint, type-check, tests, coverage, vuln scan, and build jobs exist and short-circuit correctly
- tests run on a meaningful Python version matrix; integration tests are scoped and resource-aware
- coverage thresholds are enforced per module or cell
- Dependabot / Renovate is configured
- release pipeline uses OIDC trusted publishing; no long-lived PyPI tokens
- CI summary is published on main

## Invocation Template
Use this skill with a prompt that supplies repository-specific context. Example:

```text
Use Python CI Mode together with cicd-github-actions.
Add a CI workflow to /path/to/repo/.github/workflows/ci.yml:
- Lockfile integrity (uv lock --check)
- Ruff lint + format check
- Pyright type check
- Tests on Python 3.11/3.12/3.13, Ubuntu-24.04 primary plus macOS/Windows spot checks on 3.13
- Coverage enforced on Linux + 3.13
- pip-audit, bandit, gitleaks
- uv build + twine check
Configure Dependabot for uv + github-actions. Set up OIDC trusted publishing to PyPI in a separate release workflow.
```
