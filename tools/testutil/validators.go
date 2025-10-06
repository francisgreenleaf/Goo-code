package testutil

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"


	"github.com/anthropics/anthropic-sdk-go"
)

// ValidateToolInterface ensures a tool properly implements the Tool interface
func ValidateToolInterface(t *testing.T, tool Tool) {
	t.Helper()

	// Test Name() method
	name := tool.Name()
	if name == "" {
		t.Error("Tool Name() returned empty string")
	}

	// Test Description() method
	description := tool.Description()
	if description == "" {
		t.Error("Tool Description() returned empty string")
	}

	// Test InputSchema() method
	schema := tool.InputSchema()
	// Note: We can't directly check for nil with ToolInputSchemaParam, so we validate by marshaling

	// Validate schema structure
	ValidateSchema(t, schema, fmt.Sprintf("Tool %s schema", name))
}

// ValidateSchema validates that a schema has the expected structure
func ValidateSchema(t *testing.T, schema anthropic.ToolInputSchemaParam, context string) {
	t.Helper()

	// Check if schema is effectively nil by trying to marshal it
	schemaBytes, err := json.Marshal(schema)
	if err != nil {
		t.Errorf("%s: failed to marshal schema: %v", context, err)
		return
	}
	
	if len(schemaBytes) == 0 || string(schemaBytes) == "null" {
		t.Errorf("%s: schema is nil or empty", context)
		return
	}

	var schemaMap map[string]interface{}
	if err := json.Unmarshal(schemaBytes, &schemaMap); err != nil {
		t.Errorf("%s: failed to unmarshal schema: %v", context, err)
		return
	}

	// Check required fields
	if schemaType, ok := schemaMap["type"]; !ok || schemaType != "object" {
		t.Errorf("%s: schema must have type 'object', got %v", context, schemaType)
	}

	if properties, ok := schemaMap["properties"]; !ok {
		t.Errorf("%s: schema must have 'properties' field", context)
	} else {
		// Validate properties structure
		if propertiesMap, ok := properties.(map[string]interface{}); ok {
			if len(propertiesMap) == 0 {
				t.Errorf("%s: schema properties cannot be empty", context)
			}
		} else {
			t.Errorf("%s: schema properties must be an object", context)
		}
	}
}

// AssertToolRegistered checks that a tool is properly registered in the registry
func AssertToolRegistered(t *testing.T, registry Registry, toolName string) {
	t.Helper()

	tool, exists := registry.Get(toolName)
	if !exists {
		t.Errorf("Tool %s is not registered in the registry", toolName)
		return
	}

	if tool.Name() != toolName {
		t.Errorf("Tool name mismatch: expected %s, got %s", toolName, tool.Name())
	}
}

// AssertToolNotRegistered checks that a tool is not registered in the registry
func AssertToolNotRegistered(t *testing.T, registry Registry, toolName string) {
	t.Helper()

	_, exists := registry.Get(toolName)
	if exists {
		t.Errorf("Tool %s should not be registered in the registry", toolName)
	}
}

// ValidateToolExecution tests that a tool can be executed with valid input
func ValidateToolExecution(t *testing.T, tool Tool, agent ToolContext, validInput interface{}) {
	t.Helper()

	// Marshal input to JSON
	inputBytes, err := json.Marshal(validInput)
	if err != nil {
		t.Fatalf("Failed to marshal test input: %v", err)
	}

	// Execute the tool
	result, err := tool.Execute(nil, agent, json.RawMessage(inputBytes))
	if err != nil {
		t.Errorf("Tool execution failed with valid input: %v", err)
		return
	}

	if result == "" {
		t.Error("Tool execution returned empty result")
	}
}

// ValidateToolExecutionError tests that a tool returns an error with invalid input
func ValidateToolExecutionError(t *testing.T, tool Tool, agent ToolContext, invalidInput interface{}, expectedErrorSubstring string) {
	t.Helper()

	// Marshal input to JSON
	inputBytes, err := json.Marshal(invalidInput)
	if err != nil {
		t.Fatalf("Failed to marshal test input: %v", err)
	}

	// Execute the tool
	result, err := tool.Execute(nil, agent, json.RawMessage(inputBytes))
	if err == nil {
		t.Errorf("Tool execution should have failed with invalid input, but got result: %s", result)
		return
	}

	if expectedErrorSubstring != "" && !containsString(err.Error(), expectedErrorSubstring) {
		t.Errorf("Error message should contain '%s', but got: %s", expectedErrorSubstring, err.Error())
	}
}

