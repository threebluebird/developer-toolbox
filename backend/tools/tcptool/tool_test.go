package tcptool

import (
	"context"
	"developer-toolbox/backend/models"
	"net"
	"testing"
)

func TestTCPClientLoopback(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	go func() {
		conn, _ := listener.Accept()
		if conn != nil {
			defer conn.Close()
			buffer := make([]byte, 16)
			n, _ := conn.Read(buffer)
			_, _ = conn.Write([]byte("echo:" + string(buffer[:n])))
		}
	}()
	address := listener.Addr().(*net.TCPAddr)
	out := NewTCPTool().Execute(models.NewToolContext(context.Background()), models.ToolInput{Payload: map[string]any{"host": "127.0.0.1", "port": address.Port, "input": "hi", "timeoutMs": 1000}})
	if !out.Success || out.Data.(map[string]any)["response"] != "echo:hi" {
		t.Fatalf("unexpected: %+v", out)
	}
}
