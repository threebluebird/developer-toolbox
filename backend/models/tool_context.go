package models

import "context"

// ToolConfig contains settings scoped to one tool. A map keeps the framework
// extensible without making the core model depend on each tool's options.
type ToolConfig map[string]any

// ToolContext is the host-owned execution environment passed to V0.5 tools.
// Context carries cancellation while the remaining fields expose only the
// application data tools are allowed to consume.
type ToolContext struct {
	Context   context.Context `json:"-"`
	Config    ToolConfig      `json:"config"`
	UserData  map[string]any  `json:"userData"`
	Workspace string          `json:"workspace"`
}

// NewToolContext returns an initialized context safe for immediate use.
func NewToolContext(ctx context.Context) ToolContext {
	if ctx == nil {
		ctx = context.Background()
	}
	return ToolContext{
		Context:  ctx,
		Config:   make(ToolConfig),
		UserData: make(map[string]any),
	}
}
