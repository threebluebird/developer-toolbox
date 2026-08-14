package models

// Tool describes a single developer tool exposed by the registry.
type Tool struct {
	ID           string           `json:"id"`
	Name         string           `json:"name"`
	Description  string           `json:"description"`
	Category     string           `json:"category"`
	Icon         string           `json:"icon"`
	Version      string           `json:"version"`
	Capabilities ToolCapabilities `json:"capabilities"`
	Keywords     []string         `json:"keywords,omitempty"`
}

// ToolCapabilities declares which host resources a tool may use. The zero
// value describes a pure, synchronous tool and is therefore V0.1 compatible.
type ToolCapabilities struct {
	FileInput  bool `json:"fileInput"`
	FileOutput bool `json:"fileOutput"`
	Network    bool `json:"network"`
	Streaming  bool `json:"streaming"`
	Stateful   bool `json:"stateful"`
}
