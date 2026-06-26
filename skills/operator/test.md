<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.1.0 -->
# Kubernetes Operator Test Strategy And Generation

## Purpose
Use this skill when designing or writing tests for Kubernetes operators built in Go.

Apply this skill with `go/policy.md`, `go/workflow.md`, and `go/test.md` for base Go testing standards.

## Tool Use
This skill is tool-agnostic and works with Claude Code, Codex, OpenCode, and similar assistants. Map its guidance to whatever file-reading, editing, search, and shell-execution tools your environment exposes.

- Run the tests you write. A new reconciler or envtest case is not done until `go test`, `envtest`, or the repo's runner has executed it and the result is known.
- Regenerate CRDs and deep-copy code via `make generate`/`make manifests` after API changes; do not rely on stale generated files.
- Issue independent tool calls (reading types, predicates, watch setup, existing tests) in parallel.
- Report failing envtest runs with the exact command and output that produced them.

## Test Layers
- pure Go unit tests for helpers, predicates, mapping functions, and condition logic
- focused reconciler tests for branching, retries, and error handling
- `envtest` for API server backed reconciliation behavior
- admission-control tests: CRD CEL validations (`x-kubernetes-validations`) and `ValidatingAdmissionPolicy` enforcement exercised through `envtest`, since both run in the API server
- webhook tests for validation and defaulting (for the cases that remain in webhooks)
- generation checks for CRDs and manifests when code generation is part of the repo

## Guidance
- Prioritize reconciliation invariants: idempotency, ownership, finalizers, status updates, and event ordering resilience.
- Test status and conditions as part of the contract, not as incidental output.
- Prefer deterministic fake-client tests only for narrow logic; use `envtest` when API-server semantics matter.
- Cover deletion paths, not-found paths, partial creation, and requeue-after-error behavior.
- Treat CRD version changes, admission policies, and webhooks as high-risk areas requiring explicit coverage.
- For `ValidatingAdmissionPolicy` and CRD CEL rules, drive `envtest` with the policy + binding applied and assert that violating objects are rejected and conforming objects admitted — test the CEL the same way you'd test a webhook's deny path. Confirm `failurePolicy`/`validationActions` behave as intended, and verify the envtest API-server version actually supports VAP (GA 1.30) before relying on it.

## Test Toolchain
- Provision API-server/etcd binaries with `setup-envtest` (pin a specific Kubernetes version so VAP/ratcheting support is deterministic) and point `KUBEBUILDER_ASSETS` at its output.
- Use Ginkgo v2 + Gomega for envtest suites; assert async reconciler effects with `Eventually`/`Consistently`, and prefer `EventuallyWithT`/`Gomega.Eventually(func(g Gomega){...})` so each poll re-fetches and re-asserts instead of closing over a stale object.
- Know the fake client's limits: it does **not** faithfully emulate server-side apply (field managers/`ForceOwnership`) and mishandles some subresources/status semantics. Test SSA writes, conditions freshness, and admission against `envtest`, not the fake client.
- For time-based requeue logic (`RequeueAfter`, backoff, lease timing), use `testing/synctest` (Go 1.25+) to drive a fake clock deterministically instead of real sleeps.
