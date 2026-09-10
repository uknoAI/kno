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

**Superseded in part by §2c.** Three rounds of measurement moved the binding
constraint off the domain and onto the **shape of the expected string**. This
section still describes a necessary condition; it is no longer the first
decision. Read §2c before authoring a Case.

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
2. Check the **answer format** offline, before anything else — §2c's three
   rules, which are properties of the expected strings alone and need no model,
   no Assets and no calls. This step was missing from the first three rounds and
   is what they spent money rediscovering.
3. Run **`kno eval inspect`** over them — no LLM call, no spend, and the tool
   built for exactly this question. It reports `separable_effect` per tag.
   Note that **none of its five checks tests `separable_effect` against any
   threshold** (§2b-bis); passing them says nothing about clearing Select.
4. Run a **cheap pilot** against a real inexpensive model to get an empirical
   baseline failure rate and an observed effect size — **scored against a graded
   Goal.** `exact-match` reported this pilot's baseline as 0.000 when it was
   0.358, and that single artifact cost two review rounds.
5. Compare the observed effect against the **corrected** bound the full N would
   impose — after `DefaultControlReserve` and Bonferroni, not the raw number the
   tool prints. **Only then** decide whether to author the rest.

`kno-examples`' `power-analysis` scenario is the worked example of this exact
argument. Using it before spending is the plan; discovering it afterwards is
what the first draft would have done.

### 2b. What the feasibility check and the pilot actually measured

Both halves were run. The free half is encouraging; the paid half falsified the
design.

**Free half — `kno eval inspect`, no spend.** The required separable effect is
driven more by **Asset count** than by N:

| | required effect |
|---|---|
| N=105, 3 tags | **0.269** — 0 of 5 checks flagged |
| N=105, 5 tags | 0.352 |
| N=105, 7 tags | 0.427 |
| N=150, 3 tags | 0.220 |
| N=210, 3 tags | 0.184 |

A 30-Case draft measures **0.54**, matching the review's projection almost
exactly — but that projection assumed 5 Assets. At **3 tags and ~105 Cases the
bar is 0.27** and every check passes, including the holdout floor. Going 3→7
tags costs more than going N=210→105 saves, which was in neither the plan nor
the review.

**Paid half — a 30-Case pilot against `openai:gpt-5.6-luna`, ~100 calls,
cents.** Baseline scored **0.000**, 26 of 26 wrong: invented conventions are
genuinely unguessable, as designed. Then four Assets were measured — a **rule**
and a **lookup** for each rule-derivable tag, plus an arbitrary tag only a
lookup can serve:

```
codes-lookup       +1.0000  [+0.2917, +1.2327]
escalation-lookup  +1.0000  [+0.2917, +1.2083]
escalation-rule    +1.0000  [+0.2917, +1.2083]
naming-rule        +1.0000  [+0.1958, +1.2327]
```

**All four are identical. The rule/lookup distinction this plan asserted does
not exist empirically.** A competent model applies `ceil(hours/12)` perfectly,
so a rule Asset lands at 1.0 exactly like a lookup table. The distinction was
reasoned, not measured, and one measurement retired it.

### 2b-bis. Third round: the arithmetic corrected, the pilot powered

Review found the first pass wrong in two ways and both were checked.

**The bar was understated by roughly half.** `separable_effect` is
`MinDetectableEffect` over a tag's **full dev count**, at the **raw** level. It
ignores two things Select actually applies: `DefaultControlReserve = 0.3`, which
removes 30% of dev Cases from routing before clustering, and the Bonferroni
correction over **Assets screened**. Corrected:

| N | Assets | tool says | +reserve | +Bonferroni |
|---|---|---|---|---|
| 105 | 3 | 0.269 | 0.324 | **0.396** |
| 105 | 6 | 0.269 | 0.324 | **0.436** |
| 210 | 6 | 0.184 | 0.221 | 0.313 |

So "feasible at 0.269" was wrong; the real bar at the proposed size is ~0.40.
Note also that **none of `eval inspect`'s five checks test `separable_effect`
against any threshold** — passing them says nothing about clearing Select.

**The falsification was underpowered, and now is not.** The first pilot had 5–6
pairs per Asset, where a genuinely 85%-reliable rule lands all-correct about 38%
of the time — the review was right that a null result there proves nothing. Rerun
at **~28 pairs** (105 Cases, 3 tags, 226 measurements, $0.03):

```
codes-lookup       +1.0000  [+0.6773, +1.0000]
escalation-lookup  +1.0000  [+0.6587, +1.0000]
escalation-rule    +1.0000  [+0.6587, +1.0000]
naming-rule        +1.0000  [+0.6939, +1.0000]
```

