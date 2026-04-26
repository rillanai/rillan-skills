---
name: rust-docs
description: Use when generating, reviewing, or maintaining Rust documentation — `///` and `//!` doc comments, doctests, crate-level docs, `cargo doc` HTML output, `mdbook` user guides, READMEs, ADRs, API docs, runbooks, changelogs, onboarding. Read source first; document actual behavior, not intent. Skip for logic review, security audit, architecture grading, or test design.
---

<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 0.1.0 -->
# Rust Documentation Mode

## Purpose
Use this skill when generating, reviewing, or maintaining documentation for a Rust codebase.

Apply this skill with:
- `policy.skill.md` for standards that affect public APIs and maintainability
- `workflow.skill.md` for code-grounded, tool-first execution

This mode is for documentation work that materially changes developer understanding: doc comments, crate-level docs, `cargo doc` output, `mdbook` guides, README files, ADRs, API docs, runbooks, changelogs, and onboarding materials.

## Tool Use
This skill is tool-agnostic and works with Claude Code, Codex, OpenCode, and similar assistants. Map its guidance to whatever file-reading, editing, search, and shell-execution tools your environment exposes.

- Read the source before documenting it; do not write documentation from memory or assumption.
- Verify doctests by running `cargo test --doc` — Rust doctests are real tests and will fail CI if they don't compile.
- Generate and inspect output with `cargo doc --no-deps --document-private-items --open` for local review; `--no-deps` keeps generation fast.
- Issue independent tool calls (reading multiple files, checking multiple symbols, expanding macros) in parallel.
- When documenting behavior, cite the file and item path (`crate::module::Item`) you read — do not paraphrase without grounding.

## When To Use
Use this skill for:
- crate-level docs (`//!` in `src/lib.rs` or `src/main.rs`)
- module-level docs (`//!` at the top of a module file)
- exported item doc comments (`///` on `pub` items)
- doctests (runnable examples in `///` comments)
- README updates
- `mdbook` user guides
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
- Match the repository's existing documentation style when it is coherent.
- Write for the reader who will hit the docs from `docs.rs`, the IDE hover tooltip, or `cargo doc` locally.

## Documentation Rules

### Crate-Level Documentation
- Every published crate should have a `//!` doc comment in `src/lib.rs` (or `src/main.rs`) summarizing what the crate does, when to use it, and a minimum viable example.
- Start with a one-line summary that fits on the `docs.rs` index card.
- Follow with a longer description, key concepts, and a runnable example.
- Add `#![warn(missing_docs)]` (or `#![deny(missing_docs)]` for stricter projects) at the crate root for libraries.
- Use `#![doc = include_str!("../README.md")]` to share content between README and crate-level docs when they should stay in sync.

### Module-Level Documentation
- Non-trivial modules should have a `//!` doc comment at the top describing what the module does, its key types, and how it relates to its parent module.
- Describe the *purpose* of the module, not its implementation details.
- Mention the primary types and the relationships between them.

### Exported Item Documentation
Every `pub` item should have a `///` doc comment when the crate exposes it as API. The structure that works for most items:

1. **Summary line** — first sentence describes what the item is/does. Ends with a period.
2. **Detailed description** — additional paragraphs as needed.
3. **`# Errors`** — for `Result`-returning functions, the conditions under which each error variant is returned.
4. **`# Panics`** — for functions that can panic, the conditions that trigger the panic. If a function never panics, you do not need this section.
5. **`# Safety`** — required for `unsafe fn`. Document the invariants the caller must uphold.
6. **`# Examples`** — at least one runnable example for non-trivial public items.

Rules:
- Start the first sentence with a noun phrase, not "This function" — `cargo doc` lifts the first line into summary contexts.
- Use ``` `backticks` ``` for inline code, item names, and types.
- Use intra-doc links (`` [`Type`] ``, `` [`Type::method`] ``, `` [crate::module::Item] ``) so docs stay valid across renames. `cargo doc` resolves these.
- Do not duplicate the function signature in prose; readers can already see it.
- Document zero-value behavior, side effects, async cancellation safety, panic behavior, and notable error conditions when relevant.
- Do not add noise comments that restate the obvious (`/// Returns the name.` on `fn name() -> &str`).

### Doctests
- Examples in `///` comments are compiled and run by `cargo test --doc`. Treat them as tests, not as decoration.
- Use `# `-prefixed lines for setup that should not appear in rendered docs.
- Use ` ```no_run ` for examples that need to compile but not run (e.g., examples that hit the network).
- Use ` ```compile_fail ` to demonstrate intentionally invalid code.
- Use ` ```ignore ` only as a last resort — ignored examples rot.
- Use ` ```text ` for non-Rust output samples or ASCII diagrams.
- Use `assert_eq!` to make examples self-checking; doctests pass only if the assertions pass.

Example structure:

