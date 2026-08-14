package base64tool

import (
	"developer-toolbox/backend/models"
	"strings"
	"testing"
)

func BenchmarkEncode100KB(b *testing.B) {
	tool := NewBase64Tool()
	input := models.ToolInput{Payload: map[string]any{"action": "encode", "input": strings.Repeat("x", 100*1024)}}
	b.SetBytes(100 * 1024)
	b.ReportAllocs()
	for b.Loop() {
		if out := tool.Execute(input); !out.Success {
			b.Fatal(out.Error)
		}
	}
}