Identical again, with intervals a third the width. At 28 pairs an 85% rule lands
all-correct roughly 1% of the time, so **the rule/lookup falsification holds at
power.** A competent model applies the rule perfectly, which makes a rule Asset
functionally a lookup table.

*(Those upper bounds read exactly 1.0000 because the out-of-range interval this
plan's first pilot surfaced was fixed separately — the earlier `[+0.29, +1.21]`
was a real defect in `adjustedWald`, not a curiosity.)*

### 2b-ter. The untested alternative was tested, and it was right

The previous round recorded one explanation it could not afford to check: that
**`exact-match` on short answers forces binary outcomes, so the finding was
about the Goal rather than the domain.** It stayed live because `goal.Registry`
is default-deny against a compile-time allowlist, so adding a graded Goal is a
code change, not configuration.

That gate has now been passed deliberately. `goal/tokenf1` ([PR #222](https://github.com/uknoAI/kno/pull/222))
registers a token-level F1 Goal declaring `SCORE_DOMAIN_UNIT_INTERVAL` and
making no provider call. The pilot's **same Cases, same Assets, same
`openai:gpt-5.6-luna` agent** were re-scored against it:

| | `exact-match` | `token-f1` |
|---|---|---|
| baseline | **0.000** | **0.358** |
| per-Case baseline scores strictly inside (0,1) | 0 of 85 | **42 of 85** |
| `codes-lookup` | +1.0000 | +1.0000 |
| `escalation-lookup` | +1.0000 | **+0.5000** |
| `escalation-rule` | +1.0000 | **+0.5000** |
| `naming-rule` | +1.0000 | **+0.7077** |

**§2c's conclusion was wrong.** It read a uniform +1.0000 as evidence that *the
domain* forces all-or-nothing outcomes and told the plan not to authorize
authoring until a "partially competent" domain was found. The domain was never
the constraint. A metric with two values reported two values.

### 2b-quater. What the graded numbers actually say — including against this plan

The re-scoring settles §2b-ter and immediately raises a sharper problem, and
the second half matters more than the first.

**Read the +0.5000s before trusting them.** Under token-F1, `"tier-3"` scored
against expected `"tier-1"` shares the token `"tier"` and receives F1 = 0.5 —
while being entirely the wrong tier. On the `escalation` tag, whose answers are
`tier-N`, a baseline that names *any* tier collects half credit for free. Both
+0.5000 deltas are that: **a shared-stem artifact of the answer format, not a
measurement of partial correctness.** The graded Goal's own package doc calls
this failure mode out by name; it is not a defect in the Goal, and it is not
partial credit either.

So the honest reading of the table is narrower than it looks:

- The baseline of 0.358 is real and is the important number. `exact-match`'s
  0.000 was an artifact.
- `naming-rule`'s +0.7077 is the only delta that plausibly measures graded
  coverage, because `vt-{env}-{resource}` has three independently-wrong fields.
- The two +0.5000s measure the answer format's token overlap.
- `codes-lookup`'s +1.0000 is unchanged, because `QX-NN` codes share nothing
  when wrong.

### 2c. The design constraint is the ANSWER FORMAT, not the domain

Every finding in this plan's three previous rounds points at one variable, and
it is not the one any round named.

- Round one blamed **Asset shape** (rule vs lookup). Falsified at power: a
  competent model applies `ceil(hours/12)` perfectly, so a rule Asset behaves as
  a lookup table.
- Round two blamed **the domain** ("invented conventions force implication").
  Falsified by the graded re-scoring: baseline 0.358, not 0.000.
- What survives both is the **shape of the expected string**. `tier-N` admits
  exactly two honest outcomes and one dishonest half. `QX-NN` admits two.
  `vt-{env}-{resource}` admits four, and it is the only Case shape in the pilot
  that produced a delta neither degenerate nor spurious.

**The answer format determines what any Goal can see, and it is upstream of both
domain and Asset design.** A scenario is authored by writing expected strings;
choosing them last, as an encoding detail, is what produced three rounds of
attributing the format's behavior to something else.

This reframes the authoring task. The requirement is not "find a domain where
the model is partially competent" — that was §2c's answer to a question that
turned out to be about the metric. It is:

**Every Case's expected answer must decompose into independently-verifiable
fields, enough of them that partial coverage is distinguishable from both total
coverage and none, and with no field sharing a literal token with a wrong
value of another field.**

Concretely, for this pilot's domain that means retiring `tier-N` (two values,
shared stem) in favor of a form like `escalate-billing-24h` — independent
fields, no shared stem between wrong and right answers, and a graded score that
moves for a genuine reason. `vt-{env}-{resource}` already satisfies it and is
the template.

**This is a constraint on authoring, and it is checkable before spending
anything.** It is a property of the expected strings alone — no model, no
Assets, no calls. Which means it belongs in §2a's free gate, and it is the
piece §2a was missing:

- No two distinct expected values within a tag share a token (rejects `tier-N`).
- Each tag's expected values carry ≥ 3 independent fields (so partial coverage
  has room to register).
