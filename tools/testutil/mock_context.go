package testutil

import (
	"fmt"
	"path/filepath"
	"strings"
)

// MockToolContext implements the ToolContext interface for testing
type MockToolContext struct {
	WorkingDirPath      string
	ResolveFilePathFunc func(string) (string, error)
	// Additional fields for controlling behavior
	AllowPathTraversal bool
	FileSystemRoot     string
}

// NewMockToolContext creates a new mock tool context with default behavior
func NewMockToolContext(workingDir string) *MockToolContext {
	return &MockToolContext{
		WorkingDirPath: workingDir,
		FileSystemRoot: workingDir,
	}
}

// WorkingDir returns the configured working directory
func (m *MockToolContext) WorkingDir() string {
	return m.WorkingDirPath
}

// ResolveFilePath implements path resolution with configurable behavior
func (m *MockToolContext) ResolveFilePath(relativePath string) (string, error) {
	// Use custom function if provided
	if m.ResolveFilePathFunc != nil {
		return m.ResolveFilePathFunc(relativePath)
	}

	// Default implementation mimics the real agent's behavior
	cleanPath := filepath.Clean(relativePath)

	// Check for path traversal unless explicitly allowed
	if !m.AllowPathTraversal && strings.Contains(cleanPath, "..") {
		return "", fmt.Errorf("path cannot contain '..' for security reasons")
	}

	// Join with working directory
	fullPath := filepath.Join(m.WorkingDirPath, cleanPath)

	// Ensure the resolved path is still within the working directory
	absWorkingDir, err := filepath.Abs(m.WorkingDirPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute working directory: %w", err)
	}

	absFullPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	if !m.AllowPathTraversal && !strings.HasPrefix(absFullPath, absWorkingDir) {
		return "", fmt.Errorf("path escapes working directory")
	}

	return fullPath, nil
}

// WithCustomResolver sets a custom path resolution function
func (m *MockToolContext) WithCustomResolver(fn func(string) (string, error)) *MockToolContext {
	m.ResolveFilePathFunc = fn
	return m
}

// WithPathTraversalAllowed configures whether path traversal is allowed
func (m *MockToolContext) WithPathTraversalAllowed(allowed bool) *MockToolContext {
	m.AllowPathTraversal = allowed
	return m
}
