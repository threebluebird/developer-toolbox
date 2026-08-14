package markdowntool

import (
	"developer-toolbox/backend/models"
	"strings"
	"testing"
)

func TestMarkdownHTML(t *testing.T) {
	out := NewMarkdownTool().Execute(models.ToolInput{Payload: map[string]any{"action": "toHtml", "input": "# Title\n\n**bold**"}})
	if !out.Success || !strings.Contains(out.Data.(string), "<h1>Title</h1>") || !strings.Contains(out.Data.(string), "<strong>bold</strong>") {
		t.Fatalf("unexpected: %+v", out)
	}
}