- The `token-f1` score of every wrong-but-plausible answer against its expected
  answer is 0 (the artifact check, run offline over the author's own distractor
  list).

**What is still unmeasured**, and this plan still does not authorize authoring
until it is: whether a format satisfying those three rules produces a delta that
clears the corrected bar of ~0.40 (§2b-bis). `naming-rule`'s +0.7077 is a single
observation on a single tag and is the only evidence in either direction. The
next step is the cheapest possible test of exactly that — **re-author the
`escalation` tag's ~28 Cases into a multi-field format and re-run the same
pilot**, ~$0.01, no new Cases elsewhere, one comparison against a number already
in hand.

That is a materially cheaper next step than §2c's "another pilot against a
candidate middle domain," and it tests the thing that actually varies.

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

## An interval above the metric's maximum — found here, fixed twice

Noted while running the first pilot: every delta CI extended **above 1.0** on a
bounded score (`[+0.2917, +1.2083]`). It was a defect, not a boundary curiosity,
and it took two fixes because the first was too narrow.

[#220](https://github.com/uknoAI/kno/pull/220) clamped `adjustedWald`, which is
where the out-of-range bound was observed. `compute` reaches that method only
for a binary score at a single trial, so three other branches kept reporting
impossible bounds — and §2b-ter's graded re-scoring walked straight into one, a
`SCORE_DOMAIN_UNIT_INTERVAL` delta reading `[0.813, 1.187]` on a score whose
maximum is 1.0. [#223](https://github.com/uknoAI/kno/pull/223) moved the clamp to
the dispatch in `compute`, where the domain is known, gated on the domain
actually being bounded.

Worth recording as a fact about this plan rather than about the statistics: the
scenario work has now surfaced two real defects in shipped code (this, and the
narrow first fix) before authoring a single production Case. A pilot that costs
cents and exercises paths fixtures do not is buying something beyond its stated
question.

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

### Fourth round

Still **not passed**, and the reason moved again. Each round blamed a different
variable, and each was falsified by a measurement rather than by more review:

| round | blamed | falsified by |
|---|---|---|
| 1 | sizing / power | §2b-bis — the bar is ~0.40, not 0.269, and the pilot then ran at power |
| 2 | Asset shape (rule vs lookup) | §2b — identical deltas at 28 pairs |
| 3 | the domain | §2b-ter — graded baseline 0.358, not 0.000 |
| 4 | **the answer format** | *unmeasured — §2c names the test* |

The pattern is the plan's most transferable finding: **three rounds attributed
to domain and design what belonged to the encoding of the expected string**, and
in each case cents of measurement settled what rounds of reasoning had not.
Round four's claim is not exempt. §2c ends with the specific ~$0.01 test that
would falsify it, and this plan does not authorize authoring until that test
runs.

## Accepted risks

**Accepted, pending the §2c re-author test:**

1. **The answer-format claim is one observation wide.** `naming-rule`'s +0.7077
   is the only delta in the pilot that is neither degenerate nor a shared-stem
   artifact, and one tag is not evidence that a multi-field format generally
   clears the corrected bar. Accepted only because the test that would settle it
   costs ~$0.01 and is this plan's next step, not because the claim is strong.

2. **`token-f1`'s tokenizer was chosen with this pilot's tag structure in
   view.** [PR #222](https://github.com/uknoAI/kno/pull/222) discloses this
   directly: splitting on `[A-Za-z0-9]` runs rather than the SQuAD convention
   preserves `vt-stg-database`'s internal structure, and `vt-{env}-{resource}`
   is the pilot's tag most likely to show partial credit under that scheme.
   §2c then builds a design constraint partly on that tag's result. The
   circularity is real, and it is why §2c's three offline checks are stated as
   properties of the expected strings rather than as "scores well under
   `token-f1`" — but a reviewer should weigh whether that is enough separation.
   The repository has no cheap out-of-pilot corpus to break it properly.

3. **The graded Goal changes what a delta in this scenario means, and the docs
   have not caught up.** *What the numbers mean* describes deltas over a binary
   score. If this scenario ships against `token-f1`, that page changes in the
   same PR — a graded delta of +0.7 is not "70% of Cases flipped."

Two carried unchanged from the previous round:

4. **This costs money on a schedule, forever.** A nightly live run is a recurring
   bill and a recurring source of alerts. Is the honesty worth it, or is a
   quarterly manual verification enough?

5. **A hand-authored scenario is a claim about the product made by its authors.**
   It is more honest than a rigged fake, and it is still us choosing the ground.
   Review should ask whether that is worth stating in the README beside the
   numbers.
