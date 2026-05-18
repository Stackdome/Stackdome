---
name: grill-with-docs
description: Grilling session that challenges your plan against the existing domain model, sharpens terminology, and updates documentation (CONTEXT.md, ADRs) inline as decisions crystallise. Use when user wants to stress-test a plan against their project's language and documented decisions.
---

> Spine note: `grill-with-docs` hardens a design against the project's language and past decisions. It works best on a superpowers spec (`docs/superpowers/specs/`) or plan (`docs/superpowers/plans/`) if one exists, but does not require one — it can grill a design in conversation. It is not a replacement for `superpowers:brainstorming` (which explores *what* to build); use it to pressure-test that the language and decisions hold. Glossary target is multi-context: write the correct per-context `CONTEXT.md` per root `CONTEXT-MAP.md`; maintain `CONTEXT-MAP.md`; create files lazily.

Spec ↔ ADR rule: the spec/plan is the source design doc. Extract into `docs/adr/` ONLY decisions clearing the 3-criteria bar (hard-to-reverse + surprising-without-context + real trade-off). ADRs are the durable decision residue, not duplicates. Cross-link both ways (ADR cites spec path; spec notes "decision → ADR-NNNN"). `CONTEXT.md` holds terms only — orthogonal to ADRs.

<what-to-do>

Interview me relentlessly about every aspect of this plan until we reach a shared understanding. Walk down each branch of the design tree, resolving dependencies between decisions one-by-one. For each question, provide your recommended answer.

Ask the questions one at a time, waiting for feedback on each question before continuing.

If a question can be answered by exploring the codebase, explore the codebase instead.

</what-to-do>

<supporting-info>

## Domain awareness

During codebase exploration, also look for existing documentation:

### File structure

Most repos have a single context:

```
/
├── CONTEXT.md
├── docs/
│   └── adr/
│       ├── 0001-event-sourced-orders.md
│       └── 0002-postgres-for-write-model.md
└── src/
```

If a `CONTEXT-MAP.md` exists at the root, the repo has multiple contexts. The map points to where each one lives:

```
/
├── CONTEXT-MAP.md
├── docs/
│   └── adr/                          ← system-wide decisions
├── src/
│   ├── ordering/
│   │   ├── CONTEXT.md
│   │   └── docs/adr/                 ← context-specific decisions
│   └── billing/
│       ├── CONTEXT.md
│       └── docs/adr/
```

Create files lazily — only when you have something to write. If no `CONTEXT.md` exists, create one when the first term is resolved. If no `docs/adr/` exists, create it when the first ADR is needed.

## During the session

### Challenge against the glossary

When the user uses a term that conflicts with the existing language in `CONTEXT.md`, call it out immediately. "Your glossary defines 'cancellation' as X, but you seem to mean Y — which is it?"

### Sharpen fuzzy language

When the user uses vague or overloaded terms, propose a precise canonical term. "You're saying 'account' — do you mean the Customer or the User? Those are different things."

### Discuss concrete scenarios

When domain relationships are being discussed, stress-test them with specific scenarios. Invent scenarios that probe edge cases and force the user to be precise about the boundaries between concepts.

### Cross-reference with code

When the user states how something works, check whether the code agrees. If you find a contradiction, surface it: "Your code cancels entire Orders, but you just said partial cancellation is possible — which is right?"

### Update CONTEXT.md inline

When a term is resolved, update `CONTEXT.md` right there. Don't batch these up — capture them as they happen. Use the format in [CONTEXT-FORMAT.md](./references/CONTEXT-FORMAT.md).

`CONTEXT.md` should be totally devoid of implementation details. Do not treat `CONTEXT.md` as a spec, a scratch pad, or a repository for implementation decisions. It is a glossary and nothing else.

When multiple contexts exist (indicated by `CONTEXT-MAP.md`), infer which context the current topic relates to and update the correct per-context `CONTEXT.md`. If unclear, ask. Always keep `CONTEXT-MAP.md` up to date.

### Offer ADRs sparingly

Only offer to create an ADR when all three are true:

1. **Hard to reverse** — the cost of changing your mind later is meaningful
2. **Surprising without context** — a future reader will wonder "why did they do it this way?"
3. **The result of a real trade-off** — there were genuine alternatives and you picked one for specific reasons

If any of the three is missing, skip the ADR. Use the format in [ADR-FORMAT.md](./references/ADR-FORMAT.md).

Cross-link every ADR you create: add the spec/plan path inside the ADR, and add a note in the spec/plan referencing the ADR number (e.g. "decision → ADR-NNNN").

</supporting-info>
