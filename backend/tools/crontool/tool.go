package crontool

import (
	"fmt"
	"strconv"
	"strings"

	"developer-toolbox/backend/models"
)

type CronTool struct{}

func NewCronTool() *CronTool { return &CronTool{} }
func (t *CronTool) Info() models.Tool {
	return models.Tool{ID: "cron", Name: "Cron Tool", Description: "Parse and generate five-field cron expressions.", Category: "developer", Icon: "cron", Version: "0.5.0", Keywords: []string{"cron", "schedule", "parser", "generator", "job"}}
}
func (t *CronTool) Execute(input models.ToolInput) models.ToolOutput {
	action := stringValue(input.Payload["action"], "parse")
	if action == "generate" {
		return generate(input.Payload)
	}
	raw := strings.TrimSpace(stringValue(input.Payload["input"], ""))
	fields := strings.Fields(raw)
	if len(fields) != 5 {
		return failure("cron expression must contain 5 fields")
	}
	ranges := [][2]int{{0, 59}, {0, 23}, {1, 31}, {1, 12}, {0, 7}}
	for i, field := range fields {
		if err := validateField(field, ranges[i][0], ranges[i][1]); err != nil {
			return failure(fmt.Sprintf("field %d: %v", i+1, err))
		}
	}
	return success(map[string]any{"expression": raw, "description": describe(fields), "minute": fields[0], "hour": fields[1], "day": fields[2], "month": fields[3], "weekday": fields[4]})
}
func generate(p map[string]any) models.ToolOutput {
	minute := stringValue(p["minute"], "*")
	hour := stringValue(p["hour"], "*")
	day := stringValue(p["day"], "*")
	month := stringValue(p["month"], "*")
	weekday := stringValue(p["weekday"], "*")
	fields := []string{minute, hour, day, month, weekday}
	ranges := [][2]int{{0, 59}, {0, 23}, {1, 31}, {1, 12}, {0, 7}}
	for i, field := range fields {
		if err := validateField(field, ranges[i][0], ranges[i][1]); err != nil {
			return failure(fmt.Sprintf("field %d: %v", i+1, err))
		}
	}
	expression := strings.Join(fields, " ")
	return success(map[string]any{"expression": expression, "description": describe(fields)})
}
func validateField(field string, min, max int) error {
	if field == "*" {
		return nil
	}
	for _, part := range strings.Split(field, ",") {
		base := part
		if strings.Contains(part, "/") {
			pieces := strings.Split(part, "/")
			if len(pieces) != 2 {
				return fmt.Errorf("invalid step")
			}
			step, err := strconv.Atoi(pieces[1])
			if err != nil || step < 1 {
				return fmt.Errorf("invalid step")
			}
			base = pieces[0]
		}
		if base == "*" {
			continue
		}
		ends := strings.Split(base, "-")
		if len(ends) > 2 {
			return fmt.Errorf("invalid range")
		}
		for _, value := range ends {
			number, err := strconv.Atoi(value)
			if err != nil || number < min || number > max {
				return fmt.Errorf("value must be %d-%d", min, max)
			}
		}
	}
	return nil
}
func describe(f []string) string {
	if strings.HasPrefix(f[0], "*/") && f[1] == "*" && f[2] == "*" && f[3] == "*" && f[4] == "*" {
		return "Every " + strings.TrimPrefix(f[0], "*/") + " minutes"
	}
	if f[0] == "0" && f[2] == "*" && f[3] == "*" && f[4] == "1-5" {
		return fmt.Sprintf("Every weekday at %s:00", pad(f[1]))
	}
	if f[2] == "*" && f[3] == "*" && f[4] == "*" {
		return fmt.Sprintf("Every day at %s:%s", pad(f[1]), pad(f[0]))
	}
	return fmt.Sprintf("At minute %s, hour %s, day %s, month %s, weekday %s", f[0], f[1], f[2], f[3], f[4])
}
func pad(v string) string {
	if n, err := strconv.Atoi(v); err == nil {
		return fmt.Sprintf("%02d", n)
	}
	return v
}
func stringValue(v any, f string) string {
	if s, ok := v.(string); ok && s != "" {
		return s
	}
	return f
}
func success(v any) models.ToolOutput    { return models.ToolOutput{Success: true, Data: v} }
func failure(s string) models.ToolOutput { return models.Failure("CRON_ERROR", s) }
