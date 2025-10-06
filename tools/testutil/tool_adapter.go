package testutil

import (
	"context"
	"encoding/json"

	"anthropic-chat/tools"
	"github.com/anthropics/anthropic-sdk-go"
)

// WrapMainTool wraps a main tools.Tool to work with our test interfaces
func WrapMainTool(mainTool tools.Tool) Tool {
	return &mainToolWrapper{tool: mainTool}
}

// mainToolWrapper adapts a main Tool to our test Tool interface
type mainToolWrapper struct {
	tool tools.Tool
}

func (w *mainToolWrapper) Name() string {
	return w.tool.Name()
}

func (w *mainToolWrapper) Description() string {
	return w.tool.Description()
}

func (w *mainToolWrapper) InputSchema() anthropic.ToolInputSchemaParam {
	return w.tool.InputSchema()
}

func (w *mainToolWrapper) Execute(ctx context.Context, agent ToolContext, input json.RawMessage) (string, error) {
	// Create an adapter for the agent
	agentAdapter := &toolContextAdapter{agent: agent}
	return w.tool.Execute(ctx, agentAdapter, input)
}
