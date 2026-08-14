// Package tcptool 实现带超时和响应上限的 TCP 调试客户端。
package tcptool

import (
	"developer-toolbox/backend/models"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

const maxTCPResponse = 1024 * 1024

type TCPTool struct {
	dial func(models.ToolContext, string, time.Duration) (net.Conn, error)
}

func NewTCPTool() *TCPTool {
	return &TCPTool{dial: func(tc models.ToolContext, address string, timeout time.Duration) (net.Conn, error) {
		return (&net.Dialer{Timeout: timeout}).DialContext(tc.Context, "tcp", address)
	}}
}
func (t *TCPTool) Info() models.Tool {
	return models.Tool{ID: "tcp", Name: "TCP Client", Description: "Send text or hexadecimal data to a TCP endpoint.", Category: "network", Icon: "tcp", Version: "0.5.0", Capabilities: models.ToolCapabilities{Network: true}, Keywords: []string{"tcp", "socket", "client", "hex", "network"}}
}

// Execute 只接受明确的主机和端口，不经过 shell；读取最多 1 MiB。
func (t *TCPTool) Execute(tc models.ToolContext, input models.ToolInput) models.ToolOutput {
	host := strings.TrimSpace(stringValue(input.Payload["host"], ""))
	port := intValue(input.Payload["port"], 0)
	if err := validateEndpoint(host, port); err != nil {
		return fail(err.Error())
	}
	timeout := time.Duration(clamp(intValue(input.Payload["timeoutMs"], 3000), 100, 30000)) * time.Millisecond
	payload, err := decodePayload(stringValue(input.Payload["input"], ""), stringValue(input.Payload["encoding"], "text"))
	if err != nil {
		return fail(err.Error())
	}
	start := time.Now()
	conn, err := t.dial(tc, net.JoinHostPort(host, strconv.Itoa(port)), timeout)
	if err != nil {
		return fail(err.Error())
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(timeout))
	if _, err = conn.Write(payload); err != nil {
		return fail(err.Error())
	}
	if tcp, ok := conn.(*net.TCPConn); ok {
		_ = tcp.CloseWrite()
	}
	data, readErr := io.ReadAll(io.LimitReader(conn, maxTCPResponse+1))
	if readErr != nil {
		if ne, ok := readErr.(net.Error); !ok || !ne.Timeout() {
			return fail(readErr.Error())
		}
	}
	truncated := len(data) > maxTCPResponse
	if truncated {
		data = data[:maxTCPResponse]
	}
	return models.ToolOutput{Success: true, Data: map[string]any{"sentBytes": len(payload), "receivedBytes": len(data), "response": encodeResponse(data, stringValue(input.Payload["responseEncoding"], "text")), "latencyMs": time.Since(start).Milliseconds(), "truncated": truncated}}
}
func validateEndpoint(host string, port int) error {
	if host == "" || port < 1 || port > 65535 {
		return fmt.Errorf("主机不能为空，端口必须为 1-65535")
	}
	for _, r := range host {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '.' || r == ':' || r == '-') {
			return fmt.Errorf("主机格式无效")
		}
	}
	return nil
}
func decodePayload(raw, encoding string) ([]byte, error) {
	if encoding == "hex" {
		clean := strings.NewReplacer(" ", "", "\n", "", "\r", "").Replace(raw)
		data, err := hex.DecodeString(clean)
		if err != nil {
			return nil, fmt.Errorf("十六进制数据无效")
		}
		return data, nil
	}
	return []byte(raw), nil
}
func encodeResponse(data []byte, encoding string) string {
	if encoding == "hex" {
		return strings.ToUpper(hex.EncodeToString(data))
	}
	return string(data)
}
func stringValue(v any, f string) string {
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
func fail(s string) models.ToolOutput { return models.Failure("TCP_ERROR", s) }
