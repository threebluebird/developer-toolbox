package hiittimer

import (
	"testing"

	"developer-toolbox/backend/models"
)

func TestTimerPlan(t *testing.T) {
	result := NewTool().Execute(models.ToolInput{Payload: map[string]any{
		"workSeconds": 20.0,
		"restSeconds": 10.0,
		"rounds":      8.0,
	}})
	if !result.Success {
		t.Fatalf("expected success: %#v", result)
	}
	data := result.Data.(map[string]any)
	if data["totalSeconds"] != 230 || data["intervals"] != 15 {
		t.Fatalf("unexpected plan: %#v", data)
	}
}

func TestTimerPlanRejectsInvalidConfig(t *testing.T) {
	result := NewTool().Execute(models.ToolInput{Payload: map[string]any{"rounds": 0.0}})
	if result.Success || result.Error == nil || result.Error.Code != "INVALID_HIIT_CONFIG" {
		t.Fatalf("unexpected result: %#v", result)
	}
}
