package urlparser

import (
	"developer-toolbox/backend/models"
	"net/url"
	"strings"
)

type URLParserTool struct{}

func NewURLParserTool() *URLParserTool { return &URLParserTool{} }
func (t *URLParserTool) Info() models.Tool {
	return models.Tool{ID: "url-parser", Name: "URL Parser", Description: "Inspect URL components and query parameters.", Category: "network", Icon: "url-parser", Version: "0.5.0", Keywords: []string{"url", "parse", "scheme", "host", "port", "query", "fragment"}}
}
func (t *URLParserTool) Execute(input models.ToolInput) models.ToolOutput {
	raw := strings.TrimSpace(stringValue(input.Payload["input"], ""))
	parsed, err := url.ParseRequestURI(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return bad("invalid absolute URL")
	}
	query := map[string][]string{}
	for k, v := range parsed.Query() {
		query[k] = v
	}
	username := ""
	passwordSet := false
	if parsed.User != nil {
		username = parsed.User.Username()
		_, passwordSet = parsed.User.Password()
	}
	return models.ToolOutput{Success: true, Data: map[string]any{"scheme": parsed.Scheme, "host": parsed.Hostname(), "port": parsed.Port(), "path": parsed.EscapedPath(), "query": query, "rawQuery": parsed.RawQuery, "fragment": parsed.Fragment, "username": username, "passwordSet": passwordSet}}
}
func stringValue(v any, f string) string {
	if s, ok := v.(string); ok {
		return s
	}
	return f
}
func bad(s string) models.ToolOutput { return models.Failure("URL_PARSE_ERROR", s) }
