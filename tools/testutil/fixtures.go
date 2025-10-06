package testutil

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestWorkspace represents a temporary workspace for testing
type TestWorkspace struct {
	Dir     string
	t       *testing.T
	cleanup []func() error
}

// NewTestWorkspace creates a new temporary workspace for testing
func NewTestWorkspace(t *testing.T) *TestWorkspace {
	t.Helper()
	
	tempDir, err := os.MkdirTemp("", "goocode-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	workspace := &TestWorkspace{
		Dir: tempDir,
		t:   t,
	}

	// Register cleanup
	t.Cleanup(func() {
		workspace.Cleanup()
	})

	return workspace
}

// CreateFile creates a file with the given content in the workspace
func (w *TestWorkspace) CreateFile(relativePath, content string) string {
	w.t.Helper()
	
	fullPath := filepath.Join(w.Dir, relativePath)
	
	// Create parent directories if needed
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		w.t.Fatalf("Failed to create parent directories for %s: %v", relativePath, err)
	}
	
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		w.t.Fatalf("Failed to create file %s: %v", relativePath, err)
	}
	
	return fullPath
}

// CreateDir creates a directory in the workspace
func (w *TestWorkspace) CreateDir(relativePath string) string {
	w.t.Helper()
	
	fullPath := filepath.Join(w.Dir, relativePath)
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		w.t.Fatalf("Failed to create directory %s: %v", relativePath, err)
	}
	
	return fullPath
}

// CreateSymlink creates a symbolic link in the workspace
func (w *TestWorkspace) CreateSymlink(relativePath, target string) string {
	w.t.Helper()
	
	fullPath := filepath.Join(w.Dir, relativePath)
	
	// Create parent directories if needed
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		w.t.Fatalf("Failed to create parent directories for %s: %v", relativePath, err)
	}
	
	if err := os.Symlink(target, fullPath); err != nil {
		w.t.Fatalf("Failed to create symlink %s -> %s: %v", relativePath, target, err)
	}
	
	return fullPath
}

// FileExists checks if a file exists in the workspace
func (w *TestWorkspace) FileExists(relativePath string) bool {
	fullPath := filepath.Join(w.Dir, relativePath)
	_, err := os.Stat(fullPath)
	return err == nil
}

// ReadFile reads a file from the workspace
func (w *TestWorkspace) ReadFile(relativePath string) string {
	w.t.Helper()
	
	fullPath := filepath.Join(w.Dir, relativePath)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		w.t.Fatalf("Failed to read file %s: %v", relativePath, err)
	}
	
	return string(content)
}

// ListFiles returns all files in the workspace (recursively)
func (w *TestWorkspace) ListFiles() []string {
	w.t.Helper()
	
	var files []string
	err := filepath.WalkDir(w.Dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		if !d.IsDir() {
			relPath, err := filepath.Rel(w.Dir, path)
			if err != nil {
				return err
			}
			files = append(files, relPath)
		}
		
		return nil
	})
	
	if err != nil {
		w.t.Fatalf("Failed to list files: %v", err)
	}
	
	return files
}

// GetMockContext returns a MockToolContext configured for this workspace
func (w *TestWorkspace) GetMockContext() *MockToolContext {
	return NewMockToolContext(w.Dir)
}

// Cleanup removes the temporary workspace
func (w *TestWorkspace) Cleanup() {
	// Run custom cleanup functions first
	for _, fn := range w.cleanup {
		if err := fn(); err != nil {
			w.t.Logf("Cleanup function failed: %v", err)
		}
	}
	
	if err := os.RemoveAll(w.Dir); err != nil {
		w.t.Logf("Failed to cleanup workspace %s: %v", w.Dir, err)
	}
}

// AddCleanup adds a custom cleanup function
func (w *TestWorkspace) AddCleanup(fn func() error) {
	w.cleanup = append(w.cleanup, fn)
}

// CreateProjectStructure creates a typical project structure for testing
func (w *TestWorkspace) CreateProjectStructure() {
	w.t.Helper()
	
	// Create common directories
	w.CreateDir("src")
	w.CreateDir("tests")
	w.CreateDir("docs")
	w.CreateDir("config")
	
	// Create common files
	w.CreateFile("README.md", "# Test Project\n\nThis is a test project.")
	w.CreateFile("go.mod", "module test-project\n\ngo 1.21")
	w.CreateFile("main.go", `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
`)
	w.CreateFile("src/utils.go", `package main

func Add(a, b int) int {
	return a + b
}
`)
	w.CreateFile("tests/main_test.go", `package main

import "testing"

func TestAdd(t *testing.T) {
	result := Add(2, 3)
	if result != 5 {
		t.Errorf("Expected 5, got %d", result)
	}
}
`)
}

