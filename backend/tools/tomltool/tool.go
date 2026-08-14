package tomltool

import (
	"developer-toolbox/backend/models"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type TOMLTool struct{}

func NewTOMLTool() *TOMLTool { return &TOMLTool{} }
func (t *TOMLTool) Info() models.Tool {
	return models.Tool{ID: "toml", Name: "TOML Toolkit", Description: "Format, validate, and convert TOML and JSON.", Category: "data", Icon: "toml", Version: "0.5.0", Keywords: []string{"toml", "json", "config", "format", "validate"}}
}
func (t *TOMLTool) Execute(input models.ToolInput) models.ToolOutput {
	raw, ok := input.Payload["input"].(string)
	if !ok {
		return bad("input must be a string")
	}
	action, _ := input.Payload["action"].(string)
	switch action {
	case "format", "validate", "tomlToJson":
		v, err := parse(raw)
		if err != nil {
			return bad(err.Error())
		}
		if action == "validate" {
			return good("valid")
		}
		if action == "format" {
			return good(encode(v))
		}
		b, _ := json.MarshalIndent(v, "", "  ")
		return good(string(b))
	case "jsonToToml":
		var v map[string]any
		if err := json.Unmarshal([]byte(raw), &v); err != nil {
			return bad(err.Error())
		}
		return good(encode(v))
	default:
		return bad("unknown action")
	}
}
func parse(raw string) (map[string]any, error) {
	root := map[string]any{}
	current := root
	for lineNo, line := range strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n") {
		line = strings.TrimSpace(strings.SplitN(line, "#", 2)[0])
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			current = root
			for _, part := range strings.Split(line[1:len(line)-1], ".") {
				next, ok := current[part].(map[string]any)
				if !ok {
					next = map[string]any{}
					current[part] = next
				}
				current = next
			}
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("line %d: expected key = value", lineNo+1)
		}
		key := strings.TrimSpace(parts[0])
		if key == "" {
			return nil, fmt.Errorf("line %d: empty key", lineNo+1)
		}
		value, err := scalar(strings.TrimSpace(parts[1]))
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNo+1, err)
		}
		current[key] = value
	}
	return root, nil
}
func scalar(s string) (any, error) {
	if strings.HasPrefix(s, "\"") {
		v, err := strconv.Unquote(s)
		return v, err
	}
	if s == "true" || s == "false" {
		return s == "true", nil
	}
	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		parts := strings.Split(strings.TrimSpace(s[1:len(s)-1]), ",")
		result := []any{}
		if len(parts) == 1 && strings.TrimSpace(parts[0]) == "" {
			return result, nil
		}
		for _, p := range parts {
			v, err := scalar(strings.TrimSpace(p))
			if err != nil {
				return nil, err
			}
			result = append(result, v)
		}
		return result, nil
	}
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i, nil
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f, nil
	}
	return nil, fmt.Errorf("unsupported value %q", s)
}
func encode(root map[string]any) string {
	var b strings.Builder
	writeTable(&b, "", root)
	return strings.TrimSpace(b.String())
}
func writeTable(b *strings.Builder, path string, v map[string]any) {
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if _, ok := v[k].(map[string]any); !ok {
			fmt.Fprintf(b, "%s = %s\n", k, format(v[k]))
		}
	}
	for _, k := range keys {
		if child, ok := v[k].(map[string]any); ok {
			section := k
			if path != "" {
				section = path + "." + k
			}
			fmt.Fprintf(b, "\n[%s]\n", section)
			writeTable(b, section, child)
		}
	}
}
func format(v any) string {
	switch x := v.(type) {
	case string:
		return strconv.Quote(x)
	case bool:
		return strconv.FormatBool(x)
	case []any:
		parts := make([]string, len(x))
		for i, item := range x {
			parts[i] = format(item)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	default:
		return fmt.Sprint(x)
	}
}
func good(v any) models.ToolOutput   { return models.ToolOutput{Success: true, Data: v} }
func bad(s string) models.ToolOutput { return models.Failure("INVALID_TOML", s) }
