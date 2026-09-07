package parse

import (
	"strings"
	"testing"
)

func TestItems_JSONArray(t *testing.T) {
	raw := `[
		{"question": "What is the deadline?", "retrieved_context": "syllabus.pdf", "generated_answer": "March 3rd."}
	]`
	items, err := Items(strings.NewReader(raw), "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Question != "What is the deadline?" {
		t.Errorf("question = %q", items[0].Question)
	}
}

func TestItems_JSONWrapped(t *testing.T) {
	raw := `{"items": [{"question": "Q1", "retrieved_context": "C1", "generated_answer": "A1"}]}`
	items, err := Items(strings.NewReader(raw), "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].GeneratedAnswer != "A1" {
		t.Fatalf("unexpected items: %+v", items)
	}
}

func TestItems_JSONInvalid(t *testing.T) {
	if _, err := Items(strings.NewReader("not json"), "json"); err == nil {
		t.Fatal("expected an error for invalid JSON")
	}
}

func TestItems_CSV(t *testing.T) {
	raw := "question,retrieved_context,generated_answer\n" +
		"What is the deadline?,syllabus.pdf,March 3rd.\n" +
		"Is X allowed?,policy.pdf,Yes with instructor approval.\n"

	items, err := Items(strings.NewReader(raw), "csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[1].Question != "Is X allowed?" {
		t.Errorf("row 2 question = %q", items[1].Question)
	}
}

func TestItems_CSVMissingColumn(t *testing.T) {
	raw := "question,generated_answer\nQ1,A1\n"
	if _, err := Items(strings.NewReader(raw), "csv"); err == nil {
		t.Fatal("expected an error for missing retrieved_context column")
	}
}

func TestItems_CSVColumnOrderIndependent(t *testing.T) {
	raw := "generated_answer,question,retrieved_context\nA1,Q1,C1\n"
	items, err := Items(strings.NewReader(raw), "csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if items[0].Question != "Q1" || items[0].RetrievedContext != "C1" || items[0].GeneratedAnswer != "A1" {
		t.Fatalf("columns not mapped correctly: %+v", items[0])
	}
}

func TestItems_UnsupportedFormat(t *testing.T) {
	if _, err := Items(strings.NewReader("x"), "xml"); err == nil {
		t.Fatal("expected an error for an unsupported format")
	}
}
