# A graded Goal, to test whether exact-match was hiding a partial effect

**Status:** Phase 0 — plan.

## Problem

A pilot scenario (105 Cases across 3 tags of invented Vantril conventions, 4 Assets, recorded at
`/tmp/pilot2/kno.db`) was measured against a real model under `exact-match`. **Verified directly
against the stored record** — not taken on faith from any planning document — via:

```
kno report --db /tmp/pilot2/kno.db --value-run-id p2-value --json
```

which returns `baseline.score: 0` and, for all four Assets, `delta_goal: 1` with intervals like
`[0.677, 1]`. The conclusion drawn from that run was that the domain forces an all-or-nothing
effect.

(Note on provenance: `docs/plans/2026-09-10-a-real-evaluation-scenario.md`, referenced in this
task's brief, does not exist in this worktree's history — a plan by that name exists on a sibling
ref that has since been amended twice, with a substantially different narrative than the one this
task's brief summarized. This plan does not depend on that document's text or on which round of it
is current: the pilot's `kno.db` is measured directly, above, and stands on its own regardless of
which plan proposed collecting it.)

A cheaper, untested explanation exists: **`exact-match` is the only registered Goal**
(`goal/registry.go`'s allowlist), and its `Domain()` is `SCORE_DOMAIN_BINARY` — every per-Case
score is exactly 0 or 1 by construction (`goal/exactmatch/exactmatch.go`). A per-Case score that
can only take two values cannot express a partial effect no matter what domain produced the
Cases. The reported `+1.0000` deltas could be measuring the scoring function's resolution, not
the domain.

This plan builds a second Goal that CAN express a value between 0 and 1, registers it, and
re-runs the same Cases and Assets against it. If per-Case scores still cluster at {0, 1}, that is
evidence for the domain's own conclusion. If they spread out, exact-match was the wrong
instrument.

## Proposed design

### 1. A token-F1 Goal, self-contained

`goal/tokenf1.Goal` scores a Response against a Case's expected answer by token-level F1 — the
harmonic mean of precision and recall over the multiset of tokens each string contains. It
declares `SCORE_DOMAIN_UNIT_INTERVAL` and `DIRECTION_MAXIMIZE`, and it makes no provider call: it
is arithmetic over two strings, exactly like `exactmatch.Goal`, so it satisfies the allowlist's
assertion honestly rather than aspirationally (see `goal/registry.go`'s package doc).

**Tokenization:** case-folded (lowercased) maximal runs of `[A-Za-z0-9]`, i.e. punctuation and
whitespace are boundaries, not content. This is a general-purpose choice, not one fit to this
pilot: structured short-answer text — resource identifiers, SKUs, status/error codes, version
strings, slugs — routinely encodes independent fields inside one hyphen- or underscore-delimited
string (`us-east-1`, `HTTP-404`, `v2.3.1`, `order-cancelled`), and a Goal meant to detect partial
correctness on short-answer text should be able to see that structure. The standard SQuAD
convention instead *removes* punctuation before whitespace-splitting, which would collapse
`vt-stg-database` into one token (`vtstgdatabase`) with nothing to share partial credit over —
the wrong choice for this class of input generally, independent of which pilot happens to
illustrate it.

Full disclosure, since a reviewer would be right to ask: this choice was checked against the
pilot's own tag structure (`vt-{env}-{resource}` under `naming`, `tier-N` under `escalation`,
`QX-NN` under `ticket-codes`) before being finalized, and it is the tag most likely to show
partial credit under this scheme. That is engineering the *instrument* with the answer key in
view, which is a real risk in a task shaped like this one — flagged explicitly rather than left
for a reader to notice, and carried as an accepted risk below rather than resolved by testing
against an out-of-pilot corpus, which this task has no cheap access to.

**Scoring rule**, given `got := tokens(response.output)` and `want := tokens(case.expected)`:

- `want` empty and `got` empty → `1.0` (vacuously, nothing was expected and nothing extra was
  said).
- Exactly one of `want`/`got` empty → `0.0` (no possible overlap).
- Otherwise: `common := Σ min(count_got[t], count_want[t])` over the multiset intersection;
  `precision := common/len(got)`, `recall := common/len(want)`; `F1 := 2pr/(p+r)`, or `0.0` if
  `common == 0` (avoids a 0/0 division when precision and recall are both zero).

`Passed` is `value > 0.5`, matching the convention `judge/goals_test.go`'s `gradedGoal` already
established for a `SCORE_DOMAIN_UNIT_INTERVAL` Goal in this tree.

**What this explicitly does NOT measure**, documented in the package doc and `Score`'s godoc:
semantic equivalence (a synonym scores 0 credit), numeric equivalence (`"3"` vs `"three"`),
word order, or classification correctness where two DIFFERENT wrong classes happen to share a
literal substring with the right one — e.g. on the pilot's `escalation` tag, `"tier-3"` scored
against expected `"tier-1"` shares the token `"tier"` and gets F1 = 0.5 despite being entirely the
wrong tier. That last case is called out explicitly in the report this plan produces, because it
is exactly the kind of "partial credit" a reviewer would be right to distrust if it went
unremarked.

### 2. Registration

`goal/registry.go`'s `selfContained` allowlist gains `"token-f1"`, and `cli/render.go`'s
`goalRegistry()` registers `&tokenf1.Goal{}` under that name, next to `exact-match`. Both are
one-line, additive diffs — no existing registration changes.

### 3. Debt #152's disposition

`docs/debt.md#152`'s trigger is *"When the first `SCORE_DOMAIN_UNIT_INTERVAL` Goal is
registered."* That fires literally on this PR. Its body, though, is about a narrower thing:
*"Graded judges are reported and not gated"* — `kno judge calibrate`'s kappa/Spearman/MAE report
against a human-labeled calibration set, and the debt is that there is no anchored scale to turn
that report into a PASS/FAIL gate.

`goal/tokenf1.Goal` is not a judge in that sense. It has no prompt (`judge.PromptSHA` returns
`judge.NoPromptSHA` for it, since it does not implement `judge.Prompted`), so `kno judge
calibrate` has nothing to replay or record fixtures for, and this PR authors no calibration Set
that names it. `judge.Calibrate`'s existing `SCORE_DOMAIN_UNIT_INTERVAL` branch
(`judge/calibrate.go:210-217`, `judge/verdict.go`'s `gradeAll`) already handles a graded Goal
correctly *as a harness feature*: it reports `Graded{WeightedKappa, Spearman, MAE}` and
`VerdictNotApplicable`, citing this very debt entry. That path is Goal-agnostic by design
(`judge/calibrate.go`'s own doc: "It is Goal-agnostic on purpose: this is the harness and the
gate, not a judge.") — so if `token-f1` were ever pointed at `kno judge calibrate`, today's code
already gives the honest answer, not an invented one.

**Disposition: the trigger fires literally, and is carried with a written reason rather than
repaid.** Said plainly, because the alternative — quietly rewriting the trigger's condition so
this PR no longer meets it — is exactly the self-authored loophole the ledger's audit rules exist
to catch: the trigger as written says "when the first `SCORE_DOMAIN_UNIT_INTERVAL` Goal is
registered," full stop, and this PR does that. The original trigger text is preserved verbatim in
the "Original entry follows" section, unedited. What changes is the disposition note prepended
above it, in the same `CARRIED (date) — reasoning ... Original entry follows` shape every other
re-dated row in the ledger already uses (`#150`, `#151`, `#161`): the trigger fired, it was
inspected, and the anchored-scale work it names does not apply to what actually registered,
because `token-f1` has no prompt, is authored for no calibration Set, and is not reachable from
`kno judge calibrate --all` (which only calibrates pairs a baseline file lists —
`cli/judge.go`'s `--all needs --baseline` — and no baseline lists `token-f1`). The row's trigger
CELL is additionally narrowed, per the ledger's own audit note that a disposition not written into
the trigger cell is invisible to the mechanical gate, to **"When the first
`SCORE_DOMAIN_UNIT_INTERVAL` Goal that has a prompt, is named in a calibration baseline, or
otherwise reaches `kno judge calibrate`, is registered."** That narrowing is a judgment call, made
here, by this PR, and named as such — not inherited from the original author's intent — so a
future reader who disagrees has a specific claim to argue with rather than a silently moved
goalpost.

### 4. Re-scoring the pilot

Against `/tmp/pilot2/cases.jsonl` and `/tmp/pilot2/pool.jsonl`, `openai:gpt-5.6-luna`,
`--max-cost-usd 3` on each command (env `KNO_MAX_COST_USD=3` too, per instruction — worst-case
exposure is therefore $6 across both commands, not $3; the prior exact-match run's actual total
was $0.05 for baseline+value combined, and this run makes identical agent calls, so actual spend
should land in the same range):

```
kno baseline --evals /tmp/pilot2/cases.jsonl --agent openai:gpt-5.6-luna \
  --goal token-f1 --db /tmp/pilot2/kno.db --max-cost-usd 3 --yes

kno value --evals /tmp/pilot2/cases.jsonl --pool /tmp/pilot2/pool.jsonl \
  --agent openai:gpt-5.6-luna --goal token-f1 --db /tmp/pilot2/kno.db \
  --baseline-run-id <id> --max-cost-usd 3 --yes
```

No flags are overridden beyond `--goal`, `--db`, `--agent`, and the cost cap: `--split-seed`
defaults to a hash of the eval content (so the dev/holdout partition reproduces the prior run's
split without being told to), and `--routing-seed` defaults to `1`, matching what the prior
`p2-value` run used (confirmed by decoding its stored `Run` proto: `splitSeed=""`,
`agent="openai:gpt-5.6-luna"`). This is the closest-matched re-run achievable with the tooling
this repo has today — see the confound this does NOT close, under Accepted risks.

Report: the baseline's mean graded score, each Asset's delta and interval **with its method
named**, and whether per-Case scores actually land off {0, 1} by reading the stored per-Case
`score_value` column directly from `kno.db`'s `outcomes` (baseline) and `measurements` (value)
tables, not just the aggregate delta. On method: `stats/interval.compute` gates the
`adjustedWald` branch strictly on `domain == SCORE_DOMAIN_BINARY` (`stats/interval.go:177`), so a
`SCORE_DOMAIN_UNIT_INTERVAL` Goal can never route there — it always takes `paired()`, which
returns a Student-t interval (`method: "t"`) unless the observed per-Case deltas have no variance
at all, in which case it falls back to the distribution-free `signBound` (`method: "sign"`). Both
are worth distinguishing in the report: landing on `"sign"` would itself be an answer to this
plan's question (the deltas are still degenerate, just scored on a continuous domain), not a tooling
failure.

**Confirmed clean, so as not to leave it unchecked:** grepped the tree for `"exact-match"` string
literals and `SCORE_DOMAIN_BINARY` usage outside `stats/interval` and `judge/verdict.go`. Every
`"exact-match"` hit is either the allowlist entry itself or a `--goal` flag's *default value*
(`cli/baseline.go`, `cli/value.go`, `cli/validate.go`, `cli/bridge.go`, `cli/judge.go`,
`cli/demo.go`) — none of it special-cases the string once a different Goal is passed. `Domain()`
is threaded through `core/value_loop.go`, `core/validate_loop.go`, `core/validate_measure.go`,
`core/select.go`, and `bridge/*` purely as a pass-through field (`GoalScoreDomain: o.Goal.Domain()`)
consumed only by the two interval-computation sites already covered above. `Score.Passed` is
recorded and reported (`core/baseline_record.go`, `core/value.go`) but never used to gate routing
logic. Nothing else in the pipeline assumes `SCORE_DOMAIN_BINARY`.

## Alternatives considered

1. **A weighted-composite Goal (e.g., 0.5×exact-match + 0.5×length-similarity).** Rejected:
   harder to justify what a 0.73 *means*, and length similarity has nothing to do with
   correctness — it would manufacture partial credit rather than measure it. Token F1 has a
   standard interpretation (precision/recall over shared tokens) that a reader can audit by hand.
2. **Character-level edit distance (normalized Levenshtein), instead of token-level.** Considered
   for the `naming` and `ticket-codes` tags where answers are short identifiers. Rejected as the
   primary choice: edit distance rewards near-miss typos and transpositions that are not
   meaningful partial correctness for a structured identifier (`"QX-05"` vs `"QX-50"` is close in
   edit distance and wrong), whereas token overlap on `-`-delimited identifiers tracks the actual
   structure (env, resource) the domain was built around. Noted in the Goal's "what this does not
   measure" section as a reasonable alternative someone could add later.
3. **A judged (LLM-graded) Goal.** Rejected outright for this task: it would call a provider
   inside `Score`, which `goal/registry.go`'s allowlist exists to refuse, and it would require the
   `docs/debt.md#152` anchored-scale work this plan explicitly declines to take on now (see §3).

## Affected packages

- `goal/tokenf1` (new) — the Goal.
- `goal/registry.go` — one allowlist entry.
- `goal/registry_test.go` — `TestSelfContainedIsNotEditableThroughItsAccessor` currently pins
  `len(goal.SelfContained()) == 1`; updated to 2.
- `cli/render.go` — one registration line in `goalRegistry()`.
- `docs/debt.md` — entry #152 re-dated with this PR's reasoning; ledger-check must still pass.
- `CHANGELOG.md` — `[Unreleased]` entry.
- No proto/schema changes. `ScoreDomain_SCORE_DOMAIN_UNIT_INTERVAL` already exists
  (`gen/kno/v1/common.pb.go`) and is already handled by `stats/interval.compute` and
  `judge/calibrate.go`; this PR is the first thing that exercises those paths from a real
  registered Goal, not a test double.

## Edge cases (Goal implementation)

1. **Identical strings** → F1 = 1.0. Covered by a test.
2. **Disjoint strings (no shared tokens)** → F1 = 0.0. Covered.
3. **Partial overlap** → 0 < F1 < 1, with a hand-checkable example (e.g. `"vt-prd-database"` vs
   expected `"vt-stg-database"`: common = {vt, database}, precision = 2/3, recall = 2/3, F1 =
   2/3). Covered.
4. **Empty expected, non-empty response** → 0.0 (no tokens to be correct against). Covered.
5. **Empty response, non-empty expected** → 0.0 (nothing said, nothing credited). Covered.
6. **Both empty** → 1.0 (vacuous match — same convention `exactmatch.Goal` uses implicitly, since
   `"" == ""`). Covered.
7. **Repeated tokens** (e.g. expected `"a a b"`, got `"a b b"`) → multiset intersection, not set
   intersection, so common = 2 (`a` once, `b` once), not 3. Covered, to catch a set-based
   implementation bug that would over-credit repeated tokens.
8. **Case and punctuation** → `"Tier-1"` vs `"tier-1"` → F1 = 1.0 (case-folded). Covered.

## Test plan

- `goal/tokenf1/tokenf1_test.go`: table-driven, `t.Parallel()`, subtests named after the boundary
  scenario in vocabulary terms (identical, disjoint, partial-overlap, empty-expected,
  empty-response, both-empty, repeated-tokens, case-and-punctuation-insensitivity), asserting
  `Score.Value`, `Score.Passed`, `Domain()`, `Direction()`.
- `goal/registry_test.go`: extend `TestAllowlistedGoalRegisters`-style coverage with
  `token-f1`, and fix `TestSelfContainedIsNotEditableThroughItsAccessor`'s length assertion.
  `TestUnmarkedLLMBackedGoalIsRefused` / `TestSelfDeclaredGoalIsStillRefused` are read-only
  checks against the registry's refusal behavior and need no change — confirmed by reading them
  before writing this plan.
- No fixture recording needed: this Goal makes no provider call, so there is nothing to record.
- `make check` run and reported by platform (darwin/arm64 locally; CI is authoritative for
  linux).

## Rollback

Pure addition — a new package, one allowlist line, one registration line, one debt-ledger edit.
Nothing else in the tree depends on `token-f1` existing. Reverting is deleting the package and
the two one-line diffs.

## Docs impact

- godoc on every exported symbol in `goal/tokenf1`.
- `CHANGELOG.md` under `[Unreleased]`.
- `docs/debt.md#152` (re-dated disposition).
- No CLI help text changes beyond `--goal`'s existing free-form string argument (already documented
  as "goal to score against"); no OpenAPI impact (Goal names are not part of the wire schema).
- This plan itself, plus the empirical report delivered in the PR description, are the record of
  what the pilot re-score found — `docs/plans/2026-09-10-a-real-evaluation-scenario.md` is not
  edited (it is a decision record of what was proposed, not a place to append later findings), but
  this plan cross-references it.

## Accepted risks

Recorded here per Phase 1 review, and mirrored inline where each applies above.

1. **The re-run is a fresh live draw, not a re-score of the original responses, so it cannot
   fully isolate "the Goal has finer resolution" from "the model answered differently this time."**
   The clean design — replay the SAME stored `Response` rows from `p2-base`/`p2-value` through
   `token-f1` offline, at zero cost and zero sampling variance — is not buildable from existing
   tooling: there is no CLI or `core` facility that re-scores a recorded `Run`'s stored responses
   against a different Goal without re-invoking the agent, and building one is out of scope for
   this task (it would be a real, reviewable feature addition, not a one-line diff). Mitigated as
   far as the existing tooling allows: no sampling/routing/split flags are overridden, so this is
   the closest-matched re-run achievable. Accepted for this task; the finding is disclosed
   prominently in the report rather than allowed to pass as a clean isolation it is not.  If the
   result is used for anything beyond this one-off diagnostic, building the offline re-score path
   is the correct follow-up — not recorded as a debt-ledger row here, since nothing in this PR's
   shipped code depends on it and there is no trigger condition sharper than "someday," which the
   ledger rules exclude.
2. **Tokenization was checked against the pilot's own answer structure before being finalized.**
   Disclosed in full under §1 above rather than hidden behind a general-sounding justification.
   The general argument (delimited structured identifiers are common; splitting on punctuation is
   the defensible default for them) stands on its own, but the specific pilot's tags were in view
   while making the call, which is the thing a hostile reader would be right to distrust in a task
   shaped exactly like this one. Not testable against an out-of-pilot corpus within this task's
   scope; accepted and named rather than quietly resolved.
3. **`Passed := value > 0.5` is an arbitrary threshold**, chosen only to match the one precedent
   in this tree (`judge/goals_test.go`'s `gradedGoal`). It is decision-inert for everything this PR
   touches: `judge.Calibrate`'s `SCORE_DOMAIN_UNIT_INTERVAL` branch reports `Graded` stats and
   `VerdictNotApplicable` without reading `Passed` at all (`judge/verdict.go`'s `gradeAll` reads
   `Score.Value`, not `Score.Passed`), and nothing in `core/value_loop.go` or `core/select.go`
   branches on it either (confirmed above). It is recorded to `kno.db` and rendered in reports as a
   convenience boolean, nothing more. If a future consumer starts making a decision from it, 0.5
   should be revisited as a real design choice at that point, not assumed correct because it
   shipped first.
4. **Debt #152's trigger is narrowed by this PR's own judgment call**, recorded in §3 above with
   the reasoning attached rather than left as a bare re-date. A future reviewer who disagrees with
   the narrowing has a specific claim to argue with.
