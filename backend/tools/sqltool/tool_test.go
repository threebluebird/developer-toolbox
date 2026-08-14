package sqltool

import (
	"developer-toolbox/backend/models"
	"strings"
	"testing"
)

func TestFormatAndMinify(t *testing.T) {
	tool := NewSQLTool()
	out := tool.Execute(models.ToolInput{Payload: map[string]any{"action": "format", "input": "select * from users where id = 1", "keyword": "upper"}})
	if !out.Success || !strings.Contains(out.Data.(string), "\n  FROM ") {
		t.Fatalf("unexpected: %+v", out)
	}
	min := tool.Execute(models.ToolInput{Payload: map[string]any{"action": "minify", "input": "SELECT *\n FROM users"}})
	if min.Data != "SELECT * FROM users" {
		t.Fatalf("got %q", min.Data)
	}
}
