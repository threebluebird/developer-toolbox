package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestStandardLoggerFiltersAndLabels(t *testing.T) {
	var output bytes.Buffer
	log := New(&output, InfoLevel)
	log.Debug("hidden")
	log.Info("processed %d items", 3)
	if strings.Contains(output.String(), "hidden") {
		t.Fatal("debug output should be filtered")
	}
	if !strings.Contains(output.String(), "[INFO] processed 3 items") {
		t.Fatalf("unexpected output: %q", output.String())
	}
}

func TestParseLevelDefaultsToInfo(t *testing.T) {
	if ParseLevel("debug") != DebugLevel || ParseLevel("unknown") != InfoLevel {
		t.Fatal("level parsing did not return expected values")
	}
}
