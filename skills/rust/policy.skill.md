---
name: rust-policy
description: Use as the Rust standards layer (DRY/KISS/YAGNI, crate and module structure, ownership and borrowing discipline, error handling, unsafe policy, async lifecycle, review bar) for any Rust implementation, review, refactor, migration, documentation, or audit work. Pair with `rust-workflow` for execution discipline. Repo-specific stricter rules win when explicit and defensible.
---

<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 0.1.0 -->
# Rust Engineering Policy

## Purpose
Use this skill when Rust work needs a clear definition of what good, correct, maintainable engineering looks like.

This skill is the standards layer for Rust tasks. It defines coding rules, architectural expectations, ownership and lifetime discipline, error-handling conventions, unsafe policy, async lifecycle expectations, review priorities, and quality bars. It does not define the execution workflow; pair it with `workflow.skill.md` for that.

## Skill Use
- Load this skill for Rust implementation, review, refactor, migration, documentation, and testing work when standards matter.
- Treat this skill as the default Rust quality contract unless the repository has stricter local rules.
- Repository-specific conventions may override this skill when they are explicit, coherent, and defensible.
- When this skill conflicts with casual convenience, follow this skill.

## Tool Use
This skill is tool-agnostic and works with Claude Code, Codex, OpenCode, and similar assistants. Map its guidance to whatever file-reading, editing, search, and shell-execution tools your environment exposes.

- Invoke tools to check standards against real code; do not apply rules from memory to code you have not read.
- Issue independent tool calls in parallel rather than sequentially.
- Verify claims about API shape, lifetimes, trait bounds, or boundaries against the current source before asserting them.

## Core Principles

### DRY
- Extract repeated logic into named functions, methods, or trait helpers when repetition is real and local.
- Prefer small, clear duplication over a brittle generic abstraction with too many type parameters.
- Use shared `const` or `static` items for meaningful repeated values; prefer `const fn` constructors when applicable.

### KISS
- Choose the simplest solution that solves the actual problem.
- Prefer straightforward control flow and concrete types over deep generics or trait-object indirection.
- Split functions or modules that mix unrelated responsibilities.

### YAGNI
- Do not add traits, generic parameters, or feature flags without a real consumer.
- Do not add speculative utility crates or modules.
- Do not introduce plugin or provider patterns for a single implementation.

## Design Priorities
- Make ownership and data flow easy to trace from entrypoint to side effect.
- Encode invariants in the type system: newtypes, enums, `NonZero*`, `NonEmpty`, sealed states.
- Keep crate and module boundaries explicit and directional.
- Separate transport, domain, persistence, and infrastructure concerns when the workspace size justifies it.
- Prefer designs that are easy to test, observe, and operate.
- Choose operational clarity over abstraction density.

## Project Structure
- Single-binary apps: `src/main.rs` stays thin (parse config, build dependencies, hand off to a `run` function in a library crate). Real logic belongs in `src/lib.rs` or a separate `crate-name-core` library.
- Multi-binary apps: put binaries in `src/bin/<name>.rs` and share logic via a sibling library crate.
- Workspaces: `Cargo.toml [workspace]` with `members = [...]`. Use `[workspace.dependencies]` to unify versions across members. Use `[workspace.lints]` for shared lint configuration.
- Default to a library crate plus a thin binary crate over a single binary; tests, docs, and reuse all benefit.
- Reserve `crate-name-internal` or `*-core` naming for crates that are intentionally not part of the public API surface.

## Module And Boundary Rules
- Keep modules cohesive with one primary reason to change.
- Make public surface area explicit: `pub` only what consumers need; default to private.
- Use `pub(crate)` for cross-module sharing within a crate; `pub(super)` for narrow parent-module exposure; reserve `pub` for true API.
- Prevent inward layers (domain) from importing outward layers (transport, persistence). Crate boundaries enforce this better than module boundaries.
- Make boundary crossings explicit in names and types (`UserId` vs. `String`).
- Avoid reusing persistence types as transport DTOs unless the tradeoff is deliberate; `From`/`TryFrom` impls between layers keep the seam visible.

