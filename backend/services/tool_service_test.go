package services

import (
	"context"
	"testing"

	"developer-toolbox/backend/models"
	"developer-toolbox/backend/registry"
)

type testTool struct {
	id string
}

type contextualTestTool struct{ received models.ToolContext }
type panicTestTool struct{}

func (panicTestTool) Info() models.Tool                          { return models.Tool{ID: "panic"} }
func (panicTestTool) Execute(models.ToolInput) models.ToolOutput { panic("boom") }

func (t *contextualTestTool) Info() models.Tool { return models.Tool{ID: "contextual"} }
func (t *contextualTestTool) Execute(ctx models.ToolContext, input models.ToolInput) models.ToolOutput {
	t.received = ctx
	return models.ToolOutput{Success: true, Data: input.Payload}
}

func (t testTool) Info() models.Tool {
	return models.Tool{ID: t.id, Name: "Test Tool", Description: "Test"}
}

func TestToolService_ContextualToolReceivesHostContext(t *testing.T) {
	service := NewToolService(registry.NewDefaultRegistry())
	tool := &contextualTestTool{}
	service.SetContextFactory(func(ctx context.Context) models.ToolContext {
		result := models.NewToolContext(ctx)
		result.Workspace = "workspace"
		return result
	})
	if err := service.Register(tool); err != nil {
		t.Fatal(err)
	}
	if output := service.ExecuteToolContext(context.Background(), models.ToolInput{ToolID: "contextual"}); !output.Success {
		t.Fatalf("expected success: %+v", output)
	}
	if tool.received.Workspace != "workspace" || tool.received.Context == nil {
		t.Fatalf("context was not propagated: %+v", tool.received)
	}
}

func (t testTool) Execute(input models.ToolInput) models.ToolOutput {
	return models.ToolOutput{Success: true, Data: input.Payload}
}

func TestToolService_RegisterAndExecute(t *testing.T) {
	r := NewToolService(registry.NewDefaultRegistry())
	tool := testTool{id: "test"}

	if err := r.Register(tool); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}

	output := r.ExecuteTool(models.ToolInput{ToolID: "test", Payload: map[string]any{"value": 1}})
	if !output.Success {
		t.Fatalf("expected success, got %s", output.Error)
	}
}

func TestToolService_ExecuteMissingTool(t *testing.T) {
	r := NewToolService(registry.NewDefaultRegistry())
	output := r.ExecuteTool(models.ToolInput{ToolID: "missing"})
	if output.Success {
		t.Fatal("expected failure for missing tool")
	}
}

func TestToolServiceRecoversToolPanic(t *testing.T) {
	service := NewToolService(registry.NewDefaultRegistry())
	if err := service.Register(panicTestTool{}); err != nil {
		t.Fatal(err)
	}
	output := service.ExecuteTool(models.ToolInput{ToolID: "panic"})
	if output.Success || output.Error == nil || output.Error.Code != "TOOL_PANIC" {
		t.Fatalf("unexpected output: %+v", output)
	}
}
