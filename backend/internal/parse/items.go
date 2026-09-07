// Package parse turns an uploaded batch file (JSON or CSV) into
// models.ItemInput values, matching the output format of the research
// harness: one row/object per (question, retrieved_context,
// generated_answer) triple.
package parse

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"consensus/internal/models"
)

// Items parses raw batch data as either JSON or CSV, chosen by format
// ("json" or "csv"). JSON may be either a bare array of items or an object
// of the form {"items": [...]}.
func Items(r io.Reader, format string) ([]models.ItemInput, error) {
	switch strings.ToLower(format) {
	case "json":
		return parseJSON(r)
	case "csv":
		return parseCSV(r)
	default:
		return nil, fmt.Errorf("unsupported format %q: expected \"json\" or \"csv\"", format)
	}
}

func parseJSON(r io.Reader) ([]models.ItemInput, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var items []models.ItemInput
	if err := json.Unmarshal(raw, &items); err == nil {
		return items, nil
	}

	var wrapped struct {
		Items []models.ItemInput `json:"items"`
	}
	if err := json.Unmarshal(raw, &wrapped); err != nil {
		return nil, fmt.Errorf("invalid JSON: expected an array of items or {\"items\": [...]}: %w", err)
	}
	return wrapped.Items, nil
}

// expected CSV headers, matched case-insensitively.
var csvColumns = []string{"question", "retrieved_context", "generated_answer"}

func parseCSV(r io.Reader) ([]models.ItemInput, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true

	header, err := reader.Read()
	if err == io.EOF {
		return nil, fmt.Errorf("CSV file is empty")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	colIndex := make(map[string]int, len(header))
	for i, col := range header {
		colIndex[strings.ToLower(strings.TrimSpace(col))] = i
	}
	for _, required := range csvColumns {
		if _, ok := colIndex[required]; !ok {
			return nil, fmt.Errorf("CSV is missing required column %q", required)
		}
	}

	var items []models.ItemInput
	rowNum := 1
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV row %d: %w", rowNum, err)
		}
		rowNum++

		items = append(items, models.ItemInput{
			Question:         field(row, colIndex, "question"),
			RetrievedContext: field(row, colIndex, "retrieved_context"),
			GeneratedAnswer:  field(row, colIndex, "generated_answer"),
		})
	}

	return items, nil
}

func field(row []string, colIndex map[string]int, name string) string {
	idx, ok := colIndex[name]
	if !ok || idx >= len(row) {
		return ""
	}
	return row[idx]
}
