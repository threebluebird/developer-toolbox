package base64tool

import (
	"testing"

	"developer-toolbox/backend/models"
)

func TestBase64RoundTripAndInvalidInput(t *testing.T) {
	tool := NewBase64Tool()
	encoded := tool.Execute(models.ToolInput{Payload: map[string]any{"input": "你好", "action": "encode"}})
	decoded := tool.Execute(models.ToolInput{Payload: map[string]any{"input": encoded.Data, "action": "decode"}})
	if !decoded.Success || decoded.Data != "你好" {
		t.Fatalf("round trip failed: %#v", decoded)
	}
	invalid := tool.Execute(models.ToolInput{Payload: map[string]any{"input": "%%%", "action": "decode"}})
	if invalid.Success || invalid.Error.Code != "INVALID_BASE64" {
		t.Fatalf("expected structured error: %#v", invalid)
	}
}
