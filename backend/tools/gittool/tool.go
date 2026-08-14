package gittool

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"developer-toolbox/backend/models"
)

type GitTool struct{}

func NewGitTool() *GitTool { return &GitTool{} }
func (t *GitTool) Info() models.Tool {
	return models.Tool{ID: "git", Name: "Git Tools", Description: "Generate gitignore files, commands, and parse Git URLs.", Category: "developer", Icon: "git", Version: "0.5.0", Keywords: []string{"git", "gitignore", "command", "repository", "github", "url"}}
}
func (t *GitTool) Execute(input models.ToolInput) models.ToolOutput {
	action := stringValue(input.Payload["action"], "gitignore")
	switch action {
	case "gitignore":
		return gitignore(input.Payload["templates"])
	case "command":
		name := stringValue(input.Payload["command"], "")
		command, ok := commands[name]
		if !ok {
			return bad("unsupported command")
		}
		return good(map[string]any{"name": name, "command": command})
	case "parseUrl":
		return parseGitURL(stringValue(input.Payload["input"], ""))
	default:
		return bad("unknown action")
	}
}

var ignores = map[string][]string{"go": {"*.exe", "*.test", "*.out", "vendor/"}, "node": {"node_modules/", "npm-debug.log*", "dist/", ".env"}, "python": {"__pycache__/", "*.py[cod]", ".venv/", "dist/"}, "java": {"*.class", "target/", "*.jar"}, "cpp": {"*.o", "*.obj", "*.exe", "build/"}, "rust": {"target/", "Cargo.lock"}, "react": {"node_modules/", "build/", ".env.local"}, "vue": {"node_modules/", "dist/", ".env.local"}, "vscode": {".vscode/"}, "intellij": {".idea/", "*.iml"}, "windows": {"Thumbs.db", "Desktop.ini"}, "macos": {".DS_Store"}, "linux": {"*~", ".directory"}}
var commands = map[string]string{"undo-last-commit": "git reset --soft HEAD~1", "discard-last-commit": "git reset --hard HEAD~1", "amend-commit": "git commit --amend", "create-branch": "git switch -c <branch>", "delete-branch": "git branch -d <branch>", "stash": "git stash push -m \"work in progress\"", "unstash": "git stash pop", "uncommit-file": "git restore --staged <file>", "show-history": "git log --oneline --graph --decorate --all"}

func gitignore(value any) models.ToolOutput {
	items := []string{}
	switch x := value.(type) {
	case []any:
		for _, v := range x {
			items = append(items, strings.ToLower(fmt.Sprint(v)))
		}
	case []string:
		items = x
	case string:
		items = strings.Split(strings.ToLower(x), ",")
	}
	sort.Strings(items)
	seen := map[string]bool{}
	var b strings.Builder
	for _, name := range items {
		patterns, ok := ignores[strings.TrimSpace(name)]
		if !ok {
			continue
		}
		fmt.Fprintf(&b, "# %s\n", name)
		for _, pattern := range patterns {
			if !seen[pattern] {
				seen[pattern] = true
				b.WriteString(pattern + "\n")
			}
		}
		b.WriteByte('\n')
	}
	return good(strings.TrimSpace(b.String()))
}
func parseGitURL(raw string) models.ToolOutput {
	raw = strings.TrimSpace(raw)
	protocol := "ssh"
	host, repoPath := "", ""
	if strings.HasPrefix(raw, "git@") {
		parts := strings.SplitN(strings.TrimPrefix(raw, "git@"), ":", 2)
		if len(parts) != 2 {
			return bad("invalid Git URL")
		}
		host, repoPath = parts[0], parts[1]
	} else {
		parsed, err := url.Parse(raw)
		if err != nil || parsed.Host == "" {
			return bad("invalid Git URL")
		}
		protocol = parsed.Scheme
		host = parsed.Hostname()
		repoPath = strings.TrimPrefix(parsed.Path, "/")
	}
	repoPath = strings.TrimSuffix(repoPath, ".git")
	parts := strings.Split(repoPath, "/")
	if len(parts) < 2 {
		return bad("Git URL must include owner and repository")
	}
	return good(map[string]any{"host": host, "owner": strings.Join(parts[:len(parts)-1], "/"), "repository": parts[len(parts)-1], "protocol": protocol})
}
func stringValue(v any, f string) string {
	if s, ok := v.(string); ok && s != "" {
		return s
	}
	return f
}
func good(v any) models.ToolOutput   { return models.ToolOutput{Success: true, Data: v} }
func bad(s string) models.ToolOutput { return models.Failure("GIT_TOOL_ERROR", s) }
