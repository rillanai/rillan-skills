<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.1.0 -->
# Kubernetes Operator Development Guidance

## Purpose
Use this skill when building or modifying Kubernetes operators in Go. This skill is a controller-specific overlay and should be used together with `go/policy.md` and `go/workflow.md`.

This guidance is centered on `kubebuilder`, `controller-runtime`, and `achilles-sdk`. `sdk.md` is not the primary workflow; if a repository already uses it, preserve local conventions rather than forcing a migration.

## Tool Use
This skill is tool-agnostic and works with Claude Code, Codex, OpenCode, and similar assistants. Map its guidance to whatever file-reading, editing, search, and shell-execution tools your environment exposes.

- Invoke tools to read API types, reconcilers, watches, and generated CRDs before editing — do not describe behavior you have not traced.
- Run `make generate`, `make manifests`, `go test ./...`, and `envtest` yourself; do not claim a controller change is verified without tool output.
- Prefer structural tooling (LSP, `gopls`) for reference and implementation lookups across the reconciler, its owned resources, and tests.
- Issue independent tool calls (reading types, watches, RBAC markers, CRD manifests) in parallel.

## Scope
- CRD and API type design
- reconcilers and watch wiring
- status and condition management
- finalizers and deletion flow
- admission control: CRD schema validation, ValidatingAdmissionPolicy (VAP), and webhooks for validation/defaulting
- controller-runtime manager setup
- achilles-sdk integration where present

## Core Principles
- Treat reconciliation as a level-triggered convergence loop, not an imperative script.
- Make status truthful, minimal, and useful to operators.
- Keep spec, status, metadata, and external side effects clearly separated.
- Prefer idempotent writes and explicit ownership of created resources.
- Avoid hidden controller behavior in helper layers that obscure requeue, error, or condition semantics.

## Admission Control And Validation
Prefer the cheapest mechanism that can enforce the rule. Reach for a webhook only when nothing in-process can express it. As of Kubernetes 1.30, `ValidatingAdmissionPolicy` (VAP) is GA and is the default choice for admission-time validation that used to require a validating webhook — it runs CEL in-process in the API server, with no webhook server, no TLS certificates, no extra network hop, and no risk of a down webhook wedging the API path.

**Validation — preference order:**
1. **CRD schema validation** (OpenAPI constraints plus CEL `x-kubernetes-validations`, via `+kubebuilder:validation:XValidation`). For single-object field rules, immutability, and cross-field constraints, embed the rule in the CRD schema — zero extra infrastructure, validated by the API server on write.
2. **ValidatingAdmissionPolicy (VAP)** for cross-cutting, policy-style, or parameterized validation that the CRD schema can't express on its own. Bind with `ValidatingAdmissionPolicyBinding`; drive variation through `paramRef` parameter resources; use CEL `variables`, `matchConditions`, and the `authorizer` for richer checks. This is the modern replacement for most validating webhooks.
3. **ValidatingWebhookConfiguration** only when the check genuinely cannot be expressed in CEL — arbitrary external lookups, calls to other systems, or stateful Go logic. Treat a new validating webhook as a thing to justify, not a default.

**Defaulting — preference order:**
1. **CRD schema defaults** (`+kubebuilder:default=`, server-side defaulting) for static or simple defaults.
2. **Mutating webhook** for defaulting that needs logic the schema can't carry. `MutatingAdmissionPolicy` (CEL-based mutation) is the emerging in-process replacement but is still feature-gated (alpha/beta depending on cluster version) — adopt it only when the target clusters enable it, and keep a mutating webhook as the portable fallback.

**When VAP is the right call, design it well:**
- Keep policy CEL focused and fail-closed where safety demands it (`failurePolicy: Fail`); use `failurePolicy: Ignore` only when a missing evaluation is genuinely acceptable.
- Use `validationActions` (`Deny`, `Warn`, `Audit`) deliberately — `Audit`/`Warn` let you roll a policy out in observe-only mode before enforcing `Deny`.
- Ship the `ValidatingAdmissionPolicy` and its `ValidatingAdmissionPolicyBinding` as managed manifests/Kustomize alongside the CRDs, and grant the API server read access to any `paramRef` resources.
- Validate CEL cost and correctness before shipping; an expensive or wrong CEL expression fails admission for real traffic.

