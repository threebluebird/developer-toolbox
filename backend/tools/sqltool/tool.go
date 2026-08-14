package sqltool

import (
	"developer-toolbox/backend/models"
	"regexp"
	"strings"
)

type SQLTool struct{}

func NewSQLTool() *SQLTool { return &SQLTool{} }
func (t *SQLTool) Info() models.Tool {
	return models.Tool{ID: "sql", Name: "SQL Formatter", Description: "Format, beautify, and minify SQL.", Category: "data", Icon: "sql", Version: "0.5.0", Keywords: []string{"sql", "format", "beautify", "minify", "select", "query"}}
}

var whitespace = regexp.MustCompile(`\s+`)
var keywords = []string{"select", "from", "where", "insert into", "values", "update", "set", "delete from", "left join", "right join", "inner join", "outer join", "join", "on", "group by", "order by", "having", "limit", "offset", "union", "and", "or"}

func (t *SQLTool) Execute(input models.ToolInput) models.ToolOutput {
	raw, ok := input.Payload["input"].(string)
	if !ok {
		return bad("input must be a string")
	}
	action, _ := input.Payload["action"].(string)
	if strings.TrimSpace(raw) == "" {
		return good("")
	}
	if action == "minify" {
		return good(strings.TrimSpace(whitespace.ReplaceAllString(raw, " ")))
	}
	if action != "format" && action != "beautify" {
		return bad("unknown action")
	}
	indent := 2
	if v, ok := input.Payload["indent"].(float64); ok && int(v) == 4 {
		indent = 4
	}
	if v, ok := input.Payload["indent"].(int); ok && v == 4 {
		indent = 4
	}
	style, _ := input.Payload["keyword"].(string)
	comma, _ := input.Payload["comma"].(string)
	sql := strings.TrimSpace(whitespace.ReplaceAllString(raw, " "))
	for _, kw := range keywords {
		re := regexp.MustCompile(`(?i)\b` + strings.ReplaceAll(kw, " ", `\s+`) + `\b`)
		replacement := strings.ToUpper(kw)
		if style == "lower" {
			replacement = strings.ToLower(kw)
		}
		sql = re.ReplaceAllString(sql, replacement)
	}
	breaks := []string{" FROM ", " WHERE ", " LEFT JOIN ", " RIGHT JOIN ", " INNER JOIN ", " OUTER JOIN ", " JOIN ", " GROUP BY ", " ORDER BY ", " HAVING ", " LIMIT ", " UNION ", " VALUES ", " SET "}
	for _, token := range breaks {
		actual := token
		if style == "lower" {
			actual = strings.ToLower(token)
		}
		sql = strings.ReplaceAll(sql, actual, "\n"+strings.Repeat(" ", indent)+strings.TrimSpace(actual)+" ")
	}
	if comma == "leading" {
		sql = regexp.MustCompile(`,\s*`).ReplaceAllString(sql, "\n"+strings.Repeat(" ", indent)+", ")
	}
	return good(strings.TrimSpace(sql))
}
func good(v any) models.ToolOutput   { return models.ToolOutput{Success: true, Data: v} }
func bad(s string) models.ToolOutput { return models.Failure("INVALID_SQL", s) }
