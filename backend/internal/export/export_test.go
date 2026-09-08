package export

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"consensus/internal/models"
)

func testData() Data {
	itemID := primitive.NewObjectID()
	session := models.Session{
		Name:   "Test Session",
		Rubric: models.DefaultRubric(),
	}
	item := models.Item{
		ID:                    itemID,
		Question:              "What is the deadline?",
		RetrievedContext:      "syllabus.pdf",
		GeneratedAnswer:       "March 3rd.",
		OrderIndex:            0,
		Flagged:               true,
		DisagreeingDimensions: []string{"Retrieval Quality"},
	}
	scores := []models.Score{
		{
			ItemID:        itemID,
			ReviewerEmail: "alice@unc.edu",
			RubricScores: map[string]float64{
				"Retrieval Quality":      4,
				"Course-Scope Adherence": 5,
				"Safety/Appropriateness": 1,
			},
			SubmittedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
		},
		{
			ItemID:        itemID,
			ReviewerEmail: "bob@unc.edu",
			RubricScores: map[string]float64{
				"Retrieval Quality":      1,
				"Course-Scope Adherence": 5,
				"Safety/Appropriateness": 0,
			},
			SubmittedAt: time.Date(2026, 1, 2, 3, 5, 0, 0, time.UTC),
		},
	}
	comments := []models.Comment{
		{ItemID: itemID, AuthorEmail: "alice@unc.edu", Text: "hi", CreatedAt: time.Now()},
	}

	return Data{Session: session, Items: []models.Item{item}, Scores: scores, Comments: comments}
}

func TestWriteJSON_RoundTrips(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteJSON(&buf, testData()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded jsonExport
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if decoded.Session.Name != "Test Session" {
		t.Errorf("session name = %q", decoded.Session.Name)
	}
	if len(decoded.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(decoded.Items))
	}
	if len(decoded.Items[0].Scores) != 2 {
		t.Errorf("expected 2 scores nested under the item, got %d", len(decoded.Items[0].Scores))
	}
	if len(decoded.Items[0].Comments) != 1 {
		t.Errorf("expected 1 comment nested under the item, got %d", len(decoded.Items[0].Comments))
	}
}

func TestWriteJSON_ItemWithNoScoresGetsEmptySlice(t *testing.T) {
	data := testData()
	data.Scores = nil
	data.Comments = nil

	var buf bytes.Buffer
	if err := WriteJSON(&buf, data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded jsonExport
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	// Must be present as [], not omitted or null, so a consumer doesn't
	// need to special-case "no scores yet".
	if decoded.Items[0].Scores == nil {
		t.Error("expected an empty (non-nil) Scores slice")
	}
}

func TestWriteCSV_OneRowPerScore(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteCSV(&buf, testData()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rows, err := csv.NewReader(&buf).ReadAll()
	if err != nil {
		t.Fatalf("output is not valid CSV: %v", err)
	}

	// header + 2 score rows
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows (header + 2 scores), got %d: %v", len(rows), rows)
	}

	header := rows[0]
	wantCols := []string{
		"order_index", "question", "retrieved_context", "generated_answer",
		"flagged", "disagreeing_dimensions", "reviewer_email",
		"Retrieval Quality", "Course-Scope Adherence", "Safety/Appropriateness",
		"submitted_at",
	}
	if len(header) != len(wantCols) {
		t.Fatalf("header = %v, want %v", header, wantCols)
	}
	for i, col := range wantCols {
		if header[i] != col {
			t.Errorf("header[%d] = %q, want %q", i, header[i], col)
		}
	}

	reviewerCol := indexOf(header, "reviewer_email")
	if rows[1][reviewerCol] != "alice@unc.edu" || rows[2][reviewerCol] != "bob@unc.edu" {
		t.Errorf("unexpected reviewer order: %q, %q", rows[1][reviewerCol], rows[2][reviewerCol])
	}

	safetyCol := indexOf(header, "Safety/Appropriateness")
	if rows[1][safetyCol] != "pass" {
		t.Errorf("alice's Safety/Appropriateness = %q, want \"pass\"", rows[1][safetyCol])
	}
	if rows[2][safetyCol] != "fail" {
		t.Errorf("bob's Safety/Appropriateness = %q, want \"fail\"", rows[2][safetyCol])
	}

	flaggedCol := indexOf(header, "flagged")
	if rows[1][flaggedCol] != "true" {
		t.Errorf("flagged = %q, want \"true\"", rows[1][flaggedCol])
	}
}

func TestWriteCSV_UnscoredItemGetsOneBlankRow(t *testing.T) {
	data := testData()
	data.Scores = nil

	var buf bytes.Buffer
	if err := WriteCSV(&buf, data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rows, err := csv.NewReader(&buf).ReadAll()
	if err != nil {
		t.Fatalf("output is not valid CSV: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 1 header + 1 blank row, got %d: %v", len(rows), rows)
	}

	reviewerCol := indexOf(rows[0], "reviewer_email")
	if rows[1][reviewerCol] != "" {
		t.Errorf("expected a blank reviewer_email for an unscored item, got %q", rows[1][reviewerCol])
	}
}

func indexOf(header []string, name string) int {
	for i, h := range header {
		if h == name {
			return i
		}
	}
	panic("column not found: " + name)
}
