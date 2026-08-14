package models

// ToolInput is the common request payload for executing a tool.
type ToolInput struct {
	ToolID  string         `json:"toolId"`
	Payload map[string]any `json:"payload"`
}
