package file

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"anthropic-chat/tools/schemas"
	"anthropic-chat/tools/testutil"
)

func TestReadFileTool_Interface(t *testing.T) {
	tool := NewReadFileTool()
	wrappedTool := testutil.WrapMainTool(tool)
	
	// Validate that it properly implements the Tool interface
	testutil.ValidateToolInterface(t, wrappedTool)
	
	// Check specific values
	if tool.Name() != "read_file" {
		t.Errorf("Expected name 'read_file', got '%s'", tool.Name())
	}
	
	expectedDesc := "Read file contents from relative path within working directory."
	if tool.Description() != expectedDesc {
		t.Errorf("Expected description '%s', got '%s'", expectedDesc, tool.Description())
	}
}

func TestReadFileTool_Schema(t *testing.T) {
	tool := NewReadFileTool()
	schema := tool.InputSchema()
	
	// Validate schema structure
	testutil.ValidateSchema(t, schema, "ReadFileTool")
	
	// Test that schema matches the expected structure
	schemaBytes, err := json.Marshal(schema)
	if err != nil {
		t.Fatalf("Failed to marshal schema: %v", err)
	}
	
	var schemaMap map[string]interface{}
	if err := json.Unmarshal(schemaBytes, &schemaMap); err != nil {
		t.Fatalf("Failed to unmarshal schema: %v", err)
	}
	
	// Check that 'path' property exists
	properties, ok := schemaMap["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema properties should be an object")
	}
	
	if _, exists := properties["path"]; !exists {
		t.Error("Schema should have 'path' property")
	}
	
	// Check if required fields exist (they may not be present if all fields are required by default)
	if required, ok := schemaMap["required"].([]interface{}); ok {
		foundPath := false
		for _, field := range required {
			if field == "path" {
				foundPath = true
				break
			}
		}
		
		if !foundPath {
			t.Error("'path' should be a required field")
		}
	}
	// If no required field is present, that's also acceptable for this schema
}

func TestReadFileTool_Execute_Success(t *testing.T) {
	tool := NewReadFileTool()
	wrappedTool := testutil.WrapMainTool(tool)
	workspace := testutil.NewTestWorkspace(t)
	agent := workspace.GetMockContext()
	
	// Create a test file
	testContent := "Hello, World!\nThis is a test file."
	workspace.CreateFile("test.txt", testContent)
	
	// Test reading the file
	input := schemas.ReadFileInput{
		Path: "test.txt",
	}
	
	testutil.ValidateToolExecution(t, wrappedTool, agent, input)
	
	// Test the actual result
	inputBytes, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal input: %v", err)
	}
	
	result, err := tool.Execute(context.Background(), agent, json.RawMessage(inputBytes))
	if err != nil {
		t.Errorf("Execution failed: %v", err)
	}
	
	if result != testContent {
		t.Errorf("Expected content '%s', got '%s'", testContent, result)
	}
}

func TestReadFileTool_Execute_FileNotFound(t *testing.T) {
	tool := NewReadFileTool()
	wrappedTool := testutil.WrapMainTool(tool)
	workspace := testutil.NewTestWorkspace(t)
	agent := workspace.GetMockContext()
	
	input := schemas.ReadFileInput{
		Path: "nonexistent.txt",
	}
	
	testutil.ValidateToolExecutionError(t, wrappedTool, agent, input, "failed to read file")
}

func TestReadFileTool_Execute_PathTraversal(t *testing.T) {
	tool := NewReadFileTool()
	wrappedTool := testutil.WrapMainTool(tool)
	workspace := testutil.NewTestWorkspace(t)
	agent := workspace.GetMockContext()
	
	// Test various path traversal attempts
	pathTraversalInputs := []string{
		"../../../etc/passwd",
		"..\\..\\windows\\system32\\config\\sam",
		"../outside.txt",
		"subdir/../../outside.txt",
	}
	
	for _, path := range pathTraversalInputs {
		t.Run("PathTraversal_"+path, func(t *testing.T) {
			input := schemas.ReadFileInput{
				Path: path,
			}
			
			testutil.ValidateToolExecutionError(t, wrappedTool, agent, input, "security reasons")
		})
	}
}

func TestReadFileTool_Execute_InvalidInput(t *testing.T) {
	tool := NewReadFileTool()
	wrappedTool := testutil.WrapMainTool(tool)
	workspace := testutil.NewTestWorkspace(t)
	agent := workspace.GetMockContext()
	
	testCases := []testutil.ToolTestCase{
		{
			Name:         "EmptyPath",
			Input:        schemas.ReadFileInput{Path: ""},
			ShouldError:  true,
			ErrorContains: "failed to read file",
		},
		{
			Name:         "InvalidJSON",
			Input:        "invalid json",
			ShouldError:  true,
			ErrorContains: "failed to parse input",
		},
		{
			Name:         "MissingPath",
			Input:        map[string]interface{}{},
			ShouldError:  true,
			ErrorContains: "failed to read file",
		},
	}
	
	testutil.TestToolWithAllInputs(t, wrappedTool, agent, testCases)
}

