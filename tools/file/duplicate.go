package file

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"anthropic-chat/tools"
	"anthropic-chat/tools/schemas"

	"github.com/anthropics/anthropic-sdk-go"
)

// DuplicateFileTool implements the duplicate_file tool
type DuplicateFileTool struct{}

// NewDuplicateFileTool creates a new DuplicateFile tool instance
func NewDuplicateFileTool() *DuplicateFileTool {
	return &DuplicateFileTool{}
}

// Name returns the tool name
func (t *DuplicateFileTool) Name() string {
	return "duplicate_file"
}

// Description returns the tool description
func (t *DuplicateFileTool) Description() string {
	return "Duplicate a file with [filename](1) naming pattern. Creates a copy of the file with automatic naming that adds '(1)' before the file extension."
}

// InputSchema returns the input schema for this tool
func (t *DuplicateFileTool) InputSchema() anthropic.ToolInputSchemaParam {
	return schemas.DuplicateFileInputSchema
}

// Execute performs the duplicate file operation
func (t *DuplicateFileTool) Execute(ctx context.Context, agent tools.ToolContext, input json.RawMessage) (string, error) {
	var duplicateInput schemas.DuplicateFileInput
	if err := json.Unmarshal(input, &duplicateInput); err != nil {
		return "", fmt.Errorf("failed to parse input: %w", err)
	}

	// Resolve the source file path using the agent's security validation
	sourcePath, err := agent.ResolveFilePath(duplicateInput.Path)
	if err != nil {
		return "", err
	}

	// Check if source file exists
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("source file does not exist: %s", duplicateInput.Path)
		}
		return "", fmt.Errorf("failed to access source file %s: %w", duplicateInput.Path, err)
	}

	if sourceInfo.IsDir() {
		return "", fmt.Errorf("source path is a directory, not a file: %s", duplicateInput.Path)
	}

	// Generate destination filename with (1) pattern
	dir := filepath.Dir(sourcePath)
	filename := filepath.Base(duplicateInput.Path)
	ext := filepath.Ext(filename)
	nameWithoutExt := strings.TrimSuffix(filename, ext)

	destinationFilename := fmt.Sprintf("%s(1)%s", nameWithoutExt, ext)
	destinationPath := filepath.Join(dir, destinationFilename)

	// Copy the file
	sourceContent, err := os.ReadFile(sourcePath)
	if err != nil {
		return "", fmt.Errorf("failed to read source file %s: %w", duplicateInput.Path, err)
	}

	if err := os.WriteFile(destinationPath, sourceContent, sourceInfo.Mode()); err != nil {
		return "", fmt.Errorf("failed to write destination file %s: %w", destinationFilename, err)
	}

	return fmt.Sprintf("Successfully duplicated %s to %s", duplicateInput.Path, destinationFilename), nil
}
