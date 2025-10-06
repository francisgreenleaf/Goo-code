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

// ListFilesTool implements the list_files tool
type ListFilesTool struct{}

// NewListFilesTool creates a new ListFiles tool instance
func NewListFilesTool() *ListFilesTool {
	return &ListFilesTool{}
}

// Name returns the tool name
func (t *ListFilesTool) Name() string {
	return "list_files"
}

// Description returns the tool description
func (t *ListFilesTool) Description() string {
	return "List files and directories at specified path with optional recursion, filtering, and limits to manage output size."
}

// InputSchema returns the input schema for this tool
func (t *ListFilesTool) InputSchema() anthropic.ToolInputSchemaParam {
	return schemas.ListFilesInputSchema
}

// Execute performs the list files operation
func (t *ListFilesTool) Execute(ctx context.Context, agent tools.ToolContext, input json.RawMessage) (string, error) {
	var listInput schemas.ListFilesInput
	if err := json.Unmarshal(input, &listInput); err != nil {
		return "", fmt.Errorf("failed to parse input: %w", err)
	}

	// Set defaults
	if listInput.Limit <= 0 {
		listInput.Limit = 100 // Default limit to prevent runaway output
	}

	// Determine directory to list
	dir := agent.WorkingDir()
	if listInput.Path != "" {
		var err error
		dir, err = agent.ResolveFilePath(listInput.Path)
		if err != nil {
			return "", err
		}
	}

	// Collect files and directories
	var files []string
	var err error

	if listInput.Recursive {
		files, err = t.listFilesRecursive(dir, listInput.Pattern, listInput.Limit)
	} else {
		files, err = t.listFilesTopLevel(dir, listInput.Pattern, listInput.Limit)
	}

	if err != nil {
		return "", fmt.Errorf("failed to list files in %s: %w", listInput.Path, err)
	}

	// Convert to JSON for output
	result, err := json.Marshal(files)
	if err != nil {
		return "", fmt.Errorf("failed to marshal file list: %w", err)
	}

	return string(result), nil
}

// listFilesTopLevel lists files in the specified directory only (non-recursive)
func (t *ListFilesTool) listFilesTopLevel(dir, pattern string, limit int) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if len(files) >= limit {
			break
		}

		name := entry.Name()

		// Apply pattern filtering if specified
		if pattern != "" {
			matched, err := filepath.Match(pattern, name)
			if err != nil {
				return nil, fmt.Errorf("invalid pattern %s: %w", pattern, err)
			}
			if !matched {
				continue
			}
		}

		// Add directory indicator
		if entry.IsDir() {
			files = append(files, name+"/")
		} else {
			files = append(files, name)
		}
	}

	return files, nil
}

// listFilesRecursive lists files recursively with limits and pattern filtering
func (t *ListFilesTool) listFilesRecursive(dir, pattern string, limit int) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Stop if we've hit the limit
		if len(files) >= limit {
			return filepath.SkipDir
		}

		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		// Skip the current directory entry
		if relPath == "." {
			return nil
		}

		// Apply pattern filtering if specified
		if pattern != "" {
			// For pattern matching, use just the filename, not the full path
			filename := filepath.Base(relPath)
			matched, err := filepath.Match(pattern, filename)
			if err != nil {
				return fmt.Errorf("invalid pattern %s: %w", pattern, err)
			}
			if !matched {
				// Skip this file/directory, but continue walking if it's a directory
				if info.IsDir() {
					return nil // Don't skip the directory traversal, just don't include it in results
				}
				return nil
			}
		}

		// Add to results with directory indicator
		if info.IsDir() {
			files = append(files, relPath+"/")
		} else {
			files = append(files, relPath)
		}

		return nil
	})

	// Handle early termination due to limit
	if err != nil && strings.Contains(err.Error(), "file name too long") {
		// This might happen in some edge cases, but we can still return what we found
		return files, nil
	}

	return files, err
}
