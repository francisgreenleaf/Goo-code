package file

import (
	"context"
	"encoding/json"
	"testing"

	"anthropic-chat/tools/schemas"
	"anthropic-chat/tools/testutil"
)

func TestDeleteFileTool_Interface(t *testing.T) {
	tool := NewDeleteFileTool()
	wrappedTool := testutil.WrapMainTool(tool)
	
	testutil.ValidateToolInterface(t, wrappedTool)
	
	if tool.Name() != "delete_file" {
		t.Errorf("Expected name 'delete_file', got '%s'", tool.Name())
	}
	
	expectedDesc := "Delete a file or directory within the working directory."
	if tool.Description() != expectedDesc {
		t.Errorf("Expected description '%s', got '%s'", expectedDesc, tool.Description())
	}
}

func TestDeleteFileTool_Schema(t *testing.T) {
	tool := NewDeleteFileTool()
	schema := tool.InputSchema()
	
	testutil.ValidateSchema(t, schema, "DeleteFileTool")
	
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
	
	if _, exists := properties["path"]; !exists {
		t.Error("Schema should have 'path' property")
	}
	
	if _, exists := properties["recursive"]; !exists {
		t.Error("Schema should have 'recursive' property")
	}
}

func TestDeleteFileTool_Execute_DeleteFile(t *testing.T) {
	tool := NewDeleteFileTool()
	workspace := testutil.NewTestWorkspace(t)
	agent := workspace.GetMockContext()
	
	// Create a test file
	workspace.CreateFile("test.txt", "Content to delete")
	
	if !workspace.FileExists("test.txt") {
		t.Fatal("Test file should exist before deletion")
	}
	
	input := schemas.DeleteFileInput{
		Path:      "test.txt",
		Recursive: false,
	}
	
	inputBytes, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal input: %v", err)
	}
	
	result, err := tool.Execute(context.Background(), agent, json.RawMessage(inputBytes))
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}
	
	// Verify file is deleted
	if workspace.FileExists("test.txt") {
		t.Error("File should not exist after deletion")
	}
	
	if result == "" {
		t.Error("Result should not be empty")
	}
}

func TestDeleteFileTool_Execute_DeleteNonExistent(t *testing.T) {
	tool := NewDeleteFileTool()
	wrappedTool := testutil.WrapMainTool(tool)
	workspace := testutil.NewTestWorkspace(t)
	agent := workspace.GetMockContext()
	
	input := schemas.DeleteFileInput{
		Path:      "nonexistent.txt",
		Recursive: false,
	}
	
	testutil.ValidateToolExecutionError(t, wrappedTool, agent, input, "does not exist")
}

func TestDeleteFileTool_Execute_DeleteEmptyDirectory(t *testing.T) {
	tool := NewDeleteFileTool()
	workspace := testutil.NewTestWorkspace(t)
	agent := workspace.GetMockContext()
	
	// Create empty directory
	workspace.CreateDir("emptydir")
	
	input := schemas.DeleteFileInput{
		Path:      "emptydir",
		Recursive: false,
	}
	
	inputBytes, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal input: %v", err)
	}
	
	result, err := tool.Execute(context.Background(), agent, json.RawMessage(inputBytes))
	if err != nil {
		t.Errorf("Delete empty directory failed: %v", err)
	}
	
	if workspace.FileExists("emptydir") {
		t.Error("Directory should not exist after deletion")
	}
	
	if result == "" {
		t.Error("Result should not be empty")
	}
}

func TestDeleteFileTool_Execute_DeleteDirectoryNonRecursive(t *testing.T) {
	tool := NewDeleteFileTool()
	workspace := testutil.NewTestWorkspace(t)
	agent := workspace.GetMockContext()
	
	// Create directory with files
	workspace.CreateFile("dir/file.txt", "Content")
	
	input := schemas.DeleteFileInput{
		Path:      "dir",
		Recursive: false,
	}
	
	inputBytes, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal input: %v", err)
	}
	
	_, err = tool.Execute(context.Background(), agent, json.RawMessage(inputBytes))
	if err == nil {
		t.Error("Should error when trying to delete non-empty directory without recursive flag")
	}
}

