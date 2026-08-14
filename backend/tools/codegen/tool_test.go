package codegen

import (
	"developer-toolbox/backend/models"
	"strings"
	"testing"
)

func TestGenerateGoAndTypeScript(t *testing.T) {
	tool := NewCodeGeneratorTool()
	raw := `{"name":"Tom","age":30,"active":true}`
	goOut := tool.Execute(models.ToolInput{Payload: map[string]any{"input": raw, "name": "User", "language": "go"}})
	if !goOut.Success || !strings.Contains(goOut.Data.(string), "Age int64 `json:\"age\"`") {
		t.Fatalf("unexpected: %+v", goOut)
	}
	ts := tool.Execute(models.ToolInput{Payload: map[string]any{"input": raw, "name": "User", "language": "typescript"}})
	if !ts.Success || !strings.Contains(ts.Data.(string), "age: number;") {
		t.Fatalf("unexpected: %+v", ts)
	}
}
func TestNestedObject(t *testing.T) {
	out := NewCodeGeneratorTool().Execute(models.ToolInput{Payload: map[string]any{"input": `{"profile":{"city":"Shanghai"}}`, "name": "User", "language": "go"}})
	if !out.Success || !strings.Contains(out.Data.(string), "type Profile struct") {
		t.Fatalf("unexpected: %+v", out)
	}
}
