package timestamptool

import (
	"developer-toolbox/backend/models"
	"errors"
	"strconv"
	"strings"
	"time"
)

type TimestampTool struct{}

func NewTimestampTool() *TimestampTool {
	return &TimestampTool{}
}

func (t *TimestampTool) Info() models.Tool {
	return models.Tool{
		ID:          "timestamp",
		Name:        "Timestamp",
		Description: "Convert between timestamps and dates.",
		Category:    "time",
		Icon:        "timestamp",
		Version:     "0.1.0",
		Keywords:    []string{"timestamp", "unix", "epoch", "date", "time"},
	}
}

func (t *TimestampTool) Execute(input models.ToolInput) models.ToolOutput {
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
		action = "toTimestamp"
	}

	zone, _ := payload["zone"].(string)
	if zone == "" {
		zone = "local"
	}

	switch action {
	case "toTimestamp":
		return t.dateToTimestamp(raw, zone)
	case "toDate":
		return t.timestampToDate(raw, zone)
	case "now":
		return t.now(zone)
	default:
		return errorResponse("unknown action")
	}
}

func (t *TimestampTool) dateToTimestamp(raw, zone string) models.ToolOutput {
	parsed, err := parseTimestampInput(raw, zone)
	if err != nil {
		return errorResponse(err.Error())
	}

	return models.ToolOutput{
		Success: true,
		Data: map[string]any{
			"unix":     parsed.Unix(),
			"unixMs":   parsed.UnixMilli(),
			"iso":      parsed.Format(time.RFC3339),
			"local":    parsed.Local().Format(time.RFC3339),
			"utc":      parsed.UTC().Format(time.RFC3339),
			"timezone": zone,
		},
	}
}

func (t *TimestampTool) timestampToDate(raw, zone string) models.ToolOutput {
	v, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil {
		return errorResponse("invalid timestamp")
	}

	var ts time.Time
	if isMillisecondTimestamp(v) {
		ts = time.UnixMilli(v)
	} else {
		ts = time.Unix(v, 0)
	}

	if zone == "utc" {
		ts = ts.UTC()
	}

	return models.ToolOutput{
		Success: true,
		Data: map[string]any{
			"unix":     ts.Unix(),
			"unixMs":   ts.UnixMilli(),
			"iso":      ts.Format(time.RFC3339),
			"local":    ts.Local().Format(time.RFC3339),
			"utc":      ts.UTC().Format(time.RFC3339),
			"timezone": zone,
		},
	}
}

func isMillisecondTimestamp(value int64) bool {
	// 以绝对值 1e11 为界，同时正确识别 1970 年之前的负数毫秒时间戳。
	return value >= 100_000_000_000 || value <= -100_000_000_000
}

func (t *TimestampTool) now(zone string) models.ToolOutput {
	ts := time.Now()
	if zone == "utc" {
		ts = ts.UTC()
	}

	return models.ToolOutput{
		Success: true,
		Data: map[string]any{
			"unix":     ts.Unix(),
			"unixMs":   ts.UnixMilli(),
			"iso":      ts.Format(time.RFC3339),
			"local":    ts.Local().Format(time.RFC3339),
			"utc":      ts.UTC().Format(time.RFC3339),
			"timezone": zone,
		},
	}
}

func parseTimestampInput(raw, zone string) (time.Time, error) {
	// 解析顺序：纯数字时间戳 → RFC3339 → 常用无时区日期格式。
	text := strings.TrimSpace(raw)
	if text == "" {
		return time.Time{}, errors.New("input cannot be empty")
	}

	if ts, err := strconv.ParseInt(text, 10, 64); err == nil {
		if isMillisecondTimestamp(ts) {
			return time.UnixMilli(ts), nil
		}
		return time.Unix(ts, 0), nil
	}

	formats := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, layout := range formats {
		var parsed time.Time
		var err error
		// 无时区的普通日期按本地墙上时间解析，避免先按 UTC 再转本地造成时差。
		if zone == "utc" || strings.Contains(text, "T") {
			parsed, err = time.Parse(layout, text)
		} else {
			parsed, err = time.ParseInLocation(layout, text, time.Local)
		}
		if err == nil {
			if zone == "utc" {
				return parsed.UTC(), nil
			}
			return parsed, nil
		}
	}

	return time.Time{}, errors.New("unsupported date format")
}

func errorResponse(message string) models.ToolOutput {
	return models.Failure("INVALID_TIMESTAMP", message)
}
