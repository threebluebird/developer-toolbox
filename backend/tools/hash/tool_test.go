package hashtool

import (
	"os"
	"path/filepath"
	"testing"

	"developer-toolbox/backend/models"
)

func TestHashTextAndFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "input.txt")
	if err := os.WriteFile(path, []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}
	tool := NewHashTool()
	text := tool.Execute(models.ToolInput{Payload: map[string]any{"input": "hello", "algorithm": "sha256"}})
	file := tool.Execute(models.ToolInput{Payload: map[string]any{"filePath": path, "algorithm": "sha256"}})
	if !text.Success || !file.Success || text.Data != file.Data {
		t.Fatalf("hash mismatch: %#v %#v", text, file)
	}
}
