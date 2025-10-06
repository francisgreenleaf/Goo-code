package tools_test

import (
	"context"
	"encoding/json"
	"testing"

	"anthropic-chat/tools"
	"anthropic-chat/tools/testutil"

	"github.com/anthropics/anthropic-sdk-go"
)

func TestRegistry_NewRegistry(t *testing.T) {
	registry := tools.NewRegistry()
	
	if registry == nil {
		t.Fatal("NewRegistry() returned nil")
	}
	
	// Should start empty
	definitions := registry.All()
	if len(definitions) != 0 {
		t.Errorf("New registry should be empty, got %d tools", len(definitions))
	}
}

func TestRegistry_Register(t *testing.T) {
	registry := tools.NewRegistry()
	adapter := testutil.NewRegistryAdapter(registry)
	mockTool := &MockTool{
		name:        "test_tool",
		description: "A test tool",
		schema:      createMockSchema(),
	}
	
	// Register the tool
	adapter.Register(mockTool)
	
	// Verify it's registered
	testutil.AssertToolRegistered(t, adapter, "test_tool")
	
	// Verify it can be retrieved
	retrievedTool, exists := adapter.Get("test_tool")
	if !exists {
		t.Error("Tool should exist after registration")
	}
	
	if retrievedTool.Name() != mockTool.Name() {
		t.Error("Retrieved tool should have the same name")
	}
}

func TestRegistry_Get(t *testing.T) {
	registry := tools.NewRegistry()
	adapter := testutil.NewRegistryAdapter(registry)
	mockTool := &MockTool{
		name:        "test_tool",
		description: "A test tool",
		schema:      createMockSchema(),
	}
	
	// Test getting non-existent tool
	_, exists := adapter.Get("nonexistent")
	if exists {
		t.Error("Non-existent tool should not exist")
	}
	
	// Register and test getting existing tool
	adapter.Register(mockTool)
	tool, exists := adapter.Get("test_tool")
	if !exists {
		t.Error("Registered tool should exist")
	}
	
	if tool.Name() != "test_tool" {
		t.Errorf("Expected tool name 'test_tool', got '%s'", tool.Name())
	}
}

func TestRegistry_All(t *testing.T) {
	registry := tools.NewRegistry()
	adapter := testutil.NewRegistryAdapter(registry)
	
	// Test empty registry
	definitions := adapter.All()
	if len(definitions) != 0 {
		t.Errorf("Empty registry should return 0 definitions, got %d", len(definitions))
	}
	
	// Add some tools
	mockTools := []*MockTool{
		{name: "tool1", description: "First tool", schema: createMockSchema()},
		{name: "tool2", description: "Second tool", schema: createMockSchema()},
		{name: "tool3", description: "Third tool", schema: createMockSchema()},
	}
	
	for _, tool := range mockTools {
		adapter.Register(tool)
	}
	
	// Test populated registry
	definitions = adapter.All()
	if len(definitions) != 3 {
		t.Errorf("Expected 3 definitions, got %d", len(definitions))
	}
	
	// Verify all tools are present
	foundNames := make(map[string]bool)
	for _, def := range definitions {
		foundNames[def.Name] = true
		
		// Verify definition structure
		if def.Name == "" {
			t.Error("Tool definition has empty name")
		}
		if def.Description == "" {
			t.Error("Tool definition has empty description")
		}
		// Check if schema is effectively nil
		if schemaBytes, err := json.Marshal(def.InputSchema); err != nil || len(schemaBytes) == 0 || string(schemaBytes) == "null" {
			t.Error("Tool definition has nil or invalid schema")
		}
	}
	
	expectedNames := []string{"tool1", "tool2", "tool3"}
	for _, name := range expectedNames {
		if !foundNames[name] {
			t.Errorf("Tool %s not found in definitions", name)
		}
	}
}

