<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.2.0 -->
# Go CI Mode

## Purpose
Use this skill when designing, writing, or refactoring CI for a Go project. It defines the job set, toolchain, and conventions that produce high-signal, reproducible Go pipelines.

This skill is the Go-specific CI layer. Pair it with `cicd/core.md` for platform-agnostic principles, a platform-specific skill (`cicd/github-actions.md` or `cicd/azure-devops.md`) for wiring, `cicd/supply-chain.md` for release integrity, and `test.md` / `policy.md` for test-shape and quality rules.

## Skill Use
- Load this skill when the task involves authoring or modifying CI for a Go codebase.
- Treat this skill as the governing contract for Go-specific CI jobs (lint, vet, test, coverage, race, vuln, static analysis, benchmark, release build).
- Keep project-specific toolchain preferences (Taskfile vs. Makefile vs. mage, Ginkgo vs. stdlib testing, goreleaser vs. hand-rolled) in the invoking prompt.

## Tool Use
This skill is tool-agnostic and works with Claude Code, Codex, OpenCode, and similar assistants. Map its guidance to whatever file-reading, editing, search, and shell-execution tools your environment exposes.

- Read `go.mod`, `Taskfile.yml` / `Makefile`, existing CI workflows, and the repository's tool pins before proposing changes. Go CI is usually a thin wrapper over project-defined tasks.
- Run the jobs locally first: `task ci`, `make ci`, or the equivalent. A CI change that is never executed locally is a CI change that will fail on the runner.
- Use `actionlint` / `az pipelines validate` to validate the workflow; use `go vet`, `go test -race`, `staticcheck`, `govulncheck` to validate the jobs.
- Issue independent tool calls (reading workflows, taskfiles, go.mod, coverage scripts) in parallel.

## Opinionated Default Toolchain
This is the default toolchain for a modern Go project. Substitute only when the project has a clear reason.

