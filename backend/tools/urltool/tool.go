package urltool

import (
	"developer-toolbox/backend/models"
	"net/url"
)

type URLTool struct{}

func NewURLTool() *URLTool {
	return &URLTool{}
}

func (t *URLTool) Info() models.Tool {
	return models.Tool{
		ID:          "url",
		Name:        "URL Encoder",
		Description: "Encode and decode URLs and query strings.",
		Category:    "encoding",
		Icon:        "url",
		Version:     "0.1.0",
		Keywords:    []string{"url", "uri", "encode", "decode", "query"},
	}
}

func (t *URLTool) Execute(input models.ToolInput) models.ToolOutput {
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
		return models.ToolOutput{Success: true, Data: url.QueryEscape(raw)}
	case "decode":
		decoded, err := url.QueryUnescape(raw)
		if err != nil {
			return errorResponse("invalid url encoding")
		}
		return models.ToolOutput{Success: true, Data: decoded}
	case "encodeQuery":
		values, err := parseQuery(raw)
		if err != nil {
			return errorResponse("invalid query string")
		}
		return models.ToolOutput{Success: true, Data: values.Encode()}
	case "decodeQuery":
		values, err := parseQuery(raw)
		if err != nil {
			return errorResponse("invalid query string")
		}
		return models.ToolOutput{Success: true, Data: queryValuesToMap(values)}
	default:
		return errorResponse("unknown action")
	}
}

func parseQuery(raw string) (url.Values, error) {
	return url.ParseQuery(raw)
}

func queryValuesToMap(values url.Values) map[string]any {
	result := make(map[string]any, len(values))
	for key, list := range values {
		if len(list) == 1 {
			result[key] = list[0]
		} else {
			result[key] = list
		}
	}
	return result
}

func errorResponse(message string) models.ToolOutput {
	return models.Failure("INVALID_URL", message)
}
