package pingtool

import (
	"context"
	"developer-toolbox/backend/models"
	"testing"
)

func TestPingParsesOutputWithoutNetwork(t *testing.T) {
	tool := NewPingTool()
	tool.run = func(context.Context, string, ...string) ([]byte, error) {
		return []byte("Reply from 127.0.0.1: bytes=32 time=2ms TTL=128\nReply from 127.0.0.1: bytes=32 time<1ms TTL=128"), nil
	}
	out := tool.Execute(models.NewToolContext(context.Background()), models.ToolInput{Payload: map[string]any{"input": "127.0.0.1", "count": 2}})
	if !out.Success {
		t.Fatalf("unexpected: %+v", out)
	}
	summary := out.Data.(map[string]any)["summary"].(map[string]any)
	if summary["received"] != 2 || summary["packetLoss"] != float64(0) {
		t.Fatalf("unexpected: %#v", summary)
	}
}
