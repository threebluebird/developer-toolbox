package gittool

import (
	"developer-toolbox/backend/models"
	"strings"
	"testing"
)

func TestGitignoreAndCommand(t *testing.T) {
	tool := NewGitTool()
	out := tool.Execute(models.ToolInput{Payload: map[string]any{"action": "gitignore", "templates": []any{"go", "node"}}})
	if !out.Success || !strings.Contains(out.Data.(string), "node_modules/") || !strings.Contains(out.Data.(string), "*.exe") {
		t.Fatalf("unexpected: %+v", out)
	}
	command := tool.Execute(models.ToolInput{Payload: map[string]any{"action": "command", "command": "undo-last-commit"}})
	if command.Data.(map[string]any)["command"] != "git reset --soft HEAD~1" {
		t.Fatalf("unexpected: %+v", command)
	}
}
func TestParseGitURLs(t *testing.T) {
	tool := NewGitTool()
	for _, raw := range []string{"https://github.com/user/project.git", "git@github.com:user/project.git"} {
		out := tool.Execute(models.ToolInput{Payload: map[string]any{"action": "parseUrl", "input": raw}})
		if !out.Success {
			t.Fatalf("unexpected: %+v", out)
		}
		v := out.Data.(map[string]any)
		if v["host"] != "github.com" || v["owner"] != "user" || v["repository"] != "project" {
			t.Fatalf("unexpected: %#v", v)
		}
	}
}
