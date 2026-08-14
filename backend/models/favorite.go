package models

// Favorite stores a tool ID that has been marked as a favorite.
type Favorite struct {
	ID      string         `json:"id,omitempty"`
	ToolID  string         `json:"toolId"`
	Kind    string         `json:"kind,omitempty"`
	Name    string         `json:"name,omitempty"`
	Payload map[string]any `json:"payload,omitempty"`
	AddedAt string         `json:"addedAt,omitempty"`
}