func TestDeleteFileTool_Execute_DeleteDirectoryRecursive(t *testing.T) {
	tool := NewDeleteFileTool()
	workspace := testutil.NewTestWorkspace(t)
	agent := workspace.GetMockContext()
	
	// Create directory with nested files
	workspace.CreateFile("dir/file1.txt", "Content 1")
	workspace.CreateFile("dir/subdir/file2.txt", "Content 2")
	workspace.CreateFile("dir/subdir/file3.txt", "Content 3")
	
	input := schemas.DeleteFileInput{
		Path:      "dir",
		Recursive: true,
	}
	
	inputBytes, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal input: %v", err)
	}
	
	result, err := tool.Execute(context.Background(), agent, json.RawMessage(inputBytes))
	if err != nil {
		t.Errorf("Recursive delete failed: %v", err)
	}
	
	// Verify entire directory is gone
	if workspace.FileExists("dir") {
		t.Error("Directory should not exist after recursive deletion")
	}
	if workspace.FileExists("dir/file1.txt") {
		t.Error("Nested file should not exist after recursive deletion")
	}
	if workspace.FileExists("dir/subdir") {
		t.Error("Subdirectory should not exist after recursive deletion")
	}
	
	if result == "" {
		t.Error("Result should not be empty")
	}
}

func TestDeleteFileTool_Execute_PathTraversal(t *testing.T) {
	tool := NewDeleteFileTool()
	wrappedTool := testutil.WrapMainTool(tool)
	workspace := testutil.NewTestWorkspace(t)
	agent := workspace.GetMockContext()
	
	pathTraversalInputs := []string{
		"../../../etc/passwd",
		"..\\..\\windows\\system32",
		"../outside.txt",
		"subdir/../../outside.txt",
	}
	
	for _, path := range pathTraversalInputs {
		t.Run("PathTraversal_"+path, func(t *testing.T) {
			input := schemas.DeleteFileInput{
				Path:      path,
				Recursive: false,
			}
			
			testutil.ValidateToolExecutionError(t, wrappedTool, agent, input, "security reasons")
		})
	}
}

func TestDeleteFileTool_Execute_InvalidInput(t *testing.T) {
	tool := NewDeleteFileTool()
	wrappedTool := testutil.WrapMainTool(tool)
	workspace := testutil.NewTestWorkspace(t)
	agent := workspace.GetMockContext()
	
	testCases := []testutil.ToolTestCase{
		{
			Name:        "EmptyPath",
			Input:       schemas.DeleteFileInput{Path: "", Recursive: false},
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

func TestDeleteFileTool_Execute_NestedFiles(t *testing.T) {
	tool := NewDeleteFileTool()
	workspace := testutil.NewTestWorkspace(t)
	agent := workspace.GetMockContext()
	
	// Create nested file structure
	workspace.CreateFile("dir1/dir2/dir3/deep.txt", "Deep file")
	
	input := schemas.DeleteFileInput{
		Path:      "dir1/dir2/dir3/deep.txt",
		Recursive: false,
	}
	
	inputBytes, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal input: %v", err)
	}
	
	result, err := tool.Execute(context.Background(), agent, json.RawMessage(inputBytes))
	if err != nil {
		t.Errorf("Delete nested file failed: %v", err)
	}
	
	if workspace.FileExists("dir1/dir2/dir3/deep.txt") {
		t.Error("Nested file should not exist after deletion")
	}
	
	// Parent directories should still exist
	if !workspace.FileExists("dir1/dir2/dir3") {
		t.Error("Parent directory should still exist")
	}
	
	if result == "" {
		t.Error("Result should not be empty")
	}
}

func TestDeleteFileTool_Execute_SpecialCharacters(t *testing.T) {
	tool := NewDeleteFileTool()
	workspace := testutil.NewTestWorkspace(t)
	agent := workspace.GetMockContext()
	
	specialFiles := []string{
		"file with spaces.txt",
		"file-with-dashes.txt",
		"file_with_underscores.txt",
		"file.with.dots.txt",
		"file123numbers.txt",
	}
	
	for _, filename := range specialFiles {
		t.Run("SpecialChar_"+filename, func(t *testing.T) {
			workspace.CreateFile(filename, "Content")
			
			input := schemas.DeleteFileInput{
				Path:      filename,
				Recursive: false,
			}
			
			inputBytes, err := json.Marshal(input)
			if err != nil {
				t.Fatalf("Failed to marshal input: %v", err)
			}
			
			_, err = tool.Execute(context.Background(), agent, json.RawMessage(inputBytes))
			if err != nil {
				t.Errorf("Delete failed for %s: %v", filename, err)
			}
			
			if workspace.FileExists(filename) {
				t.Errorf("File %s should not exist after deletion", filename)
			}
		})
	}
}