A validating webhook that exists only to do field/shape validation expressible in CEL is now technical debt — prefer migrating it to a VAP (see `migrate.md`).

## Writes And Ownership (Server-Side Apply)
This is how you actually deliver the "idempotent writes + explicit ownership" the Core Principles demand — stop hand-rolling get-then-create-or-update.
- Make SSA the default write idiom for owned resources: `client.Patch(ctx, obj, client.Apply, client.FieldOwner("<controller-name>"), client.ForceOwnership)`. The field manager declares ownership; `ForceOwnership` resolves conflicts on fields this controller is authoritative for.
- Patch status the same way: `r.Status().Patch(ctx, obj, client.Apply, client.FieldOwner("<controller-name>"))` rather than read-modify-`Update`, which races on the whole object.
- Prefer the generated `ApplyConfiguration` types (`applyconfiguration` gen, or `controllerutil` apply helpers) so you send only the fields you own — sending a fully-populated typed object via apply silently claims ownership of every set field, including defaults.
- SSA replaces the create-or-update dance and removes most resourceVersion-conflict retries; reserve optimistic `Update` for cases where you must read-modify-write under a specific resourceVersion.

## Watch Wiring (Typed controller-runtime)
Default to the typed builder surface; the untyped forms are legacy.
- Build with `builder.TypedBuilder` / `ctrl.NewControllerManagedBy(mgr).For(&T{})` and implement `reconcile.TypedReconciler[T]`.
- Watch sources with the generic `source.Kind[client.Object](cache, &T{}, handler, predicates...)`, not `source.Kind(cache, &T{})`.
- Use `handler.TypedEventHandler` / `handler.TypedEnqueueRequestForObject[T]()` and `handler.TypedEnqueueRequestsFromMapFunc`, not the untyped `handler.EnqueueRequestForObject{}`.
- Filter with typed `predicate.TypedPredicate[T]` (e.g. `predicate.TypedGenerationChangedPredicate`).
- Warn on sight of `source.Kind(cache, &T{})` or `handler.EnqueueRequestForObject{}` — these untyped forms are deprecated; migrate to the generic signatures.

## Manager Setup
- **Metrics auth — kube-rbac-proxy is gone.** Do not add the kube-rbac-proxy sidecar; serve authenticated metrics in-process: `metricsserver.Options{BindAddress: ":8443", SecureServing: true, FilterProvider: filters.WithAuthenticationAndAuthorization}` (from `sigs.k8s.io/controller-runtime/pkg/metrics/...` and `.../metrics/filters`). This uses `TokenReview`/`SubjectAccessReview` against the API server — no extra container.
- **Cache scoping** sets memory footprint and tenancy. Use `cache.Options.ByObject` with field/label `Selectors` to watch only relevant objects, `DefaultNamespaces` to scope per namespace (multi-tenant), and `DefaultTransform` to strip fields you never read (e.g. `managedFields`, large annotations) before they hit cache.
- **Concurrency** is per-controller: `.WithOptions(controller.Options{MaxConcurrentReconciles: N, RateLimiter: ...})`. Tune `MaxConcurrentReconciles` for throughput and supply a per-controller workqueue `RateLimiter` (e.g. `workqueue.NewTypedMaxOfRateLimiter`) to bound retry storms.
- **Health + readiness:** wire `mgr.AddHealthzCheck("healthz", healthz.Ping)` and `mgr.AddReadyzCheck("readyz", ...)` so the probes actually reflect manager state.
- **Leader election:** beyond `LeaderElection: true`, tune `LeaseDuration`/`RenewDeadline`/`RetryPeriod` for your API-server latency, and set `LeaderElectionReleaseOnCancel: true` so a graceful shutdown releases the lease immediately instead of forcing the next leader to wait out the lease.

