package file

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"anthropic-chat/tools"
	"anthropic-chat/tools/schemas"

	"github.com/anthropics/anthropic-sdk-go"
)

// FindFilesTool implements the find_files tool
type FindFilesTool struct{}

// NewFindFilesTool creates a new FindFiles tool instance
func NewFindFilesTool() *FindFilesTool {
	return &FindFilesTool{}
}

// Name returns the tool name
func (t *FindFilesTool) Name() string {
	return "find_files"
}

// Description returns the tool description
func (t *FindFilesTool) Description() string {
	return "Find files and directories matching specified criteria within the working directory."
}

// InputSchema returns the input schema for this tool
func (t *FindFilesTool) InputSchema() anthropic.ToolInputSchemaParam {
	return schemas.FindFilesInputSchema
}

// Execute performs the find files operation
func (t *FindFilesTool) Execute(ctx context.Context, agent tools.ToolContext, input json.RawMessage) (string, error) {
	var findInput schemas.FindFilesInput
	if err := json.Unmarshal(input, &findInput); err != nil {
		return "", fmt.Errorf("failed to parse input: %w", err)
	}

	// Validate input
	if findInput.Pattern == "" {
		return "", fmt.Errorf("pattern cannot be empty")
	}

	// Check for path traversal in pattern
	if strings.Contains(findInput.Pattern, "..") {
		return "", fmt.Errorf("pattern cannot contain '..' for security reasons")
	}

	// Set default limit if not specified
	limit := findInput.Limit
	if limit <= 0 {
		limit = 100
	}

	workingDir := agent.WorkingDir()
	var matches []string
	count := 0

	if findInput.Recursive {
		// Recursive search
		err := filepath.WalkDir(workingDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil // Skip errors and continue
			}

			// Get relative path
			relPath, err := filepath.Rel(workingDir, path)
			if err != nil {
				return nil
			}

			// Skip the root directory itself
			if relPath == "." {
				return nil
			}

			// Check if the file matches the pattern
			matched, err := filepath.Match(findInput.Pattern, filepath.Base(path))
			if err != nil {
				return nil // Skip invalid patterns
			}

			if matched {
				if count >= limit {
					return filepath.SkipAll // Stop searching once limit is reached
				}
				
				itemType := "file"
				if d.IsDir() {
					itemType = "dir"
				}
				matches = append(matches, fmt.Sprintf("%s (%s)", relPath, itemType))
				count++
			}

			return nil
		})

		if err != nil {
			return "", fmt.Errorf("failed to search files: %w", err)
		}
	} else {
		// Non-recursive search (top-level only)
		entries, err := filepath.Glob(filepath.Join(workingDir, findInput.Pattern))
		if err != nil {
			return "", fmt.Errorf("invalid pattern: %w", err)
		}

		for _, entry := range entries {
			if count >= limit {
				break
			}

			relPath, err := filepath.Rel(workingDir, entry)
			if err != nil {
				continue
			}

			// Use os.Stat to check if it's a directory
			stat, err := os.Stat(entry)
			if err != nil {
				continue
			}

			itemType := "file"
			if stat.IsDir() {
				itemType = "dir"
			}
			
			matches = append(matches, fmt.Sprintf("%s (%s)", relPath, itemType))
			count++
		}
	}

	// Build result
	if len(matches) == 0 {
		return "No files found matching pattern: " + findInput.Pattern, nil
	}

	result := fmt.Sprintf("Found %d file(s) matching pattern '%s':\n", len(matches), findInput.Pattern)
	for _, match := range matches {
		result += "  " + match + "\n"
	}

	if count >= limit {
		result += fmt.Sprintf("\n(Limited to %d results. There may be more matches.)", limit)
	}

	return result, nil
}
