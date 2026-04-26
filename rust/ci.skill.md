---
name: rust-ci
description: Use when designing or modifying CI for a Rust project — `.github/workflows/*.yml`, `azure-pipelines.yml`, `Justfile` / `Taskfile.yml` / `xtask` crate. Defines the DCO / fmt / clippy / test (matrix) / coverage / docs / supply-chain (`cargo deny`/`cargo audit`) / `miri` / MSRV / build / release job set. Default toolchain is rustfmt + clippy + cargo-nextest + cargo-deny + cargo-audit + cargo-llvm-cov + cargo-dist (or release-plz). Pair with `cicd-*` skills and `rust-test` / `rust-policy`.
---

<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 0.1.0 -->
# Rust CI Mode

## Purpose
Use this skill when designing, writing, or refactoring CI for a Rust project. It defines the job set, toolchain, and conventions that produce high-signal, reproducible Rust pipelines.

This skill is the Rust-specific CI layer. Pair it with `cicd-core` for platform-agnostic principles, a platform-specific skill (`cicd-github-actions` or `cicd-azure-devops`) for wiring, `cicd-supply-chain` for release integrity, and `rust-test` / `rust-policy` for test-shape and quality rules.

## Skill Use
- Load this skill when the task involves authoring or modifying CI for a Rust codebase.
- Treat this skill as the governing contract for Rust-specific CI jobs (fmt, clippy, test, coverage, docs, MSRV, miri, supply-chain, release build).
- Keep project-specific toolchain preferences (`Justfile` vs. `xtask` vs. `Taskfile.yml`, `cargo-dist` vs. `release-plz` vs. hand-rolled, `cargo-nextest` vs. `cargo test`) in the invoking prompt.

## Tool Use
This skill is tool-agnostic and works with Claude Code, Codex, OpenCode, and similar assistants. Map its guidance to whatever file-reading, editing, search, and shell-execution tools your environment exposes.

- Read `Cargo.toml`, `Cargo.lock`, `rust-toolchain.toml`, `.cargo/config.toml`, existing CI workflows, and the repository's tool pins before proposing changes. Rust CI is usually a thin wrapper over project-defined tasks.
- Run the jobs locally first: `cargo xtask ci`, `just ci`, or the equivalent. A CI change that is never executed locally is a CI change that will fail on the runner.
- Use `actionlint` / `az pipelines validate` to validate the workflow; use `cargo fmt --check`, `cargo clippy -- -D warnings`, `cargo test`, `cargo deny check`, `cargo audit` to validate the jobs.
- Issue independent tool calls (reading workflows, `Cargo.toml` files, `xtask`, `deny.toml`) in parallel.

## Opinionated Default Toolchain
This is the default toolchain for a modern Rust project. Substitute only when the project has a clear reason.