```rust
/// Parses a configuration string and returns a `Config`.
///
/// # Errors
/// Returns [`ConfigError::Syntax`] if the input is not valid TOML.
/// Returns [`ConfigError::Validation`] if required keys are missing.
///
/// # Examples
/// ```
/// use mycrate::Config;
/// let cfg = Config::parse("name = \"test\"").unwrap();
/// assert_eq!(cfg.name(), "test");
/// ```
pub fn parse(s: &str) -> Result<Config, ConfigError> { /* ... */ }
```

### Unsafe Documentation
- Every `unsafe fn` must document `# Safety` describing the invariants the caller must uphold.
- Every `unsafe` block needs a `// SAFETY:` line-comment justifying *why* the local invariants hold. This is policy, but the doc comment is the API contract.
- For `unsafe trait` (e.g., `Send`, `Sync` implementations), document `# Safety` on the trait itself listing the invariants implementors must uphold.

### Macros
- Document macros with `///` on the `#[macro_export]`-ed `macro_rules!` definition or on the proc-macro `pub fn`.
- Show at least one expansion example so readers see the input → output relationship.
- Document any input forms the macro accepts and any compile errors it can produce.

### Feature Flags
- Use `#[cfg_attr(docsrs, doc(cfg(feature = "x")))]` on items behind feature flags so `docs.rs` shows the gating badge.
- Add `[package.metadata.docs.rs] all-features = true` (or a curated set) so `docs.rs` builds the full surface.
- Document feature flags in the crate-level docs: what each one enables, what's default, and what combinations are supported.

### README Files
A Rust crate README should usually include:
1. crate name + one-line description (matches the `Cargo.toml` `description`)
2. crates.io / docs.rs / CI badges
3. overview
4. installation (`cargo add <crate>`)
5. quickstart
6. feature flag reference
7. MSRV note
8. license + contribution link

Rules:
- Do not duplicate `cargo doc` output in the README. Link to docs.rs instead.
- Keep the quickstart short and self-contained; ideally it's the same example shown in the crate-level doc comment.
- Update the README when public API, MSRV, or feature flags change.

### `mdbook` User Guides
Use `mdbook` for guides that are too long for a README and too narrative for `cargo doc`:
- conceptual overviews
- multi-page tutorials
- design rationale
- migration guides

Rules:
- Keep `mdbook` content in the repo (`docs/` or `book/`); CI builds and publishes it.
- Make code samples runnable — use `mdbook-keeper` (or compile snippets in CI) so examples don't rot.
- Cross-link to `docs.rs` items; do not re-explain the API surface.

### ADRs
Use:
- `Status`
- `Context`
- `Decision`
- `Consequences`
- `Alternatives Considered`

Rules:
- Keep ADRs concise.
- Store them in the repository's decisions directory.
- Supersede accepted ADRs with new ADRs rather than rewriting history.

### API Docs
- Keep API docs in sync with handlers, request/response types, and auth requirements.
- For HTTP/RPC services, prefer code-near canonical sources: `utoipa` derives OpenAPI from types; `axum` + `aide` does the same; `tonic` generates from `.proto`.
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
- Use numbered operational steps.
- Include actual commands or queries.
- Include expected outputs when practical.

### Changelogs
- Follow the repository's changelog format; `Keep a Changelog` is a strong default.
- Group entries under user-visible categories such as `Added`, `Changed`, `Fixed`, `Deprecated`, `Removed`, and `Security`.
- For libraries, call out semver impact: a `Removed` or signature-changed entry implies a major bump.
- Do not treat internal cleanup as a public changelog item unless it changes user-visible behavior.
- `cargo-release` and `release-plz` automate changelog updates from conventional commits — adopt one if the project releases regularly.

## Evidence Rules
- Verify docs against source, config, tests, and handlers.
- Verify defaults by reading the code that loads or constructs them, not existing docs.
- Verify error variants by reading the error type definition and the call sites that produce each variant.
- Verify safety invariants by reading the `unsafe` block and the surrounding contract.
- Label planned work explicitly as planned or proposed.

## Quality Checklist
Before considering documentation work complete, verify:
- `cargo doc --no-deps --document-private-items` succeeds with no broken intra-doc links
- `cargo test --doc` passes — every example compiles and any assertions pass
- item and crate names in docs match the current source
- examples compile and use the current API
- configuration docs match real defaults
- API docs match actual handlers and types
- docs do not describe removed items, modules, features, or behavior
- `# Errors`, `# Panics`, `# Safety` sections exist where applicable
- intra-doc links resolve (`cargo doc` will warn on broken ones if `#![warn(rustdoc::broken_intra_doc_links)]` is set)
- feature-gated items have `#[cfg_attr(docsrs, doc(cfg(...)))]` annotations

## Anti-Patterns To Reject
- stale documentation
- aspirational documentation presented as implemented behavior
- comment duplication of the signature
- undocumented public APIs (`#![warn(missing_docs)]` is the friend here)
- examples that don't compile (`ignore` everywhere)
- doctests that hit external services without `no_run`
- `unsafe fn` without `# Safety`
- `unsafe trait` impls without justification
- intra-doc links that don't resolve (silent rot)
- README that drifts from the crate-level doc comment
- screenshots of text where searchable text should exist

## Invocation Template
Use this skill with a prompt that supplies repository-specific context. Example:

```text
Use Rust Documentation Mode with Rust Engineering Policy and Rust Engineering Workflow.
Update crate-level and module documentation for /path/to/repo/crates/auth.
Add doctests for every pub function on the public API surface.
Ground every statement in code or config.
Verify with cargo doc --no-deps --document-private-items and cargo test --doc.
```
