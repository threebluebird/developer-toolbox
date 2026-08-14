package colortool

import (
	"developer-toolbox/backend/models"
	"testing"
)

func TestColorConversions(t *testing.T) {
	out := NewColorTool().Execute(models.ToolInput{Payload: map[string]any{"input": "#1890FF"}})
	if !out.Success {
		t.Fatalf("unexpected: %+v", out)
	}
	v := out.Data.(map[string]any)
	if v["rgb"] != "rgb(24, 144, 255)" || v["hex"] != "#1890FF" {
		t.Fatalf("unexpected: %#v", v)
	}
}
func TestHSLInput(t *testing.T) {
	out := NewColorTool().Execute(models.ToolInput{Payload: map[string]any{"input": "hsl(0, 100%, 50%)"}})
	if !out.Success || out.Data.(map[string]any)["hex"] != "#FF0000" {
		t.Fatalf("unexpected: %+v", out)
	}
}
