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

### 2. Sizing — and the half of the arithmetic the first draft missed

`split.MinHoldout = 20` at `DefaultHoldoutFrac = 0.2` means **N ≥ 100 Cases**.

**That is only half the calculation, and Phase 1 review showed the other half is
the whole problem.** N sets the interval's width; nothing in the first draft
asked what **effect size** must clear it.

Work it with the repo's own committed numbers. At N=100: dev ≈ 80,
`DefaultControlReserve = 0.3` reserves 24 for the control arm, leaving ≈56
routable. With 5 Assets screened, `correctedLevel` gives 1 − 0.05/5 = 0.99, and
`portfolio.Correct` widens the interval by roughly 1.3–1.5×. The evidence for
what `n_pairs` a per-tag cluster actually yields at this scale is in the sibling
repo: `support-refunds` (12 Cases) measures **5 pairs** at a raw half-width of
**±0.396**; `power-analysis` needs **75 pairs** to reach ±0.036.

A cluster at N=100 sits nearer the first regime than the second. The delta needed
to clear a Bonferroni-corrected bound is therefore somewhere around **0.5–0.6 in
accuracy space** — which is not "the model sometimes gets the policy wrong and
the Asset fixes it." It is "wrong on essentially every routed Case without the
Asset, right with it."

**That is the rejected plan's all-or-nothing shape, reproduced with real money.**
Nobody would set out to reintroduce circularity; the power budget would force it
under schedule pressure.

So the first step is **not authoring**. It is a free, offline feasibility check
(§2a). If that check says the required effect is implausible for a
genuinely-missing-knowledge domain at any Asset count the budget affords, then
**the honest zero is the right answer**, and this plan should conclude that in
writing rather than leave it as the thing nobody checked.

### 2a. The feasibility gate — free, offline, and first

Before a hundredth Case is written:

1. Draft **20–30** Cases, not 100.
2. Run **`kno eval inspect`** over them — no LLM call, no spend, and the tool
   built for exactly this question. It reports `separable_effect` per tag.
3. Run a **cheap pilot** against a real inexpensive model to get an empirical
   baseline failure rate and an observed effect size.
4. Compare the observed effect against the corrected bound the full N would
   impose. **Only then** decide whether to author the rest.

`kno-examples`' `power-analysis` scenario is the worked example of this exact
argument. Using it before spending is the plan; discovering it afterwards is
what the first draft would have done.

### 2b. Invented facts, not plausible policy

The domain should turn on **arbitrary invented conventions** — the shape
`diy-ablation` uses with its invented API — rather than real-world policy like
refund windows.

A model can hit "5 business days" by prior or luck at a non-trivial rate, which
dilutes the effect exactly where the power budget is thinnest. An invented
convention has a controllable near-zero baseline hit rate, which is what makes
the measured delta mean what it says.

### 3. Both Kinds, because the bridge needs behavior Assets

The pool carries **`kind: behavior`** Assets — few-shot demonstrations, the thing
a tuning set is made of — alongside knowledge ones. Without behavior Assets there
is no `DESTINATION_TUNING_SET` entry, no Portfolio for `kno bridge` to ablate,
and `#161` stays open regardless of credentials.

### 3a. Its relationship to `support-refunds`

`kno-examples/scenarios/support-refunds` already exists with the same four tags
and the same shape — and its README says outright that **no scenario in that
repository can show an Asset earning its place.**

The first draft proposed a support/policy domain without noticing. Two authors
converging on the same domain from different directions is itself informative
about what is easy to make legible, and §2b now argues for invented conventions
instead. Whatever domain is chosen, the plan must state plainly whether this is
that scenario's live-model successor — named so a reader is not left guessing
which is "the real one" — or a deliberately different thing, and why.

### 4. Where it lives

**`uknoAI/kno-examples`**, as a scenario beside the existing ones, not in `kno`.
It is data with a verification harness, `kno-examples` already has `cmd/verify`
and a scenario runner, and `kno`'s own fixtures are deliberately tiny and
deliberately show nothing.

The consequence to accept: the quickstart GIF lives in `kno` and would need to
reference a scenario from another repository, or carry a trimmed copy. Review
should decide which.

### 5. Verification — human-gated, not nightly

**The first draft proposed a nightly live run. `kno-examples` has already
rejected that pattern in writing, for this exact failure mode, one directory
from where this would land.**

Its `nightly.yml` header: *"NOTHING HERE SPENDS MONEY. Every command runs
against `fake:`, offline, with committed data... Prime directive 4 applies to
our own CI."* And `VERIFICATION.md`: *"nightly runs against eight third-party
SaaS products... would spend money unattended and redden on every vendor's
outage, training everyone to ignore the signal. We took the honest ceiling over
the impressive-looking one."*