func TestRegistry_Execute(t *testing.T) {
	registry := tools.NewRegistry()
	adapter := testutil.NewRegistryAdapter(registry)
	workspace := testutil.NewTestWorkspace(t)
	agent := workspace.GetMockContext()
	
	mockTool := &MockTool{
		name:        "test_tool",
		description: "A test tool",
		schema:      createMockSchema(),
		executeFunc: func(ctx context.Context, agent testutil.ToolContext, input json.RawMessage) (string, error) {
			return "test result", nil
		},
	}
	
	adapter.Register(mockTool)
	
	// Test successful execution
	input := json.RawMessage(`{"test": "value"}`)
	result, err := adapter.Execute(context.Background(), agent, "test_tool", input)
	if err != nil {
		t.Errorf("Execution should succeed, got error: %v", err)
	}
	
	if result != "test result" {
		t.Errorf("Expected 'test result', got '%s'", result)
	}
	
	// Test execution of non-existent tool
	_, err = adapter.Execute(context.Background(), agent, "nonexistent", input)
	if err == nil {
		t.Error("Execution of non-existent tool should fail")
	}
	
	// Verify it's a ToolNotFoundError
	if _, ok := err.(*tools.ToolNotFoundError); !ok {
		t.Errorf("Expected ToolNotFoundError, got %T", err)
	}
}

func TestRegistry_DuplicateRegistration(t *testing.T) {
	registry := tools.NewRegistry()
	adapter := testutil.NewRegistryAdapter(registry)
	
	tool1 := &MockTool{name: "duplicate", description: "First", schema: createMockSchema()}
	tool2 := &MockTool{name: "duplicate", description: "Second", schema: createMockSchema()}
	
	adapter.Register(tool1)
	adapter.Register(tool2) // This should overwrite the first
	
	retrievedTool, exists := adapter.Get("duplicate")
	if !exists {
		t.Error("Tool should exist")
	}
	
	// Should be the second tool (last registered wins)
	if retrievedTool.Description() != "Second" {
		t.Errorf("Expected second tool description, got '%s'", retrievedTool.Description())
	}
	
	// Should only have one definition
	definitions := adapter.All()
	count := 0
	for _, def := range definitions {
		if def.Name == "duplicate" {
			count++
		}
	}
	
	if count != 1 {
		t.Errorf("Expected 1 definition for duplicate name, got %d", count)
	}
}

func TestRegistry_ConsistencyValidation(t *testing.T) {
	registry := tools.NewRegistry()
	adapter := testutil.NewRegistryAdapter(registry)
	
	// Add some valid tools
	mockTools := []*MockTool{
		{name: "valid1", description: "Valid tool 1", schema: createMockSchema()},
		{name: "valid2", description: "Valid tool 2", schema: createMockSchema()},
	}
	
	for _, tool := range mockTools {
		adapter.Register(tool)
	}
	
	// Use our validator to check consistency
	testutil.ValidateRegistryConsistency(t, adapter)
}

func TestToolNotFoundError(t *testing.T) {
	err := &tools.ToolNotFoundError{Name: "missing_tool"}
	
	expectedMsg := "tool missing_tool not found"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

// MockTool implements the testutil.Tool interface for testing
type MockTool struct {
	name        string
	description string
	schema      anthropic.ToolInputSchemaParam
	executeFunc func(context.Context, testutil.ToolContext, json.RawMessage) (string, error)
}

func (m *MockTool) Name() string {
	return m.name
}

func (m *MockTool) Description() string {
	return m.description
}

func (m *MockTool) InputSchema() anthropic.ToolInputSchemaParam {
	return m.schema
}

func (m *MockTool) Execute(ctx context.Context, agent testutil.ToolContext, input json.RawMessage) (string, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, agent, input)
	}
	return "mock result", nil
}

// Helper function to create a mock schema
func createMockSchema() anthropic.ToolInputSchemaParam {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"test": map[string]interface{}{
				"type":        "string",
				"description": "A test parameter",
			},
		},
		"required": []string{"test"},
	}
}
