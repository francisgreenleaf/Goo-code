package file

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"anthropic-chat/tools"
	"anthropic-chat/tools/schemas"

	"github.com/anthropics/anthropic-sdk-go"
)

// SearchFilesTool implements the search_files tool
type SearchFilesTool struct{}

// NewSearchFilesTool creates a new SearchFiles tool instance
func NewSearchFilesTool() *SearchFilesTool {
	return &SearchFilesTool{}
}

// Name returns the tool name
func (t *SearchFilesTool) Name() string {
	return "search_files"
}

// Description returns the tool description
func (t *SearchFilesTool) Description() string {
	return "Perform regex search across files in a directory with context-rich results. Supports Rust regex syntax and file pattern filtering."
}

// InputSchema returns the input schema for this tool
func (t *SearchFilesTool) InputSchema() anthropic.ToolInputSchemaParam {
	return schemas.SearchFilesInputSchema
}

// Execute performs the file search operation
func (t *SearchFilesTool) Execute(ctx context.Context, agent tools.ToolContext, input json.RawMessage) (string, error) {
	var searchInput schemas.SearchFilesInput
	if err := json.Unmarshal(input, &searchInput); err != nil {
		return "", fmt.Errorf("failed to parse input: %w", err)
	}

	// Resolve the search path using the agent's security validation
	searchPath, err := agent.ResolveFilePath(searchInput.Path)
	if err != nil {
		return "", err
	}

	// Validate path is a directory
	fileInfo, err := os.Stat(searchPath)
	if err != nil {
		return "", fmt.Errorf("failed to access directory %s: %w", searchInput.Path, err)
	}

	if !fileInfo.IsDir() {
		return "", fmt.Errorf("path is not a directory: %s", searchInput.Path)
	}

	// Compile regex pattern
	regex, err := regexp.Compile(searchInput.Regex)
	if err != nil {
		return "", fmt.Errorf("invalid regex pattern: %w", err)
	}

	// Walk directory and search files
	var results []string
	err = filepath.WalkDir(searchPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Apply file pattern filtering if provided
		if searchInput.FilePattern != "" {
			matched, err := filepath.Match(searchInput.FilePattern, d.Name())
			if err != nil {
				return err
			}
			if !matched {
				return nil
			}
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}

		// Search for matches
		matches := regex.FindAllStringIndex(string(content), -1)
		if len(matches) == 0 {
			return nil
		}

		// Add each match with context
		for i, match := range matches {
			start := match[0]
			lineNumber := countLinesBefore(content, start)

			// Extract context (3 lines before and after match)
			lines := strings.Split(string(content), "\n")
			startLine := max(0, lineNumber-3)
			endLine := min(len(lines), lineNumber+3)

			contextLines := lines[startLine:endLine]
			relativePath, _ := filepath.Rel(searchPath, path)

			results = append(results, fmt.Sprintf(
				"[FILE: %s] [LINE: %d] [MATCH %d/%d]:\n%s",
				relativePath,
				lineNumber+1,
				i+1,
				len(matches),
				strings.Join(contextLines, "\n"),
			))
		}

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("search failed: %w", err)
	}

	if len(results) == 0 {
		return "No matches found", nil
	}

	return strings.Join(results, "\n\n"), nil
}

// countLinesBefore counts the number of lines before a given position
func countLinesBefore(content []byte, position int) int {
	count := 0
	for i := 0; i < position; i++ {
		if content[i] == '\n' {
			count++
		}
	}
	return count
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
