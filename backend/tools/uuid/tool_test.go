package uuidtool

import (
	"testing"

	"developer-toolbox/backend/models"
)

func TestUUIDToolRejectsOutOfRangeCount(t *testing.T) {
	tool := NewUUIDTool()
	for _, count := range []any{0, 101, -1, 1.5, "10"} {
		output := tool.Execute(models.ToolInput{Payload: map[string]any{"count": count}})
		if output.Success {
			t.Fatalf("expected count %v to fail", count)
		}
	}
}

func TestUUIDToolGeneratesRequestedCount(t *testing.T) {
	output := NewUUIDTool().Execute(models.ToolInput{Payload: map[string]any{"count": float64(3)}})
	values, ok := output.Data.([]string)
	if !output.Success || !ok || len(values) != 3 {
		t.Fatalf("unexpected output: %#v", output)
	}
}
