package pinyintool

import (
	"developer-toolbox/backend/models"
	"testing"
)

func TestPinyin(t *testing.T) {
	tool := NewPinyinTool()
	full := tool.Execute(models.ToolInput{Payload: map[string]any{"input": "开发工具箱", "mode": "full", "separator": " "}})
	if full.Data != "kai fa gong ju xiang" {
		t.Fatalf("unexpected: %#v", full)
	}
	initial := tool.Execute(models.ToolInput{Payload: map[string]any{"input": "开发工具箱", "mode": "initial"}})
	if initial.Data != "kfgjx" {
		t.Fatalf("unexpected: %#v", initial)
	}
}
