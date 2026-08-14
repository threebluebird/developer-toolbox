package models

// History stores recent tool usage timestamps.
type History struct {
	ID     string         `json:"id,omitempty"`
	ToolID string         `json:"toolId"`
	UsedAt string         `json:"usedAt"`
	Input  map[string]any `json:"input,omitempty"`
	Output any            `json:"output,omitempty"`
}
