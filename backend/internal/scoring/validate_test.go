package scoring

import "testing"

func TestValidateRubricScores_Valid(t *testing.T) {
	input := map[string]float64{
		"Retrieval Quality":      4,
		"Course-Scope Adherence": 5,
		"Safety/Appropriateness": 1,
	}
	got, err := ValidateRubricScores(input, defaultRubric)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 3 || got["Retrieval Quality"] != 4 {
		t.Errorf("unexpected result: %v", got)
	}
}

func TestValidateRubricScores_MissingDimension(t *testing.T) {
	input := map[string]float64{
		"Retrieval Quality":      4,
		"Course-Scope Adherence": 5,
	}
	if _, err := ValidateRubricScores(input, defaultRubric); err == nil {
		t.Fatal("expected an error for a missing dimension")
	}
}

func TestValidateRubricScores_UnknownDimensionRejected(t *testing.T) {
	input := map[string]float64{
		"Retrieval Quality":      4,
		"Course-Scope Adherence": 5,
		"Safety/Appropriateness": 1,
		"Vibes":                  5, // not part of the rubric
	}
	if _, err := ValidateRubricScores(input, defaultRubric); err == nil {
		t.Fatal("expected an error for an unknown dimension")
	}
}

func TestValidateRubricScores_ScaleOutOfRange(t *testing.T) {
	cases := []float64{0, -1, 6, 100}
	for _, v := range cases {
		input := map[string]float64{
			"Retrieval Quality":      v,
			"Course-Scope Adherence": 5,
			"Safety/Appropriateness": 1,
		}
		if _, err := ValidateRubricScores(input, defaultRubric); err == nil {
			t.Errorf("expected an error for out-of-range scale value %v", v)
		}
	}
}

func TestValidateRubricScores_ScaleBoundsInclusive(t *testing.T) {
	for _, v := range []float64{1, 5} {
		input := map[string]float64{
			"Retrieval Quality":      v,
			"Course-Scope Adherence": 5,
			"Safety/Appropriateness": 1,
		}
		if _, err := ValidateRubricScores(input, defaultRubric); err != nil {
			t.Errorf("value %v should be valid (inclusive bounds): %v", v, err)
		}
	}
}

func TestValidateRubricScores_PassFailMustBeZeroOrOne(t *testing.T) {
	cases := []float64{-1, 0.5, 2, 5}
	for _, v := range cases {
		input := map[string]float64{
			"Retrieval Quality":      4,
			"Course-Scope Adherence": 5,
			"Safety/Appropriateness": v,
		}
		if _, err := ValidateRubricScores(input, defaultRubric); err == nil {
			t.Errorf("expected an error for pass_fail value %v", v)
		}
	}
}

func TestValidateRubricScores_ReturnsFreshMapNotSuperset(t *testing.T) {
	input := map[string]float64{
		"Retrieval Quality":      4,
		"Course-Scope Adherence": 5,
		"Safety/Appropriateness": 1,
	}
	got, err := ValidateRubricScores(input, defaultRubric)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Mutating the input afterward must not affect the returned map.
	input["Retrieval Quality"] = 999
	if got["Retrieval Quality"] != 4 {
		t.Error("ValidateRubricScores should return an independent copy")
	}
}
