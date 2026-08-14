// Package udptool 实现单次请求/响应式 UDP 调试客户端。
package udptool

import (
	"developer-toolbox/backend/models"
	"encoding/hex"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

const maxUDPDatagram = 65507

type UDPTool struct{}

func NewUDPTool() *UDPTool { return &UDPTool{} }
func (t *UDPTool) Info() models.Tool {
	return models.Tool{ID: "udp", Name: "UDP Client", Description: "Send a UDP datagram and optionally wait for a response.", Category: "network", Icon: "udp", Version: "0.5.0", Capabilities: models.ToolCapabilities{Network: true}, Keywords: []string{"udp", "datagram", "socket", "client", "hex"}}
}

// Execute 使用 DialUDP 固定远端地址，避免接收其他来源的数据报。
func (t *UDPTool) Execute(tc models.ToolContext, input models.ToolInput) models.ToolOutput {
	host := strings.TrimSpace(stringValue(input.Payload["host"], ""))
	port := intValue(input.Payload["port"], 0)
	if err := validateEndpoint(host, port); err != nil {
		return fail(err.Error())
	}
	payload, err := decode(stringValue(input.Payload["input"], ""), stringValue(input.Payload["encoding"], "text"))
	if err != nil {
		return fail("十六进制数据无效")
	}
	if len(payload) > maxUDPDatagram {
		return fail("UDP 数据报不能超过 65507 字节")
	}
	timeout := time.Duration(clamp(intValue(input.Payload["timeoutMs"], 3000), 100, 30000)) * time.Millisecond
	address, err := net.ResolveUDPAddr("udp", net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		return fail(err.Error())
	}
	dialer := net.Dialer{}
	raw, err := dialer.DialContext(tc.Context, "udp", address.String())
	if err != nil {
		return fail(err.Error())
	}
	conn := raw.(*net.UDPConn)
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(timeout))
	start := time.Now()
	if _, err = conn.Write(payload); err != nil {
		return fail(err.Error())
	}
	buffer := make([]byte, maxUDPDatagram)
	count, readErr := conn.Read(buffer)
	timedOut := false
	if readErr != nil {
		if ne, ok := readErr.(net.Error); ok && ne.Timeout() {
			timedOut = true
			count = 0
		} else {
			return fail(readErr.Error())
		}
	}
	response := buffer[:count]
	if stringValue(input.Payload["responseEncoding"], "text") == "hex" {
		return models.ToolOutput{Success: true, Data: map[string]any{"sentBytes": len(payload), "receivedBytes": count, "response": strings.ToUpper(hex.EncodeToString(response)), "latencyMs": time.Since(start).Milliseconds(), "timedOut": timedOut}}
	}
	return models.ToolOutput{Success: true, Data: map[string]any{"sentBytes": len(payload), "receivedBytes": count, "response": string(response), "latencyMs": time.Since(start).Milliseconds(), "timedOut": timedOut}}
}
func decode(raw, encoding string) ([]byte, error) {
	if encoding == "hex" {
		return hex.DecodeString(strings.NewReplacer(" ", "", "\n", "", "\r", "").Replace(raw))
	}
	return []byte(raw), nil
}
func fail(s string) models.ToolOutput { return models.Failure("UDP_ERROR", s) }

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
