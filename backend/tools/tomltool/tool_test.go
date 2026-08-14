package tomltool

import (
	"developer-toolbox/backend/models"
	"strings"
	"testing"
)

func TestTOMLConversions(t *testing.T) {
	tool := NewTOMLTool()
	out := tool.Execute(models.ToolInput{Payload: map[string]any{"action": "tomlToJson", "input": "title = \"App\"\n[server]\nport = 8080"}})
	if !out.Success || !strings.Contains(out.Data.(string), `"port": 8080`) {
		t.Fatalf("unexpected: %+v", out)
	}
	back := tool.Execute(models.ToolInput{Payload: map[string]any{"action": "jsonToToml", "input": out.Data.(string)}})
	if !back.Success || !strings.Contains(back.Data.(string), "[server]") {
		t.Fatalf("unexpected: %+v", back)
	}
}
