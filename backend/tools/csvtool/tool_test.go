package csvtool

import (
	"developer-toolbox/backend/models"
	"strings"
	"testing"
)

func TestCSVJSONRoundTrip(t *testing.T) {
	tool := NewCSVTool()
	out := tool.Execute(models.ToolInput{Payload: map[string]any{"action": "csvToJson", "input": "name,age\nTom,30"}})
	if !out.Success || !strings.Contains(out.Data.(string), `"name": "Tom"`) {
		t.Fatalf("unexpected: %+v", out)
	}
	back := tool.Execute(models.ToolInput{Payload: map[string]any{"action": "jsonToCsv", "input": out.Data.(string)}})
	if !back.Success || !strings.Contains(back.Data.(string), "30,Tom") {
		t.Fatalf("unexpected: %+v", back)
	}
}
