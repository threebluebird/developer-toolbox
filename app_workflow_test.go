package main

import "testing"

func TestSafeSnapshotRedactsSensitiveFields(t *testing.T) {
	result := safeSnapshot(map[string]any{"input": "hello", "token": "secret", "headers": map[string]any{"Authorization": "Bearer secret", "Accept": "json"}})
	if result["token"] != "••••••••" {
		t.Fatalf("token was not redacted: %#v", result)
	}
	headers := result["headers"].(map[string]any)
	if headers["Authorization"] != "••••••••" || headers["Accept"] != "json" {
		t.Fatalf("headers not safely redacted: %#v", headers)
	}
}
