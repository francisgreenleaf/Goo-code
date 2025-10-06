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

// CreateDirectoryTool implements the create_directory tool
type CreateDirectoryTool struct{}

// NewCreateDirectoryTool creates a new CreateDirectory tool instance
func NewCreateDirectoryTool() *CreateDirectoryTool {
	return &CreateDirectoryTool{}
}

// Name returns the tool name
func (t *CreateDirectoryTool) Name() string {
	return "create_directory"
}

// Description returns the tool description
func (t *CreateDirectoryTool) Description() string {
	return "Create a new directory or ensure a directory exists. Can create multiple nested directories in one operation. If the directory already exists, this operation will succeed silently."
}

// InputSchema returns the input schema for this tool
func (t *CreateDirectoryTool) InputSchema() anthropic.ToolInputSchemaParam {
	return schemas.CreateDirectoryInputSchema
}

// Execute performs the directory creation
func (t *CreateDirectoryTool) Execute(ctx context.Context, agent tools.ToolContext, input json.RawMessage) (string, error) {
	var dirInput schemas.CreateDirectoryInput
	if err := json.Unmarshal(input, &dirInput); err != nil {
		return "", fmt.Errorf("failed to parse input: %w", err)
	}

	// Resolve the directory path using the agent's security validation
	dirPath, err := agent.ResolveFilePath(dirInput.Path)
	if err != nil {
		return "", err
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", dirInput.Path, err)
	}

	return fmt.Sprintf("Directory created successfully: %s", dirInput.Path), nil
}