## Traits And Generics
- Define traits where they are consumed, not where they are implemented.
- Keep traits small (interface segregation). A trait with one method is often the right size.
- Use generics with `impl Trait` arguments for caller flexibility; use `&dyn Trait` or `Box<dyn Trait>` when you need heterogeneous collections or a stable ABI boundary.
- Prefer `impl Trait` returns for simple cases; use `Box<dyn Trait>` returns when callers need to store the value.
- Avoid trait objects for hot paths where monomorphization buys real performance.
- Do not create traits only to mock concrete types — prefer dependency injection via generics or closures, or carve out a narrow trait at the consumer.
- Mark traits `#[non_exhaustive]` or seal them (`pub trait Foo: private::Sealed`) when downstream impls would be a maintenance hazard.

## Types And Data Handling
- Start with types and invariants before implementation detail. The compiler is your design partner.
- Use newtypes (`struct UserId(Uuid)`) over primitive obsession; derive `Debug, Clone, PartialEq, Eq, Hash` as appropriate.
- Encode mutually exclusive states as enums with associated data, not as flags or sentinel values.
- Use the typestate pattern when a value moves through a sequence of states with different operations available.
- Prefer `Option<T>` over null-sentinel patterns; prefer `Result<T, E>` over panics for fallible operations.
- Validate untrusted input at the boundary closest to ingress and convert to a domain type that cannot represent the invalid state.
- Use `&str`/`&[T]` parameters and own (`String`/`Vec<T>`) return values; let the borrow checker minimize unnecessary allocation.
- Prefer concrete types over `Box<dyn Any>` or `serde_json::Value` in public APIs unless dynamic typing is genuinely required.

## Ownership And Borrowing
- Take ownership when you need it, borrow when you don't. Default to borrowing.
- Prefer `&T` over `&mut T` over owned `T` in argument position.
- Use `Cow<'_, T>` when callers may pass either an owned or borrowed value.
- Do not reach for `Rc`/`Arc` or `RefCell`/`Mutex` to dodge the borrow checker — restructure ownership instead. Reach for them when shared ownership or interior mutability is the genuine model.
- Use `Arc<Mutex<T>>` only when the data is genuinely shared across threads and contended mutation is the real requirement; channels often beat shared mutable state.
- Lifetimes on public APIs should be the minimum that compiles; explicit lifetimes are a feature, not noise to remove.

## Error Handling
- Return `Result<T, E>`; never panic for expected error conditions.
- **Libraries**: define typed errors with `thiserror` (or hand-rolled enums implementing `std::error::Error`). Each error variant has a clear cause; do not collapse everything into a `Generic(String)` variant.
- **Applications**: use `anyhow::Result` (or `eyre::Result`) at the top-level binary boundary. Add `.context()`/`.with_context()` at meaningful operation boundaries.
- Propagate with `?`; wrap with context (`.map_err(...)` or `anyhow::Context`) when crossing meaningful boundaries.
- Mark error enums `#[non_exhaustive]` if you may add variants; otherwise consumers' exhaustive matches break on minor versions.
- Do not match on error message strings; match on variants or downcast.
- `panic!`, `unwrap()`, `expect("...")` are acceptable for invariants that hold by construction. `unwrap()` in production code should pair with a comment or `expect("reason")` explaining why it cannot fail.

## Unsafe Policy
- Default to `#![forbid(unsafe_code)]` at the crate root for application code and most library code. Lift the deny to specific modules only when an unsafe block is genuinely needed.
- Every `unsafe` block requires a `// SAFETY:` comment that documents the invariants the caller (or surrounding code) must uphold.
- Wrap `unsafe` in a safe API at the smallest possible boundary. Callers should never need to know the raw pointer / FFI / transmute exists.
- Do not use `transmute`, `from_raw`/`into_raw`, or `mem::uninitialized` (removed) without a written justification and tests under `miri`.
- For FFI, define the C signature precisely (`#[repr(C)]`, `extern "C"`, exact integer widths) and use `bindgen` for non-trivial bindings.

