package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"developer-toolbox/backend/models"
)

const MaxResponseBytes int64 = 10 * 1024 * 1024

type HTTPClientTool struct{ client *http.Client }

type Response struct {
	Status     string              `json:"status"`
	StatusCode int                 `json:"statusCode"`
	Headers    map[string][]string `json:"headers"`
	Body       string              `json:"body"`
	Cookies    []string            `json:"cookies"`
	TimingMS   int64               `json:"timingMs"`
	Size       int64               `json:"size"`
	Truncated  bool                `json:"truncated"`
}

func NewHTTPClientTool() *HTTPClientTool {
	return &HTTPClientTool{client: &http.Client{Timeout: 30 * time.Second, CheckRedirect: limitRedirects}}
}
func NewHTTPClientToolWithClient(client *http.Client) *HTTPClientTool {
	return &HTTPClientTool{client: client}
}
func (t *HTTPClientTool) Info() models.Tool {
	return models.Tool{ID: "http", Name: "HTTP Client", Description: "Send HTTP requests and inspect responses.", Category: "network", Icon: "http", Version: "0.5.0", Capabilities: models.ToolCapabilities{Network: true, Stateful: true, Streaming: true}, Keywords: []string{"http", "api", "request", "response", "rest", "headers"}}
}
func (t *HTTPClientTool) Execute(toolCtx models.ToolContext, input models.ToolInput) models.ToolOutput {
	p := input.Payload
	rawURL, _ := p["url"].(string)
	if strings.TrimSpace(rawURL) == "" {
		rawURL, _ = p["input"].(string)
	}
	parsed, err := url.ParseRequestURI(strings.TrimSpace(rawURL))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return fail("invalid absolute URL")
	}
	query := stringMap(p["query"])
	values := parsed.Query()
	for k, v := range query {
		values.Set(k, v)
	}
	parsed.RawQuery = values.Encode()
	method := strings.ToUpper(stringValue(p["method"], "GET"))
	if !allowedMethod(method) {
		return fail("unsupported HTTP method")
	}
	body, contentType, err := requestBody(stringValue(p["bodyType"], "none"), stringValue(p["body"], ""))
	if err != nil {
		return fail(err.Error())
	}
	request, err := http.NewRequestWithContext(toolCtx.Context, method, parsed.String(), body)
	if err != nil {
		return fail(err.Error())
	}
	for k, v := range stringMap(p["headers"]) {
		request.Header.Set(k, v)
	}
	if contentType != "" && request.Header.Get("Content-Type") == "" {
		request.Header.Set("Content-Type", contentType)
	}
	switch stringValue(p["authType"], "none") {
	case "basic":
		request.SetBasicAuth(stringValue(p["username"], ""), stringValue(p["password"], ""))
	case "bearer":
		request.Header.Set("Authorization", "Bearer "+stringValue(p["token"], ""))
	}
	start := time.Now()
	response, err := t.client.Do(request)
	elapsed := time.Since(start).Milliseconds()
	if err != nil {
		return fail(err.Error())
	}
	defer response.Body.Close()
	data, err := io.ReadAll(io.LimitReader(response.Body, MaxResponseBytes+1))
	if err != nil {
		return fail(err.Error())
	}
	truncated := int64(len(data)) > MaxResponseBytes
	if truncated {
		data = data[:MaxResponseBytes]
	}
	cookies := make([]string, 0, len(response.Cookies()))
	for _, cookie := range response.Cookies() {
		cookies = append(cookies, cookie.String())
	}
	return models.ToolOutput{Success: true, Data: Response{Status: response.Status, StatusCode: response.StatusCode, Headers: response.Header, Body: string(data), Cookies: cookies, TimingMS: elapsed, Size: int64(len(data)), Truncated: truncated}}
}
func allowedMethod(method string) bool {
	switch method {
	case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS":
		return true
	}
	return false
}
func requestBody(kind, raw string) (io.Reader, string, error) {
	switch kind {
	case "none":
		return nil, "", nil
	case "json":
		var value any
		if err := json.Unmarshal([]byte(raw), &value); err != nil {
			return nil, "", fmt.Errorf("invalid JSON body: %w", err)
		}
		return bytes.NewBufferString(raw), "application/json", nil
	case "form":
		values, err := url.ParseQuery(raw)
		if err != nil {
			return nil, "", err
		}
		return strings.NewReader(values.Encode()), "application/x-www-form-urlencoded", nil
	case "raw":
		return strings.NewReader(raw), "text/plain", nil
	default:
		return nil, "", fmt.Errorf("unsupported body type")
	}
}
func stringMap(value any) map[string]string {
	result := map[string]string{}
	if source, ok := value.(map[string]any); ok {
		for k, v := range source {
			result[k] = fmt.Sprint(v)
		}
	}
	if source, ok := value.(map[string]string); ok {
		for k, v := range source {
			result[k] = v
		}
	}
	return result
}
func stringValue(value any, fallback string) string {
	if v, ok := value.(string); ok && v != "" {
		return v
	}
	return fallback
}
func fail(message string) models.ToolOutput { return models.Failure("HTTP_ERROR", message) }

func limitRedirects(_ *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return fmt.Errorf("stopped after 10 redirects")
	}
	return nil
}
