package jsontool

import (
	"encoding/json"
	"testing"

	"developer-toolbox/backend/models"
)

func TestJSONTool_Format(t *testing.T) {
	tool := NewJSONTool()
	input := `{"foo": [1,2], "bar": {"baz":true}}`

	output := tool.Execute(mapInput(input, "format", 2))
	if !output.Success {
		t.Fatalf("expected success, got error: %s", output.Error)
	}

	formatted, ok := output.Data.(string)
	if !ok {
		t.Fatal("expected formatted result to be a string")
	}
	if formatted == input {
		t.Fatal("expected formatted JSON to be different from raw input")
	}
}

func TestJSONToolTreePreservesNumbers(t *testing.T) {
	output := NewJSONTool().Execute(mapInput(`{"large":9007199254740993}`, "tree", 2))
	if !output.Success {
		t.Fatal(output.Error)
	}
	data := output.Data.(map[string]any)
	if data["large"].(json.Number).String() != "9007199254740993" {
		t.Fatalf("number lost precision: %#v", data)
	}
}

func TestJSONTool_Minify(t *testing.T) {
	tool := NewJSONTool()
	input := `{
  "foo": [1, 2],
  "bar": {"baz": true}
}`

	output := tool.Execute(mapInput(input, "minify", 0))
	if !output.Success {
		t.Fatalf("expected success, got error: %s", output.Error)
	}

	result, ok := output.Data.(string)
	if !ok {
		t.Fatal("expected minified result to be a string")
	}
	if result == "" {
		t.Fatal("expected minified JSON to be non-empty")
	}
	if result[0] != '{' || result[len(result)-1] != '}' {
		t.Fatalf("unexpected minified JSON structure: %s", result)
	}
}

func TestJSONTool_Validate(t *testing.T) {
	tool := NewJSONTool()
	input := `{"foo": 1}`

	output := tool.Execute(mapInput(input, "validate", 0))
	if !output.Success {
		t.Fatalf("expected success, got error: %s", output.Error)
	}
	if output.Data != "valid" {
		t.Fatalf("unexpected validate response: %v", output.Data)
	}
}

func TestJSONTool_InvalidInput(t *testing.T) {
	tool := NewJSONTool()
	input := `{"foo":` // invalid JSON

	output := tool.Execute(mapInput(input, "format", 2))
	if output.Success {
		t.Fatal("expected failure for invalid JSON")
	}
}

func mapInput(input, action string, indent int) models.ToolInput {
	return models.ToolInput{
		ToolID: "json",
		Payload: map[string]any{
			"input":  input,
			"action": action,
			"indent": indent,
		},
	}
}