## Async And Concurrency
- Pick one async runtime per workspace and stick with it. Tokio is the default; `async-std` and `smol` are valid but have smaller ecosystems.
- Spawned tasks should have an owner, a shutdown path, and a clear error-handling strategy. Detached `tokio::spawn` without join handles is the equivalent of an unsupervised goroutine.
- Use `tokio_util::sync::CancellationToken` (or `tokio::select!` with a shutdown channel) for cooperative cancellation. Do not rely on dropping `JoinHandle` to cancel tasks — `JoinHandle` drops detach.
- Use channels (`tokio::sync::mpsc`, `oneshot`, `broadcast`, `watch`) for coordination over shared mutable state when practical.
- Hold `Mutex` guards across `.await` only when intentional; prefer `tokio::sync::Mutex` for guards held across await points and `parking_lot::Mutex` (or `std::sync::Mutex`) for short-lived sync sections.
- Distinguish CPU-bound and I/O-bound work. Use `tokio::task::spawn_blocking` (or `rayon`) for CPU-bound work to avoid starving the async runtime.
- Make timeout, retry, idempotency, and backpressure behavior explicit at the call site or in the type, not implicit in the runtime.
- Avoid `.await` inside loops without backpressure or batching; each `.await` is a yield point.

## Lifecycle And Shutdown
- Handle `SIGINT` and `SIGTERM` via `tokio::signal::ctrl_c` (or platform-specific signals).
- Drain in-flight work before exit; bound the drain with an explicit timeout.
- Make background tasks participate in cancellation via `CancellationToken` or a watch channel.
- Log the shutdown path (start, drain, finish) clearly enough for operators to follow.

## Configuration
- Load configuration at startup into a typed struct. `figment`, `config-rs`, `serde` + a custom loader, or hand-rolled — pick one and keep precedence explicit.
- Validate configuration before starting the application; `TryFrom<Config> for ValidatedConfig` is a good pattern.
- Make precedence explicit: CLI flags > env vars > config file > defaults.
- Do not scatter `std::env::var` reads across the codebase.
- Use `clap` (derive macro) for CLI parsing in non-trivial binaries.

## Logging And Observability
- Use `tracing` for structured logging, not `log`. The two interoperate (`tracing-log`) but new code should target `tracing` directly.
- Initialize a single global subscriber early in `main`; pick `tracing-subscriber` with `EnvFilter` for log-level configuration.
- Instrument functions with `#[tracing::instrument]` at meaningful operation boundaries; use `skip(...)` for sensitive arguments and large payloads.
- Use spans for request lifetime; use events for point-in-time facts. Include correlation identifiers as span fields.
- Never log secrets, tokens, or sensitive payloads. Wrap secret types in a newtype with a `Debug` impl that prints `[REDACTED]`.
- For metrics, use the `metrics` facade crate; for tracing exports, `opentelemetry` + `tracing-opentelemetry` is the common path.

## Security
- Treat all external input as untrusted.
- Validate inputs before database, file, template, shell, or outbound-network use; convert to a domain type that cannot represent the invalid state.
- Use parameterized queries (`sqlx`, `diesel`, `tokio-postgres` `query!`) — never string-concatenate SQL.
- Keep secret access narrow; never expose secrets in `Debug` output, logs, or errors. Use `secrecy::SecretString` (or equivalent) for secret types.
- Make authorization checks explicit at the request boundary, not buried in the data layer.
- Be cautious with file paths (`std::path::Path::canonicalize`, traversal checks), redirects, remote fetches, and proxy-like behavior.
- Pin the toolchain via `rust-toolchain.toml` and the dependency tree via `Cargo.lock` (commit it for binaries; libraries do not commit `Cargo.lock`).

