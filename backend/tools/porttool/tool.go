package porttool

import (
	"developer-toolbox/backend/models"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

type PortTool struct {
	dial func(models.ToolContext, string, time.Duration) (net.Conn, error)
}

func NewPortTool() *PortTool {
	return &PortTool{dial: func(tc models.ToolContext, address string, timeout time.Duration) (net.Conn, error) {
		d := net.Dialer{Timeout: timeout}
		return d.DialContext(tc.Context, "tcp", address)
	}}
}
func (t *PortTool) Info() models.Tool {
	return models.Tool{ID: "port", Name: "Port Checker", Description: "Check whether a TCP port is open.", Category: "network", Icon: "port", Version: "0.5.0", Capabilities: models.ToolCapabilities{Network: true}, Keywords: []string{"port", "tcp", "open", "closed", "timeout"}}
}
func (t *PortTool) Execute(tc models.ToolContext, input models.ToolInput) models.ToolOutput {
	host := strings.TrimSpace(value(input.Payload["host"], value(input.Payload["input"], "")))
	port := intValue(input.Payload["port"], 0)
	if net.ParseIP(host) == nil {
		if err := validateHostname(host); err != nil {
			return bad(err.Error())
		}
	}
	if port < 1 || port > 65535 {
		return bad("port must be between 1 and 65535")
	}
	timeout := time.Duration(clamp(intValue(input.Payload["timeoutMs"], 3000), 100, 30000)) * time.Millisecond
	start := time.Now()
	conn, err := t.dial(tc, net.JoinHostPort(host, strconv.Itoa(port)), timeout)
	elapsed := time.Since(start).Milliseconds()
	status := "OPEN"
	detail := ""
	if err != nil {
		status = "CLOSED"
		detail = err.Error()
		if ne, ok := err.(net.Error); ok && ne.Timeout() {
			status = "TIMEOUT"
		}
	} else {
		_ = conn.Close()
	}
	return models.ToolOutput{Success: true, Data: map[string]any{"host": host, "port": port, "status": status, "latencyMs": elapsed, "detail": detail}}
}
func validateHostname(s string) error {
	if s == "" || len(s) > 253 {
		return fmt.Errorf("invalid host")
	}
	for _, part := range strings.Split(s, ".") {
		if part == "" || len(part) > 63 || strings.HasPrefix(part, "-") || strings.HasSuffix(part, "-") {
			return fmt.Errorf("invalid host")
		}
		for _, r := range part {
			if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-') {
				return fmt.Errorf("invalid host")
			}
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
func bad(s string) models.ToolOutput { return models.Failure("PORT_ERROR", s) }
