package dbdocs

import (
	"developer-toolbox/backend/models"
	"strings"
	"testing"
)

func TestDatabaseDocs(t *testing.T) {
	out := NewDatabaseDocsTool().Execute(models.ToolInput{Payload: map[string]any{"database": "postgresql", "input": "upsert"}})
	if !out.Success {
		t.Fatalf("unexpected: %+v", out)
	}
	value := out.Data.(map[string]any)
	if value["count"] != 1 || value["entries"].([]Entry)[0].Title != "UPSERT" {
		t.Fatalf("unexpected: %#v", value)
	}
}

func TestDatabaseDocsIncludesCommonFunctionCategories(t *testing.T) {
	databases := []string{"postgresql", "mysql", "oracle", "sqlserver"}
	categories := []string{"字符串", "数组", "JSON", "日期时间"}
	tool := NewDatabaseDocsTool()

	for _, database := range databases {
		for _, category := range categories {
			t.Run(database+"/"+category, func(t *testing.T) {
				out := tool.Execute(models.ToolInput{Payload: map[string]any{"database": database, "input": category}})
				if !out.Success {
					t.Fatalf("unexpected failure: %+v", out)
				}
				entries := out.Data.(map[string]any)["entries"].([]Entry)
				found := false
				for _, entry := range entries {
					if strings.Contains(entry.Title, "函数") && entry.Syntax != "" && entry.Example != "" {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("missing %s function documentation for %s: %#v", category, database, entries)
				}
			})
		}
	}
}