// CreateGitRepo initializes a git repository in the workspace
func (w *TestWorkspace) CreateGitRepo() {
	w.t.Helper()
	
	// Create .git directory structure (minimal)
	w.CreateDir(".git")
	w.CreateDir(".git/refs")
	w.CreateDir(".git/refs/heads")
	w.CreateDir(".git/objects")
	
	// Create basic git files
	w.CreateFile(".git/HEAD", "ref: refs/heads/main\n")
	w.CreateFile(".git/config", `[core]
	repositoryformatversion = 0
	filemode = true
	bare = false
	logallrefupdates = true
[branch "main"]
	remote = origin
	merge = refs/heads/main
`)
	w.CreateFile(".gitignore", "*.log\n*.tmp\n.env\n")
}

// Sample content generators for different file types
var (
	SampleGoCode = `package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: program <name>")
		return
	}
	
	name := os.Args[1]
	fmt.Printf("Hello, %s!\n", name)
}

func greet(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}
`

	SampleJavaScriptCode = `const express = require('express');
const app = express();
const port = 3000;

app.get('/', (req, res) => {
  res.send('Hello World!');
});

app.get('/api/users/:id', (req, res) => {
  const userId = req.params.id;
  res.json({ id: userId, name: 'John Doe' });
});

app.listen(port, () => {
  console.log('Server running on port ' + port);
});
`

	SamplePythonCode = `#!/usr/bin/env python3
"""
A simple Python script for demonstration.
"""

import sys
import json
from typing import List, Dict

def process_data(data: List[Dict]) -> Dict:
    """Process a list of dictionaries and return summary."""
    result = {
        'count': len(data),
        'keys': set()
    }
    
    for item in data:
        result['keys'].update(item.keys())
    
    result['keys'] = list(result['keys'])
    return result

def main():
    if len(sys.argv) < 2:
        print("Usage: script.py <json_file>")
        return 1
    
    try:
        with open(sys.argv[1], 'r') as f:
            data = json.load(f)
        
        result = process_data(data)
        print(json.dumps(result, indent=2))
        
    except Exception as e:
        print(f"Error: {e}")
        return 1
    
    return 0

if __name__ == '__main__':
    sys.exit(main())
`

	SampleJSONConfig = `{
  "name": "test-project",
  "version": "1.0.0",
  "description": "A test project for tool testing",
  "main": "index.js",
  "scripts": {
    "start": "node index.js",
    "test": "jest",
    "build": "webpack --mode production"
  },
  "dependencies": {
    "express": "^4.18.0",
    "lodash": "^4.17.21"
  },
  "devDependencies": {
    "jest": "^29.0.0",
    "webpack": "^5.74.0"
  }
}
`

	SampleMarkdown = `# Test Document

This is a sample markdown document for testing purposes.

## Features

- **Bold text**
- *Italic text*
- [Links](https://example.com)
- Code blocks

## Code Example

` + "```go" + `
func main() {
    fmt.Println("Hello, World!")
}
` + "```" + `

## Lists

1. First item
2. Second item
3. Third item

- Bullet point
- Another bullet
- Last bullet

## Table

| Column 1 | Column 2 | Column 3 |
|----------|----------|----------|
| Value 1  | Value 2  | Value 3  |
| Value 4  | Value 5  | Value 6  |
`
)

// CreateSampleFiles creates various sample files for testing
func (w *TestWorkspace) CreateSampleFiles() {
	w.t.Helper()
	
	w.CreateFile("main.go", SampleGoCode)
	w.CreateFile("server.js", SampleJavaScriptCode)
	w.CreateFile("script.py", SamplePythonCode)
	w.CreateFile("package.json", SampleJSONConfig)
	w.CreateFile("README.md", SampleMarkdown)
	
	// Create some nested files
	w.CreateFile("src/utils/helper.go", `package utils

func IsEmpty(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}
`)
	
	w.CreateFile("tests/integration/api_test.go", `package integration

import (
	"testing"
	"net/http"
)

func TestAPIEndpoint(t *testing.T) {
	resp, err := http.Get("http://localhost:3000/api/health")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}
`)
}

// AssertFileContent checks that a file has the expected content
func (w *TestWorkspace) AssertFileContent(t *testing.T, relativePath, expectedContent string) {
	t.Helper()
	
	actualContent := w.ReadFile(relativePath)
	if actualContent != expectedContent {
		t.Errorf("File %s content mismatch.\nExpected:\n%s\n\nActual:\n%s", 
			relativePath, expectedContent, actualContent)
	}
}

// AssertFileContains checks that a file contains the expected substring
func (w *TestWorkspace) AssertFileContains(t *testing.T, relativePath, expectedSubstring string) {
	t.Helper()
	
	actualContent := w.ReadFile(relativePath)
	if !strings.Contains(actualContent, expectedSubstring) {
		t.Errorf("File %s does not contain expected substring.\nExpected substring: %s\nActual content:\n%s", 
			relativePath, expectedSubstring, actualContent)
	}
}
