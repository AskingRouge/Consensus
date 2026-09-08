package scoring

import "consensus/internal/models"

// Result describes whether an item needs discussion, and which specific
// rubric dimensions triggered that — so the UI can show reviewers exactly
// where they disagree, not just that they do.
type Result struct {
	Flagged               bool
	DisagreeingDimensions []string
}

// Detect computes disagreement across every reviewer's score on one item.
// A dimension counts as disagreeing when:
//   - scale: the spread between the highest and lowest rating meets or
//     exceeds threshold (the session's configurable
//     disagreement_threshold);
//   - pass_fail: reviewers didn't all give the same verdict.
//
// Disagreement is only meaningful once at least two reviewers have scored
// the item — scores holds at most one entry per reviewer (the caller is
// expected to pass the current, deduplicated-by-reviewer set, which is
// what an upsert-by-reviewer storage model naturally produces), and a
// dimension needs at least two of those reviewers to have rated it before
// it can disagree at all.
func Detect(scores []models.Score, rubric []models.RubricDimension, threshold float64) Result {
	if len(scores) < 2 {
		return Result{}
	}

	var disagreeing []string
	for _, dim := range rubric {
		values := valuesFor(scores, dim.Name)
		if len(values) < 2 {
			continue
		}

		switch dim.Type {
		case models.RubricTypeScale:
			if spread(values) >= threshold {
				disagreeing = append(disagreeing, dim.Name)
			}
		case models.RubricTypePassFail:
			if !allSame(values) {
				disagreeing = append(disagreeing, dim.Name)
			}
		}
	}

	return Result{Flagged: len(disagreeing) > 0, DisagreeingDimensions: disagreeing}
}

func valuesFor(scores []models.Score, dimension string) []float64 {
	var out []float64
	for _, s := range scores {
		if v, ok := s.RubricScores[dimension]; ok {
			out = append(out, v)
		}
	}
	return out
}

func spread(values []float64) float64 {
	min, max := values[0], values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return max - min
}

func allSame(values []float64) bool {
	for _, v := range values[1:] {
		if v != values[0] {
			return false
		}
	}
	return true
}
