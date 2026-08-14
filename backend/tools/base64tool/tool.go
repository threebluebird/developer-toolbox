package base64tool

import (
	"developer-toolbox/backend/models"
	"encoding/base64"
)

type Base64Tool struct{}

func NewBase64Tool() *Base64Tool {
	return &Base64Tool{}
}

func (t *Base64Tool) Info() models.Tool {
	return models.Tool{
		ID:          "base64",
		Name:        "Base64",
		Description: "Encode and decode Base64 text.",
		Category:    "encoding",
		Icon:        "base64",
		Version:     "0.1.0",
		Keywords:    []string{"base64", "encode", "decode"},
	}
}

func (t *Base64Tool) Execute(input models.ToolInput) models.ToolOutput {
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
		action = "encode"
	}

	switch action {
	case "encode":
		return encodeResponse(raw)
	case "decode":
		return decodeResponse(raw)
	default:
		return errorResponse("unknown action")
	}
}

func encodeResponse(raw string) models.ToolOutput {
	return models.ToolOutput{Success: true, Data: base64.StdEncoding.EncodeToString([]byte(raw))}
}

func decodeResponse(raw string) models.ToolOutput {
	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return errorResponse("invalid base64 input")
	}
	return models.ToolOutput{Success: true, Data: string(decoded)}
}

func errorResponse(message string) models.ToolOutput {
	return models.Failure("INVALID_BASE64", message)
}
