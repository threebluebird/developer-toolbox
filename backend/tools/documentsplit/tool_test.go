package documentsplit

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"developer-toolbox/backend/models"
	"github.com/pdfcpu/pdfcpu/pkg/api"
)

func TestSplitPDFIntoOddAndEvenFiles(t *testing.T) {
	dir := t.TempDir()
	jsonFile := filepath.Join(dir, "pages.json")
	inputFile := filepath.Join(dir, "sample.pdf")
	content := []byte(`{"pages":{"1":{"content":{}},"2":{"content":{}},"3":{"content":{}},"4":{"content":{}},"5":{"content":{}}}}`)
	if err := os.WriteFile(jsonFile, content, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := api.CreateFile("", jsonFile, inputFile, nil); err != nil {
		t.Fatal(err)
	}

	result, err := Split(context.Background(), inputFile, dir)
	if err != nil {
		t.Fatal(err)
	}
	if result.TotalPages != 5 || result.OddPages != 3 || result.EvenPages != 2 {
		t.Fatalf("unexpected result: %#v", result)
	}
	if count, err := api.PageCountFile(result.OddFile); err != nil || count != 3 {
		t.Fatalf("odd output: count=%d err=%v", count, err)
	}
	if count, err := api.PageCountFile(result.EvenFile); err != nil || count != 2 {
		t.Fatalf("even output: count=%d err=%v", count, err)
	}
}

func TestSplitRejectsUnsupportedInput(t *testing.T) {
	file := filepath.Join(t.TempDir(), "notes.txt")
	if err := os.WriteFile(file, []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Split(context.Background(), file, ""); err == nil {
		t.Fatal("expected unsupported extension error")
	}
}

func TestSplitSinglePageDoesNotOverwriteExistingOutput(t *testing.T) {
	dir := t.TempDir()
	jsonFile := filepath.Join(dir, "single.json")
	inputFile := filepath.Join(dir, "single.pdf")
	if err := os.WriteFile(jsonFile, []byte(`{"pages":{"1":{"content":{}}}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := api.CreateFile("", jsonFile, inputFile, nil); err != nil {
		t.Fatal(err)
	}
	existing := filepath.Join(dir, "single_odd.pdf")
	if err := os.WriteFile(existing, []byte("keep me"), 0o600); err != nil {
		t.Fatal(err)
	}

	result, err := Split(context.Background(), inputFile, dir)
	if err != nil {
		t.Fatal(err)
	}
	if result.OddFile != filepath.Join(dir, "single_odd_2.pdf") || result.EvenFile != "" {
		t.Fatalf("unexpected output paths: %#v", result)
	}
	if raw, err := os.ReadFile(existing); err != nil || string(raw) != "keep me" {
		t.Fatalf("existing output was changed: %q err=%v", string(raw), err)
	}
}

func TestToolMetadataAndMissingInput(t *testing.T) {
	tool := NewTool()
	if !tool.Info().Capabilities.FileInput || !tool.Info().Capabilities.FileOutput {
		t.Fatal("document split capabilities are incomplete")
	}
	result := tool.Execute(models.NewToolContext(context.Background()), models.ToolInput{})
	if result.Success || result.Error == nil || result.Error.Code != "DOCUMENT_SPLIT_FAILED" {
		t.Fatalf("unexpected missing input result: %#v", result)
	}
}
