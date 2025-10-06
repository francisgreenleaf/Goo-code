package file

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"anthropic-chat/tools/schemas"
	"anthropic-chat/tools/testutil"
)

func TestFindFilesTool_Interface(t *testing.T) {
	tool := NewFindFilesTool()
	wrappedTool := testutil.WrapMainTool(tool)
	
	testutil.ValidateToolInterface(t, wrappedTool)
	
	if tool.Name() != "find_files" {
		t.Errorf("Expected name 'find_files', got '%s'", tool.Name())
	}
	
	expectedDesc := "Find files and directories matching specified criteria within the working directory."
	if tool.Description() != expectedDesc {
		t.Errorf("Expected description '%s', got '%s'", expectedDesc, tool.Description())
	}
}

func TestFindFilesTool_Schema(t *testing.T) {
	tool := NewFindFilesTool()
	schema := tool.InputSchema()
	
	testutil.ValidateSchema(t, schema, "FindFilesTool")
	
	schemaBytes, err := json.Marshal(schema)
	if err != nil {
		t.Fatalf("Failed to marshal schema: %v", err)
	}
	
	var schemaMap map[string]interface{}
	if err := json.Unmarshal(schemaBytes, &schemaMap); err != nil {
		t.Fatalf("Failed to unmarshal schema: %v", err)
	}
	
	properties, ok := schemaMap["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema properties should be an object")
	}
	
	requiredProps := []string{"pattern"}
	for _, prop := range requiredProps {
		if _, exists := properties[prop]; !exists {
			t.Errorf("Schema should have '%s' property", prop)
		}
	}
}

func TestFindFilesTool_Execute_FindByNamePattern(t *testing.T) {
	tool := NewFindFilesTool()
	workspace := testutil.NewTestWorkspace(t)
	agent := workspace.GetMockContext()
	
	// Create test files
	workspace.CreateFile("test.go", "package main")
	workspace.CreateFile("test.txt", "text file")
	workspace.CreateFile("main.go", "package main")
	workspace.CreateFile("helper.go", "package helper")
	workspace.CreateFile("readme.md", "# Readme")
	
	input := schemas.FindFilesInput{
		Pattern:   "*.go",
		Recursive: true,
	}
	
	inputBytes, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal input: %v", err)
	}
	
	result, err := tool.Execute(context.Background(), agent, json.RawMessage(inputBytes))
	if err != nil {
		t.Errorf("Find failed: %v", err)
	}
	
	// Verify result contains all .go files
	if !strings.Contains(result, "test.go") {
		t.Error("Result should contain test.go")
	}
	if !strings.Contains(result, "main.go") {
		t.Error("Result should contain main.go")
	}
	if !strings.Contains(result, "helper.go") {
		t.Error("Result should contain helper.go")
	}
	
	// Verify result doesn't contain non-.go files
	if strings.Contains(result, "test.txt") {
		t.Error("Result should not contain test.txt")
	}
	if strings.Contains(result, "readme.md") {
		t.Error("Result should not contain readme.md")
	}
}

func TestFindFilesTool_Execute_FindRecursive(t *testing.T) {
	tool := NewFindFilesTool()
	workspace := testutil.NewTestWorkspace(t)
	agent := workspace.GetMockContext()
	
	// Create nested structure
	workspace.CreateFile("root.txt", "root")
	workspace.CreateFile("dir1/file1.txt", "file1")
	workspace.CreateFile("dir1/dir2/file2.txt", "file2")
	workspace.CreateFile("dir1/dir2/dir3/file3.txt", "file3")
	
	input := schemas.FindFilesInput{
		Pattern:   "*.txt",
		Recursive: true,
	}
	
	inputBytes, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal input: %v", err)
	}
	
	result, err := tool.Execute(context.Background(), agent, json.RawMessage(inputBytes))
	if err != nil {
		t.Errorf("Find failed: %v", err)
	}
	
	// All files should be found
	expectedFiles := []string{"root.txt", "dir1/file1.txt", "dir1/dir2/file2.txt", "dir1/dir2/dir3/file3.txt"}
	for _, file := range expectedFiles {
		if !strings.Contains(result, file) {
			t.Errorf("Result should contain %s", file)
		}
	}
}

func TestFindFilesTool_Execute_FindNonRecursive(t *testing.T) {
	tool := NewFindFilesTool()
	workspace := testutil.NewTestWorkspace(t)
	agent := workspace.GetMockContext()
	
	// Create nested structure
	workspace.CreateFile("root.txt", "root")
	workspace.CreateFile("dir1/file1.txt", "file1")
	workspace.CreateFile("dir1/dir2/file2.txt", "file2")
	
	input := schemas.FindFilesInput{
		Pattern:   "*.txt",
		Recursive: false,
	}
	
	inputBytes, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal input: %v", err)
	}
	
	result, err := tool.Execute(context.Background(), agent, json.RawMessage(inputBytes))
	if err != nil {
		t.Errorf("Find failed: %v", err)
	}
	
	// Only root level file should be found
	if !strings.Contains(result, "root.txt") {
		t.Error("Result should contain root.txt")
	}
	
	// Nested files should not be found
	if strings.Contains(result, "dir1/file1.txt") {
		t.Error("Result should not contain nested files in non-recursive mode")
	}
}

