package jsontool

import (
	"developer-toolbox/backend/models"
	"encoding/json"
	"strings"
)

// JSONTool is a simple JSON formatter and validator tool.
type JSONTool struct{}

// NewJSONTool creates a new JSONTool instance.
func NewJSONTool() *JSONTool {
	return &JSONTool{}
}

// Info returns metadata about the JSON tool.
func (t *JSONTool) Info() models.Tool {
	return models.Tool{
		ID:          "json",
		Name:        "JSON Formatter",
		Description: "Format, minify, and validate JSON.",
		Category:    "encoding",
		Icon:        "json",
		Version:     "0.1.0",
		Keywords:    []string{"json", "format", "pretty", "beautify", "minify", "validate", "tree"},
	}
}

// Execute runs the JSON tool with the provided input payload.
func (t *JSONTool) Execute(input models.ToolInput) models.ToolOutput {
	payload := input.Payload
	if payload == nil {
		return errorResponse("payload required")
	}

	raw, ok := payload["input"].(string)
	if !ok {
		return errorResponse("input must be a string")
	}

	action, ok := payload["action"].(string)
	if !ok || action == "" {
		action = "format"
	}

	indent := 2
	if rawIndent, ok := payload["indent"]; ok {
		if v, ok := asInt(rawIndent); ok && v > 0 {
			indent = v
		}
	}

	switch action {
	case "format":
		result, err := FormatJSON(raw, indent)
		if err != nil {
			return errorResponse(err.Error())
		}
		return models.ToolOutput{Success: true, Data: result}
	case "minify":
		result, err := MinifyJSON(raw)
		if err != nil {
			return errorResponse(err.Error())
		}
		return models.ToolOutput{Success: true, Data: result}
	case "validate":
		err := ValidateJSON(raw)
		if err != nil {
			return errorResponse(err.Error())
		}
		return models.ToolOutput{Success: true, Data: "valid"}
	case "tree":
		result, err := ParseJSONTree(raw)
		if err != nil {
			return errorResponse(err.Error())
		}
		return models.ToolOutput{Success: true, Data: result}
	default:
		return errorResponse("unknown action")
	}
}

func errorResponse(message string) models.ToolOutput {
	return models.Failure("INVALID_JSON", message)
}

func ParseJSONTree(raw string) (any, error) {
	var value any
	decoder := json.NewDecoder(strings.NewReader(raw))
	// UseNumber 防止较大的 JSON 整数先转成 float64 而丢失精度。
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	return value, nil
}

func asInt(value any) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int8:
		return int(v), true
	case int16:
		return int(v), true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	case float32:
		return int(v), true
	case float64:
		return int(v), true
	default:
		return 0, false
	}
}
