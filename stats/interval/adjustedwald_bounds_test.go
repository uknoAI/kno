package interval_test

import (
	"testing"

	knov1 "github.com/knograph/kno/gen/kno/v1"
	"github.com/knograph/kno/stats/interval"
)

// TestAdjustedWaldStaysInsideItsEstimandsRange pins a bound that was reported
// outside the range the estimand can take.
//
// A paired difference of binary scores lies in [-1, +1] by construction: every
// pair contributes -1, 0 or +1. The adjusted-Wald interval is a normal
// approximation around an adjusted point estimate, and normal approximations
// run off the end of a bounded parameter space at extreme observed rates — the
// textbook failure that retired the naive Wald interval for a single
// proportion, here in its paired form.
//
// It was systematic, not a small-sample curiosity: with every pair improving,
// the upper bound read 1.2327 at n=5 and was still 1.0362 at n=50. Nothing
// downstream caught it — build knows no domain, portfolio.Correct widens the
// half and pushes an out-of-range bound further out, and select only asks
// whether an interval crosses zero. So the decision was right and the number
// was impossible.
func TestAdjustedWaldStaysInsideItsEstimandsRange(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		delta float64
	}{
		{"every pair improved", 1},
		{"every pair regressed", -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			for _, n := range []int{5, 6, 10, 20, 50, 200} {
				d := make([]float64, n)
				for i := range d {
					d[i] = tc.delta
				}
				ci := interval.Paired(d, knov1.ScoreDomain_SCORE_DOMAIN_BINARY, 1, interval.DefaultLevel)
				if ci == nil {
					t.Fatalf("n=%d: no interval for a sample of identical binary differences", n)
				}
				if ci.GetHigh() > 1 || ci.GetLow() < -1 {
					t.Errorf("n=%d: [%.4f, %.4f] leaves [-1, +1]. A paired difference of "+
						"binary scores cannot take a value outside that range, so this "+
						"interval asserts something the estimand cannot do",
						n, ci.GetLow(), ci.GetHigh())
				}
			}
		})
	}

	// The clamp must not flatten an interval that was already inside the
	// range — a guard that always fires is not a guard.
	t.Run("an interval already inside its range is untouched", func(t *testing.T) {
		t.Parallel()

		d := []float64{1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0}
		ci := interval.Paired(d, knov1.ScoreDomain_SCORE_DOMAIN_BINARY, 1, interval.DefaultLevel)
		if ci == nil {
			t.Fatal("no interval for a mixed sample")
		}
		if ci.GetHigh() >= 1 || ci.GetLow() <= -1 {
			t.Errorf("a mixed sample produced [%.4f, %.4f], which sits on the clamp "+
				"rather than inside it — the clamp is doing more than confining",
				ci.GetLow(), ci.GetHigh())
		}
	})
}
