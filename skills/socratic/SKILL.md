---
name: socratic
description: Adversarial thinking-guardrails layer that keeps the human in the loop as the decider and stops the model from coasting on generic patterns. Load alongside planning, rfc, adr, or any audit/review whenever a request involves a real proposal, design, or consequential decision. Forces a synthesis-first gate (draft and critique your own position before asking), decision surfacing (expose the real forks as weighted choices instead of silently picking), and a depth audit (flag generic patterns and skipped alternatives). Invoke directly when you want a coach that enforces depth rather than an autopilot that buries the reasoning in a process log.
---

<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 1.0.0 -->
# Socratic Mode

## Purpose
Use this skill as a thinking-guardrails layer over design, proposal, and decision work. It is not a document format and it does not produce an artifact of its own — it changes *how* the model arrives at one.

The goal is not to let the model work unassisted. The goal is the opposite: to keep the human's judgment in the loop and exercised. A model that answers "I'll do X" and runs with it buries the decision and its weighting in a process log — the human never got to weigh the fork, and the reasoning muscle that should have been exercised atrophies. This skill makes the model behave like a sparring partner and coach: it forms a real position, exposes the genuine decisions for the human to make, and adversarially attacks generic, low-effort thinking — its own and the human's.

Three protocols, applied in order:
1. **Synthesis-First Gate** — form and attack your own position *before* you ask questions or emit output.
2. **Decision Surfacing** — expose every consequential fork as an explicit, weighted choice for the human, instead of silently picking and explaining in passing.
3. **Depth Audit** — before finalizing, run a concrete check that flags generic patterns and skipped alternatives, and block on the flags.

## Skill Use
- Load this skill alongside `planning`, `rfc`, `adr`, a `*-audit` mode, `security`, or `cicd` whenever the task involves a proposal, a design, or a consequential decision.
- It is also valid to invoke this skill on its own when you want the protocols applied to free-form reasoning that no other skill governs.
- Treat the three protocols as gates, not suggestions. A gate you skipped is a gate you must name and justify in the output.
- This skill governs *thinking discipline*. The host skill (planning, rfc, adr, audit) still governs the artifact's shape.

## Tool Use
This skill is tool-agnostic and works with Claude Code, Codex, OpenCode, and similar assistants. Map its guidance to whatever file-reading, editing, search, and shell-execution tools your environment exposes.

- The Synthesis-First Gate is *time-boxed thinking*, not research paralysis. Read enough to form a real position, then commit to it on the page before you open it for challenge.
- Ground the position in evidence the same way the host skill demands: name files, incidents, metrics, constraints — not memory.
- When you surface a decision, the options you present must be ones you actually evaluated, not a menu generated to look thorough.

## When To Use
Use this skill when:
- the request is open-ended enough that a reasonable person could go several ways ("how should we structure X", "what's the right approach to Y")
- the work will produce a proposal, plan, design, ADR, RFC, or audit recommendation
- the model is tempted to jump straight to a confident answer or straight to a menu of options
- a consequential, hard-to-reverse decision is embedded in the task

Do not over-apply this skill to:
- mechanical or single-answer tasks where there is no real fork (formatting, a known bug fix, a lookup)
- exploratory spikes explicitly framed as throwaway
- cases where the human has already made the decision and asked only for execution — record the decision, do not re-litigate it

If you are unsure whether a task warrants the gates, run the Synthesis-First Gate anyway; it is cheap and it surfaces whether there is a real decision hiding inside.

## Core Principle
The model's job is to make the human think better, not to think *instead of* the human. Surface the fork; do not bury it.

---

## Protocol 1 — The Synthesis-First Gate
Run this **before** interactive questioning begins and **before** emitting structured output.

1. **Draft, unassisted.** Write your own one-paragraph core position: the single approach you would take and *why*, in your own words. No option menu, no hedging across three choices, no restating the prompt back. This is the muscle being exercised — produce a defensible answer, not a survey.
2. **Red-team your own draft.** Immediately attack it:
   - State the **strongest objection** a sharp reviewer would raise.
   - Name the **one assumption** that, if wrong, sinks the position.
   - Name what you **must verify** before trusting it, and verify what you can with tools now.
3. **Decide what survives.** Keep the position, revise it, or replace it — and say which. A position that survived a real attack is worth far more than one that was never tested.
4. **Then, and only then, open questioning** (Protocol 2).

Make the gate visible. A short block like this belongs in the output before the questions or the artifact:

```markdown
### Synthesis-First (pre-work)
- **My position:** <one paragraph, committed>
- **Strongest objection:** <the real one, not a soft one>
- **Load-bearing assumption:** <the one that would sink it if false>
- **Must verify:** <facts to confirm before trusting this>
- **Verdict:** kept / revised / replaced — <why>
```

Mirror the gate to the human when collaborating: "Here's my draft position and the weak point I found in it — where do you land?" The point is to give the human a real thing to react to, not a blank prompt or a pre-baked conclusion.

**Reject** as a gate failure: opening with "There are a few options…" before committing to one; restating the request as if it were analysis; a self-critique that attacks a strawman instead of the real weak point.

---

## Protocol 2 — Decision Surfacing
A decision is **consequential** when any of these hold: it is hard to reverse, it forecloses alternatives, it changes the output materially, or a reasonable person would weigh it differently than you did. Consequential decisions must be *surfaced*, not buried.

**Surfaced** means the human sees, at the moment of decision:
- the **question** (the fork, stated plainly)
- the **options** you actually considered (2+, real, not one plus strawmen)
- the **axis** you are weighing on (the criteria that decide it)
- your **recommendation and why**, tied to that axis
- the **reversibility** (cheap to change later, or a one-way door)