- **Toolchain pinning**: `rust-toolchain.toml` is the source of truth (channel, components, profile, optional `targets`). CI uses a setup action that respects it — never hardcode the channel in the workflow.
- **Toolchain installer in CI**: [`actions-rust-lang/setup-rust-toolchain`](https://github.com/actions-rust-lang/setup-rust-toolchain) (preferred for Actions) or `dtolnay/rust-toolchain` (minimal, fast, by-toolchain action). Both honor `rust-toolchain.toml`.
- **Cache**: [`Swatinem/rust-cache`](https://github.com/Swatinem/rust-cache) — keys by `Cargo.lock`, handles `target/` invalidation correctly, much better than rolling your own.
- **Format**: `rustfmt` via `cargo fmt --all -- --check` in CI; let developers run `cargo fmt --all` locally. Config in `rustfmt.toml`.
- **Lint**: `clippy` via `cargo clippy --workspace --all-targets --all-features -- -D warnings`. Config in `clippy.toml` and via `[workspace.lints]` in `Cargo.toml`.
- **Test runner**: [`cargo-nextest`](https://nexte.st) (preferred for medium and large workspaces) or `cargo test`. Doc tests still need `cargo test --doc` even when nextest is the primary runner.
- **Coverage**: [`cargo-llvm-cov`](https://github.com/taiki-e/cargo-llvm-cov) — fast, accurate, integrates with `cargo-nextest`. Older `tarpaulin` is fine for stable-only setups but slower.
- **Supply-chain advisories**: [`cargo-audit`](https://github.com/rustsec/rustsec) for the RustSec advisory DB.
- **Supply-chain policy**: [`cargo-deny`](https://github.com/EmbarkStudios/cargo-deny) for licenses, advisories, banned crates, source allowlists, and duplicate-dependency control. Config in `deny.toml`.
- **Unused deps**: [`cargo-machete`](https://github.com/bnjbvr/cargo-machete) (stable-friendly) or `cargo-udeps` (requires nightly).
- **Docs**: `cargo doc --workspace --no-deps --document-private-items` with `RUSTDOCFLAGS="-D warnings"` to fail on broken intra-doc links.
- **MSRV verification**: [`cargo-msrv`](https://github.com/foresterre/cargo-msrv) verify (matches the declared MSRV against a check), or matrix-based MSRV cells.
- **`miri` for unsafe**: `cargo +nightly miri test` for crates that exercise `unsafe`. Slow; gate behind a path filter or schedule.
- **Release**: [`cargo-dist`](https://github.com/axodotdev/cargo-dist) (handles cross-compilation, archives, attestations, GitHub Releases) or [`release-plz`](https://github.com/release-plz/release-plz) (automated release PRs with changelog from conventional commits). Many projects use both: `release-plz` for the version/PR flow, `cargo-dist` for the artifact pipeline.
- **License / attribution**: REUSE tool (`pipx run reuse lint`) when SPDX headers are required.
- **Developer Certificate of Origin**: verify `Signed-off-by:` trailers on PR commits when DCO is required.

Pin every CI tool's version explicitly. Floating versions defeat reproducibility.

## Pipeline Job Set
A full Rust CI pipeline usually has these jobs. Skip a job only when it demonstrably doesn't apply, and say so.

### 1. DCO Check (if required)
For projects that require the Developer Certificate of Origin, verify `Signed-off-by:` on every PR commit.

```yaml
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

### 3. Format Check
Format drift is faster to catch than complex logic bugs. Check (don't auto-fix) in CI:

```yaml
- name: rustfmt
  run: cargo fmt --all -- --check
```

### 4. Lint (clippy)
Run clippy as the workflow's first heavy-feedback tier — the diagnostics are the highest-quality lint output in any major language.

```yaml
- name: clippy
  run: cargo clippy --workspace --all-targets --all-features -- -D warnings
```

Rules:
- `--all-targets` covers `lib`, `bin`, `test`, `bench`, and `example`. Without it, you miss real bugs in test code.
- `--all-features` is the safer default; if features are non-additive (rare and a smell), substitute a curated subset.
- Treat `-D warnings` as non-negotiable in CI. New warnings should fail the build.
- Repo-wide lint config goes in `[workspace.lints.clippy]` in the root `Cargo.toml`. Avoid scattering `#[allow(...)]` across the codebase.

### 5. Build (Cargo Check)
Cheap full-workspace check:

```yaml
- name: cargo check
  run: cargo check --workspace --all-targets --all-features
```

If features are non-trivial, also test combinations:

```yaml
- name: feature powerset
  run: cargo hack check --feature-powerset --no-dev-deps --workspace
```

### 6. Unit + Integration Tests (OS Matrix)
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
    - uses: actions-rust-lang/setup-rust-toolchain@<sha>
      with:
        toolchain: stable
    - uses: Swatinem/rust-cache@<sha>
    - name: Install nextest
      uses: taiki-e/install-action@<sha>
      with:
        tool: cargo-nextest
    - name: Run tests
      run: cargo nextest run --workspace --all-features --no-fail-fast
    - name: Doc tests (Linux only)
      if: matrix.os == 'ubuntu-24.04'
      run: cargo test --workspace --doc --all-features
```

Rules:
- `fail-fast: false` on the test matrix — knowing which platform failed is the point.
- Doc tests on one OS is fine; they exercise the `///` examples, not OS-specific code.
- `cargo-nextest` does not run doc tests; keep `cargo test --doc` as a separate step.
- Run the test matrix on every push and PR.

### 7. Coverage
Coverage on a single OS (usually Linux) is sufficient for trend tracking; full-matrix coverage is rarely worth the runtime cost.

```yaml
coverage:
  runs-on: ubuntu-24.04
  steps:
    - uses: actions/checkout@<sha>
    - uses: actions-rust-lang/setup-rust-toolchain@<sha>
      with:
        toolchain: stable
        components: llvm-tools-preview
    - uses: Swatinem/rust-cache@<sha>
    - uses: taiki-e/install-action@<sha>
      with:
        tool: cargo-llvm-cov,cargo-nextest
    - name: Generate coverage
      run: cargo llvm-cov nextest --workspace --all-features --lcov --output-path lcov.info
    - name: Upload to Codecov (optional)
      uses: codecov/codecov-action@<sha>
      with:
        files: lcov.info
        fail_ci_if_error: false
```

Set per-package thresholds in a script that reads `lcov.info` (or use `cargo llvm-cov --fail-under-lines NN`); blanket repo percentages are noise.

### 8. Documentation Build
Catch broken intra-doc links and `cargo doc` failures before they hit `docs.rs`:

```yaml
docs:
  runs-on: ubuntu-24.04
  env:
    RUSTDOCFLAGS: "-D warnings -D rustdoc::broken_intra_doc_links"
  steps:
    - uses: actions/checkout@<sha>
    - uses: actions-rust-lang/setup-rust-toolchain@<sha>
    - uses: Swatinem/rust-cache@<sha>
    - name: Build docs
      run: cargo doc --workspace --no-deps --all-features
```

Use the nightly toolchain for `--cfg docsrs` rendering if the project uses `#[cfg_attr(docsrs, doc(cfg(...)))]`:

```yaml
- name: Build docs (docs.rs config)
  run: cargo +nightly doc --workspace --no-deps --all-features
  env:
    RUSTDOCFLAGS: "--cfg docsrs -D warnings"
```

### 9. Supply-Chain (cargo-deny + cargo-audit)
`cargo-audit` reports advisories from the RustSec database. `cargo-deny` enforces license, advisory, banned-crate, source, and duplicate policy.

```yaml
deny:
  runs-on: ubuntu-24.04
  steps:
    - uses: actions/checkout@<sha>
    - uses: EmbarkStudios/cargo-deny-action@<sha>
      with:
        command: check advisories bans licenses sources
```

```yaml
audit:
  runs-on: ubuntu-24.04
  steps:
    - uses: actions/checkout@<sha>
    - uses: rustsec/audit-check@<sha>
      with:
        token: ${{ secrets.GITHUB_TOKEN }}
```

Run on every PR and on a daily schedule. A clean dependency yesterday isn't guaranteed clean today.

### 10. MSRV Check (libraries)
For libraries that publish a Minimum Supported Rust Version, verify it stays accurate:

```yaml
msrv:
  runs-on: ubuntu-24.04
  steps:
    - uses: actions/checkout@<sha>
    - uses: actions-rust-lang/setup-rust-toolchain@<sha>
      with:
        toolchain: "1.X"   # match `rust-version` in Cargo.toml
    - uses: Swatinem/rust-cache@<sha>
    - name: Check on MSRV
      run: cargo check --workspace --all-features
```

Or use `cargo-msrv verify`:

```yaml
- uses: taiki-e/install-action@<sha>
  with:
    tool: cargo-msrv
- run: cargo msrv verify
```

### 11. Miri (Unsafe Soundness)
For crates that exercise `unsafe`, run `miri` to catch UB. Miri is slow; consider scheduling it for nightly runs or gating behind a path filter.

```yaml
miri:
  runs-on: ubuntu-24.04
  steps:
    - uses: actions/checkout@<sha>
    - uses: actions-rust-lang/setup-rust-toolchain@<sha>
      with:
        toolchain: nightly
        components: miri
    - uses: Swatinem/rust-cache@<sha>
    - name: Setup miri
      run: cargo miri setup
    - name: Run miri
      run: cargo miri test --workspace --all-features
```

If the workspace has crates without `unsafe`, scope miri to crates that need it: `cargo miri test -p <crate>`.

### 12. Benchmarks (Performance Baseline)
Track performance across commits with `criterion`. A lightweight run on PRs, a full run on main-branch pushes:

```yaml
perf:
  runs-on: ubuntu-24.04
  steps:
    - uses: actions/checkout@<sha>
    - uses: actions-rust-lang/setup-rust-toolchain@<sha>
    - uses: Swatinem/rust-cache@<sha>
    - name: Quick benchmarks (PR)
      if: github.event_name == 'pull_request'
      run: cargo bench --workspace -- --profile-time=1 --quick
    - name: Full benchmarks (main)
      if: github.event_name == 'push' && github.ref == 'refs/heads/main'
      run: cargo bench --workspace
    - name: Upload perf artifacts
      if: github.event_name == 'push' && github.ref == 'refs/heads/main'
      uses: actions/upload-artifact@<sha>
      with:
        name: bench-${{ github.sha }}
        path: target/criterion
        retention-days: 30
```

Treat benchmark history as an artifact; do not gate PRs on absolute timing (machine variance), but flag significant regressions.

### 13. Cross-Compilation Smoke (libraries)
For libraries that claim to support `no_std` or non-host targets, smoke-test a representative target:

```yaml
- run: rustup target add thumbv7em-none-eabihf
- run: cargo check --target thumbv7em-none-eabihf --no-default-features
```

### 14. CI Summary
Publish a summary of all job results. Failures should be scannable in the PR view:

```yaml
summary:
  needs: [dco, reuse, fmt, clippy, test, coverage, docs, deny, audit, msrv, miri, perf]
  if: always() && github.event_name == 'push' && github.ref == 'refs/heads/main'
  runs-on: ubuntu-24.04
  steps:
    - name: Publish summary
      env:
        FMT_RESULT: ${{ needs.fmt.result }}
        # ... other jobs
      run: |
        {
          echo "## CI Summary"
          echo
          echo "| Job | Result |"
          echo "| --- | --- |"
          echo "| fmt | $FMT_RESULT |"
          # ...
        } >> "$GITHUB_STEP_SUMMARY"
```

ADO equivalent: `##vso[task.uploadsummary]<file>`.

## Release Pipeline
A separate pipeline, triggered on tags (`v*.*.*`) or by `release-plz` PR merges, handles signed releases. It overlaps with CI but adds:

- Cross-platform builds (handled by `cargo-dist` if used)
- `cargo publish` for crate registry publication (or `--dry-run` first)
- SBOM generation (`cargo cyclonedx`)
- Provenance attestation (SLSA generator)
- Cosign keyless signing of archives, checksums, and images
- Registry push (GHCR, ACR, ECR) for any container artifacts
- GitHub Release creation with changelog and artifacts

See `cicd-supply-chain` for signing and provenance, and the platform-specific skill (`cicd-github-actions` or `cicd-azure-devops`) for wiring.

### Versioning
[`release-plz`](https://github.com/release-plz/release-plz) opens release PRs that bump versions and update changelogs from conventional commits. It's effectively `svu` + `goreleaser` rolled into one for Rust:

```yaml
release-plz:
  if: github.ref == 'refs/heads/main'
  runs-on: ubuntu-24.04
  steps:
    - uses: actions/checkout@<sha>
      with:
        fetch-depth: 0
    - uses: actions-rust-lang/setup-rust-toolchain@<sha>
    - uses: release-plz/action@<sha>
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        CARGO_REGISTRY_TOKEN: ${{ secrets.CARGO_REGISTRY_TOKEN }}
```

For binary distribution at the same time, [`cargo-dist`](https://github.com/axodotdev/cargo-dist) generates a tag-triggered release workflow that builds matrix archives and publishes a GitHub Release.

## Coverage Discipline
- Set per-package minimum thresholds, not a blanket repository percentage. Some crates have more branching and deserve higher coverage; some (binaries, glue main, generated code) don't.
- Enforce thresholds with a script that reads `lcov.info` (or use `cargo llvm-cov --fail-under-lines NN`) and fails the job on regression.
- Publish HTML coverage (`cargo llvm-cov --html`) as an artifact on main for browsability.
- A coverage drop in a PR is a discussion, not an auto-fail — but make it visible.

## Dependencies (Dependabot / Renovate)
- Enable `cargo` and the relevant Actions/Pipelines ecosystems.
- Group related Cargo updates so PRs don't fragment.

```yaml
# .github/dependabot.yml
version: 2
updates:
  - package-ecosystem: "cargo"
    directory: "/"
    schedule:
      interval: "weekly"
    groups:
      cargo-deps:
        patterns: ["*"]
  - package-ecosystem: "github-actions"
    directory: "/"
    schedule:
      interval: "weekly"
```

Renovate has richer Cargo support (workspace-aware version unification, MSRV-aware updates) and is worth adopting on larger workspaces.

## Caching
- Use `Swatinem/rust-cache`. It keys by `Cargo.lock`, handles `target/` invalidation, supports incremental builds. Rolling your own cache is almost always a regression.
- Avoid caching `~/.cargo/registry/cache` separately; the action already covers it.
- For very large workspaces, sccache is an option but adds operational overhead.

## Runner Pinning
- Pin Ubuntu (`ubuntu-24.04`) rather than `ubuntu-latest`. `latest` rolls under you and breaks reproducibility.
- Bump runner version deliberately; treat it like a dependency update with its own PR.

## Anti-Patterns To Reject
- Hardcoded toolchain channel in the workflow instead of reading from `rust-toolchain.toml`
- `actions/cache` instead of `Swatinem/rust-cache` (loses correctness on `target/` invalidation)
- Skipping `--all-targets` on clippy (misses test/example/bench code)
- Skipping `--all-features` (or a curated feature set) — feature combinations consumers will use must be tested
- `cargo test ./...`-style invocation copied from Go; in Rust it's `cargo test --workspace`
- Coverage thresholds applied as a global blanket percentage
- `cargo install` to install CI tools (slow, no caching); use `taiki-e/install-action` or `cargo-binstall`
- `fail-fast: true` on the test matrix (hides cross-platform failures)
- Auto-formatting commits pushed from CI (violates signed-commit / DCO flow)
- `release-plz` and `cargo-dist` versions unpinned or on a major range
- Running `miri` on every PR for a workspace where most crates don't have `unsafe` (wastes CI minutes)
- Ignoring `cargo audit` advisories without a tracked `RUSTSEC-XXXX-YYYY` exception in `deny.toml`
- Benchmark runs that drop results on the floor (no history, no regression alert)
- CI summary absent — PR authors have to dig through logs to find the failure

## Completion Criteria
Do not consider a Rust CI task complete until all applicable items are true:
- toolchain channel comes from `rust-toolchain.toml`, not the workflow
- every CI tool is version-pinned (action SHAs, `taiki-e/install-action` versions, `cargo-deny` action version)
- fmt, clippy, build, test, coverage, docs, supply-chain (deny + audit), and (if applicable) miri/MSRV jobs exist and short-circuit correctly
- tests run on a meaningful OS matrix; integration tests are scoped where they belong
- coverage thresholds are enforced per-package or by a deliberate baseline
- `RUSTDOCFLAGS="-D warnings"` catches broken intra-doc links
- benchmark history is appended as an artifact on main-branch pushes
- CI summary is published on main
- Dependabot / Renovate is configured for `cargo` and the CI platform's actions/tasks
- release pipeline (separate) handles signing, SBOM, and provenance — see `cicd-supply-chain`

## Invocation Template
Use this skill with a prompt that supplies repository-specific context. Example:

```text
Use Rust CI Mode together with cicd-github-actions.
Add a CI workflow to /path/to/repo/.github/workflows/ci.yml with DCO, REUSE, rustfmt --check, clippy with -D warnings,
cargo check, cargo nextest run on ubuntu-24.04/macos/windows, cargo test --doc on Linux, cargo llvm-cov coverage on Linux,
cargo doc with broken-link checking, cargo deny check, cargo audit, MSRV verify, and miri on the unsafe-bearing crates only.
Publish a CI summary. Pin the toolchain via rust-toolchain.toml, use Swatinem/rust-cache, and pin every action by SHA.
Configure Dependabot for cargo and github-actions. Add release-plz for the release-PR flow and cargo-dist for the artifact pipeline.
```