func TestFindFilesTool_Execute_NoMatches(t *testing.T) {
	tool := NewFindFilesTool()
	workspace := testutil.NewTestWorkspace(t)
	agent := workspace.GetMockContext()
	
	// Create files that won't match
	workspace.CreateFile("test.txt", "text")
	workspace.CreateFile("readme.md", "markdown")
	
	input := schemas.FindFilesInput{
		Pattern:   "*.go",
		Recursive: true,
	}
	
	inputBytes, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal input: %v", err)
	}
	
	result, err := tool.Execute(context.Background(), agent, json.RawMessage(inputBytes))
	if err != nil {
		t.Errorf("Find should not error when no matches: %v", err)
	}
	
	if !strings.Contains(result, "No files found") && !strings.Contains(result, "0 files") {
		t.Error("Result should indicate no files were found")
	}
}

func TestFindFilesTool_Execute_MultiplePatterns(t *testing.T) {
	tool := NewFindFilesTool()
	workspace := testutil.NewTestWorkspace(t)
	agent := workspace.GetMockContext()
	
	// Create various files
	workspace.CreateFile("test.go", "go file")
	workspace.CreateFile("test.js", "js file")
	workspace.CreateFile("test.py", "python file")
	workspace.CreateFile("readme.md", "markdown")
	
	// Test finding multiple extensions
	patterns := []string{"*.go", "*.js", "*.py"}
	
	for _, pattern := range patterns {
		t.Run("Pattern_"+pattern, func(t *testing.T) {
			input := schemas.FindFilesInput{
				Pattern:   pattern,
				Recursive: true,
			}
			
			inputBytes, err := json.Marshal(input)
			if err != nil {
				t.Fatalf("Failed to marshal input: %v", err)
			}
			
			result, err := tool.Execute(context.Background(), agent, json.RawMessage(inputBytes))
			if err != nil {
				t.Errorf("Find failed for pattern %s: %v", pattern, err)
			}
			
			// Verify at least one match
			if strings.Contains(result, "No files found") {
				t.Errorf("Should find files matching pattern %s", pattern)
			}
		})
	}
}

func TestFindFilesTool_Execute_PathTraversal(t *testing.T) {
	tool := NewFindFilesTool()
	wrappedTool := testutil.WrapMainTool(tool)
	workspace := testutil.NewTestWorkspace(t)
	agent := workspace.GetMockContext()
	
	pathTraversalInputs := []string{
		"../../../*",
		"../../*.txt",
	}
	
	for _, pattern := range pathTraversalInputs {
		t.Run("PathTraversal_"+pattern, func(t *testing.T) {
			input := schemas.FindFilesInput{
				Pattern:   pattern,
				Recursive: true,
			}
			
			testutil.ValidateToolExecutionError(t, wrappedTool, agent, input, "")
		})
	}
}

func TestFindFilesTool_Execute_InvalidInput(t *testing.T) {
	tool := NewFindFilesTool()
	wrappedTool := testutil.WrapMainTool(tool)
	workspace := testutil.NewTestWorkspace(t)
	agent := workspace.GetMockContext()
	
	testCases := []testutil.ToolTestCase{
		{
			Name:        "EmptyPattern",
			Input:       schemas.FindFilesInput{Pattern: "", Recursive: true},
			ShouldError: true,
			ErrorContains: "cannot be empty",
		},
		{
			Name:        "InvalidJSON",
			Input:       "invalid json",
			ShouldError: true,
			ErrorContains: "failed to parse input",
		},
	}
	
	testutil.TestToolWithAllInputs(t, wrappedTool, agent, testCases)
}

func TestFindFilesTool_Execute_LargeResultSet(t *testing.T) {
	tool := NewFindFilesTool()
	workspace := testutil.NewTestWorkspace(t)
	agent := workspace.GetMockContext()
	
	// Create many files
	for i := 0; i < 50; i++ {
		workspace.CreateFile(fmt.Sprintf("file%d.txt", i), "content")
	}
	
	input := schemas.FindFilesInput{
		Pattern:   "*.txt",
		Recursive: true,
		Limit:     10,
	}
	
	inputBytes, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal input: %v", err)
	}
	
	result, err := tool.Execute(context.Background(), agent, json.RawMessage(inputBytes))
	if err != nil {
		t.Errorf("Find failed: %v", err)
	}
	
	// Verify the limit was applied by checking the result message
	if !strings.Contains(result, "Limited to 10 results") {
		// Count actual file entries (lines with file paths)
		lines := strings.Split(result, "\n")
		fileLines := 0
		for _, line := range lines {
			if strings.Contains(line, "file") && strings.Contains(line, ".txt") && strings.HasPrefix(strings.TrimSpace(line), "file") {
				fileLines++
			}
		}
		if fileLines > 10 {
			t.Errorf("Result should be limited to 10 files, found %d file entries", fileLines)
		}
	}
}

func TestFindFilesTool_Execute_SpecialCharacters(t *testing.T) {
	tool := NewFindFilesTool()
	workspace := testutil.NewTestWorkspace(t)
	agent := workspace.GetMockContext()
	
	specialFiles := []string{
		"file with spaces.txt",
		"file-with-dashes.txt",
		"file_with_underscores.txt",
	}
	
	for _, filename := range specialFiles {
		workspace.CreateFile(filename, "Content")
	}
	
	input := schemas.FindFilesInput{
		Pattern:   "*.txt",
		Recursive: true,
	}
	
	inputBytes, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal input: %v", err)
	}
	
	result, err := tool.Execute(context.Background(), agent, json.RawMessage(inputBytes))
	if err != nil {
		t.Errorf("Find failed: %v", err)
	}
	
	// All special character files should be found
	for _, filename := range specialFiles {
		if !strings.Contains(result, filename) {
			t.Errorf("Result should contain %s", filename)
		}
	}
}