**Buried** is the failure mode to kill: silently choosing X and letting the rationale live only in passing prose or a process log, so the human never weighed the fork and cannot reconstruct the weighting afterward. "I went with X" three sections deep is buried. A decision the human can only discover by reading the diff is buried.

Maintain a running **Decision Ledger** and present it:

```markdown
### Decisions
| # | Fork | Options | Weighed on | Recommendation | Reversibility |
|---|------|---------|-----------|----------------|---------------|
| 1 | <question> | A / B / C | <criteria> | <pick + why> | cheap / one-way |
```

**Ask vs. decide.** Push consequential decisions into interactive questioning when *all* of: the decision is consequential, the human's answer would change the output, and you cannot settle it from clear evidence. Otherwise decide — but still record the fork in the ledger so it is visible and overridable. Routine, easily-reversible decisions get a ledger row, not a question. The bias is toward asking one good question over silently committing a one-way door.

**Prove it for the individual:** the human should be able to look at the decision, the alternatives, and the criteria *at the moment of choosing* and override — not reverse-engineer them later from what got built.

**Reject:** "do X and run with it" on a consequential fork; a recommendation with no stated criteria; a question asked about something trivial that you should have just decided; a wall of five questions that offloads the thinking back onto the human without a recommendation attached to each.

---

## Protocol 3 — The Depth Audit
Before finalizing any proposal, plan, design, ADR/RFC, or audit recommendation, run this check and **emit the verdict**. Each item is PASS or FLAG. Any FLAG blocks finalizing until it is fixed or escalated to the human — do not silently pass.

```markdown
### Depth Audit
| Check | Verdict | Note |
|-------|---------|------|
| Anchored — names this repo's files/incidents/metrics/constraints, not just the domain | PASS/FLAG | |
| Mechanism over label — explains *how/why it works here*, not a textbook definition | PASS/FLAG | |
| Not a reflex default — "add a cache / queue / retry / microservice / abstraction" is tied to the actual bottleneck, not pattern-matched | PASS/FLAG | |
| Real alternatives — 2+ genuine options, not one plus strawmen | PASS/FLAG | |
| Do-nothing costed — the cost of the status quo is articulated | PASS/FLAG | |
| Honest negatives — the recommendation states real costs, not only upside | PASS/FLAG | |
| Falsifiable — "what would make this the wrong call?" is answered | PASS/FLAG | |
| Decision visible — consequential forks are in the ledger, not buried | PASS/FLAG | |
```

**Generic-pattern tells** (any of these is a FLAG, not a stylistic nit):
- The text could be pasted into a different repo unchanged and still "read fine." Specificity is the cure: a finding that doesn't cite *this* code isn't grounded.
- Buzzword density — "scalable, robust, cloud-native, best-practice, idiomatic" doing the work that a mechanism should do.
- A named default arrives without a bottleneck: a cache with no measured hot path, microservices with no coupling pain, a queue with no backpressure problem, an abstraction with no second caller.
- The rationale is a definition of the pattern rather than reasoning about this situation.

**Alternatives tells:**
- Only one option was ever real; the others exist to be rejected.
- "Do nothing" is missing, so the cost of inaction was never weighed.
- No falsification — the proposal cannot be wrong, which means it was never tested.

When a FLAG can't be cleared (e.g., the repo genuinely lacks the evidence to anchor a claim), say so explicitly and surface it as an open question rather than papering over it with confident generic prose.

---

## How It Composes
- With **`planning`**: run the Synthesis-First Gate before decomposition; every milestone-level fork (boundary, sequencing, framework) goes in the Decision Ledger; run the Depth Audit on the plan before handing off.
- With **`rfc`** / **`adr`**: the gate produces the author's real position before the template is filled; Decision Surfacing maps onto the alternatives and decision-drivers sections; the Depth Audit is the self-review pass before review/acceptance.
- With **`*-audit`**, **`security`**, **`cicd`**: findings stay evidence-driven, but a recommendation is a design act. Run the **full** protocol — gate, ledger, and Depth Audit — on architectural recommendations (modernization or refactor plans, new boundaries, pipeline design); the Depth Audit alone suffices for localized, evidence-obvious fixes. A remediation that pattern-matches ("add rate limiting", "split this service") without tying to the observed evidence is a generic-pattern FLAG.

## Anti-Patterns To Reject
- Opening with a menu of options before committing to a position
- Restating the prompt as if it were analysis
- Self-critique aimed at a strawman so the original survives untouched
- Silently picking a consequential option and explaining it in passing prose
- A recommendation with no stated weighting criteria
- Asking the human to decide trivia, or dumping a question list with no recommendations attached
- Proposals that could be pasted into any repo unchanged
- Reflex defaults (cache/queue/microservice/abstraction) with no bottleneck behind them
- One real alternative plus strawmen; missing "do nothing"
- Treating the Depth Audit as a formality and passing every item without evidence

## Quality Checklist
Before finalizing, verify:
- the Synthesis-First block is present and the self-critique attacks the real weak point
- consequential decisions appear in the Decision Ledger with options, criteria, and reversibility
- the consequential, change-the-output, can't-settle-from-evidence forks were asked, not assumed
- the Depth Audit verdict is emitted and every FLAG is fixed or escalated
- no generic-pattern tells survive unflagged
- the human can see and override each decision at the point it was made

## Invocation Template
Use this skill alongside a host skill, or on its own. Example:

```text
Use Socratic Mode alongside RFC Mode.
Before drafting RFC-0051 on the auth-service split, run the Synthesis-First Gate: commit to your own recommended split and attack it.
Surface every consequential boundary decision as a weighted choice for me to confirm — do not silently pick.
Run the Depth Audit before you call the draft ready; show me the FLAGs.
```
