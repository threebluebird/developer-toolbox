package services

import (
	"context"
	"developer-toolbox/backend/models"
	"developer-toolbox/backend/registry"
)

// ToolService orchestrates tool execution and discovery.
type ToolService struct {
	registry       registry.Registry
	contextFactory func(context.Context) models.ToolContext
}

// NewToolService creates a new tool service instance with the provided registry.
func NewToolService(reg registry.Registry) *ToolService {
	return &ToolService{
		registry:       reg,
		contextFactory: models.NewToolContext,
	}
}

// SetContextFactory configures host data supplied to contextual V0.5 tools.
func (s *ToolService) SetContextFactory(factory func(context.Context) models.ToolContext) {
	if factory != nil {
		s.contextFactory = factory
	}
}

// Register registers a tool with the underlying registry.
func (s *ToolService) Register(tool registry.ToolDefinition) error {
	return s.registry.Register(tool)
}

// ExecuteTool executes a registered tool by ID.
func (s *ToolService) ExecuteTool(input models.ToolInput) models.ToolOutput {
	return s.ExecuteToolContext(context.Background(), input)
}

// ExecuteToolContext executes a tool with cancellation and host-owned context.
// Legacy tools are adapted transparently so V0.1 behavior remains unchanged.
func (s *ToolService) ExecuteToolContext(ctx context.Context, input models.ToolInput) (output models.ToolOutput) {
	defer func() {
		if recovered := recover(); recovered != nil {
			output = models.Failure("TOOL_PANIC", "tool execution failed unexpectedly")
		}
	}()
	tool, ok := s.registry.Get(input.ToolID)
	if !ok {
		return models.Failure("TOOL_NOT_FOUND", "tool not found")
	}

	toolCtx := s.contextFactory(ctx)
	switch executor := tool.(type) {
	case registry.ToolExecutor:
		return executor.Execute(toolCtx, input)
	case registry.LegacyToolExecutor:
		return executor.Execute(input)
	default:
		return models.Failure("TOOL_NOT_EXECUTABLE", "tool does not implement an execution contract")
	}
}

// ListTools returns metadata for all registered tools.
func (s *ToolService) ListTools() []models.Tool {
	return s.registry.List()
}
