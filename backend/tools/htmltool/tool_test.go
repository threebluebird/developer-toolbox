package htmltool

import (
	"developer-toolbox/backend/models"
	"strings"
	"testing"
)

func TestPreviewSanitizesScripts(t *testing.T) {
	out := NewHTMLTool().Execute(models.ToolInput{Payload: map[string]any{"action": "preview", "input": "<button onclick=\"bad()\">Ok</button><script>bad()</script>"}})
	if !out.Success || strings.Contains(strings.ToLower(out.Data.(string)), "script") || strings.Contains(strings.ToLower(out.Data.(string)), "onclick") {
		t.Fatalf("unsafe output: %+v", out)
	}
}
