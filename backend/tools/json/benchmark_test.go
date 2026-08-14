package jsontool

import (
	"strings"
	"testing"
)

func BenchmarkFormatJSON100KB(b *testing.B) {
	raw := `{"items":["` + strings.Repeat("x", 100*1024) + `"]}`
	b.SetBytes(int64(len(raw)))
	b.ReportAllocs()
	for b.Loop() {
		if _, err := FormatJSON(raw, 2); err != nil {
			b.Fatal(err)
		}
	}
}
