package scoring

import (
	"reflect"
	"testing"

	"consensus/internal/models"
)

var defaultRubric = models.DefaultRubric() // Retrieval Quality (scale 5), Course-Scope Adherence (scale 5), Safety/Appropriateness (pass_fail)

func score(reviewer string, retrieval, scope, safety float64) models.Score {
	return models.Score{
		ReviewerEmail: reviewer,
		RubricScores: map[string]float64{
			"Retrieval Quality":      retrieval,
			"Course-Scope Adherence": scope,
			"Safety/Appropriateness": safety,
		},
	}
}

func TestDetect_FewerThanTwoScoresNeverFlags(t *testing.T) {
	cases := [][]models.Score{
		nil,
		{score("a@unc.edu", 5, 5, 1)},
	}
	for _, scores := range cases {
		result := Detect(scores, defaultRubric, 2)
		if result.Flagged {
			t.Errorf("Detect(%d scores) flagged, want not flagged", len(scores))
		}
	}
}

func TestDetect_AgreementNotFlagged(t *testing.T) {
	scores := []models.Score{
		score("a@unc.edu", 4, 5, 1),
		score("b@unc.edu", 5, 4, 1), // within threshold=2 on every dimension, same pass/fail
	}
	result := Detect(scores, defaultRubric, 2)
	if result.Flagged {
		t.Errorf("expected no flag, got disagreeing dimensions: %v", result.DisagreeingDimensions)
	}
}

func TestDetect_ScaleSpreadAtThresholdFlags(t *testing.T) {
	scores := []models.Score{
		score("a@unc.edu", 2, 5, 1),
		score("b@unc.edu", 4, 5, 1), // spread of exactly 2 on Retrieval Quality
	}
	result := Detect(scores, defaultRubric, 2)
	if !result.Flagged {
		t.Fatal("expected a flag when spread equals the threshold")
	}
	if !contains(result.DisagreeingDimensions, "Retrieval Quality") {
		t.Errorf("expected Retrieval Quality to disagree, got %v", result.DisagreeingDimensions)
	}
	if contains(result.DisagreeingDimensions, "Course-Scope Adherence") {
		t.Errorf("Course-Scope Adherence should not disagree, got %v", result.DisagreeingDimensions)
	}
}

func TestDetect_ScaleSpreadBelowThresholdNotFlagged(t *testing.T) {
	scores := []models.Score{
		score("a@unc.edu", 3, 5, 1),
		score("b@unc.edu", 4, 5, 1), // spread of 1, threshold is 2
	}
	result := Detect(scores, defaultRubric, 2)
	if result.Flagged {
		t.Errorf("spread of 1 should not flag at threshold 2, got %v", result.DisagreeingDimensions)
	}
}

func TestDetect_PassFailDisagreementFlags(t *testing.T) {
	scores := []models.Score{
		score("a@unc.edu", 5, 5, 1), // pass
		score("b@unc.edu", 5, 5, 0), // fail
	}
	result := Detect(scores, defaultRubric, 2)
	if !result.Flagged {
		t.Fatal("expected a flag on pass/fail disagreement")
	}
	if !reflect.DeepEqual(result.DisagreeingDimensions, []string{"Safety/Appropriateness"}) {
		t.Errorf("disagreeing dimensions = %v", result.DisagreeingDimensions)
	}
}

func TestDetect_PassFailAgreementNotFlagged(t *testing.T) {
	scores := []models.Score{
		score("a@unc.edu", 1, 1, 0),
		score("b@unc.edu", 1, 1, 0), // both fail: agreement, even though the score is bad
	}
	result := Detect(scores, defaultRubric, 2)
	if result.Flagged {
		t.Errorf("unanimous fail should not flag, got %v", result.DisagreeingDimensions)
	}
}

func TestDetect_ThreeReviewersOutlierFlags(t *testing.T) {
	scores := []models.Score{
		score("a@unc.edu", 5, 5, 1),
		score("b@unc.edu", 5, 5, 1),
		score("c@unc.edu", 1, 5, 1), // one outlier on Retrieval Quality
	}
	result := Detect(scores, defaultRubric, 2)
	if !result.Flagged {
		t.Fatal("expected a flag from the outlier")
	}
	if !contains(result.DisagreeingDimensions, "Retrieval Quality") {
		t.Errorf("expected Retrieval Quality to disagree, got %v", result.DisagreeingDimensions)
	}
}

func TestDetect_MultipleDisagreeingDimensions(t *testing.T) {
	scores := []models.Score{
		score("a@unc.edu", 1, 1, 0),
		score("b@unc.edu", 5, 5, 1), // disagrees on every dimension
	}
	result := Detect(scores, defaultRubric, 2)
	want := []string{"Retrieval Quality", "Course-Scope Adherence", "Safety/Appropriateness"}
	if !reflect.DeepEqual(result.DisagreeingDimensions, want) {
		t.Errorf("disagreeing dimensions = %v, want %v", result.DisagreeingDimensions, want)
	}
}

func TestDetect_PartialScoringOnlyComparesDimensionsBothRated(t *testing.T) {
	scores := []models.Score{
		{ReviewerEmail: "a@unc.edu", RubricScores: map[string]float64{"Retrieval Quality": 5}},
		{ReviewerEmail: "b@unc.edu", RubricScores: map[string]float64{"Course-Scope Adherence": 1}},
	}
	// Neither dimension has two ratings, so nothing can be compared, even
	// though these two reviewers technically "scored" the item.
	result := Detect(scores, defaultRubric, 2)
	if result.Flagged {
		t.Errorf("expected no flag when no dimension has two ratings, got %v", result.DisagreeingDimensions)
	}
}

func TestDetect_ZeroThresholdFlagsAnySpread(t *testing.T) {
	scores := []models.Score{
		score("a@unc.edu", 4, 5, 1),
		score("b@unc.edu", 5, 5, 1), // spread of 1
	}
	result := Detect(scores, defaultRubric, 0)
	if !result.Flagged {
		t.Fatal("a threshold of 0 should flag any nonzero spread")
	}
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
