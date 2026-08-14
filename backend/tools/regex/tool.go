package regex

import (
	"developer-toolbox/backend/models"
	"fmt"
	"regexp"
	"strings"
)

type RegexTool struct{}

func NewRegexTool() *RegexTool {
	return &RegexTool{}
}

func (t *RegexTool) Info() models.Tool {
	return models.Tool{
		ID:          "regex",
		Name:        "Regex Tester",
		Description: "Test regular expressions and replacements.",
		Category:    "developer",
		Icon:        "regex",
		Version:     "0.1.0",
		Keywords:    []string{"regex", "regexp", "regular expression", "match", "replace"},
	}
}

func (t *RegexTool) Execute(input models.ToolInput) models.ToolOutput {
	payload := input.Payload
	if payload == nil {
		return errorResponse("payload required")
	}

	pattern, ok := payload["pattern"].(string)
	if !ok || strings.TrimSpace(pattern) == "" {
		return errorResponse("pattern must be a string")
	}

	subject, ok := payload["input"].(string)
	if !ok {
		return errorResponse("input must be a string")
	}

	flags, _ := payload["flags"].(string)
	regex, err := compileRegex(pattern, flags)
	if err != nil {
		return errorResponse(err.Error())
	}

	action, _ := payload["action"].(string)
	if action == "" {
		action = "match"
	}

	switch action {
	case "match":
		return matchResponse(regex, subject)
	case "replace":
		replacement, _ := payload["replacement"].(string)
		return replaceResponse(regex, subject, replacement)
	default:
		return errorResponse("unknown action")
	}
}

func compileRegex(pattern, flags string) (*regexp.Regexp, error) {
	// Go regexp 没有单独的 global 模式；matchResponse 使用 FindAll 已天然全局匹配。
	// i/m/s 会转换为 RE2 内联修饰符，不支持的 JS flag 明确报错而不是静默忽略。
	modifiers := strings.Builder{}
	seen := make(map[rune]bool)
	for _, flag := range flags {
		if seen[flag] {
			continue
		}
		seen[flag] = true
		switch flag {
		case 'g':
			// Matching is global because FindAllStringSubmatch is used.
		case 'i', 'm', 's':
			modifiers.WriteRune(flag)
		default:
			return nil, fmt.Errorf("unsupported regex flag: %c", flag)
		}
	}
	if modifiers.Len() > 0 {
		pattern = "(?" + modifiers.String() + ")" + pattern
	}
	return regexp.Compile(pattern)
}

func matchResponse(regex *regexp.Regexp, subject string) models.ToolOutput {
	// 每个二维数组的第一个元素是完整匹配，之后依次为捕获组。
	matches := regex.FindAllStringSubmatch(subject, -1)
	if matches == nil {
		return models.ToolOutput{Success: true, Data: map[string]any{"matches": []any{}, "count": 0}}
	}

	results := make([]any, 0, len(matches))
	for _, match := range matches {
		capture := make([]string, len(match))
		copy(capture, match)
		results = append(results, capture)
	}

	return models.ToolOutput{Success: true, Data: map[string]any{"matches": results, "count": len(results)}}
}

func replaceResponse(regex *regexp.Regexp, subject, replacement string) models.ToolOutput {
	result := regex.ReplaceAllString(subject, replacement)
	return models.ToolOutput{Success: true, Data: result}
}

func errorResponse(message string) models.ToolOutput {
	return models.Failure("INVALID_REGEX", message)
}
