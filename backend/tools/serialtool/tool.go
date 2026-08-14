// Package serialtool 提供串口枚举以及一次性收发能力。
package serialtool

import (
	"developer-toolbox/backend/models"
	"encoding/hex"
	"fmt"
	"go.bug.st/serial"
	"io"
	"strconv"
	"strings"
	"time"
)

const maxSerialRead = 1024 * 1024

type SerialTool struct{}

func NewSerialTool() *SerialTool { return &SerialTool{} }
func (t *SerialTool) Info() models.Tool {
	return models.Tool{ID: "serial", Name: "Serial Port", Description: "List serial ports and send or receive text/hex data.", Category: "network", Icon: "serial", Version: "0.5.0", Capabilities: models.ToolCapabilities{Stateful: true}, Keywords: []string{"serial", "uart", "com", "串口", "baud", "rs232"}}
}

// Execute 的 list 动作不打开设备；transfer 使用白名单参数并确保端口及时关闭。
func (t *SerialTool) Execute(input models.ToolInput) models.ToolOutput {
	action := stringValue(input.Payload["action"], "list")
	if action == "list" {
		ports, err := serial.GetPortsList()
		if err != nil {
			return fail(err.Error())
		}
		return models.ToolOutput{Success: true, Data: ports}
	}
	if action != "transfer" {
		return fail("不支持的串口动作")
	}
	name := strings.TrimSpace(stringValue(input.Payload["portName"], ""))
	if name == "" || len(name) > 128 || strings.ContainsAny(name, "\r\n\x00") {
		return fail("串口名称无效")
	}
	baud := intValue(input.Payload["baudRate"], 9600)
	if baud < 300 || baud > 4000000 {
		return fail("波特率必须为 300-4000000")
	}
	dataBits := intValue(input.Payload["dataBits"], 8)
	if dataBits < 5 || dataBits > 8 {
		return fail("数据位必须为 5-8")
	}
	mode := &serial.Mode{BaudRate: baud, DataBits: dataBits, Parity: parity(stringValue(input.Payload["parity"], "none")), StopBits: stopBits(stringValue(input.Payload["stopBits"], "1"))}
	port, err := serial.Open(name, mode)
	if err != nil {
		return fail(err.Error())
	}
	defer port.Close()
	timeout := time.Duration(clamp(intValue(input.Payload["timeoutMs"], 1000), 100, 30000)) * time.Millisecond
	_ = port.SetReadTimeout(timeout)
	payload, err := serialPayload(stringValue(input.Payload["input"], ""), stringValue(input.Payload["encoding"], "text"))
	if err != nil {
		return fail("十六进制数据无效")
	}
	if len(payload) > maxSerialRead {
		return fail("发送数据不能超过 1 MiB")
	}
	if len(payload) > 0 {
		if _, err = port.Write(payload); err != nil {
			return fail(err.Error())
		}
	}
	buffer := make([]byte, maxSerialRead)
	count, readErr := port.Read(buffer)
	if readErr != nil && readErr != io.EOF {
		return fail(readErr.Error())
	}
	data := buffer[:count]
	response := string(data)
	if stringValue(input.Payload["responseEncoding"], "text") == "hex" {
		response = strings.ToUpper(hex.EncodeToString(data))
	}
	return models.ToolOutput{Success: true, Data: map[string]any{"port": name, "sentBytes": len(payload), "receivedBytes": count, "response": response}}
}
func parity(v string) serial.Parity {
	switch v {
	case "odd":
		return serial.OddParity
	case "even":
		return serial.EvenParity
	case "mark":
		return serial.MarkParity
	case "space":
		return serial.SpaceParity
	default:
		return serial.NoParity
	}
}
func stopBits(v string) serial.StopBits {
	if v == "2" {
		return serial.TwoStopBits
	}
	if v == "1.5" {
		return serial.OnePointFiveStopBits
	}
	return serial.OneStopBit
}
func serialPayload(v, e string) ([]byte, error) {
	if e == "hex" {
		return hex.DecodeString(strings.NewReplacer(" ", "", "\n", "", "\r", "").Replace(v))
	}
	return []byte(v), nil
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
func fail(s string) models.ToolOutput {
	return models.Failure("SERIAL_ERROR", fmt.Sprintf("串口操作失败：%s", s))
}
