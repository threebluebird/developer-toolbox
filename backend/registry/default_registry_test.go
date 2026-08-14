package registry

import (
	"testing"

	"developer-toolbox/backend/models"
)

type dummyTool struct {
	id string
}

func TestCreateDefaultRegistryIncludesExtendedTools(t *testing.T) {
	r := CreateDefaultRegistry()
	for _, id := range []string{"amount-cn", "tcp", "udp", "serial", "pinyin", "database-docs"} {
		if _, ok := r.Get(id); !ok {
			t.Errorf("extended tool %q was not registered", id)
		}
	}
}

func TestCreateDefaultRegistryIncludesPhase5Tools(t *testing.T) {
	r := CreateDefaultRegistry()
	for _, id := range []string{"cron", "color", "code-generator", "git"} {
		if _, ok := r.Get(id); !ok {
			t.Errorf("phase 5 tool %q was not registered", id)
		}
	}
}

func TestCreateDefaultRegistryIncludesPhase4Tools(t *testing.T) {
	r := CreateDefaultRegistry()
	for _, id := range []string{"http", "dns", "ping", "port", "cidr", "url-parser"} {
		if _, ok := r.Get(id); !ok {
			t.Errorf("phase 4 tool %q was not registered", id)
		}
	}
}

func TestCreateDefaultRegistryIncludesPhase3Tools(t *testing.T) {
	r := CreateDefaultRegistry()
	for _, id := range []string{"xml", "csv", "toml", "text", "sql", "markdown", "html"} {
		if _, ok := r.Get(id); !ok {
			t.Errorf("phase 3 tool %q was not registered", id)
		}
	}
}

func (d dummyTool) Info() models.Tool {
	return models.Tool{
		ID:          d.id,
		Name:        "Dummy",
		Description: "A dummy tool for testing",
		Category:    "test",
		Icon:        "dummy",
		Version:     "0.1.0",
	}
}

func (d dummyTool) Execute(input models.ToolInput) models.ToolOutput {
	return models.ToolOutput{
		Success: true,
		Data:    input.Payload,
	}
}

func TestDefaultRegistry_RegisterAndGet(t *testing.T) {
	r := NewDefaultRegistry()
	dummy := dummyTool{id: "dummy"}

	if err := r.Register(dummy); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}

	if _, ok := r.Get("dummy"); !ok {
		t.Fatal("expected registered tool to be found")
	}
}

func TestDefaultRegistry_RegisterDuplicate(t *testing.T) {
	r := NewDefaultRegistry()
	tool := dummyTool{id: "dup"}

	if err := r.Register(tool); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}

	if err := r.Register(tool); err != ErrToolAlreadyRegistered {
		t.Fatalf("expected ErrToolAlreadyRegistered, got %v", err)
	}
}

func TestDefaultRegistry_List(t *testing.T) {
	r := NewDefaultRegistry()
	first := dummyTool{id: "first"}
	second := dummyTool{id: "second"}

	if err := r.Register(first); err != nil {
		t.Fatal(err)
	}
	if err := r.Register(second); err != nil {
		t.Fatal(err)
	}

	list := r.List()
	if len(list) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(list))
	}
	if list[0].ID != "first" || list[1].ID != "second" {
		t.Fatalf("unexpected tool order: %v", list)
	}
}