The existing answer is `vendor-smoke.yml`: `workflow_dispatch` only, behind an
environment with required reviewers. **This adopts that pattern**, on a real
cadence someone owns, rather than proposing what that repo already declined.

There is a second reason nightly-live cannot work as drafted.
`kno-examples`' verification model has exactly three tiers — `executed`,
`flags-only`, `manual` — and `executed` means `cmd/verify scenario --repeat 2`:
**byte-identical output across two runs** against committed expectations. A live
LLM call cannot satisfy that. Tolerance-band pass/fail against a drifting live
target is a **fourth tier that does not exist**, and inventing one silently
inside a scenario PR is not the way to add it.

### 5a. The harm test cannot fire at this N — say so

`core/select.go` skips the REGRESSION rule entirely when `ControlUnderpowered`
is set. `power-analysis` reports that **even at N=160** with 39 control Cases the
harm bound is underpowered. At N=100 the control arm is ≈24, so it is
underpowered too.

That means the thing `kno bridge`'s own help advertises — *"does tuning on them
regress anything else"* — **cannot be demonstrated by this scenario**, and a run
where nothing regresses would read as "validated safe" when the honest statement
is "the harm test had no power to say either way."

That is precisely the confusion `underpowered-eval` exists to prevent, and
reproducing it in this plan's headline deliverable would be worse than not
shipping it. The scenario's README states it plainly, or the scenario needs a
materially larger control reserve — which tightens §2's tension further.

## Alternatives considered

**A. Keep everything synthetic and accept the zeros.** The status quo. Rejected
because four separate workstreams are blocked on it, but worth naming: it is
honest, it costs nothing, and it is what the repo chose deliberately once.

**B. Record fixtures from one real run and replay them.** Cheap to re-run,
deterministic, and no per-run spend. Rejected as the *primary* artifact and
recommended as a companion: a recording cannot detect that the live model
changed, which is the failure mode §5 exists for. Record from the live run, use
the recording in PR CI, keep the live run nightly.

**B-note:** Alternative B is **new harness work, not a checkbox.** `cmd/verify`
has five subcommands and none of them record or replay a live run with
tolerance. Scoping it as "a companion" understated it.

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
4. **Two cost classes, not one.** Eval calls (baseline, value's with/without
   passes, validate) are cents to low dollars per run. **Bridge is categorically
   different money**: `cli/bridge.go` states each ablation group is one
   fine-tuning job, *"charged when it is submitted and cannot be un-submitted"*,
   plus hosting *"per minute per endpoint, including while idle."*

   Re-verification covers the **eval tier only**. The bridge tier runs
   human-gated and infrequently, mirroring `vendor-smoke.yml`. Conflating them
   is how a recurring unattended fine-tuning bill gets approved by accident.
5. **Holdout leakage is mechanically precomputable, and a content check cannot
   catch it.** `split.AssignSplit` with an empty seed — the normal case — is a
   pure unsalted FNV-1a hash of the Case ID. An author can compute **offline,
   before writing a single Asset**, exactly which Cases land in holdout.

   A verbatim-string check catches neither a paraphrase nor the sharper exploit:
   an author who, knowing the split, tunes Asset content against what verifies
   well on **dev** while leaving holdout Cases untouched by design. Nothing in
   the Asset text references the holdout, so no content matching can see it, and
   the entire point of dev/holdout separation is gone without anyone "accessing"
   the holdout.

   **Required, not a check to be filled in later:** a non-empty `--split-seed`
   chosen and revealed only **after** the Assets are frozen and committed, so
   the split the shipped run uses is not one the author could have precomputed
   against. Authoring Cases and Assets in separate passes strengthens it
   further.

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

## Phase 1 review outcome

**Did not pass.** Three blockers, all verified against the code before amending:
the effect-size arithmetic the draft never did (§2), a verification mechanism the
target repository had already rejected in writing (§5), and a harm test that
cannot fire at the proposed N (§5a). Plus a holdout exploit that no content check
can catch, because the split is precomputable.

The review's own conclusion, adopted here: this is not "accept the zero forever"
and it is not "build it." It is **run the free, offline feasibility check first**
(§2a) — and if the effect required is implausible for an honest domain at any
affordable Asset count, conclude that in writing, as an argued finding rather
than the default nobody checked.

## Accepted risks

*To be filled by a second Phase 1 review, once §2a has produced numbers.*

Two the review should weigh:

1. **This costs money on a schedule, forever.** A nightly live run is a recurring
   bill and a recurring source of alerts. Is the honesty worth it, or is a
   quarterly manual verification enough?
2. **A hand-authored scenario is a claim about the product made by its authors.**
   It is more honest than a rigged fake, and it is still us choosing the ground.
   Review should ask whether that is worth stating in the README beside the
   numbers.
