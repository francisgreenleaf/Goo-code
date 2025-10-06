package file

import (
	"context"
	"encoding/json"
	"testing"

	"anthropic-chat/tools/schemas"
	"anthropic-chat/tools/testutil"
)

func TestMoveFileTool_Interface(t *testing.T) {
	tool := NewMoveFileTool()
	wrappedTool := testutil.WrapMainTool(tool)
	
	testutil.ValidateToolInterface(t, wrappedTool)
	
	if tool.Name() != "move_file" {
		t.Errorf("Expected name 'move_file', got '%s'", tool.Name())
	}
	
	expectedDesc := "Move or rename a file or directory within the working directory."
	if tool.Description() != expectedDesc {
		t.Errorf("Expected description '%s', got '%s'", expectedDesc, tool.Description())
	}
}

func TestMoveFileTool_Schema(t *testing.T) {
	tool := NewMoveFileTool()
	schema := tool.InputSchema()
	
	testutil.ValidateSchema(t, schema, "MoveFileTool")
	
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
	
	// Verify required properties exist
	requiredProps := []string{"source", "destination"}
	for _, prop := range requiredProps {
		if _, exists := properties[prop]; !exists {
			t.Errorf("Schema should have '%s' property", prop)
		}
	}
}

func TestMoveFileTool_Execute_RenameFile(t *testing.T) {
	tool := NewMoveFileTool()
	workspace := testutil.NewTestWorkspace(t)
	agent := workspace.GetMockContext()
	
	// Create a test file
	originalContent := "Test file content"
	workspace.CreateFile("original.txt", originalContent)
	
	input := schemas.MoveFileInput{
		Source:      "original.txt",
		Destination: "renamed.txt",
	}
	
	inputBytes, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal input: %v", err)
	}
	
	result, err := tool.Execute(context.Background(), agent, json.RawMessage(inputBytes))
	if err != nil {
		t.Errorf("Move failed: %v", err)
	}
	
	// Verify original file is gone
	if workspace.FileExists("original.txt") {
		t.Error("Original file should not exist after move")
	}
	
	// Verify new file exists with same content
	if !workspace.FileExists("renamed.txt") {
		t.Error("Renamed file should exist")
	}
	
	newContent := workspace.ReadFile("renamed.txt")
	if newContent != originalContent {
		t.Errorf("Content mismatch after rename. Expected: %q, Got: %q", originalContent, newContent)
	}
	
	if result == "" {
		t.Error("Result should not be empty")
	}
}

func TestMoveFileTool_Execute_MoveToDirectory(t *testing.T) {
	tool := NewMoveFileTool()
	workspace := testutil.NewTestWorkspace(t)
	agent := workspace.GetMockContext()
	
	// Create file and directory
	originalContent := "File to move"
	workspace.CreateFile("file.txt", originalContent)
	workspace.CreateDir("subdir")
	
	input := schemas.MoveFileInput{
		Source:      "file.txt",
		Destination: "subdir/file.txt",
	}
	
	inputBytes, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal input: %v", err)
	}
	
	result, err := tool.Execute(context.Background(), agent, json.RawMessage(inputBytes))
	if err != nil {
		t.Errorf("Move failed: %v", err)
	}
	
	// Verify original is gone
	if workspace.FileExists("file.txt") {
		t.Error("Original file should not exist")
	}
	
	// Verify file in new location
	if !workspace.FileExists("subdir/file.txt") {
		t.Error("File should exist in new location")
	}
	
	newContent := workspace.ReadFile("subdir/file.txt")
	if newContent != originalContent {
		t.Errorf("Content mismatch. Expected: %q, Got: %q", originalContent, newContent)
	}
	
	if result == "" {
		t.Error("Result should not be empty")
	}
}

