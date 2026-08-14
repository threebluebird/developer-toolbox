package yamltool

import (
	"testing"

	"developer-toolbox/backend/models"
)

func TestYAMLRoundTripAndValidation(t *testing.T) {
	tool := NewYAMLTool()
	yamlOutput := tool.Execute(models.ToolInput{Payload: map[string]any{"input": `{"name":"toolbox"}`, "action": "jsonToYaml"}})
	if !yamlOutput.Success {
		t.Fatal(yamlOutput.Error)
	}
	jsonOutput := tool.Execute(models.ToolInput{Payload: map[string]any{"input": yamlOutput.Data, "action": "yamlToJson"}})
	if !jsonOutput.Success || jsonOutput.Data != `{"name":"toolbox"}` {
		t.Fatalf("unexpected output: %#v", jsonOutput)
	}
	invalid := tool.Execute(models.ToolInput{Payload: map[string]any{"input": "a: [", "action": "validate"}})
	if invalid.Success {
		t.Fatal("expected invalid YAML to fail")
	}
}
