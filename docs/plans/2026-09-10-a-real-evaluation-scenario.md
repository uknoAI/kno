# A real evaluation scenario

**Status:** Phase 0 — plan. Not implemented.

## Problem

Four separate things are blocked on the same missing artifact, and each has
been rediscovered independently:

- **`docs/debt.md#161`** — the Together and OpenAI adapters' inference routes are
  unconfirmed against live infrastructure. `docs/status.json` carries Bridge as
  `partial` for this reason and cannot honestly say otherwise.
- **`kno bridge` cannot run at all**, even with credentials: it group-ablates
  **behavior** Assets, `destinationFor` routes only `KIND_BEHAVIOR` to
  `DESTINATION_TUNING_SET` (`core/select.go`), and **every fixture in the tree is
  `kind: knowledge`.** A live attempt returns *"Nothing to bridge: the Portfolio
  has no tuning-set entries."*
- **The README quickstart** ends with three `+0.0000` deltas and `Selected 0` —
  the demo of a tool that measures which data helps, concluding that nothing
  helps.
- **`kno validate --agent tuned:`**, v0.3's headline and the payoff for
  everything since v0.2, needs a tuned model to point at.

All four need one thing: **an eval set a real model measurably fails, and Assets
that measurably fix it.** Nothing in the repository is that.

## Why this is not the rejected quickstart plan

A plan to make the quickstart demonstrate value was written and **rejected in
review** ([PR #201](https://github.com/uknoAI/kno/pull/201), closed — the file was
never merged, so it exists only there). Its reasons are load-bearing here.

It proposed making `fake:` condition its answer on the injected Asset. Review
found the mechanism inert (`fake:` already answers every Case correctly, so both
branches were identical), found it reversed a CI-enforced decision
([the demo plan](2026-08-30-kno-demo.md)'s criterion 6, whose stated purpose is
to *"catch anyone engineering the demo into looking impressive"*), and found the
repo had already concluded — in that plan, in `cli/demodata/README.md`, and in
`kno-examples`' `naive_ablation.py` — that **no zero-cost synthetic demo can show
an Asset earning its place without begging the question.**

`naive_ablation.py` states it plainly: *"an agent that returns the expected
answer cannot be improved by anything put in front of it."*

This plan accepts that conclusion rather than working around it. The numbers here
are true because they were **measured against a real model**, not because a
fixture was arranged to produce them. That is the whole difference, and it is why
this costs money and the rejected approach did not.

## Proposed design

### 1. What the scenario has to be

A domain where a competent small model fails **for a reason context actually
fixes** — not because the expected string is arbitrary. The failure mode must be
*missing knowledge or missing convention*, since that is what an Asset supplies.

Candidate shape, to be settled in review: a support/policy domain with rules the
model cannot infer — specific refund windows, escalation thresholds, naming
conventions. A model answers plausibly and **wrongly** without them, correctly
with them. Concretely checkable, and the failure is honest rather than a
tokenizer artifact.

**Explicitly not** trivia the model happens to know, and **not** Cases whose
`expected` string is unguessable phrasing. Both produce deltas that measure the
wrong thing.

### 2. Sizing, computed rather than asserted

`split.MinHoldout = 20` and `DefaultHoldoutFrac = 0.2`, so clearing the
underpowered-holdout warning needs **N ≥ 100 Cases**.

This is worth stating precisely because the rejected plan claimed sizing was
"set by what `MinHoldout` needs" while assuming a number **more than 8× too
small**. 100 Cases at a few hundred tokens each is real authoring work and real
per-run cost, and both belong in the estimate before anyone starts.

### 3. Both Kinds, because the bridge needs behavior Assets

The pool carries **`kind: behavior`** Assets — few-shot demonstrations, the thing
a tuning set is made of — alongside knowledge ones. Without behavior Assets there
is no `DESTINATION_TUNING_SET` entry, no Portfolio for `kno bridge` to ablate,
and `#161` stays open regardless of credentials.

### 4. Where it lives

**`uknoAI/kno-examples`**, as a scenario beside the existing ones, not in `kno`.
It is data with a verification harness, `kno-examples` already has `cmd/verify`
and a scenario runner, and `kno`'s own fixtures are deliberately tiny and
deliberately show nothing.

The consequence to accept: the quickstart GIF lives in `kno` and would need to
reference a scenario from another repository, or carry a trimmed copy. Review
should decide which.

### 5. Verification, and what happens when it rots

The scenario is only worth having if its claims stay true, and they will not stay
true by themselves — models change under a fixed name, and a model that improves
may stop needing the Assets at all.

So: a **nightly live run** with a capped budget that re-measures and fails when
the effect disappears. That failure is a **finding, not a flake** — it means the
model changed, and the scenario's own README should say what to do about it
rather than leaving the next person to discover it at 3am.

## Alternatives considered

**A. Keep everything synthetic and accept the zeros.** The status quo. Rejected
because four separate workstreams are blocked on it, but worth naming: it is
honest, it costs nothing, and it is what the repo chose deliberately once.

**B. Record fixtures from one real run and replay them.** Cheap to re-run,
deterministic, and no per-run spend. Rejected as the *primary* artifact and
recommended as a companion: a recording cannot detect that the live model
changed, which is the failure mode §5 exists for. Record from the live run, use
the recording in PR CI, keep the live run nightly.

**C. Use an existing public benchmark.** Rejected: benchmarks are built so that
context does not trivially fix them, which is the opposite of what is needed
here, and licensing is a separate problem.

## Edge cases

1. **The Assets help too much.** If every routed Case flips, the interval is
   degenerate and Select reports a suspiciously perfect delta. Some Cases must
   stay unhelped, so the CI has width and the demo shows a real interval.
2. **The Assets help too little.** Below the Bonferroni-corrected threshold over
   the screened set, nothing selects and we are back where we started. The
   scenario must be verified end-to-end **before** it is committed, not after.
3. **A model that improves.** §5's nightly failure. Recorded as a finding.
4. **Cost per verification run.** N ≥ 100 Cases × arms × groups is not free, and
   the nightly cap must be set deliberately rather than inherited.
5. **Holdout leakage in authoring.** Cases and Assets are written by the same
   hand; an Asset that quotes a holdout Case's expected answer verbatim
   manufactures the result. The authoring process needs a check, not just care.

## Test plan

- The scenario runs end to end against a real provider and produces: a non-empty
  Portfolio, at least one **selected** Asset and at least one **rejected** one,
  and a `tuning_set` export with behavior Assets in it.
- `kno bridge` reaches submission against that Portfolio — which is what clears
  `#161`.
- The recorded-fixture replay (alternative B) reproduces the same verdicts in PR
  CI without spending.
- A holdout-leakage check over the authored files: no Asset contains a holdout
  Case's expected string.

## Rollback

Pure addition in another repository. Nothing in `kno` depends on it until the
quickstart is repointed, which is a separate change.

## Docs impact

`kno-examples`' scenario README and index; `docs/status.json`'s Bridge claim if
the live run succeeds; the README quickstart if it is repointed;
`docs/debt.md#161`'s disposition.

## Accepted risks

*To be filled by Phase 1 review, and mirrored to `docs/debt.md` with triggers.*

Two the review should weigh:

1. **This costs money on a schedule, forever.** A nightly live run is a recurring
   bill and a recurring source of alerts. Is the honesty worth it, or is a
   quarterly manual verification enough?
2. **A hand-authored scenario is a claim about the product made by its authors.**
   It is more honest than a rigged fake, and it is still us choosing the ground.
   Review should ask whether that is worth stating in the README beside the
   numbers.
