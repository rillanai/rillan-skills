<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.3.0 -->
# Go Development Mode

## Purpose
Use this skill when the task is to implement or modify Go code: feature work, bug fixes, targeted cleanup, or small refactors coupled to a concrete behavior change.

Use it with `policy.md` and `workflow.md`; load another mode only if the task also requires that specialty.

## Implementation Guidance

- Deliver the requested behavior with the smallest coherent change and follow repository package, generator, build-tag, and tooling conventions.
- Preserve package and public API boundaries unless the requested behavior or an evidenced design constraint requires changing them.
- Reuse existing patterns when they fit. Avoid speculative interfaces, helpers, configuration, and unrelated cleanup.
- Inspect existing coverage before editing. Add regression coverage for a bug when it prevents recurrence, and use a failing or coverage-driving test for new behavior when practical.
- Keep side effects, cancellation, ownership, and error behavior explicit at the boundary where they matter.
- Complete the behavior, run the risk-based checks in `workflow.md`, and report any concrete verification gap or material uncertainty.