## Status And Conditions
- Model conditions as `[]metav1.Condition` and mutate them only through `meta.SetStatusCondition` / `meta.RemoveStatusCondition`; read with `meta.FindStatusCondition` / `meta.IsStatusConditionTrue`. Do not hand-append condition slices.
- Write `observedGeneration` into both `status.observedGeneration` and each `metav1.Condition.ObservedGeneration` so consumers can tell a condition reflects the current spec, not a stale one.
- Expose the freshness contract explicitly: `status.observedGeneration == metadata.generation` means "controller has acted on the current spec." Gate readiness and downstream automation on it.

## CRD Ergonomics
Beyond admission validation, drive CRD UX from markers:
- `+kubebuilder:subresource:status` (required for the status-patch flow above), `+kubebuilder:printcolumn:...` for `kubectl get` output, and `+kubebuilder:resource:categories=...` so `kubectl get <category>` includes the CRD.
- **CRD validation ratcheting** (GA 1.33): the API server re-validates only fields that changed on update, so an object violating a newly tightened rule still updates as long as the offending field is untouched. This changes how you stage CEL tightening — you can land a stricter rule without breaking in-flight updates to existing objects, but must still plan a migration for the objects that do touch the field.
- Budget CEL cost per rule: each `x-kubernetes-validations` rule has an estimated cost limit, and the per-CRD total is capped. Keep expressions cheap (avoid unbounded loops/`all()` over large lists) or the API server rejects the schema.

## Logging And Observability
- Use contextual `logr` from `log.FromContext(ctx)` with keys-and-values (`logger.Info("reconciled", "namespace", ns, "name", name)`); controller-runtime already seeds reconcile request fields. Avoid `fmt`-formatted messages.
- Bridge to/from structured stdlib logging with `logr.FromSlogHandler` / `logr.ToSlogHandler` when a dependency speaks `log/slog`.
- Register custom Prometheus collectors through `sigs.k8s.io/controller-runtime/pkg/metrics` (`metrics.Registry`) rather than ad-hoc package-level counters, so they share the manager's authenticated endpoint.
- Optional: instrument reconcile with OpenTelemetry tracing (span per reconcile, keyed by request) when you need cross-component latency attribution.

## Default Workflow
1. Inspect API types, CRD markers, reconciler flow, watches, predicates, and owned resources before editing.
2. Identify the contract change: API shape, reconciliation behavior, status semantics, or lifecycle behavior.
3. Make the smallest safe change to types, reconciliation logic, or generated artifacts.
4. Re-check generation, tests, and upgrade implications before considering the task done.

## Knowledge-Graph Discovery (When Available)
If the repository carries a graphify knowledge graph (a `graphify-out/` directory), use it as a map to consult before broad text search — never as ground truth.
- Orient first from `graphify-out/GRAPH_REPORT.md` (or `graphify-out/wiki/index.md` when present): god nodes, communities, and cross-file relationships expose the API/reconciler topology and owned-resource wiring before you open a file.
- For "what watches this", "what owns that", and blast-radius questions, prefer `graphify query "<question>"`, `graphify path "<A>" "<B>"`, and `graphify explain "<symbol>"` over grep — they traverse extracted and inferred edges across package and controller boundaries that text search misses.
- Every edge is tagged `EXTRACTED`, `INFERRED`, or `AMBIGUOUS`. Treat `EXTRACTED` as structural evidence; treat `INFERRED` and `AMBIGUOUS` as leads to confirm with `gopls`, the generators, the compiler, or `envtest`. The graph never outranks executable evidence.
- After changing code, run `graphify update .` (AST-only, no API cost) to keep the graph current.

If no `graphify-out/` directory exists, ignore this section.

## Default Verification
- Run repository-standard generators for deep-copy, CRD, or manifests when needed.
- Prefer focused Go tests plus `envtest` for reconciliation behavior.
- Re-check finalizers, conditions, observed generation, and owner references for lifecycle correctness.
- Verify that controller behavior is safe across retries, duplicate events, and partial failure.

## Completion Criteria
- Reconciliation remains idempotent.
- Status and conditions reflect reality.
- API and CRD changes are intentional and migration-aware.
