package htmltool

import (
	"developer-toolbox/backend/models"
	"html"
	"regexp"
	"strings"
)

type HTMLTool struct{}

func NewHTMLTool() *HTMLTool { return &HTMLTool{} }
func (t *HTMLTool) Info() models.Tool {
	return models.Tool{ID: "html", Name: "HTML Toolkit", Description: "Format, minify, escape, and preview HTML.", Category: "text", Icon: "html", Version: "0.5.0", Keywords: []string{"html", "format", "minify", "escape", "preview"}}
}

var between = regexp.MustCompile(`>\s+<`)
var tags = regexp.MustCompile(`(?s)<[^>]+>`)

func (t *HTMLTool) Execute(input models.ToolInput) models.ToolOutput {
	raw, ok := input.Payload["input"].(string)
	if !ok {
		return bad("input must be a string")
	}
	action, _ := input.Payload["action"].(string)
	switch action {
	case "escape":
		return good(html.EscapeString(raw))
	case "minify":
		return good(strings.TrimSpace(between.ReplaceAllString(raw, "><")))
	case "preview":
		return good(sanitize(raw))
	case "format":
		return good(format(raw))
	default:
		return bad("unknown action")
	}
}
func format(raw string) string {
	compact := between.ReplaceAllString(raw, "><")
	parts := tags.FindAllStringIndex(compact, -1)
	var b strings.Builder
	depth := 0
	offset := 0
	void := regexp.MustCompile(`(?i)^<(area|base|br|col|embed|hr|img|input|link|meta|param|source|track|wbr)\b`)
	for _, loc := range parts {
		text := strings.TrimSpace(compact[offset:loc[0]])
		tag := compact[loc[0]:loc[1]]
		closing := strings.HasPrefix(tag, "</")
		if closing && depth > 0 {
			depth--
		}
		if text != "" {
			b.WriteString(strings.Repeat("  ", depth) + text + "\n")
		}
		b.WriteString(strings.Repeat("  ", depth) + tag + "\n")
		if !closing && !strings.HasSuffix(tag, "/>") && !strings.HasPrefix(tag, "<!") && !void.MatchString(tag) {
			depth++
		}
		offset = loc[1]
	}
	if tail := strings.TrimSpace(compact[offset:]); tail != "" {
		b.WriteString(strings.Repeat("  ", depth) + tail + "\n")
	}
	return strings.TrimSpace(b.String())
}
func sanitize(raw string) string {
	script := regexp.MustCompile(`(?is)<script\b[^>]*>.*?</script>`)
	dangerousBlocks := regexp.MustCompile(`(?is)<(iframe|object)\b[^>]*>.*?</(iframe|object)\s*>`)
	dangerousTags := regexp.MustCompile(`(?is)</?(iframe|object|embed|link|meta|base)\b[^>]*\/?>`)
	style := regexp.MustCompile(`(?is)<style\b[^>]*>.*?</style>`)
	events := regexp.MustCompile(`(?i)\s+on[a-z]+\s*=\s*("[^"]*"|'[^']*')`)
	jsDouble := regexp.MustCompile(`(?i)(href|src)\s*=\s*"\s*javascript:[^"]*"`)
	jsSingle := regexp.MustCompile(`(?i)(href|src)\s*=\s*'\s*javascript:[^']*'`)
	clean := script.ReplaceAllString(raw, "")
	clean = style.ReplaceAllString(clean, "")
	clean = dangerousBlocks.ReplaceAllString(clean, "")
	clean = dangerousTags.ReplaceAllString(clean, "")
	clean = events.ReplaceAllString(clean, "")
	clean = jsDouble.ReplaceAllString(clean, `$1="#"`)
	return jsSingle.ReplaceAllString(clean, `$1="#"`)
}
func good(v any) models.ToolOutput   { return models.ToolOutput{Success: true, Data: v} }
func bad(s string) models.ToolOutput { return models.Failure("INVALID_HTML", s) }
