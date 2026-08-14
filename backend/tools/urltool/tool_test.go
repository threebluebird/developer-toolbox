package urltool

import (
	"testing"

	"developer-toolbox/backend/models"
)

func TestURLQueryOperations(t *testing.T) {
	output := NewURLTool().Execute(models.ToolInput{Payload: map[string]any{"input": "a=1&a=2&name=hello+world", "action": "decodeQuery"}})
	if !output.Success {
		t.Fatal(output.Error)
	}
	if output.Data.(map[string]any)["name"] != "hello world" {
		t.Fatalf("unexpected query: %#v", output.Data)
	}
}
