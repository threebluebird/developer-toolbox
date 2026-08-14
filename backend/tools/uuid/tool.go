package uuidtool

import (
	"developer-toolbox/backend/models"
	"math"

	"github.com/google/uuid"
)

type UUIDTool struct{}

func NewUUIDTool() *UUIDTool {
	return &UUIDTool{}
}

func (t *UUIDTool) Info() models.Tool {
	return models.Tool{
		ID:          "uuid",
		Name:        "UUID",
		Description: "Generate UUID v4 values in batch or single mode.",
		Category:    "generator",
		Icon:        "uuid",
		Version:     "0.1.0",
		Keywords:    []string{"uuid", "guid", "generate", "random"},
	}
}

func (t *UUIDTool) Execute(input models.ToolInput) models.ToolOutput {
	payload := input.Payload
	if payload == nil {
		return errorResponse("payload required")
	}

	count := 1
	if rawCount, ok := payload["count"]; ok {
		// Wails/JSON 数字通常以 float64 传入，asInt 会拒绝小数并统一转换。
		v, valid := asInt(rawCount)
		if !valid || v < 1 || v > 100 {
			return errorResponse("count must be between 1 and 100")
		}
		count = v
	}

	if count == 1 {
		return models.ToolOutput{Success: true, Data: uuid.NewString()}
	}

	values := make([]string, 0, count)
	for i := 0; i < count; i++ {
		values = append(values, uuid.NewString())
	}

	return models.ToolOutput{Success: true, Data: values}
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
		if math.Trunc(float64(v)) != float64(v) {
			return 0, false
		}
		return int(v), true
	case float64:
		if math.Trunc(v) != v {
			return 0, false
		}
		return int(v), true
	default:
		return 0, false
	}
}

func errorResponse(message string) models.ToolOutput {
	return models.Failure("INVALID_UUID", message)
}
