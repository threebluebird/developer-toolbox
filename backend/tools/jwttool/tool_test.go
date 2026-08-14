package jwttool

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"developer-toolbox/backend/models"
)

func jwtSegment(value any) string {
	data, _ := json.Marshal(value)
	return base64.RawURLEncoding.EncodeToString(data)
}

func TestJWTDecodesClaims(t *testing.T) {
	token := jwtSegment(map[string]any{"alg": "none"}) + "." + jwtSegment(map[string]any{"exp": 1, "iat": 0}) + ".signature"
	output := NewJWTTool().Execute(models.ToolInput{Payload: map[string]any{"input": token}})
	if !output.Success {
		t.Fatal(output.Error)
	}
	data := output.Data.(map[string]any)
	if data["expired"] != true || data["iatInfo"] != "1970-01-01T00:00:00Z" {
		t.Fatalf("unexpected claims: %#v", data)
	}
}

func TestJWTRejectsMalformedToken(t *testing.T) {
	output := NewJWTTool().Execute(models.ToolInput{Payload: map[string]any{"input": "not-a-token"}})
	if output.Success || output.Error.Code != "INVALID_JWT" {
		t.Fatalf("unexpected output: %#v", output)
	}
}