- **Go toolchain pinning**: `go.mod`'s `go` directive (e.g., `go 1.26`) is the source of truth; CI uses `actions/setup-go@...` with `go-version-file: go.mod` — never a hardcoded version in the workflow.
- **Task runner**: [Task](https://taskfile.dev/) (`Taskfile.yml`), invoked via the Go `tool` directive: `go -C tools tool task <name>`. This avoids a separate Task install and keeps the runner version pinned in `tools/go.mod`.
- **Lint**: `golangci-lint` (via `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@<pinned-version>`). Config in `.golangci.yml` or `.golangci.yaml`.
- **Format**: `goimports` (`go run golang.org/x/tools/cmd/goimports@<pinned-version>`) + `go fmt`.
- **Static analysis**: `staticcheck` (`go run honnef.co/go/tools/cmd/staticcheck@<pinned-version>`).
- **Vulnerability scan**: `govulncheck` (`go run golang.org/x/vuln/cmd/govulncheck@<pinned-version>`).
- **Tests**: stdlib `testing` by default; Ginkgo when the project already uses it. Always run with the race detector in CI, parallelized (`t.Parallel()` by default — see `test.md`), and with `-shuffle=on`.
- **Fuzzing**: native Go fuzzing (`go test -fuzz`). Seed corpus runs as part of the normal test job; time-boxed active fuzzing runs as its own scheduled job (see job 7).
- **Coverage**: `go test -coverprofile=coverage.out ./...` with per-package thresholds enforced by a script.
- **Release**: [goreleaser](https://goreleaser.com/) for cross-platform builds, archives, and artifact publishing.
- **Versioning**: [svu](https://github.com/caarlos0/svu) for computing the next semantic version from conventional commits.
- **License / attribution**: REUSE tool (`pipx run reuse lint`) when SPDX headers are required.
- **Developer Certificate of Origin**: verify `Signed-off-by:` trailers on PR commits when DCO is required.

Pin every `go run` tool version explicitly; floating `@latest` defeats reproducibility.

## Pipeline Job Set
A full Go CI pipeline usually has these jobs. Skip a job only when it demonstrably doesn't apply, and say so.

### 1. DCO Check (if required)
For projects that require the Developer Certificate of Origin, verify `Signed-off-by:` on every PR commit.

```yaml
# GitHub Actions
- name: Verify Signed-off-by on PR commits
  if: github.event_name == 'pull_request'
  env:
    BASE_REF: ${{ github.base_ref }}
  run: |
    set -Eeuo pipefail
    git fetch --no-tags --prune origin "${BASE_REF}:${BASE_REF}"
    missing=0
    while IFS= read -r commit; do
      [ -z "$commit" ] && continue
      if ! git log -1 --pretty=%B "$commit" | grep -qi '^Signed-off-by: '; then
        echo "Missing Signed-off-by trailer: $commit"
        missing=1
      fi
    done < <(git rev-list --no-merges "origin/${BASE_REF}..HEAD")
    [ "$missing" -eq 0 ] || { echo "DCO check failed. Rebase/amend with: git commit --signoff"; exit 1; }
```

### 2. License / REUSE Compliance (if required)
If the project carries SPDX headers and uses REUSE:

```yaml
- name: REUSE lint
  run: pipx run reuse lint
```

Or via a task: `task reuse-lint`.

### 3. Lint (golangci-lint)
```yaml
- name: Lint
  run: go -C tools tool task lint
```

Rules:
- Run golangci-lint at the workflow's first fast-feedback tier — failures here should short-circuit more expensive jobs.
- `.golangci.yml` is the source of truth. Do not override enabled linters in the CI step.
- Treat `govet`, `staticcheck`, `errcheck`, `ineffassign`, `revive`, and `gosec` (where applicable) as the enabled baseline.

### 4. Format Check
Format drift is faster to catch than complex logic bugs. Check (don't auto-fix) in CI:

```yaml
- name: Format check
  run: |
    set -Eeuo pipefail
    go run golang.org/x/tools/cmd/goimports@<pinned> -l . | tee /tmp/fmt.out
    test ! -s /tmp/fmt.out || { echo "Run 'task fmt' to fix formatting."; exit 1; }
```

### 5. Unit Tests (OS Matrix)
Run tests across the OS matrix your consumers run on. For libraries: all three (`ubuntu-latest`, `macos-latest`, `windows-latest`). For services: the deployment target, optionally plus macOS for developer-local parity.

```yaml
test:
  runs-on: ${{ matrix.os }}
  strategy:
    fail-fast: false
    matrix:
      os: [ubuntu-24.04, macos-latest, windows-latest]
  steps:
    - uses: actions/checkout@<sha>  # v4.2.2
    - uses: actions/setup-go@<sha>  # v6.0.0
      with:
        go-version-file: go.mod
    - name: Run tests with coverage
      run: go -C tools tool task test-cover
    - name: Check coverage
      if: matrix.os == 'ubuntu-24.04'
      run: ./scripts/check-coverage.sh coverage.out
```

Rules:
- The underlying test task should run parallel-by-default and shuffled (`go test -shuffle=on ./...`, with `t.Parallel()` the norm per `test.md`) so the suite stays fast and order-independent.
- `fail-fast: false` on the test matrix — knowing which platform failed is the point.
- Enforce coverage thresholds on a single matrix cell (usually Linux). Running the check on all three ties you to consistent behavior across OSes that isn't guaranteed.
- Publish the `coverage.out` as an artifact on main-branch pushes so trend tools can consume it.

### 6. Race Detector
Run the race detector on concurrency-sensitive packages at minimum, on the whole repo when practical. The race detector catches a class of bugs nothing else does.

```yaml
- name: Race tests
  run: go test -race -count=1 ./...
```

If race-suite duration becomes a problem, scope to `-run=...` or race-only integration tests, but keep it running somewhere in CI.

### 7. Fuzz Tests
Native Go fuzzing has two distinct roles in CI, and the pipeline should cover both:

**Seed corpus on every PR (free, fast).** A plain `go test` already executes every `FuzzXxx` target's seed corpus and any committed `testdata/fuzz/` crashers as ordinary cases. This is regression protection at no extra cost — it runs inside the normal test job. Nothing to add beyond having fuzz targets.

**Time-boxed active fuzzing on a schedule.** Mutation-based fuzzing is open-ended, so it does not belong in the per-PR critical path. Run it as its own job on a `schedule:` (e.g. nightly) and on manual dispatch, with a wall-clock budget:

```yaml
fuzz:
  if: github.event_name == 'schedule' || github.event_name == 'workflow_dispatch'
  runs-on: ubuntu-24.04
  steps:
    - uses: actions/checkout@<sha>  # v4.2.2
    - uses: actions/setup-go@<sha>  # v6.0.0
      with:
        go-version-file: go.mod
    - name: Fuzz
      run: go -C tools tool task fuzz  # iterates targets: go test -run=^$ -fuzz=FuzzXxx -fuzztime=5m ./...
    - name: Upload crashers
      if: failure()
      uses: actions/upload-artifact@<sha>
      with:
        name: fuzz-crashers-${{ github.run_id }}
        path: "**/testdata/fuzz/**"
        if-no-files-found: ignore
```

Rules:
- `-fuzz` runs one target at a time and cannot span multiple packages — the task iterates targets/packages rather than passing `-fuzz` to `./...`.
- Always bound active fuzzing with `-fuzztime` (wall-clock) or `-fuzzminimizetime`; an unbounded `-fuzz` run never returns.
- When a scheduled run finds a crasher, upload the `testdata/fuzz/` input as an artifact and open an issue/PR that commits it. Once committed it is replayed by the normal test job forever after.
- Keep `-fuzztime` modest per target so the matrix of targets finishes within the scheduled window; raise it deliberately, like a runner bump.

### 8. Integration Tests
Integration tests usually live behind a build tag (`-tags integration`) so they don't run by default:

```yaml
- name: Integration tests
  run: go test -v -tags integration ./...
```

Run integration tests on Linux only unless the integration genuinely depends on OS-specific behavior.

### 9. Static Analysis (staticcheck)
`go vet` runs as part of `go test` by default. `staticcheck` is the additional pass:

```yaml
- name: Staticcheck
  run: go -C tools tool task staticcheck
```

### 10. Vulnerability Scan (govulncheck)
`govulncheck` is Go-specific: it reports CVEs for code paths actually reachable from your binary, which is dramatically more useful than a blind dependency match.

```yaml
- name: Vulnerability scan
  run: go -C tools tool task vuln
```

Run on every PR and on schedule. A dependency that was clean yesterday isn't guaranteed clean today.

### 11. Benchmarks (Performance Baseline)
Track performance across commits. A lightweight run on PRs, a full run on main-branch pushes:

```yaml
perf:
  runs-on: ubuntu-24.04
  steps:
    - name: Quick benchmarks (PR)
      if: github.event_name == 'pull_request'
      run: go -C tools tool task perf-bench-quick
    - name: Full benchmarks (main)
      if: github.event_name == 'push'
      run: go -C tools tool task perf-bench
    - name: Upload perf artifacts
      if: github.event_name == 'push' && github.ref == 'refs/heads/main'
      uses: actions/upload-artifact@<sha>
      with:
        name: perf-${{ github.sha }}
        path: |
          perf/history.jsonl
          perf/runs/*.txt
        if-no-files-found: error
        retention-days: 30
```

Treat performance history as an artifact: append-only JSONL, one record per commit, reviewable by diff. Flag significant regressions as part of PR review.

### 12. Build (goreleaser snapshot)
On every commit, verify every release platform builds:

```yaml
- uses: goreleaser/goreleaser-action@<sha>
  with:
    version: "~> v2"
    install-only: true
- name: Snapshot build
  run: go -C tools tool task build-ci  # invokes: goreleaser build --snapshot --clean
```

Pin goreleaser by minor version (`~> v2`); major version changes have broken config schemas.

### 13. CI Summary
Publish a summary of all job results. Failures should be scannable in the PR view:

```yaml
summary:
  needs: [dco, reuse, lint, test, integration, quality, perf, build]
  if: always() && github.event_name == 'push' && github.ref == 'refs/heads/main'
  runs-on: ubuntu-24.04
  steps:
    - name: Publish summary
      env:
        DCO_RESULT: ${{ needs.dco.result }}
        # ... other jobs
      run: |
        {
          echo "## CI Summary"
          echo
          echo "| Job | Result |"
          echo "| --- | --- |"
          echo "| dco | $DCO_RESULT |"
          # ...
        } >> "$GITHUB_STEP_SUMMARY"
```

ADO equivalent: `##vso[task.uploadsummary]<file>`.

## Release Pipeline
A separate pipeline, triggered on tags (`v*.*.*`), handles signed releases. It overlaps with CI but adds:

- Full matrix cross-compilation via goreleaser
- SBOM generation (syft / `goreleaser --attestations`)
- Provenance attestation (SLSA generator)
- Cosign keyless signing of archives, checksums, and images
- Registry push (GHCR, ACR, ECR) with OIDC
- GitHub Release (or equivalent) creation with changelog and artifacts

See `cicd/supply-chain.md` for signing and provenance, and the platform-specific skill (`cicd/github-actions.md` or `cicd/azure-devops.md`) for wiring.

### Version Computation
`svu` computes the next version from commit history:

```sh
go run github.com/caarlos0/svu/v3@<pinned> next
```

For conventional-commit-driven releases, use `svu` in a release-PR workflow that opens a PR updating CHANGELOG and version files.

## Coverage Discipline
- Set per-package minimum thresholds, not a blanket repository percentage. Some packages have more branching and deserve higher coverage; some (generated code, glue main) don't.
- Enforce thresholds with a script that reads `coverage.out` and fails the job on regression. Example signature: `./scripts/check-coverage.sh coverage.out`.
- Publish HTML coverage (`go tool cover -html=coverage.out -o coverage.html`) as an artifact on main for browsability.
- A coverage drop in a PR is a discussion, not an auto-fail — but make it visible.

## Dependencies (Dependabot / Renovate)
- Enable `gomod` and `github-actions` ecosystems.
- Separate Go module updates from Actions updates into different PR groupings; they fail in different ways.

```yaml
# .github/dependabot.yml
version: 2
updates:
  - package-ecosystem: "gomod"
    directory: "/"
    schedule:
      interval: "weekly"
    groups:
      go-deps:
        patterns: ["*"]
  - package-ecosystem: "github-actions"
    directory: "/"
    schedule:
      interval: "weekly"
```

## Build And Test Time Discipline
Pipeline wall-clock time is a first-class quality attribute: faster CI is run more often, blocks merges for less time, and costs less. Optimize it relentlessly — but never by removing coverage, dropping the race detector, or weakening gates. The goal is the same signal in less time, achieved through parallelism, caching, and avoided rework.

- **Parallelize jobs by default.** Independent jobs (lint, test, staticcheck, vuln, build) should run as concurrent jobs, not sequential steps. Use `needs:` only to express real ordering — e.g. gate the expensive matrix behind a fast lint job so a formatting error doesn't burn the full matrix. Fast-fail, then fan out.
- **Parallelize within the test job.** `go test ./...` already runs distinct packages concurrently (`-p`, defaults to `GOMAXPROCS`); `t.Parallel()` (the default per `test.md`) parallelizes cases within a package (`-parallel`). The slow packages are usually the ones not yet internally parallel — fix those first.
- **Lean on the build and test cache.** `actions/setup-go`'s cache (keyed on `go.sum`) plus Go's own build/test cache means unchanged packages aren't recompiled or re-run. Do not pass `-count=1` to the whole suite by reflex — it defeats the test cache and re-runs everything. Reserve `-count=1` for the cases where a fresh run is the point (race, flaky-hunt).
- **Tier the work.** Cheap, high-signal jobs (lint, format, vet) run first and short-circuit the rest. Open-ended work (active fuzzing, full benchmarks) runs on a schedule, not on every PR. Keep the per-PR path lean; push the long tail off the critical path.
- **Shard only when a single job is the bottleneck.** If one test job dominates wall-clock even after internal parallelism, split it by package group or use a test-splitting runner across matrix shards. Shard as a last resort, after parallelism and caching — sharding adds orchestration cost and result-merging complexity.
- **Measure before optimizing.** Use the CI summary's per-job timing (or `gotestsum`/`go test -json`) to find the actual long pole. Don't micro-optimize a 20-second job while a 6-minute one sits unparallelized.
- **Never trade signal for speed.** Keep `-race`, keep coverage, keep the OS matrix where it earns its place. If something must be cut for time, cut it from the per-PR tier and move it to a scheduled run — don't delete it.

## Caching
- Use `actions/setup-go` built-in cache. It keys by `go.sum` and handles invalidation correctly.
- Additional caches for `~/.cache/go-build` and `~/go/pkg/mod` beyond the default are rarely worth the complexity.

## Runner Pinning
- Pin Ubuntu (`ubuntu-24.04`) rather than `ubuntu-latest`. `latest` rolls under you and breaks reproducibility.
- Bump runner version deliberately; treat it like a dependency update with its own PR.

## Anti-Patterns To Reject
- Hardcoded Go version in the workflow instead of reading from `go.mod`
- Unpinned `go run` tool versions (`@latest`)
- Skipping the race detector in CI "for speed"
- `go test ./...` without any coverage or vuln scan in the pipeline
- Separate format / lint / vet steps that duplicate golangci-lint's coverage
- `fail-fast: true` on the test matrix (hides cross-platform failures)
- Benchmark runs that drop results on the floor (no history, no regression alert)
- Auto-formatting commits pushed from CI (violates signed-commit / DCO flow)
- `goreleaser` version unpinned or on a major range
- Coverage thresholds applied as a global blanket percentage
- Using `go install` to install CI tools (pollutes the cache; prefer `go run` with a pinned module path)
- Running integration tests on every OS in the matrix when they only validate Linux behavior
- CI summary absent — PR authors have to dig through logs to find the failure
- No fuzz coverage on parsers/validators/deserializers, or fuzz targets that exist but are never exercised in CI (not even their seed corpus)
- Unbounded `-fuzz` runs (no `-fuzztime`) that hang the job, or active fuzzing wedged into the per-PR critical path instead of a schedule
- Discovered fuzz crashers never committed to `testdata/fuzz/`, so the regression isn't replayed
- Serializing independent jobs that could run concurrently, inflating wall-clock for no signal
- Blanket `-count=1` across the whole suite, defeating the test cache and re-running unchanged packages
- Trading away `-race`, coverage, or the OS matrix to make CI faster instead of parallelizing/caching/scheduling

## Completion Criteria
Do not consider a Go CI task complete until all applicable items are true:
- Go toolchain version comes from `go.mod`, not the workflow
- every `go run` tool is version-pinned
- lint, vet/staticcheck, race tests, vuln scan, coverage, and build jobs exist and short-circuit correctly
- tests run on a meaningful OS matrix; integration tests are tagged and scoped
- tests run parallel-by-default (`t.Parallel()`) and with `-shuffle=on`; the suite is not needlessly serialized
- fuzz targets exist for parsers/validators/deserializers; their seed corpus runs on every PR and time-boxed active fuzzing runs on a schedule
- pipeline wall-clock is kept lean via job-level parallelism, caching, and tiering — without dropping `-race`, coverage, or the matrix
- coverage thresholds are enforced per-package
- benchmark history is appended as an artifact on main-branch pushes
- CI summary is published on main
- Dependabot / Renovate is configured for `gomod` and the CI platform's actions/tasks
- release pipeline (separate) handles signing, SBOM, and provenance — see `cicd/supply-chain.md`

## Invocation Template
Use this skill with a prompt that supplies repository-specific context. Example:

```text
Use Go CI Mode together with cicd-github-actions.
Add a CI workflow to /path/to/repo/.github/workflows/ci.yml with DCO, REUSE, lint, format check,
tests on ubuntu-24.04/macos/windows (parallel, -shuffle=on), race detector on Linux,
the fuzz seed corpus on every PR plus a scheduled time-boxed -fuzz job, integration tests with -tags integration,
staticcheck, govulncheck, benchmarks (quick on PR, full on main), and goreleaser snapshot build.
Run independent jobs concurrently and gate the matrix behind fast lint for lean wall-clock.
Publish a CI summary. Pin the Go toolchain via go.mod, use a Taskfile in tools/, and pin every tool version.
Configure Dependabot for gomod and github-actions.
```
