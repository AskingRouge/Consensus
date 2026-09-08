// Package export renders a session's items, scores, and comments as JSON
// or CSV for the research team to feed back into their evaluation
// pipeline. It's kept free of HTTP/database concerns — handlers.
// ExportSession gathers the data, this package only knows how to render
// it — so the rendering logic can be tested directly.
package export

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"consensus/internal/models"
)

// Data is everything one session's export is built from.
type Data struct {
	Session  models.Session
	Items    []models.Item
	Scores   []models.Score
	Comments []models.Comment
}

// itemExport is one item's JSON shape: its own fields plus every score
// and comment that belongs to it.
type itemExport struct {
	models.Item
	Scores   []models.Score   `json:"scores"`
	Comments []models.Comment `json:"comments"`
}

type jsonExport struct {
	Session models.Session `json:"session"`
	Items   []itemExport   `json:"items"`
}

// WriteJSON renders the full export — session, and every item nested
// with its own scores and comments — as indented JSON.
func WriteJSON(w io.Writer, data Data) error {
	scoresByItem := groupScores(data.Scores)
	commentsByItem := groupComments(data.Comments)

	items := make([]itemExport, len(data.Items))
	for i, item := range data.Items {
		// scoresByItem/commentsByItem return nil, not [], for an item
		// with no entries — coerce to an empty slice so it marshals as
		// [] rather than null, and a consumer never needs a null check.
		scores := scoresByItem[item.ID]
		if scores == nil {
			scores = []models.Score{}
		}
		comments := commentsByItem[item.ID]
		if comments == nil {
			comments = []models.Comment{}
		}
		items[i] = itemExport{Item: item, Scores: scores, Comments: comments}
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(jsonExport{Session: data.Session, Items: items})
}

// csvHeader is fixed for every row; the rubric dimension columns are
// appended dynamically since a session's rubric is configurable.
var csvBaseHeader = []string{
	"order_index", "question", "retrieved_context", "generated_answer",
	"flagged", "disagreeing_dimensions", "reviewer_email",
}

// WriteCSV renders one row per (item, reviewer) score, with a rubric
// dimension per column — the tabular shape a spreadsheet or analysis
// script expects. An item nobody has scored yet still gets a single row
// with blank score columns, so the export always covers every item.
//
// Comments aren't part of the CSV: a discussion thread doesn't have a
// natural one-row-per-item shape the way a score does. Use the JSON
// export for comments.
func WriteCSV(w io.Writer, data Data) error {
	scoresByItem := groupScores(data.Scores)

	cw := csv.NewWriter(w)

	header := append([]string{}, csvBaseHeader...)
	for _, dim := range data.Session.Rubric {
		header = append(header, dim.Name)
	}
	header = append(header, "submitted_at")
	if err := cw.Write(header); err != nil {
		return err
	}

	for _, item := range data.Items {
		scores := scoresByItem[item.ID]
		if len(scores) == 0 {
			if err := cw.Write(itemRow(item, data.Session.Rubric, nil)); err != nil {
				return err
			}
			continue
		}
		for _, s := range scores {
			s := s
			if err := cw.Write(itemRow(item, data.Session.Rubric, &s)); err != nil {
				return err
			}
		}
	}

	cw.Flush()
	return cw.Error()
}

func itemRow(item models.Item, rubric []models.RubricDimension, score *models.Score) []string {
	row := []string{
		strconv.Itoa(item.OrderIndex),
		item.Question,
		item.RetrievedContext,
		item.GeneratedAnswer,
		strconv.FormatBool(item.Flagged),
		strings.Join(item.DisagreeingDimensions, "; "),
	}

	if score == nil {
		blanks := make([]string, len(rubric)+2) // reviewer_email + rubric columns + submitted_at
		return append(row, blanks...)
	}

	row = append(row, score.ReviewerEmail)
	for _, dim := range rubric {
		v, ok := score.RubricScores[dim.Name]
		if !ok {
			row = append(row, "")
			continue
		}
		row = append(row, formatScoreValue(dim, v))
	}
	row = append(row, score.SubmittedAt.Format(time.RFC3339))

	return row
}

func formatScoreValue(dim models.RubricDimension, v float64) string {
	if dim.Type == models.RubricTypePassFail {
		if v == 1 {
			return "pass"
		}
		return "fail"
	}
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func groupScores(scores []models.Score) map[primitive.ObjectID][]models.Score {
	out := map[primitive.ObjectID][]models.Score{}
	for _, s := range scores {
		out[s.ItemID] = append(out[s.ItemID], s)
	}
	return out
}

func groupComments(comments []models.Comment) map[primitive.ObjectID][]models.Comment {
	out := map[primitive.ObjectID][]models.Comment{}
	for _, c := range comments {
		out[c.ItemID] = append(out[c.ItemID], c)
	}
	return out
}
