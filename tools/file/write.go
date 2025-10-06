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

// WriteFileTool implements the write_file tool
type WriteFileTool struct{}

// NewWriteFileTool creates a new WriteFile tool instance
func NewWriteFileTool() *WriteFileTool {
	return &WriteFileTool{}
}

// Name returns the tool name
func (t *WriteFileTool) Name() string {
	return "write_file"
}

// Description returns the tool description
func (t *WriteFileTool) Description() string {
	return "Create or overwrite a file with complete content. Automatically creates directories if needed."
}

// InputSchema returns the input schema for this tool
func (t *WriteFileTool) InputSchema() anthropic.ToolInputSchemaParam {
	return schemas.WriteFileInputSchema
}

// Execute performs the file writing operation
func (t *WriteFileTool) Execute(ctx context.Context, agent tools.ToolContext, input json.RawMessage) (string, error) {
	var writeInput schemas.WriteFileInput
	if err := json.Unmarshal(input, &writeInput); err != nil {
		return "", fmt.Errorf("failed to parse input: %w", err)
	}

	// Resolve the file path using the agent's security validation
	filePath, err := agent.ResolveFilePath(writeInput.Path)
	if err != nil {
		return "", err
	}

	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return "", fmt.Errorf("failed to create directory for %s: %w", writeInput.Path, err)
	}

	// Write content to file
	if err := os.WriteFile(filePath, []byte(writeInput.Content), 0644); err != nil {
		return "", fmt.Errorf("failed to write file %s: %w", writeInput.Path, err)
	}

	return fmt.Sprintf("File written successfully: %s", writeInput.Path), nil
}
