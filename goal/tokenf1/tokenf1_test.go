package tokenf1_test

import (
	"context"
	"math"
	"testing"

	"github.com/knograph/kno/core"
	knov1 "github.com/knograph/kno/gen/kno/v1"
	"github.com/knograph/kno/goal/tokenf1"
)

// scoreCase builds a minimal Case/Response pair and scores it, so every
// subtest below exercises only what it varies.
func scoreCase(t *testing.T, expected, output string) *core.Score {
	t.Helper()
	g := &tokenf1.Goal{}
	c := &knov1.Case{Id: "c1", Expected: expected}
	r := &knov1.Response{CaseId: "c1", Output: output}
	score, err := g.Score(context.Background(), c, r)
	if err != nil {
		t.Fatalf("Score: %v", err)
	}
	return score
}

const epsilon = 1e-9

func almostEqual(a, b float64) bool { return math.Abs(a-b) < epsilon }

func TestScoreBoundaries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		expected  string
		output    string
		wantValue float64
	}{
		{
			name:      "identical strings",
			expected:  "vt-stg-database",
			output:    "vt-stg-database",
			wantValue: 1.0,
		},
		{
			name:      "disjoint strings share no tokens",
			expected:  "tier-1",
			output:    "completely unrelated words",
			wantValue: 0.0,
		},
		{
			name:      "partial overlap credits shared tokens",
			expected:  "vt-stg-database",
			output:    "vt-prd-database",
			wantValue: 2.0 / 3.0, // common={vt,database}; precision=recall=2/3
		},
		{
			name:      "empty expected with non-empty response scores zero",
			expected:  "",
			output:    "tier-1",
			wantValue: 0.0,
		},
		{
			name:      "empty response with non-empty expected scores zero",
			expected:  "tier-1",
			output:    "",
			wantValue: 0.0,
		},
		{
			name:      "both empty is a vacuous match",
			expected:  "",
			output:    "",
			wantValue: 1.0,
		},
		{
			name:      "repeated tokens use multiset not set intersection",
			expected:  "a a b",
			output:    "a b b",
			wantValue: 2.0 / 3.0, // common=2 (one 'a', one 'b'), not 3
		},
		{
			name:      "case and punctuation are not content",
			expected:  "tier-1",
			output:    "Tier 1",
			wantValue: 1.0,
		},
		{
			name:      "escalation tag near miss shares a literal token, not the classification",
			expected:  "tier-1",
			output:    "tier-3",
			wantValue: 0.5, // common={tier}; precision=recall=1/2 — see the Goal's own doc
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			score := scoreCase(t, tt.expected, tt.output)
			if !almostEqual(score.GetValue(), tt.wantValue) {
				t.Errorf("Value = %v, want %v", score.GetValue(), tt.wantValue)
			}
			wantPassed := tt.wantValue > 0.5
			if score.GetPassed() != wantPassed {
				t.Errorf("Passed = %v, want %v (value %.4f)", score.GetPassed(), wantPassed, tt.wantValue)
			}
			if score.GetRationale() == "" {
				t.Error("Rationale is empty; a graded score without an explanation cannot be audited by hand")
			}
			if score.GetJudgeModel() != "" {
				t.Errorf("JudgeModel = %q, want empty: this Goal consults no judge", score.GetJudgeModel())
			}
			if score.GetCaseId() != "c1" {
				t.Errorf("CaseId = %q, want %q", score.GetCaseId(), "c1")
			}
		})
	}
}

// TestScoreValueBounds pins the value to [0, 1] over a spread of inputs, so a
// future edit to the tokenizer or the F1 arithmetic cannot silently produce a
// value outside the domain this Goal declares.
func TestScoreValueBounds(t *testing.T) {
	t.Parallel()

	pairs := []struct{ expected, output string }{
		{"vt-stg-database", "vt-stg-database"},
		{"vt-stg-database", "vt-prd-cache"},
		{"a b c d e", "a b"},
		{"a b", "a b c d e"},
		{"", ""},
		{"", "x"},
		{"x", ""},
	}
	for _, p := range pairs {
		score := scoreCase(t, p.expected, p.output)
		if v := score.GetValue(); v < 0 || v > 1 {
			t.Errorf("expected=%q output=%q: Value = %v, out of [0,1]", p.expected, p.output, v)
		}
	}
}

func TestDomain(t *testing.T) {
	t.Parallel()

	g := &tokenf1.Goal{}
	if got := g.Domain(); got != knov1.ScoreDomain_SCORE_DOMAIN_UNIT_INTERVAL {
		t.Errorf("Domain() = %v, want SCORE_DOMAIN_UNIT_INTERVAL", got)
	}
}

func TestDirection(t *testing.T) {
	t.Parallel()

	g := &tokenf1.Goal{}
	if got := g.Direction(); got != knov1.Direction_DIRECTION_MAXIMIZE {
		t.Errorf("Direction() = %v, want DIRECTION_MAXIMIZE", got)
	}
}

func TestName(t *testing.T) {
	t.Parallel()

	g := &tokenf1.Goal{}
	if got := g.Name(); got != "token-f1" {
		t.Errorf("Name() = %q, want %q", got, "token-f1")
	}
}

// TestScoreMakesNoProviderCall is the acceptance criterion this Goal exists
// to satisfy honestly: goal/registry.go's allowlist admits a Goal only on the
// assertion that Score never calls a provider. There is no provider handle
// anywhere in this Goal's construction or Score to call — this test pins that
// by construction: a zero-value Goal, scored against a background context
// with no client, transport or credential reachable from it, must still
// succeed.
func TestScoreMakesNoProviderCall(t *testing.T) {
	t.Parallel()

	g := &tokenf1.Goal{}
	c := &knov1.Case{Id: "c1", Expected: "tier-1"}
	r := &knov1.Response{CaseId: "c1", Output: "tier-1"}
	if _, err := g.Score(context.Background(), c, r); err != nil {
		t.Fatalf("Score: %v", err)
	}
}
