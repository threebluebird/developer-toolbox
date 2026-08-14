package urlparser

import (
	"developer-toolbox/backend/models"
	"testing"
)

func TestURLParser(t *testing.T) {
	out := NewURLParserTool().Execute(models.ToolInput{Payload: map[string]any{"input": "https://tom@example.com:8443/a%20b?q=one&q=two#part"}})
	if !out.Success {
		t.Fatalf("unexpected: %+v", out)
	}
	v := out.Data.(map[string]any)
	if v["host"] != "example.com" || v["port"] != "8443" || v["path"] != "/a%20b" || v["username"] != "tom" {
		t.Fatalf("unexpected: %#v", v)
	}
}
