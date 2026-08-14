package crontool

import (
	"developer-toolbox/backend/models"
	"testing"
)

func TestParseAndGenerate(t *testing.T) {
	tool := NewCronTool()
	parsed := tool.Execute(models.ToolInput{Payload: map[string]any{"action": "parse", "input": "*/5 * * * *"}})
	if !parsed.Success || parsed.Data.(map[string]any)["description"] != "Every 5 minutes" {
		t.Fatalf("unexpected: %+v", parsed)
	}
	generated := tool.Execute(models.ToolInput{Payload: map[string]any{"action": "generate", "minute": "0", "hour": "9", "day": "*", "month": "*", "weekday": "1-5"}})
	if generated.Data.(map[string]any)["expression"] != "0 9 * * 1-5" {
		t.Fatalf("unexpected: %+v", generated)
	}
}
func TestRejectInvalidCron(t *testing.T) {
	if NewCronTool().Execute(models.ToolInput{Payload: map[string]any{"input": "99 * * * *"}}).Success {
		t.Fatal("expected failure")
	}
}
