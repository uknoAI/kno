// Package tokenf1 scores a Response by token-level F1 against the Case's
// expected answer.
//
// It exists because exact-match's Domain is SCORE_DOMAIN_BINARY: every
// per-Case score is exactly 0 or 1, so a paired difference of two such scores
// can only ever land in {-1, 0, +1}. That is the right instrument for a Case
// whose answer either is or is not correct, but it cannot express a PARTIAL
// effect no matter what domain produced the Cases — a model that gets the
// right resource but the wrong environment in "vt-stg-database" scores
// identically to one that answers nothing at all. Goal.tokenf1 gives that
// case a number between 0 and 1 instead of collapsing it to the same 0
// exact-match would report.
//
// See docs/plans/2026-09-10-graded-goal-token-f1.md for the design record,
// including what this Goal explicitly does NOT measure.
package tokenf1

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/knograph/kno/core"
	knov1 "github.com/knograph/kno/gen/kno/v1"
)

// tokenPattern extracts maximal runs of ASCII letters and digits.
//
// Punctuation and whitespace are token BOUNDARIES, not content — deliberately
// the opposite of the SQuAD convention of removing punctuation before
// whitespace-splitting. Removing punctuation from a structured identifier
// like "vt-stg-database" collapses it to one token ("vtstgdatabase") with
// nothing left to share partial credit over. Splitting ON punctuation instead
// keeps the fields a delimited identifier encodes (environment, resource,
// version, status code, ...) visible to the overlap computation, which is
// exactly the class of short-answer text this Goal is for.
var tokenPattern = regexp.MustCompile(`[A-Za-z0-9]+`)

// Goal scores a Response by the harmonic mean of precision and recall over
// the multiset of tokens the Response's output and the Case's expected
// answer share.
//
// It is case-folded and self-contained: Score is arithmetic over two
// strings and makes no provider call, which is what lets it sit on
// goal/registry.go's self-contained allowlist honestly rather than
// aspirationally — see that package's doc for why the allowlist exists.
//
// What this Goal explicitly does NOT measure: semantic equivalence (a
// correct synonym scores no credit), numeric equivalence ("3" vs "three"),
// word order, or classification correctness in general — two DIFFERENT wrong
// answers that happen to share a literal token score partial credit against
// each other's shared word even though the classification itself is entirely
// wrong (e.g. "tier-3" scored against an expected "tier-1" shares the token
// "tier" and gets F1 = 0.5 despite naming the wrong tier). Token overlap is a
// STRING-STRUCTURE signal, not a correctness judgment, and should be read as
// one.
type Goal struct{}

// Name identifies this Goal in reports and on the Run.
func (g *Goal) Name() string { return "token-f1" }

// Score computes token-level F1 between the Response's output and the
// Case's expected answer.
//
// Both strings are tokenized into maximal runs of [A-Za-z0-9], case-folded.
// Given got := tokens(response.output) and want := tokens(case.expected):
//
//   - want empty and got empty: F1 = 1.0. Vacuously — nothing was expected
//     and nothing extra was said, so there is nothing to be wrong about.
//   - Exactly one of want/got empty: F1 = 0.0. No possible overlap.
//   - Otherwise: common is the multiset intersection (Σ min(count in got,
//     count in want) over every distinct token), precision = common/len(got),
//     recall = common/len(want), and F1 = 2·precision·recall/(precision+recall),
//     or 0.0 if common is 0 (avoiding a 0/0 division when precision and
//     recall are both zero).
//
// Passed reports value > 0.5, matching the convention this tree's one other
// SCORE_DOMAIN_UNIT_INTERVAL Goal fixture uses (judge/goals_test.go's
// gradedGoal). It is not read by anything that gates a decision in this
// codebase today — see the design plan's accepted risks — so 0.5 is a
// reasonable default rather than a load-bearing threshold.
func (g *Goal) Score(_ context.Context, c *core.Case, r *core.Response) (*core.Score, error) {
	got := tokenize(r.GetOutput())
	want := tokenize(c.GetExpected())

	value, precision, recall, common := f1(got, want)

	return &knov1.Score{
		CaseId:    c.GetId(),
		Value:     value,
		Passed:    value > 0.5,
		Rationale: rationale(value, precision, recall, common, len(got), len(want)),
		// No judge model: this Goal does not use one, and naming a model it
		// did not consult would misrepresent how the number was produced.
	}, nil
}

// Direction reports that higher is better.
//
// Without this the sign of every delta measured against this Goal would be
// uninterpretable.
func (g *Goal) Direction() core.Direction { return knov1.Direction_DIRECTION_MAXIMIZE }

// Domain reports that this Goal's Scores span the unit interval.
//
// Unlike exact-match's SCORE_DOMAIN_BINARY, a token-F1 score can land
// anywhere in [0, 1], so a paired difference of two such scores is
// continuous rather than confined to {-1, 0, +1}. stats/interval.compute
// routes a SCORE_DOMAIN_UNIT_INTERVAL Goal to the Student-t interval (or its
// distribution-free sign-test fallback when the observed deltas carry no
// variance at all) rather than the McNemar-style adjusted-Wald interval
// exact-match uses — see stats/interval/interval.go's compute and paired.
func (g *Goal) Domain() core.ScoreDomain {
	return knov1.ScoreDomain_SCORE_DOMAIN_UNIT_INTERVAL
}

var _ core.Goal = (*Goal)(nil)

// tokenize lowercases s and splits it into maximal runs of [A-Za-z0-9].
func tokenize(s string) []string {
	return tokenPattern.FindAllString(strings.ToLower(s), -1)
}

// f1 computes token-level F1 over two token slices, treating each as a
// multiset (repeated tokens count more than once, on both sides
// independently) rather than a set — a set-based intersection would
// over-credit an answer that repeats a correct token more times than the
// expected answer does.
//
// Returns the F1 value along with precision, recall and the common-token
// count, so callers can build a rationale without recomputing them.
func f1(got, want []string) (value, precision, recall float64, common int) {
	if len(want) == 0 && len(got) == 0 {
		return 1, 1, 1, 0
	}
	if len(want) == 0 || len(got) == 0 {
		return 0, 0, 0, 0
	}

	gotCounts := counts(got)
	wantCounts := counts(want)
	for token, wantN := range wantCounts {
		if gotN := gotCounts[token]; gotN < wantN {
			common += gotN
		} else {
			common += wantN
		}
	}
	if common == 0 {
		return 0, 0, 0, 0
	}

	precision = float64(common) / float64(len(got))
	recall = float64(common) / float64(len(want))
	value = 2 * precision * recall / (precision + recall)
	return value, precision, recall, common
}

// counts builds a token -> occurrence-count multiset.
func counts(tokens []string) map[string]int {
	m := make(map[string]int, len(tokens))
	for _, t := range tokens {
		m[t]++
	}
	return m
}

// rationale renders a human-auditable explanation of one score, so a token-F1
// number can be checked by hand the same way exact-match's pass/fail can.
func rationale(value, precision, recall float64, common, gotN, wantN int) string {
	if gotN == 0 && wantN == 0 {
		return "both the response and the expected answer were empty: vacuous match"
	}
	if gotN == 0 {
		return "the response was empty against a non-empty expected answer"
	}
	if wantN == 0 {
		return "the expected answer was empty against a non-empty response"
	}
	return fmt.Sprintf(
		"token F1 %.4f: %d of %d expected tokens matched (recall %.4f), "+
			"%d of %d response tokens were expected (precision %.4f)",
		value, common, wantN, recall, common, gotN, precision)
}
