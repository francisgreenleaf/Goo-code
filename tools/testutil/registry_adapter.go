package testutil

import (
	"context"
	"encoding/json"

	"anthropic-chat/tools"
	"github.com/anthropics/anthropic-sdk-go"
)

// RegistryAdapter adapts the main tools.Registry to work with our test interfaces
type RegistryAdapter struct {
	registry *tools.Registry
}

// NewRegistryAdapter creates a new registry adapter
func NewRegistryAdapter(registry *tools.Registry) *RegistryAdapter {
	return &RegistryAdapter{registry: registry}
}

// Register adapts a main tool to our test interface and registers it
func (r *RegistryAdapter) Register(tool Tool) {
	// Create an adapter that wraps our test tool to implement the main Tool interface
	adapter := &toolAdapter{tool: tool}
	r.registry.Register(adapter)
}

// Get retrieves a tool and wraps it in our test interface
func (r *RegistryAdapter) Get(name string) (Tool, bool) {
	tool, exists := r.registry.Get(name)
	if !exists {
		return nil, false
	}
	
	// If it's already wrapped, unwrap it
	if adapter, ok := tool.(*toolAdapter); ok {
		return adapter.tool, true
	}
	
	// Otherwise wrap it
	return &mainToolAdapter{tool: tool}, true
}

// All returns all tool definitions
func (r *RegistryAdapter) All() []ToolDefinition {
	mainDefs := r.registry.All()
	testDefs := make([]ToolDefinition, len(mainDefs))
	
	for i, def := range mainDefs {
		testDefs[i] = ToolDefinition{
			Name:        def.Name,
			Description: def.Description,
			InputSchema: def.InputSchema,
			Function:    def.Function,
		}
	}
	
	return testDefs
}

// Execute executes a tool
func (r *RegistryAdapter) Execute(ctx context.Context, agent ToolContext, toolName string, input json.RawMessage) (string, error) {
	// Create an adapter for the agent
	agentAdapter := &toolContextAdapter{agent: agent}
	return r.registry.Execute(ctx, agentAdapter, toolName, input)
}

// toolAdapter adapts our test Tool interface to the main Tool interface
type toolAdapter struct {
	tool Tool
}

func (t *toolAdapter) Name() string {
	return t.tool.Name()
}

func (t *toolAdapter) Description() string {
	return t.tool.Description()
}

func (t *toolAdapter) InputSchema() anthropic.ToolInputSchemaParam {
	return t.tool.InputSchema()
}

func (t *toolAdapter) Execute(ctx context.Context, agent tools.ToolContext, input json.RawMessage) (string, error) {
	// Create an adapter for the agent
	agentAdapter := &mainToolContextAdapter{agent: agent}
	return t.tool.Execute(ctx, agentAdapter, input)
}

// mainToolAdapter adapts a main Tool to our test Tool interface
type mainToolAdapter struct {
	tool tools.Tool
}

func (t *mainToolAdapter) Name() string {
	return t.tool.Name()
}

func (t *mainToolAdapter) Description() string {
	return t.tool.Description()
}

func (t *mainToolAdapter) InputSchema() anthropic.ToolInputSchemaParam {
	return t.tool.InputSchema()
}

func (t *mainToolAdapter) Execute(ctx context.Context, agent ToolContext, input json.RawMessage) (string, error) {
	// Create an adapter for the agent
	agentAdapter := &toolContextAdapter{agent: agent}
	return t.tool.Execute(ctx, agentAdapter, input)
}

// toolContextAdapter adapts our test ToolContext to the main ToolContext interface
type toolContextAdapter struct {
	agent ToolContext
}

func (t *toolContextAdapter) WorkingDir() string {
	return t.agent.WorkingDir()
}

func (t *toolContextAdapter) ResolveFilePath(relativePath string) (string, error) {
	return t.agent.ResolveFilePath(relativePath)
}

// mainToolContextAdapter adapts a main ToolContext to our test ToolContext interface
type mainToolContextAdapter struct {
	agent tools.ToolContext
}

func (t *mainToolContextAdapter) WorkingDir() string {
	return t.agent.WorkingDir()
}

func (t *mainToolContextAdapter) ResolveFilePath(relativePath string) (string, error) {
	return t.agent.ResolveFilePath(relativePath)
}