func TestMoveFileTool_Execute_MoveDirectory(t *testing.T) {
	tool := NewMoveFileTool()
	workspace := testutil.NewTestWorkspace(t)
	agent := workspace.GetMockContext()
	
	// Create directory with files
	workspace.CreateDir("olddir")
	workspace.CreateFile("olddir/file1.txt", "Content 1")
	workspace.CreateFile("olddir/file2.txt", "Content 2")
	
	input := schemas.MoveFileInput{
		Source:      "olddir",
		Destination: "newdir",
	}
	
	inputBytes, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal input: %v", err)
	}
	
	result, err := tool.Execute(context.Background(), agent, json.RawMessage(inputBytes))
	if err != nil {
		t.Errorf("Move directory failed: %v", err)
	}
	
	// Verify old directory is gone
	if workspace.FileExists("olddir") {
		t.Error("Old directory should not exist")
	}
	
	// Verify new directory and contents
	if !workspace.FileExists("newdir/file1.txt") {
		t.Error("File1 should exist in new directory")
	}
	if !workspace.FileExists("newdir/file2.txt") {
		t.Error("File2 should exist in new directory")
	}
	
	if result == "" {
		t.Error("Result should not be empty")
	}
}

func TestMoveFileTool_Execute_SourceNotFound(t *testing.T) {
	tool := NewMoveFileTool()
	wrappedTool := testutil.WrapMainTool(tool)
	workspace := testutil.NewTestWorkspace(t)
	agent := workspace.GetMockContext()
	
	input := schemas.MoveFileInput{
		Source:      "nonexistent.txt",
		Destination: "destination.txt",
	}
	
	testutil.ValidateToolExecutionError(t, wrappedTool, agent, input, "")
}

func TestMoveFileTool_Execute_DestinationExists(t *testing.T) {
	tool := NewMoveFileTool()
	workspace := testutil.NewTestWorkspace(t)
	agent := workspace.GetMockContext()
	
	// Create both source and destination
	workspace.CreateFile("source.txt", "Source content")
	workspace.CreateFile("destination.txt", "Destination content")
	
	input := schemas.MoveFileInput{
		Source:      "source.txt",
		Destination: "destination.txt",
	}
	
	inputBytes, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal input: %v", err)
	}
	
	_, err = tool.Execute(context.Background(), agent, json.RawMessage(inputBytes))
	if err == nil {
		t.Error("Should error when destination exists")
	}
}

func TestMoveFileTool_Execute_PathTraversal(t *testing.T) {
	tool := NewMoveFileTool()
	wrappedTool := testutil.WrapMainTool(tool)
	workspace := testutil.NewTestWorkspace(t)
	agent := workspace.GetMockContext()
	
	workspace.CreateFile("test.txt", "content")
	
	pathTraversalTests := []struct {
		name        string
		source      string
		destination string
	}{
		{"SourceTraversal", "../../../etc/passwd", "dest.txt"},
		{"DestinationTraversal", "test.txt", "../../../tmp/evil.txt"},
		{"BothTraversal", "../source.txt", "../dest.txt"},
		{"SubdirTraversal", "test.txt", "subdir/../../outside.txt"},
	}
	
	for _, tt := range pathTraversalTests {
		t.Run(tt.name, func(t *testing.T) {
			input := schemas.MoveFileInput{
				Source:      tt.source,
				Destination: tt.destination,
			}
			
			testutil.ValidateToolExecutionError(t, wrappedTool, agent, input, "security reasons")
		})
	}
}

func TestMoveFileTool_Execute_NestedPaths(t *testing.T) {
	tool := NewMoveFileTool()
	workspace := testutil.NewTestWorkspace(t)
	agent := workspace.GetMockContext()
	
	// Create nested structure
	workspace.CreateFile("dir1/dir2/file.txt", "Nested content")
	workspace.CreateDir("dir3/dir4")
	
	input := schemas.MoveFileInput{
		Source:      "dir1/dir2/file.txt",
		Destination: "dir3/dir4/file.txt",
	}
	
	inputBytes, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal input: %v", err)
	}
	
	result, err := tool.Execute(context.Background(), agent, json.RawMessage(inputBytes))
	if err != nil {
		t.Errorf("Move nested file failed: %v", err)
	}
	
	if workspace.FileExists("dir1/dir2/file.txt") {
		t.Error("Original file should not exist")
	}
	
	if !workspace.FileExists("dir3/dir4/file.txt") {
		t.Error("File should exist in new location")
	}
	
	if result == "" {
		t.Error("Result should not be empty")
	}
}

