package texttool

import (
	"developer-toolbox/backend/models"
	"testing"
)

func run(action, input string, extra map[string]any) models.ToolOutput {
	p := map[string]any{"action": action, "input": input}
	for k, v := range extra {
		p[k] = v
	}
	return NewTextTool().Execute(models.ToolInput{Payload: p})
}
func TestFormattingAndStatistics(t *testing.T) {
	if got := run("deduplicate", "b\na\nb", nil).Data; got != "b\na" {
		t.Fatalf("got %q", got)
	}
	out := run("statistics", "你好 world", nil)
	stats := out.Data.(map[string]any)
	if stats["unicodeCharacters"] != 8 || stats["bytes"] != 12 {
		t.Fatalf("unexpected stats: %#v", stats)
	}
}
func TestEscape(t *testing.T) {
	if got := run("htmlEscape", "<b>", nil).Data; got != "&lt;b&gt;" {
		t.Fatalf("got %q", got)
	}
}
