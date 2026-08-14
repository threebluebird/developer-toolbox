package jwttool

import (
	"developer-toolbox/backend/models"
	"encoding/base64"
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

type JWTTool struct{}

func NewJWTTool() *JWTTool {
	return &JWTTool{}
}

func (t *JWTTool) Info() models.Tool {
	return models.Tool{
		ID:          "jwt",
		Name:        "JWT Decoder",
		Description: "Decode JWT header and payload without verifying signatures.",
		Category:    "security",
		Icon:        "jwt",
		Version:     "0.1.0",
		Keywords:    []string{"jwt", "token", "decode", "claims", "header", "payload"},
	}
}

func (t *JWTTool) Execute(input models.ToolInput) models.ToolOutput {
	payload := input.Payload
	if payload == nil {
		return errorResponse("payload required")
	}

	raw, ok := payload["input"].(string)
	if !ok || strings.TrimSpace(raw) == "" {
		return errorResponse("input must be a JWT string")
	}

	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return errorResponse("invalid jwt format")
	}

	header, err := decodeSegment(parts[0])
	if err != nil {
		return errorResponse("invalid jwt header")
	}

	claims, err := decodeSegment(parts[1])
	if err != nil {
		return errorResponse("invalid jwt payload")
	}

	var payloadData map[string]any
	if err := json.Unmarshal([]byte(claims), &payloadData); err != nil {
		return errorResponse("invalid jwt payload json")
	}

	expired, expMessage := parseExpiration(payloadData)
	iatMessage := parseClaimTime(payloadData, "iat")

	return models.ToolOutput{Success: true, Data: map[string]any{
		"header":    jsonOrString(header),
		"payload":   payloadData,
		"signature": parts[2],
		"expired":   expired,
		"expInfo":   expMessage,
		"iatInfo":   iatMessage,
	}}
}

func decodeSegment(segment string) (string, error) {
	// JWT 通常省略 Base64URL padding；若输入包含 padding，再回退到标准 URL 编码。
	decoded, err := base64.RawURLEncoding.DecodeString(segment)
	if err != nil {
		decoded, err = base64.URLEncoding.DecodeString(segment)
		if err != nil {
			return "", err
		}
	}
	return string(decoded), nil
}

func parseExpiration(payload map[string]any) (bool, string) {
	// JSON 数字默认解码为 float64，同时兼容内部调用和部分外部输入的 int64/string。
	value, ok := payload["exp"]
	if !ok {
		return false, "no exp claim"
	}

	switch v := value.(type) {
	case float64:
		exp := time.Unix(int64(v), 0)
		return time.Now().After(exp), exp.Format(time.RFC3339)
	case int64:
		exp := time.Unix(v, 0)
		return time.Now().After(exp), exp.Format(time.RFC3339)
	case string:
		if ts, err := strconv.ParseInt(v, 10, 64); err == nil {
			exp := time.Unix(ts, 0)
			return time.Now().After(exp), exp.Format(time.RFC3339)
		}
	}

	return false, "unsupported exp value"
}

func jsonOrString(raw string) any {
	var value any
	if err := json.Unmarshal([]byte(raw), &value); err == nil {
		return value
	}
	return raw
}

func errorResponse(message string) models.ToolOutput {
	return models.Failure("INVALID_JWT", message)
}

func parseClaimTime(payload map[string]any, name string) string {
	// exp、iat 等 NumericDate 均按 Unix 秒解释，并统一输出 UTC ISO 8601。
	value, ok := payload[name]
	if !ok {
		return "no " + name + " claim"
	}
	var timestamp int64
	switch v := value.(type) {
	case float64:
		timestamp = int64(v)
	case int64:
		timestamp = v
	case string:
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return "unsupported " + name + " value"
		}
		timestamp = parsed
	default:
		return "unsupported " + name + " value"
	}
	return time.Unix(timestamp, 0).UTC().Format(time.RFC3339)
}