## Handlers And Entrypoints
- Keep handlers in a consistent flow: extract → validate → call service → encode response.
- Inject dependencies through application state (`axum::extract::State`, `actix-web::web::Data`) or a constructed router struct, not globals.
- Do not leak raw domain or database errors directly to clients; map to a stable HTTP/RPC error type at the handler boundary.
- Use `tower` middleware (or framework-native middleware) for cross-cutting concerns: logging, request IDs, timeouts, rate limits, CORS.

## Cargo Features
- Use features to gate optional functionality, not to express incompatible configurations of the same code.
- Make features additive: enabling one feature must not break another. CI must test the feature matrix that consumers will combine.
- Default features should reflect the most common use case. Document non-default features in the crate README.
- Avoid `default-features = false` in workspace dependencies unless the workspace genuinely needs the slim form everywhere.

## Edition And MSRV
- New crates: target the latest stable edition (Edition 2024 currently). Keep edition consistent across a workspace.
- Declare MSRV explicitly in `Cargo.toml` (`rust-version = "1.X"`) for libraries you publish; pick a version that's at least 6 months old for ecosystem stability.
- Verify MSRV in CI with `cargo-msrv` or by running `cargo +<msrv> check` and `cargo +<msrv> test`.
- Pin the project's working toolchain in `rust-toolchain.toml`; this is what `cargo`/`rustup` will use without a flag.

## Refactoring Rules
- Refactor toward simpler ownership, fewer lifetimes, and clearer module boundaries.
- Keep refactors incremental unless a larger redesign is explicitly requested.
- Make the smallest structural change that solves the real problem.
- Leave touched code more testable, traceable, and observable than you found it.

## Review Standard
Evaluate Rust changes in this order:
1. correctness (compiler-confirmed, then test-confirmed)
2. error handling and panic safety
3. ownership, lifetimes, and `unsafe` correctness
4. async lifecycle (cancellation, drop behavior, blocking calls)
5. API and boundary design (`pub`, trait surface, semver implications)
6. testability
7. simplicity (avoid speculative generics, trait objects, features)
8. performance where it is obvious and material

## Shared Quality Checklist
Before considering Rust work complete, verify all applicable items:
- behavior changes have relevant tests or an explicit gap explanation
- async / concurrent changes hold up under cancellation, drop, and partial failure
- errors are typed (libraries) or contextual (apps), and handled at the right level
- no new `unsafe` block exists without a `// SAFETY:` comment
- no new `unwrap()`/`expect()` exists in fallible production paths without justification
- no dead code, stray `dbg!`, `println!` debugging output, or `TODO`s without tracking context were introduced
- naming follows Rust conventions (snake_case items, CamelCase types, SCREAMING_SNAKE_CASE consts)
- exported API surface has the needed `///` doc comments
- `cargo fmt`, `cargo clippy`, and `cargo test` pass; repository-standard lints were not silenced silently

## Anti-Patterns To Reject
- god structs and oversized traits
- public APIs built around `Box<dyn Any>`, `serde_json::Value`, or `Box<dyn Error>` without strong reason
- `unwrap()` / `expect()` in fallible paths to "make the borrow checker happy"
- `Arc<Mutex<T>>` reached for to avoid restructuring ownership
- `unsafe` blocks without a `// SAFETY:` comment
- `panic!` for expected errors; converting `Result` to `Option` to dodge error handling
- detached `tokio::spawn` with no join, no cancellation, and no logged failure path
- holding a `std::sync::Mutex` guard across `.await`
- string-concatenated SQL or shell commands
- crates named `util`, `common`, or `helpers` without a narrow purpose
- generic functions with five+ type parameters
- features that subtract behavior (non-additive features)
- `from_raw` / `transmute` / `mem::*` tricks without `miri` coverage
- catch-all `From<E>` impls that lose the original error variant

## Invocation Template
Use this skill with a prompt that supplies repository-specific context. Example:

```text
Use Rust Engineering Policy with Rust Engineering Workflow.
Apply Rust standards while reviewing and updating /path/to/repo.
Keep crate boundaries explicit, encode invariants in the type system, avoid speculative generics, and prioritize correctness over convenience.
```