// ValidateJSONSchema validates that input conforms to the tool's schema
func ValidateJSONSchema(t *testing.T, tool Tool, input interface{}) {
	t.Helper()

	// This is a basic validation - in a real implementation you might want
	// to use a proper JSON schema validator
	inputBytes, err := json.Marshal(input)
	if err != nil {
		t.Errorf("Input cannot be marshaled to JSON: %v", err)
		return
	}

	var inputMap map[string]interface{}
	if err := json.Unmarshal(inputBytes, &inputMap); err != nil {
		t.Errorf("Input is not valid JSON object: %v", err)
		return
	}

	// Get the schema
	schema := tool.InputSchema()
	schemaBytes, err := json.Marshal(schema)
	if err != nil {
		t.Errorf("Schema cannot be marshaled: %v", err)
		return
	}

	var schemaMap map[string]interface{}
	if err := json.Unmarshal(schemaBytes, &schemaMap); err != nil {
		t.Errorf("Schema is not valid JSON: %v", err)
		return
	}

	// Basic validation - check that all required fields are present
	if required, ok := schemaMap["required"].([]interface{}); ok {
		for _, reqField := range required {
			if fieldName, ok := reqField.(string); ok {
				if _, exists := inputMap[fieldName]; !exists {
					t.Errorf("Required field '%s' is missing from input", fieldName)
				}
			}
		}
	}
}

// CompareToolDefinitions compares two tool definitions for equality
func CompareToolDefinitions(t *testing.T, expected, actual ToolDefinition) {
	t.Helper()

	if expected.Name != actual.Name {
		t.Errorf("Tool name mismatch: expected %s, got %s", expected.Name, actual.Name)
	}

	if expected.Description != actual.Description {
		t.Errorf("Tool description mismatch: expected %s, got %s", expected.Description, actual.Description)
	}

	// Compare schemas by marshaling to JSON
	expectedSchemaBytes, err := json.Marshal(expected.InputSchema)
	if err != nil {
		t.Errorf("Failed to marshal expected schema: %v", err)
		return
	}

	actualSchemaBytes, err := json.Marshal(actual.InputSchema)
	if err != nil {
		t.Errorf("Failed to marshal actual schema: %v", err)
		return
	}

	if !reflect.DeepEqual(expectedSchemaBytes, actualSchemaBytes) {
		t.Errorf("Tool schema mismatch:\nExpected: %s\nActual: %s", 
			string(expectedSchemaBytes), string(actualSchemaBytes))
	}
}

// ValidateRegistryConsistency checks that all tools in registry are properly configured
func ValidateRegistryConsistency(t *testing.T, registry Registry) {
	t.Helper()

	definitions := registry.All()
	if len(definitions) == 0 {
		t.Error("Registry has no tools registered")
		return
	}

	seenNames := make(map[string]bool)
	for _, def := range definitions {
		// Check for duplicate names
		if seenNames[def.Name] {
			t.Errorf("Duplicate tool name found: %s", def.Name)
		}
		seenNames[def.Name] = true

		// Validate each definition
		if def.Name == "" {
			t.Error("Tool definition has empty name")
		}
		if def.Description == "" {
			t.Errorf("Tool %s has empty description", def.Name)
		}
		// Check if schema is effectively nil
		if schemaBytes, err := json.Marshal(def.InputSchema); err != nil || len(schemaBytes) == 0 || string(schemaBytes) == "null" {
			t.Errorf("Tool %s has nil or invalid input schema", def.Name)
		}

		// Verify tool can be retrieved from registry
		tool, exists := registry.Get(def.Name)
		if !exists {
			t.Errorf("Tool %s in definitions but not retrievable from registry", def.Name)
			continue
		}

		// Verify consistency between definition and tool
		if tool.Name() != def.Name {
			t.Errorf("Tool name inconsistency: definition says %s, tool says %s", 
				def.Name, tool.Name())
		}
		if tool.Description() != def.Description {
			t.Errorf("Tool description inconsistency for %s: definition says %s, tool says %s", 
				def.Name, def.Description, tool.Description())
		}
	}
}

// Helper function to check if a string contains a substring (case-insensitive)
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    len(s) > len(substr) && 
		    (s[:len(substr)] == substr || 
		     s[len(s)-len(substr):] == substr || 
		     containsString(s[1:], substr)))
}

// TestToolWithAllInputs is a helper that tests a tool with various input scenarios
func TestToolWithAllInputs(t *testing.T, tool Tool, agent ToolContext, testCases []ToolTestCase) {
	t.Helper()

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			inputBytes, err := json.Marshal(tc.Input)
			if err != nil {
				t.Fatalf("Failed to marshal input: %v", err)
			}

			result, err := tool.Execute(nil, agent, json.RawMessage(inputBytes))

			if tc.ShouldError {
				if err == nil {
					t.Errorf("Expected error but got result: %s", result)
				} else if tc.ErrorContains != "" && !containsString(err.Error(), tc.ErrorContains) {
					t.Errorf("Error should contain '%s', got: %s", tc.ErrorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else if tc.ResultContains != "" && !containsString(result, tc.ResultContains) {
					t.Errorf("Result should contain '%s', got: %s", tc.ResultContains, result)
				}
			}
		})
	}
}

// ToolTestCase represents a test case for tool execution
type ToolTestCase struct {
	Name           string
	Input          interface{}
	ShouldError    bool
	ErrorContains  string
	ResultContains string
}