func TestReadFileTool_Execute_DifferentFileTypes(t *testing.T) {
	tool := NewReadFileTool()
	workspace := testutil.NewTestWorkspace(t)
	agent := workspace.GetMockContext()
	
	// Test different file types and contents
	testFiles := map[string]string{
		"text.txt":       "Simple text content",
		"empty.txt":      "",
		"json.json":      `{"key": "value", "number": 42}`,
		"code.go":        testutil.SampleGoCode,
		"markdown.md":    testutil.SampleMarkdown,
		"unicode.txt":    "Hello 世界! 🌍 Émojis and ñoñó",
		"multiline.txt":  "Line 1\nLine 2\nLine 3\n\nLine 5",
		"binary-like.txt": "\x00\x01\x02\x03\xFF", // Binary-like content
	}
	
	for filename, content := range testFiles {
		t.Run("FileType_"+filename, func(t *testing.T) {
			workspace.CreateFile(filename, content)
			
			input := schemas.ReadFileInput{
				Path: filename,
			}
			
			inputBytes, err := json.Marshal(input)
			if err != nil {
				t.Fatalf("Failed to marshal input: %v", err)
			}
			
			result, err := tool.Execute(context.Background(), agent, json.RawMessage(inputBytes))
			if err != nil {
				t.Errorf("Failed to read %s: %v", filename, err)
				return
			}
			
			if result != content {
				t.Errorf("Content mismatch for %s.\nExpected: %q\nGot: %q", filename, content, result)
			}
		})
	}
}

func TestReadFileTool_Execute_NestedPaths(t *testing.T) {
	tool := NewReadFileTool()
	workspace := testutil.NewTestWorkspace(t)
	agent := workspace.GetMockContext()
	
	// Create nested directory structure
	nestedFiles := map[string]string{
		"root.txt":                    "Root level file",
		"dir1/file1.txt":             "File in dir1",
		"dir1/dir2/file2.txt":        "File in nested dir",
		"dir1/dir2/dir3/deep.txt":    "Deep nested file",
		"src/main.go":                "package main\n\nfunc main() {}",
		"tests/unit/helper_test.go":  "package unit\n\nimport \"testing\"",
	}
	
	for path, content := range nestedFiles {
		workspace.CreateFile(path, content)
	}
	
	// Test reading all nested files
	for path, expectedContent := range nestedFiles {
		t.Run("NestedPath_"+path, func(t *testing.T) {
			input := schemas.ReadFileInput{
				Path: path,
			}
			
			inputBytes, err := json.Marshal(input)
			if err != nil {
				t.Fatalf("Failed to marshal input: %v", err)
			}
			
			result, err := tool.Execute(context.Background(), agent, json.RawMessage(inputBytes))
			if err != nil {
				t.Errorf("Failed to read nested file %s: %v", path, err)
				return
			}
			
			if result != expectedContent {
				t.Errorf("Content mismatch for %s.\nExpected: %q\nGot: %q", path, expectedContent, result)
			}
		})
	}
}

func TestReadFileTool_Execute_LargeFile(t *testing.T) {
	tool := NewReadFileTool()
	workspace := testutil.NewTestWorkspace(t)
	agent := workspace.GetMockContext()
	
	// Create a large file (1MB of content)
	largeContent := ""
	line := "This is line number %d with some additional content to make it longer.\n"
	for i := 0; i < 10000; i++ {
		largeContent += fmt.Sprintf(line, i)
	}
	
	workspace.CreateFile("large.txt", largeContent)
	
	input := schemas.ReadFileInput{
		Path: "large.txt",
	}
	
	inputBytes, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal input: %v", err)
	}
	
	result, err := tool.Execute(context.Background(), agent, json.RawMessage(inputBytes))
	if err != nil {
		t.Errorf("Failed to read large file: %v", err)
		return
	}
	
	if result != largeContent {
		t.Error("Large file content mismatch")
	}
	
	if len(result) != len(largeContent) {
		t.Errorf("Large file length mismatch: expected %d, got %d", len(largeContent), len(result))
	}
}

func TestReadFileTool_Execute_SpecialCharacters(t *testing.T) {
	tool := NewReadFileTool()
	workspace := testutil.NewTestWorkspace(t)
	agent := workspace.GetMockContext()
	
	// Test files with special characters in names
	specialFiles := map[string]string{
		"file with spaces.txt":    "Content with spaces in filename",
		"file-with-dashes.txt":    "Content with dashes in filename",
		"file_with_underscores.txt": "Content with underscores in filename",
		"file.with.dots.txt":      "Content with dots in filename",
		"file123numbers.txt":      "Content with numbers in filename",
	}
	
	for filename, content := range specialFiles {
		workspace.CreateFile(filename, content)
		
		t.Run("SpecialChar_"+filename, func(t *testing.T) {
			input := schemas.ReadFileInput{
				Path: filename,
			}
			
			inputBytes, err := json.Marshal(input)
			if err != nil {
				t.Fatalf("Failed to marshal input: %v", err)
			}
			
			result, err := tool.Execute(context.Background(), agent, json.RawMessage(inputBytes))
			if err != nil {
				t.Errorf("Failed to read file with special characters %s: %v", filename, err)
				return
			}
			
			if result != content {
				t.Errorf("Content mismatch for %s.\nExpected: %q\nGot: %q", filename, content, result)
			}
		})
	}
}

func TestReadFileTool_Execute_Directory(t *testing.T) {
	tool := NewReadFileTool()
	wrappedTool := testutil.WrapMainTool(tool)
	workspace := testutil.NewTestWorkspace(t)
	agent := workspace.GetMockContext()
	
	// Create a directory
	workspace.CreateDir("testdir")
	
	input := schemas.ReadFileInput{
		Path: "testdir",
	}
	
	testutil.ValidateToolExecutionError(t, wrappedTool, agent, input, "failed to read file")
}
