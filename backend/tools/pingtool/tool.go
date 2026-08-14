package pingtool

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"developer-toolbox/backend/models"
)

type runner func(context.Context, string, ...string) ([]byte, error)
type PingTool struct{ run runner }

func NewPingTool() *PingTool {
	return &PingTool{run: func(ctx context.Context, name string, args ...string) ([]byte, error) {
		return exec.CommandContext(ctx, name, args...).CombinedOutput()
	}}
}
func (t *PingTool) Info() models.Tool {
	return models.Tool{ID: "ping", Name: "Ping", Description: "Measure host reachability and latency.", Category: "network", Icon: "ping", Version: "0.5.0", Capabilities: models.ToolCapabilities{Network: true}, Keywords: []string{"ping", "latency", "packet", "ttl", "loss"}}
}
func (t *PingTool) Execute(tc models.ToolContext, input models.ToolInput) models.ToolOutput {
	host := strings.TrimSpace(value(input.Payload["host"], value(input.Payload["input"], "")))
	if err := validateHost(host); err != nil {
		return bad(err.Error())
	}
	count := clamp(intValue(input.Payload["count"], 4), 1, 10)
	timeout := clamp(intValue(input.Payload["timeoutMs"], 3000), 500, 10000)
	interval := clamp(intValue(input.Payload["intervalMs"], 1000), 200, 5000)
	name, args := command(host, count, timeout, interval)
	ctx, cancel := context.WithTimeout(tc.Context, time.Duration(count*interval+timeout+2000)*time.Millisecond)
	defer cancel()
	output, err := t.run(ctx, name, args...)
	packets, summary := parse(string(output), count)
	if err != nil && len(packets) == 0 {
		summary["error"] = strings.TrimSpace(string(output))
		if ctx.Err() != nil {
			summary["error"] = ctx.Err().Error()
		}
	}
	return models.ToolOutput{Success: true, Data: map[string]any{"host": host, "packets": packets, "summary": summary}}
}
func command(host string, count, timeout, interval int) (string, []string) {
	if runtime.GOOS == "windows" {
		return "ping", []string{"-n", strconv.Itoa(count), "-w", strconv.Itoa(timeout), host}
	}
	return "ping", []string{"-c", strconv.Itoa(count), "-W", strconv.Itoa(max(1, timeout/1000)), "-i", fmt.Sprintf("%.1f", float64(interval)/1000), host}
}

var latencyPattern = regexp.MustCompile(`(?i)time[=<]\s*([0-9.]+)\s*ms`)
var ttlPattern = regexp.MustCompile(`(?i)ttl[= ]([0-9]+)`)

func parse(output string, sent int) ([]map[string]any, map[string]any) {
	packets := []map[string]any{}
	latencies := []float64{}
	for _, line := range strings.Split(output, "\n") {
		match := latencyPattern.FindStringSubmatch(line)
		if len(match) < 2 {
			continue
		}
		latency, _ := strconv.ParseFloat(match[1], 64)
		ttl := 0
		if m := ttlPattern.FindStringSubmatch(line); len(m) > 1 {
			ttl, _ = strconv.Atoi(m[1])
		}
		latencies = append(latencies, latency)
		packets = append(packets, map[string]any{"packet": len(packets) + 1, "ttl": ttl, "latencyMs": latency})
	}
	received := len(latencies)
	summary := map[string]any{"sent": sent, "received": received, "packetLoss": float64(sent-received) * 100 / float64(sent)}
	if received > 0 {
		min, maxV, total := latencies[0], latencies[0], 0.0
		for _, v := range latencies {
			if v < min {
				min = v
			}
			if v > maxV {
				maxV = v
			}
			total += v
		}
		summary["minMs"] = min
		summary["maxMs"] = maxV
		summary["averageMs"] = total / float64(received)
	}
	return packets, summary
}
func validateHost(s string) error {
	if s == "" || len(s) > 253 {
		return fmt.Errorf("invalid host")
	}
	for _, r := range s {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '.' || r == ':' || r == '-') {
			return fmt.Errorf("invalid host")
		}
	}
	return nil
}
func value(v any, f string) string {
	if s, ok := v.(string); ok && s != "" {
		return s
	}
	return f
}
func intValue(v any, f int) int {
	switch x := v.(type) {
	case int:
		return x
	case float64:
		return int(x)
	case string:
		n, e := strconv.Atoi(x)
		if e == nil {
			return n
		}
	}
	return f
}
func clamp(v, a, b int) int {
	if v < a {
		return a
	}
	if v > b {
		return b
	}
	return v
}
func bad(s string) models.ToolOutput { return models.Failure("PING_ERROR", s) }