func TestMoveFileTool_Execute_SpecialCharacters(t *testing.T) {
	tool := NewMoveFileTool()
	workspace := testutil.NewTestWorkspace(t)
	agent := workspace.GetMockContext()
	
	specialFiles := map[string]string{
		"file with spaces.txt":    "renamed with spaces.txt",
		"file-with-dashes.txt":    "new-name-dashes.txt",
		"file_underscores.txt":    "new_underscores.txt",
		"file.multiple.dots.txt":  "new.dots.txt",
	}
	
	for source, dest := range specialFiles {
		t.Run("SpecialChar_"+source, func(t *testing.T) {
			content := "Content for " + source
			workspace.CreateFile(source, content)
			
			input := schemas.MoveFileInput{
				Source:      source,
				Destination: dest,
			}
			
			inputBytes, err := json.Marshal(input)
			if err != nil {
				t.Fatalf("Failed to marshal input: %v", err)
			}
			
			_, err = tool.Execute(context.Background(), agent, json.RawMessage(inputBytes))
			if err != nil {
				t.Errorf("Move failed for %s: %v", source, err)
			}
			
			if workspace.FileExists(source) {
				t.Errorf("Original file %s should not exist", source)
			}
			
			if !workspace.FileExists(dest) {
				t.Errorf("Destination file %s should exist", dest)
			}
			
			newContent := workspace.ReadFile(dest)
			if newContent != content {
				t.Errorf("Content mismatch for %s", dest)
			}
		})
	}
}

func TestMoveFileTool_Execute_InvalidInput(t *testing.T) {
	tool := NewMoveFileTool()
	wrappedTool := testutil.WrapMainTool(tool)
	workspace := testutil.NewTestWorkspace(t)
	agent := workspace.GetMockContext()
	
	testCases := []testutil.ToolTestCase{
		{
			Name:        "EmptySource",
			Input:       schemas.MoveFileInput{Source: "", Destination: "dest.txt"},
			ShouldError: true,
		},
		{
			Name:        "EmptyDestination",
			Input:       schemas.MoveFileInput{Source: "source.txt", Destination: ""},
			ShouldError: true,
		},
		{
			Name:        "BothEmpty",
			Input:       schemas.MoveFileInput{Source: "", Destination: ""},
			ShouldError: true,
		},
		{
			Name:        "SameSourceAndDestination",
			Input:       schemas.MoveFileInput{Source: "file.txt", Destination: "file.txt"},
			ShouldError: true,
		},
	}
	
	testutil.TestToolWithAllInputs(t, wrappedTool, agent, testCases)
}

func TestMoveFileTool_Execute_CreateDestinationDirectory(t *testing.T) {
	tool := NewMoveFileTool()
	workspace := testutil.NewTestWorkspace(t)
	agent := workspace.GetMockContext()
	
	workspace.CreateFile("file.txt", "Content")
	
	// Destination directory doesn't exist yet
	input := schemas.MoveFileInput{
		Source:      "file.txt",
		Destination: "newdir/subdir/file.txt",
	}
	
	inputBytes, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal input: %v", err)
	}
	
	result, err := tool.Execute(context.Background(), agent, json.RawMessage(inputBytes))
	if err != nil {
		t.Errorf("Move should create destination directories: %v", err)
	}
	
	if !workspace.FileExists("newdir/subdir/file.txt") {
		t.Error("File should exist in new location with created directories")
	}
	
	if result == "" {
		t.Error("Result should not be empty")
	}
}
