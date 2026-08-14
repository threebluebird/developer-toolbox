package httpclient

import (
	"context"
	"developer-toolbox/backend/models"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHTTPClientRequestAndResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Query().Get("page") != "1" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL)
		}
		if r.Header.Get("Authorization") != "Bearer secret" {
			t.Errorf("missing auth")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Add("Set-Cookie", "session=ok")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()
	tool := NewHTTPClientTool()
	out := tool.Execute(models.NewToolContext(context.Background()), models.ToolInput{Payload: map[string]any{"url": server.URL, "method": "POST", "query": map[string]any{"page": "1"}, "bodyType": "json", "body": `{"hello":"world"}`, "authType": "bearer", "token": "secret"}})
	if !out.Success {
		t.Fatalf("unexpected: %+v", out)
	}
	data := out.Data.(Response)
	if data.StatusCode != 200 || !strings.Contains(data.Body, "true") || len(data.Cookies) != 1 {
		t.Fatalf("unexpected response: %+v", data)
	}
}

func TestHTTPStatusesAndRedirect(t *testing.T) {
	for _, status := range []int{400, 401, 404, 500} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(status) }))
		out := NewHTTPClientTool().Execute(models.NewToolContext(context.Background()), models.ToolInput{Payload: map[string]any{"url": server.URL}})
		server.Close()
		if !out.Success || out.Data.(Response).StatusCode != status {
			t.Fatalf("status %d: %+v", status, out)
		}
	}
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/start" {
			http.Redirect(w, r, "/done", http.StatusFound)
			return
		}
		_, _ = w.Write([]byte("done"))
	}))
	defer server.Close()
	out := NewHTTPClientTool().Execute(models.NewToolContext(context.Background()), models.ToolInput{Payload: map[string]any{"url": server.URL + "/start"}})
	if !out.Success || out.Data.(Response).Body != "done" {
		t.Fatalf("redirect failed: %+v", out)
	}
}

func TestHTTPLargeResponseIsTruncated(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("x", int(MaxResponseBytes)+10)))
	}))
	defer server.Close()
	out := NewHTTPClientTool().Execute(models.NewToolContext(context.Background()), models.ToolInput{Payload: map[string]any{"url": server.URL}})
	if !out.Success {
		t.Fatalf("unexpected: %+v", out)
	}
	response := out.Data.(Response)
	if !response.Truncated || response.Size != MaxResponseBytes {
		t.Fatalf("unexpected: %+v", response)
	}
}

func TestHTTPTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { time.Sleep(50 * time.Millisecond); w.WriteHeader(200) }))
	defer server.Close()
	out := NewHTTPClientToolWithClient(&http.Client{Timeout: 5 * time.Millisecond}).Execute(models.NewToolContext(context.Background()), models.ToolInput{Payload: map[string]any{"url": server.URL}})
	if out.Success {
		t.Fatal("expected timeout failure")
	}
}
func TestHTTPClientRejectsInvalidJSONBody(t *testing.T) {
	out := NewHTTPClientTool().Execute(models.NewToolContext(context.Background()), models.ToolInput{Payload: map[string]any{"url": "https://example.test", "method": "POST", "bodyType": "json", "body": "{"}})
	if out.Success {
		t.Fatal("expected failure")
	}
}
func TestResponseJSONContract(t *testing.T) {
	data, _ := json.Marshal(Response{StatusCode: 200, TimingMS: 3})
	if !strings.Contains(string(data), `"timingMs":3`) {
		t.Fatal(string(data))
	}
}
