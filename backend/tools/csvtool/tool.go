package csvtool

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"developer-toolbox/backend/models"
)

type CSVTool struct{}

func NewCSVTool() *CSVTool { return &CSVTool{} }
func (t *CSVTool) Info() models.Tool {
	return models.Tool{ID: "csv", Name: "CSV Toolkit", Description: "View, format, and convert CSV and JSON.", Category: "data", Icon: "csv", Version: "0.5.0", Keywords: []string{"csv", "json", "table", "convert", "format"}}
}
func (t *CSVTool) Execute(input models.ToolInput) models.ToolOutput {
	raw, valid := input.Payload["input"].(string)
	if !valid {
		return fail("input must be a string")
	}
	action, _ := input.Payload["action"].(string)
	comma := separator(input.Payload["separator"])
	switch action {
	case "view", "format", "csvToJson":
		r := csv.NewReader(strings.NewReader(raw))
		r.Comma = comma
		r.FieldsPerRecord = -1
		records, err := r.ReadAll()
		if err != nil {
			return fail(err.Error())
		}
		if action == "view" {
			return ok(records)
		}
		if action == "csvToJson" {
			return csvJSON(records, boolValue(input.Payload["header"], true))
		}
		var b bytes.Buffer
		w := csv.NewWriter(&b)
		w.Comma = comma
		w.UseCRLF = false
		if err := w.WriteAll(records); err != nil {
			return fail(err.Error())
		}
		return ok(strings.TrimSuffix(b.String(), "\n"))
	case "jsonToCsv":
		return jsonCSV(raw, comma, boolValue(input.Payload["header"], true))
	default:
		return fail("unknown action")
	}
}
func separator(v any) rune {
	s, _ := v.(string)
	if s == "tab" || s == "\\t" {
		return '\t'
	}
	r, _, _, err := strconvRune(s)
	if err == nil {
		return r
	}
	return ','
}
func strconvRune(s string) (rune, int, string, error) {
	if s == "" {
		return 0, 0, "", io.EOF
	}
	r := []rune(s)
	if len(r) != 1 {
		return 0, 0, "", fmt.Errorf("separator must be one character")
	}
	return r[0], 1, "", nil
}
func boolValue(v any, fallback bool) bool {
	b, ok := v.(bool)
	if ok {
		return b
	}
	return fallback
}
func csvJSON(records [][]string, header bool) models.ToolOutput {
	if len(records) == 0 {
		return ok([]any{})
	}
	var value any
	if header {
		keys := records[0]
		rows := make([]map[string]string, 0, len(records)-1)
		for _, record := range records[1:] {
			row := map[string]string{}
			for i, key := range keys {
				if i < len(record) {
					row[key] = record[i]
				}
			}
			rows = append(rows, row)
		}
		value = rows
	} else {
		value = records
	}
	b, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fail(err.Error())
	}
	return ok(string(b))
}
func jsonCSV(raw string, comma rune, header bool) models.ToolOutput {
	var rows []map[string]any
	if err := json.Unmarshal([]byte(raw), &rows); err != nil {
		return fail("JSON must be an array of objects")
	}
	keys := []string{}
	seen := map[string]bool{}
	for _, row := range rows {
		for k := range row {
			if !seen[k] {
				seen[k] = true
				keys = append(keys, k)
			}
		}
	}
	sort.Strings(keys)
	var b bytes.Buffer
	w := csv.NewWriter(&b)
	w.Comma = comma
	if header {
		_ = w.Write(keys)
	}
	for _, row := range rows {
		record := make([]string, len(keys))
		for i, k := range keys {
			record[i] = fmt.Sprint(row[k])
		}
		_ = w.Write(record)
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return fail(err.Error())
	}
	return ok(strings.TrimSuffix(b.String(), "\n"))
}
func ok(data any) models.ToolOutput         { return models.ToolOutput{Success: true, Data: data} }
func fail(message string) models.ToolOutput { return models.Failure("INVALID_CSV", message) }
