package cidrtool

import (
	"developer-toolbox/backend/models"
	"testing"
)

func TestIPv4CIDR(t *testing.T) {
	out := NewCIDRTool().Execute(models.ToolInput{Payload: map[string]any{"input": "192.168.1.42/24"}})
	if !out.Success {
		t.Fatalf("unexpected: %+v", out)
	}
	v := out.Data.(map[string]any)
	if v["broadcast"] != "192.168.1.255" || v["firstUsable"] != "192.168.1.1" || v["usableHosts"] != uint64(254) {
		t.Fatalf("unexpected: %#v", v)
	}
}
