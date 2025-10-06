package file

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"anthropic-chat/tools"
	"anthropic-chat/tools/schemas"

	"github.com/anthropics/anthropic-sdk-go"
)

// DeleteFileTool implements the delete_file tool
type DeleteFileTool struct{}

// NewDeleteFileTool creates a new DeleteFile tool instance
func NewDeleteFileTool() *DeleteFileTool {
	return &DeleteFileTool{}
}

// Name returns the tool name
func (t *DeleteFileTool) Name() string {
	return "delete_file"
}

// Description returns the tool description
func (t *DeleteFileTool) Description() string {
	return "Delete a file or directory within the working directory."
}

// InputSchema returns the input schema for this tool
func (t *DeleteFileTool) InputSchema() anthropic.ToolInputSchemaParam {
	return schemas.DeleteFileInputSchema
}

// Execute performs the delete file operation
func (t *DeleteFileTool) Execute(ctx context.Context, agent tools.ToolContext, input json.RawMessage) (string, error) {
	var deleteInput schemas.DeleteFileInput
	if err := json.Unmarshal(input, &deleteInput); err != nil {
		return "", fmt.Errorf("failed to parse input: %w", err)
	}

	// Validate input
	if deleteInput.Path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	// Resolve the path using the agent's security validation
	fullPath, err := agent.ResolveFilePath(deleteInput.Path)
	if err != nil {
		return "", err
	}

	// Check if file/directory exists
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("file or directory does not exist: %s", deleteInput.Path)
		}
		return "", fmt.Errorf("failed to access path: %w", err)
	}

	// Determine what we're deleting
	itemType := "file"
	if fileInfo.IsDir() {
		itemType = "directory"
		
		// Check if directory is empty
		entries, err := os.ReadDir(fullPath)
		if err != nil {
			return "", fmt.Errorf("failed to read directory: %w", err)
		}
		
		// If directory is not empty and recursive is false, error
		if len(entries) > 0 && !deleteInput.Recursive {
			return "", fmt.Errorf("directory is not empty. Use recursive=true to delete non-empty directories")
		}
	}

	// Perform the deletion
	if deleteInput.Recursive && fileInfo.IsDir() {
		// Use RemoveAll for recursive deletion
		if err := os.RemoveAll(fullPath); err != nil {
			return "", fmt.Errorf("failed to recursively delete %s: %w", deleteInput.Path, err)
		}
	} else {
		// Use Remove for single file or empty directory
		if err := os.Remove(fullPath); err != nil {
			return "", fmt.Errorf("failed to delete %s: %w", deleteInput.Path, err)
		}
	}

	return fmt.Sprintf("Successfully deleted %s: %s", itemType, deleteInput.Path), nil
}
