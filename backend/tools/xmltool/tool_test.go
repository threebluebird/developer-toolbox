package xmltool

import (
	"developer-toolbox/backend/models"
	"strings"
	"testing"
)

func TestXMLJSONConversions(t *testing.T) {
	tool := NewXMLTool()
	out := tool.Execute(models.ToolInput{Payload: map[string]any{"action": "xmlToJson", "input": "<user><name>Tom</name></user>"}})
	if !out.Success || !strings.Contains(out.Data.(string), `"name": "Tom"`) {
		t.Fatalf("unexpected: %+v", out)
	}
	back := tool.Execute(models.ToolInput{Payload: map[string]any{"action": "jsonToXml", "input": out.Data.(string)}})
	if !back.Success || !strings.Contains(back.Data.(string), "<user>") {
		t.Fatalf("unexpected: %+v", back)
	}
}
