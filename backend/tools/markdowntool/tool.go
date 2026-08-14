package markdowntool

import (
	"developer-toolbox/backend/models"
	"fmt"
	"html"
	"regexp"
	"strings"
)

type MarkdownTool struct{}

func NewMarkdownTool() *MarkdownTool { return &MarkdownTool{} }
func (t *MarkdownTool) Info() models.Tool {
	return models.Tool{ID: "markdown", Name: "Markdown", Description: "Edit, preview, and convert Markdown to HTML.", Category: "text", Icon: "markdown", Version: "0.5.0", Keywords: []string{"markdown", "md", "html", "preview", "editor"}}
}

var linkPattern = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
var boldPattern = regexp.MustCompile(`\*\*([^*]+)\*\*`)
var italicPattern = regexp.MustCompile(`\*([^*]+)\*`)
var codePattern = regexp.MustCompile("`([^`]+)`")

func (t *MarkdownTool) Execute(input models.ToolInput) models.ToolOutput {
	raw, ok := input.Payload["input"].(string)
	if !ok {
		return bad("input must be a string")
	}
	action, _ := input.Payload["action"].(string)
	if action != "toHtml" && action != "preview" {
		return bad("unknown action")
	}
	return good(render(raw))
}
func render(raw string) string {
	lines := strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n")
	var b strings.Builder
	inList := false
	for _, line := range lines {
		escaped := html.EscapeString(line)
		escaped = codePattern.ReplaceAllString(escaped, "<code>$1</code>")
		escaped = boldPattern.ReplaceAllString(escaped, "<strong>$1</strong>")
		escaped = italicPattern.ReplaceAllString(escaped, "<em>$1</em>")
		escaped = linkPattern.ReplaceAllString(escaped, `<a href="$2" rel="noreferrer">$1</a>`)
		if strings.HasPrefix(line, "- ") {
			if !inList {
				b.WriteString("<ul>")
				inList = true
			}
			fmt.Fprintf(&b, "<li>%s</li>", escaped[2:])
			continue
		}
		if inList {
			b.WriteString("</ul>")
			inList = false
		}
		count := 0
		for count < len(line) && count < 6 && line[count] == '#' {
			count++
		}
		if count > 0 && len(line) > count && line[count] == ' ' {
			fmt.Fprintf(&b, "<h%d>%s</h%d>", count, escaped[count+1:], count)
		} else if strings.TrimSpace(line) == "" {
			b.WriteString("<br>")
		} else {
			fmt.Fprintf(&b, "<p>%s</p>", escaped)
		}
	}
	if inList {
		b.WriteString("</ul>")
	}
	return b.String()
}
func good(v any) models.ToolOutput   { return models.ToolOutput{Success: true, Data: v} }
func bad(s string) models.ToolOutput { return models.Failure("INVALID_MARKDOWN", s) }
