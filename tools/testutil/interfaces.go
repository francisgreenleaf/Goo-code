package testutil

import (
	"context"
	"encoding/json"

	"github.com/anthropics/anthropic-sdk-go"
)

// Tool interface for testing - mirrors the main Tool interface to avoid import cycles
type Tool interface {
	Name() string
	Description() string
	InputSchema() anthropic.ToolInputSchemaParam
	Execute(ctx context.Context, agent ToolContext, input json.RawMessage) (string, error)
}

// ToolContext interface for testing - mirrors the main ToolContext interface
type ToolContext interface {
	WorkingDir() string
	ResolveFilePath(relativePath string) (string, error)
}

// Registry interface for testing - mirrors the main Registry interface
type Registry interface {
	Register(tool Tool)
	Get(name string) (Tool, bool)
	All() []ToolDefinition
	Execute(ctx context.Context, agent ToolContext, toolName string, input json.RawMessage) (string, error)
}

// ToolDefinition represents a complete tool definition for registration
type ToolDefinition struct {
	Name        string
	Description string
	InputSchema anthropic.ToolInputSchemaParam
	Function    func(json.RawMessage) (string, error)
}
