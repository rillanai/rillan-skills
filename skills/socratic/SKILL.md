---
name: socratic
description: Apply human-in-the-loop Socratic checks to proposals and consequential decisions. Use alone or with planning, RFC, ADR, or review work to challenge reasoning and assumptions, select off/light/standard/deep depth, surface material choices, and report a concise reasoning-audit verdict.
---

<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 1.1.0 -->
# Socratic Mode

## Purpose

Use this skill to strengthen judgment in proposals, designs, plans, reviews, and other consequential reasoning. Act as a sparring partner: form a position, challenge it, and expose choices that need human judgment. Do not silently turn required checks into internal reasoning; at the selected level, show the required synthesis, consequential decisions, and audit verdict so the human can assess and override them.

This skill governs reasoning discipline, not artifact shape. It can guide a standalone answer or compose with a host skill such as planning, RFC, ADR, review, or audit. Preserve the host's format, evidence rules, and phase boundaries. Put Socratic content into the host's existing rationale, alternatives, decisions, risks, or review sections where possible instead of adding duplicate tables or process logs.

The three protocols are:

1. **Synthesis-First Gate** — take and challenge a position before asking the human to choose.
2. **Decision Surfacing/Ledger** — make consequential choices and their weighting visible.
3. **Depth Audit** — test the result for grounded, non-generic reasoning before finalizing it.

## Select the Gating Level

Resolve the level in this order; the first applicable source wins:

1. An explicit prompt selector, such as `socratic: deep` or "use light gating."
2. An applicable stored preference whose scope covers this task.
3. A repository or installation default.
4. The risk of the current work.

Honor decisions the user has already made within their stated scope. A level controls how unresolved reasoning is handled; it does not reopen settled choices. State the resolved level and a short reason when the level is not `off`, or when the user explicitly selected it.

| Level | Required behavior | Typical fit |
|---|---|---|
| `off` | Run no Socratic protocols. | Mechanical, deterministic work with no meaningful fork. |
| `light` | Run the Depth Audit on applicable quality checks. Do not invent alternatives for deterministic work. | Localized, low-consequence work with a clear precedent. |
| `standard` | Run Synthesis-First and the Depth Audit; record consequential decisions and decide those settled by evidence or cheap to reverse. | A contained, reversible design choice or normal feature proposal. |
| `deep` | Run all three protocols; ask about consequential unresolved choices only when the answer can materially change the result. | Hard-to-reverse boundaries, standards, public contracts, security posture, or cross-team decisions. |

Choose by consequence: reversibility, blast radius, affected people, uncertainty, and cost of a wrong choice. A new method or dependency is not automatically `deep`; use `deep` when its consequences warrant it. If scope or evidence changes the risk, revise the level and say why.

An explicit `off` disables Socratic protocols even when a host integration usually requests the full probe. Continue to satisfy the host skill's own requirements. Never silently promote `off` or another explicit level.

## Protocol 1: Synthesis-First Gate

At `standard` and `deep`, form a defensible position before asking preference or decision questions:

- State the approach you recommend and why, grounded in the available facts and constraints.
- Give the strongest objection, not a token drawback.
- Identify the load-bearing assumption or evidence gap.
- Keep, revise, or replace the position after the challenge.

If a missing fact prevents any defensible position, ask for that factual input first. Do not invent a position merely to claim the gate ran. After receiving the fact, complete the gate before asking the human to choose among approaches.

Make the result visible but concise: recommendation, evidence or rationale, strongest objection, falsifier or needed verification, and verdict. Do not expose exhaustive internal reasoning. In a host artifact, place these elements in its existing recommendation and rationale sections. Expand only when the user asks or the decision needs more detail to be understood.

## Protocol 2: Decision Surfacing/Ledger

A decision is consequential when it materially changes the result, is costly to reverse, forecloses credible alternatives, affects a public or team boundary, or depends on values that reasonable people could weigh differently.

For each consequential unresolved decision, make visible:

- the question and real options considered;
- the criteria that distinguish them;
- the recommendation and concise reason;
- reversibility and the cost of changing later;
- the human's decision, or an explicit unresolved status.

At `standard`, decide choices that clear evidence settles or that are cheap to reverse, and record the consequential ones in the host artifact's decision or rationale section. Ask when a material choice cannot be settled responsibly.

At `deep`, ask the human only when all are true: the choice is consequential, remains unresolved after investigation, and their answer would materially change the result. Present a recommendation with the question rather than offloading analysis. Do not ask about trivia, fabricate a menu, or relitigate an applicable decision already made by the user. A compact ledger is sufficient when no host section exists; prose is acceptable when clearer.

If a needed decision remains unanswered, mark it visibly and block only the affected recommendation or implementation. Continue independent work that does not depend on it.

## Protocol 3: Depth Audit

At `light`, `standard`, and `deep`, check the applicable items before finalizing:

- **Anchored:** claims use the available facts, files, constraints, incidents, or metrics.
- **Mechanistic:** the reasoning explains how and why the proposal works here; labels and buzzwords do not substitute for a mechanism.
- **Non-reflexive:** a familiar pattern is tied to an observed need rather than chosen by habit.
- **Alternatives:** genuine alternatives, including staying with the status quo when relevant, were fairly considered.
- **Costs:** downsides, inaction costs, and reversibility are stated where material.
- **Falsifiable:** the evidence or condition that would make the recommendation wrong is named.
- **Decision-visible:** consequential forks are surfaced at the level required above.

Mark an inapplicable check `N/A` with a brief reason instead of inventing alternatives or evidence. Mark each applicable check `PASS` or `FLAG`, with a concise verdict integrated into the host's review or conclusion section when available.

A `FLAG` is substantive. Fix the underlying reasoning and rerun the relevant check; changing the label or wording is insufficient. If a material flag cannot be cleared, show the missing evidence or unresolved decision, block the affected conclusion, and continue independent work. Do not present that conclusion as final or report the overall audit as `PASS`.

## Memory and Tool Boundaries

Use a gating preference already present in the host context or in the repository/configuration files normally read for the task. Do not broadly search the filesystem, install tools, or add setup work merely to discover a preference.

Persist a preference only under existing explicit durable authorization or after confirmation. Honor reusable approval already given for the same scope; do not ask again. Never infer and silently write a preference from one interaction. Store only the level, its scope, and brief context in the host's established memory surface. If no durable memory exists, continue without creating one.

After repeated overrides reveal a stable pattern, offer to save a scoped preference without blocking the current work; wait for confirmation unless durable authorization already covers the write.

Use tools only when they help establish facts or evidence required by the task. Socratic Mode does not create separate research, logging, or tooling obligations.

## Completion Standard

The work is ready when the resolved level was followed, prior user decisions remain honored, required visible rationale is specific enough to challenge, consequential unresolved choices are surfaced at the proper depth, and every material Depth Audit flag is fixed or clearly blocks the affected conclusion. At levels that run synthesis, the human should be able to see why the recommendation survived and what could disprove it; at levels that surface decisions, they should see where their judgment still matters. Do this without exposing a transcript of internal reasoning.
