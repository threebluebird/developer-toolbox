package timestamptool

import (
	"testing"
	"time"

	"developer-toolbox/backend/models"
)

func TestLocalDateUsesLocalWallClock(t *testing.T) {
	originalLocal := time.Local
	time.Local = time.FixedZone("UTC+8", 8*60*60)
	t.Cleanup(func() { time.Local = originalLocal })

	parsed, err := parseTimestampInput("1970-01-01 08:00:00", "local")
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Unix() != 0 {
		t.Fatalf("expected local wall clock to resolve to epoch, got %d", parsed.Unix())
	}
}

func TestNegativeMillisecondTimestamp(t *testing.T) {
	output := NewTimestampTool().Execute(models.ToolInput{Payload: map[string]any{
		"input": "-2208988800000", "action": "toDate", "zone": "utc",
	}})
	if !output.Success {
		t.Fatalf("unexpected failure: %s", output.Error)
	}
	data := output.Data.(map[string]any)
	if data["unix"] != int64(-2208988800) {
		t.Fatalf("milliseconds were not detected: %#v", data)
	}
}
