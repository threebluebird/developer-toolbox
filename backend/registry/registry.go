package registry

import "developer-toolbox/backend/models"

// ToolDefinition is the stable portion shared by V0.1 and V0.5 tools.
type ToolDefinition interface {
	Info() models.Tool
}

// ToolExecutor is the V0.5 execution contract.
type ToolExecutor interface {
	ToolDefinition
	Execute(ctx models.ToolContext, input models.ToolInput) models.ToolOutput
}

// LegacyToolExecutor keeps V0.1 tools operational during the incremental
// migration to contextual execution.
type LegacyToolExecutor interface {
	ToolDefinition
	Execute(input models.ToolInput) models.ToolOutput
}

// Registry provides access to registered tools.
type Registry interface {
	Register(tool ToolDefinition) error
	Get(toolID string) (ToolDefinition, bool)
	List() []models.Tool
}
