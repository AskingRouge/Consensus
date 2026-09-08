// Package scoring holds the business logic around rubric scores: input
// validation and disagreement detection. It's kept free of any HTTP or
// database concerns specifically so it can be tested directly, per the
// project's non-functional requirements.
package scoring

import (
	"fmt"

	"consensus/internal/models"
)

// ValidateRubricScores checks a submitted rubric_scores map against a
// session's rubric: every dimension must be present (no partial scoring —
// a reviewer either scores the whole item or hasn't scored it yet, which
// keeps "has this item been scored" unambiguous), no unknown dimensions
// are allowed, scale values must fall within [1, dim.ScaleMax], and
// pass_fail values must be exactly 0 (fail) or 1 (pass).
//
// On success it returns a fresh map containing exactly the rubric's
// dimensions — never a superset — so a stray or renamed key in the
// request can never sneak into storage.
func ValidateRubricScores(input map[string]float64, rubric []models.RubricDimension) (map[string]float64, error) {
	if len(input) != len(rubric) {
		return nil, fmt.Errorf("must provide a score for exactly the %d rubric dimensions", len(rubric))
	}

	out := make(map[string]float64, len(rubric))
	for _, dim := range rubric {
		v, ok := input[dim.Name]
		if !ok {
			return nil, fmt.Errorf("missing score for rubric dimension %q", dim.Name)
		}

		switch dim.Type {
		case models.RubricTypeScale:
			if v < 1 || v > float64(dim.ScaleMax) {
				return nil, fmt.Errorf("%q must be between 1 and %d", dim.Name, dim.ScaleMax)
			}
		case models.RubricTypePassFail:
			if v != 0 && v != 1 {
				return nil, fmt.Errorf("%q must be 0 (fail) or 1 (pass)", dim.Name)
			}
		default:
			return nil, fmt.Errorf("session has an unknown rubric dimension type %q", dim.Type)
		}

		out[dim.Name] = v
	}

	return out, nil
}
