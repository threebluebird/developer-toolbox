package services

import (
	"testing"

	"developer-toolbox/backend/models"
	"developer-toolbox/backend/registry"
)

type searchTestTool struct{ info models.Tool }

func (t searchTestTool) Info() models.Tool { return t.info }
func (t searchTestTool) Execute(input models.ToolInput) models.ToolOutput {
	return models.ToolOutput{Success: true, Data: input.Payload}
}

func TestSearchServiceMatchesMetadata(t *testing.T) {
	service := NewSearchService(NewToolService(registry.CreateDefaultRegistry()))
	for _, query := range []string{"json", "security", "convert between"} {
		if results := service.Search(query); len(results) == 0 {
			t.Fatalf("expected results for %q", query)
		}
	}
	if results := service.Search("definitely-missing"); len(results) != 0 {
		t.Fatalf("expected no results, got %d", len(results))
	}
}

func TestSearchToolsByMetadataKeyword(t *testing.T) {
	registry := registry.NewDefaultRegistry()
	if err := registry.Register(searchTestTool{info: models.Tool{ID: "json", Name: "JSON", Keywords: []string{"beautify"}}}); err != nil {
		t.Fatal(err)
	}
	results := NewSearchService(NewToolService(registry)).Search("beautify")
	if len(results) != 1 || results[0].ID != "json" {
		t.Fatalf("expected keyword match, got %+v", results)
	}
}
