package regex

import (
	"testing"

	"developer-toolbox/backend/models"
)

func TestRegexFlagsAndMatchCount(t *testing.T) {
	output := NewRegexTool().Execute(models.ToolInput{Payload: map[string]any{
		"pattern": "^hello", "input": "HELLO\nhello", "flags": "gim",
	}})
	if !output.Success {
		t.Fatalf("unexpected failure: %s", output.Error)
	}
	data := output.Data.(map[string]any)
	if data["count"] != 2 {
		t.Fatalf("expected 2 matches, got %#v", data)
	}
}

func TestRegexRejectsUnsupportedFlag(t *testing.T) {
	output := NewRegexTool().Execute(models.ToolInput{Payload: map[string]any{
		"pattern": ".", "input": "x", "flags": "y",
	}})
	if output.Success {
		t.Fatal("expected unsupported flag to fail")
	}
}
