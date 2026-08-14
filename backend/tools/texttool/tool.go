package texttool

import (
	"encoding/json"
	"html"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"developer-toolbox/backend/models"
)

type TextTool struct{}

func NewTextTool() *TextTool { return &TextTool{} }
func (t *TextTool) Info() models.Tool {
	return models.Tool{ID: "text", Name: "Text Toolkit", Description: "Format, escape, and inspect text.", Category: "text", Icon: "text", Version: "0.5.0", Keywords: []string{"text", "trim", "lines", "escape", "statistics", "sort"}}
}
func (t *TextTool) Execute(input models.ToolInput) models.ToolOutput {
	raw, ok := input.Payload["input"].(string)
	if !ok {
		return failure("input must be a string")
	}
	action, _ := input.Payload["action"].(string)
	switch action {
	case "trim":
		return success(strings.TrimSpace(raw))
	case "removeEmpty":
		return success(filterLines(raw, func(line string) bool { return strings.TrimSpace(line) != "" }))
	case "deduplicate":
		seen := map[string]bool{}
		return success(filterLines(raw, func(line string) bool {
			if seen[line] {
				return false
			}
			seen[line] = true
			return true
		}))
	case "sort":
		lines := strings.Split(normalize(raw), "\n")
		sort.Strings(lines)
		return success(strings.Join(lines, "\n"))
	case "reverse":
		lines := strings.Split(normalize(raw), "\n")
		for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
			lines[i], lines[j] = lines[j], lines[i]
		}
		return success(strings.Join(lines, "\n"))
	case "prefix", "suffix":
		value, _ := input.Payload["value"].(string)
		lines := strings.Split(normalize(raw), "\n")
		for i := range lines {
			if action == "prefix" {
				lines[i] = value + lines[i]
			} else {
				lines[i] += value
			}
		}
		return success(strings.Join(lines, "\n"))
	case "jsonEscape":
		encoded, _ := json.Marshal(raw)
		return success(string(encoded[1 : len(encoded)-1]))
	case "htmlEscape":
		return success(html.EscapeString(raw))
	case "urlEscape":
		return success(url.QueryEscape(raw))
	case "unicodeEscape":
		var b strings.Builder
		for _, r := range raw {
			if r > 127 {
				if r <= 0xffff {
					b.WriteString(fmtUnicode(r, 4))
				} else {
					b.WriteString(fmtUnicode(r, 8))
				}
			} else {
				b.WriteRune(r)
			}
		}
		return success(b.String())
	case "statistics":
		lines := 0
		if raw != "" {
			lines = strings.Count(normalize(raw), "\n") + 1
		}
		return success(map[string]any{"characters": len(raw), "words": len(strings.Fields(raw)), "lines": lines, "bytes": len([]byte(raw)), "unicodeCharacters": utf8.RuneCountInString(raw)})
	default:
		return failure("unknown action")
	}
}
func normalize(s string) string { return strings.ReplaceAll(s, "\r\n", "\n") }
func filterLines(raw string, keep func(string) bool) string {
	result := []string{}
	for _, line := range strings.Split(normalize(raw), "\n") {
		if keep(line) {
			result = append(result, line)
		}
	}
	return strings.Join(result, "\n")
}
func fmtUnicode(r rune, width int) string {
	return "\\U" + strings.Repeat("0", width-len(strconv.FormatInt(int64(r), 16))) + strings.ToUpper(strconv.FormatInt(int64(r), 16))
}
func success(data any) models.ToolOutput       { return models.ToolOutput{Success: true, Data: data} }
func failure(message string) models.ToolOutput { return models.Failure("INVALID_TEXT", message) }
