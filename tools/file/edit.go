package file

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"anthropic-chat/tools"
	"anthropic-chat/tools/schemas"

	"github.com/anthropics/anthropic-sdk-go"
)

// EditFileTool implements the edit_file tool
type EditFileTool struct{}

// NewEditFileTool creates a new EditFile tool instance
func NewEditFileTool() *EditFileTool {
	return &EditFileTool{}
}

// Name returns the tool name
func (t *EditFileTool) Name() string {
	return "edit_file"
}

// Description returns the tool description
func (t *EditFileTool) Description() string {
	return "Edit files with insert, delete, replace, append operations. Supports line-based positioning and multiple operations per call."
}

// InputSchema returns the input schema for this tool
func (t *EditFileTool) InputSchema() anthropic.ToolInputSchemaParam {
	return schemas.EditFileInputSchema
}

// Execute performs the edit file operation
func (t *EditFileTool) Execute(ctx context.Context, agent tools.ToolContext, input json.RawMessage) (string, error) {
	var editInput schemas.EditFileInput
	if err := json.Unmarshal(input, &editInput); err != nil {
		return "", fmt.Errorf("failed to parse input: %w", err)
	}

	// Resolve the file path using the agent's security validation
	fullPath, err := agent.ResolveFilePath(editInput.Path)
	if err != nil {
		return "", err
	}

	// Read the current file content
	var lines []string
	if _, err := os.Stat(fullPath); err == nil {
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return "", fmt.Errorf("failed to read file %s: %w", editInput.Path, err)
		}
		lines = strings.Split(string(content), "\n")
	} else if os.IsNotExist(err) {
		// File doesn't exist, create empty content
		lines = []string{}
	} else {
		return "", fmt.Errorf("failed to check file %s: %w", editInput.Path, err)
	}

	// Apply each operation in order
	for i, op := range editInput.Operations {
		var opErr error
		lines, opErr = t.applyOperation(lines, op)
		if opErr != nil {
			return "", fmt.Errorf("operation %d failed: %w", i+1, opErr)
		}
	}

	// Write the modified content back to the file
	newContent := strings.Join(lines, "\n")
	if err := os.WriteFile(fullPath, []byte(newContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write file %s: %w", editInput.Path, err)
	}

	return fmt.Sprintf("Successfully applied %d operation(s) to %s", len(editInput.Operations), editInput.Path), nil
}

// applyOperation applies a single edit operation to the lines
func (t *EditFileTool) applyOperation(lines []string, op schemas.EditOperation) ([]string, error) {
	switch op.Type {
	case "insert":
		return t.insertOperation(lines, op)
	case "delete":
		return t.deleteOperation(lines, op)
	case "replace":
		return t.replaceOperation(lines, op)
	case "append":
		return t.appendOperation(lines, op)
	default:
		return lines, fmt.Errorf("unknown operation type: %s", op.Type)
	}
}

// insertOperation inserts content at the specified line
func (t *EditFileTool) insertOperation(lines []string, op schemas.EditOperation) ([]string, error) {
	if op.Line == nil {
		return lines, fmt.Errorf("insert operation requires line number")
	}

	lineNum := *op.Line
	if lineNum < 1 {
		return lines, fmt.Errorf("line number must be >= 1")
	}

	// Convert to 0-based index
	insertIndex := lineNum - 1

	// Handle inserting beyond the end of file
	if insertIndex > len(lines) {
		// Pad with empty lines if necessary
		for len(lines) < insertIndex {
			lines = append(lines, "")
		}
	}

	// Split content by newlines
	contentLines := strings.Split(op.Content, "\n")

	// Insert the content lines
	if insertIndex >= len(lines) {
		lines = append(lines, contentLines...)
	} else {
		// Insert at the specified position
		newLines := make([]string, 0, len(lines)+len(contentLines))
		newLines = append(newLines, lines[:insertIndex]...)
		newLines = append(newLines, contentLines...)
		newLines = append(newLines, lines[insertIndex:]...)
		lines = newLines
	}

	return lines, nil
}

// deleteOperation deletes lines in the specified range
func (t *EditFileTool) deleteOperation(lines []string, op schemas.EditOperation) ([]string, error) {
	if op.Line == nil {
		return lines, fmt.Errorf("delete operation requires line number")
	}

	startLine := *op.Line
	endLine := startLine
	if op.EndLine != nil {
		endLine = *op.EndLine
	}

	if startLine < 1 || endLine < 1 {
		return lines, fmt.Errorf("line numbers must be >= 1")
	}

	if startLine > endLine {
		return lines, fmt.Errorf("start line must be <= end line")
	}

	// Convert to 0-based indices
	startIndex := startLine - 1
	endIndex := endLine - 1

	if startIndex >= len(lines) {
		// Nothing to delete
		return lines, nil
	}

	// Ensure endIndex doesn't exceed bounds
	if endIndex >= len(lines) {
		endIndex = len(lines) - 1
	}

	// Delete the specified range
	newLines := make([]string, 0, len(lines)-(endIndex-startIndex+1))
	newLines = append(newLines, lines[:startIndex]...)
	newLines = append(newLines, lines[endIndex+1:]...)

	return newLines, nil
}

// replaceOperation replaces lines in the specified range with new content
func (t *EditFileTool) replaceOperation(lines []string, op schemas.EditOperation) ([]string, error) {
	if op.Line == nil {
		return lines, fmt.Errorf("replace operation requires line number")
	}

	startLine := *op.Line
	endLine := startLine
	if op.EndLine != nil {
		endLine = *op.EndLine
	}

	if startLine < 1 || endLine < 1 {
		return lines, fmt.Errorf("line numbers must be >= 1")
	}

	if startLine > endLine {
		return lines, fmt.Errorf("start line must be <= end line")
	}

	// Convert to 0-based indices
	startIndex := startLine - 1
	endIndex := endLine - 1

	// Handle replacing beyond the end of file
	if startIndex >= len(lines) {
		// Pad with empty lines if necessary
		for len(lines) <= startIndex {
			lines = append(lines, "")
		}
	}

	// Ensure endIndex doesn't exceed bounds
	if endIndex >= len(lines) {
		endIndex = len(lines) - 1
	}

	// Split content by newlines
	contentLines := strings.Split(op.Content, "\n")

	// Replace the specified range
	newLines := make([]string, 0, len(lines)-(endIndex-startIndex+1)+len(contentLines))
	newLines = append(newLines, lines[:startIndex]...)
	newLines = append(newLines, contentLines...)
	newLines = append(newLines, lines[endIndex+1:]...)

	return newLines, nil
}

// appendOperation appends content to the end of the file
func (t *EditFileTool) appendOperation(lines []string, op schemas.EditOperation) ([]string, error) {
	// Split content by newlines
	contentLines := strings.Split(op.Content, "\n")

	// Append to the end
	lines = append(lines, contentLines...)

	return lines, nil
}
