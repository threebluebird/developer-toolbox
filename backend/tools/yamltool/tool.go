package yamltool

import (
	"developer-toolbox/backend/models"
	"encoding/json"

	"gopkg.in/yaml.v3"
)

type YAMLTool struct{}

func NewYAMLTool() *YAMLTool {
	return &YAMLTool{}
}

func (t *YAMLTool) Info() models.Tool {
	return models.Tool{
		ID:          "yaml",
		Name:        "JSON ↔ YAML",
		Description: "Convert between JSON and YAML formats.",
		Category:    "encoding",
		Icon:        "yaml",
		Version:     "0.1.0",
		Keywords:    []string{"yaml", "json", "convert", "format"},
	}
}

func (t *YAMLTool) Execute(input models.ToolInput) models.ToolOutput {
	payload := input.Payload
	if payload == nil {
		return errorResponse("payload required")
	}

	raw, ok := payload["input"].(string)
	if !ok {
		return errorResponse("input must be a string")
	}

	action, _ := payload["action"].(string)
	if action == "" {
		action = "jsonToYaml"
	}

	switch action {
	case "jsonToYaml":
		return t.jsonToYaml(raw)
	case "yamlToJson":
		return t.yamlToJson(raw)
	case "validate":
		return t.validate(raw)
	default:
		return errorResponse("unknown action")
	}
}

func (t *YAMLTool) jsonToYaml(raw string) models.ToolOutput {
	var data any
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return errorResponse("invalid json")
	}

	yamlBytes, err := yaml.Marshal(data)
	if err != nil {
		return errorResponse("failed to convert to yaml")
	}

	return models.ToolOutput{Success: true, Data: string(yamlBytes)}
}

func (t *YAMLTool) yamlToJson(raw string) models.ToolOutput {
	var data any
	if err := yaml.Unmarshal([]byte(raw), &data); err != nil {
		return errorResponse("invalid yaml")
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return errorResponse("failed to convert to json")
	}

	return models.ToolOutput{Success: true, Data: string(jsonBytes)}
}

func (t *YAMLTool) validate(raw string) models.ToolOutput {
	var data any
	if err := yaml.Unmarshal([]byte(raw), &data); err != nil {
		return errorResponse("invalid yaml")
	}
	return models.ToolOutput{Success: true, Data: "valid"}
}

func errorResponse(message string) models.ToolOutput {
	return models.Failure("INVALID_YAML", message)
}
