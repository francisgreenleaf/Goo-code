package file

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"anthropic-chat/tools"
	"anthropic-chat/tools/schemas"

	"github.com/anthropics/anthropic-sdk-go"
)

// MoveFileTool implements the move_file tool
type MoveFileTool struct{}

// NewMoveFileTool creates a new MoveFile tool instance
func NewMoveFileTool() *MoveFileTool {
	return &MoveFileTool{}
}

// Name returns the tool name
func (t *MoveFileTool) Name() string {
	return "move_file"
}

// Description returns the tool description
func (t *MoveFileTool) Description() string {
	return "Move or rename a file or directory within the working directory."
}

// InputSchema returns the input schema for this tool
func (t *MoveFileTool) InputSchema() anthropic.ToolInputSchemaParam {
	return schemas.MoveFileInputSchema
}

// Execute performs the move file operation
func (t *MoveFileTool) Execute(ctx context.Context, agent tools.ToolContext, input json.RawMessage) (string, error) {
	var moveInput schemas.MoveFileInput
	if err := json.Unmarshal(input, &moveInput); err != nil {
		return "", fmt.Errorf("failed to parse input: %w", err)
	}

	// Validate inputs
	if moveInput.Source == "" {
		return "", fmt.Errorf("source path cannot be empty")
	}
	if moveInput.Destination == "" {
		return "", fmt.Errorf("destination path cannot be empty")
	}
	if moveInput.Source == moveInput.Destination {
		return "", fmt.Errorf("source and destination cannot be the same")
	}

	// Resolve both paths using the agent's security validation
	sourcePath, err := agent.ResolveFilePath(moveInput.Source)
	if err != nil {
		return "", fmt.Errorf("invalid source path: %w", err)
	}

	destPath, err := agent.ResolveFilePath(moveInput.Destination)
	if err != nil {
		return "", fmt.Errorf("invalid destination path: %w", err)
	}

	// Check if source exists
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("source file or directory does not exist: %s", moveInput.Source)
		}
		return "", fmt.Errorf("failed to access source: %w", err)
	}

	// Check if destination already exists
	if _, err := os.Stat(destPath); err == nil {
		return "", fmt.Errorf("destination already exists: %s", moveInput.Destination)
	}

	// Create destination directory if it doesn't exist
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Perform the move operation
	if err := os.Rename(sourcePath, destPath); err != nil {
		return "", fmt.Errorf("failed to move %s to %s: %w", moveInput.Source, moveInput.Destination, err)
	}

	// Build result message
	itemType := "file"
	if sourceInfo.IsDir() {
		itemType = "directory"
	}

	return fmt.Sprintf("Successfully moved %s from %s to %s", itemType, moveInput.Source, moveInput.Destination), nil
}
