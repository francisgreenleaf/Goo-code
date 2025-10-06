# New Tools Implementation Plan

This document provides a structured plan for implementing Phases 2 and 3 of the tool expansion roadmap.

## Testing Infrastructure Status

✅ **Completed**:
- Mock tool context for isolated testing
- Test workspace with temporary file operations
- Comprehensive validators for tool interfaces and execution
- Tool adapters to bridge main tools with test interfaces
- Complete test suite for `read_file` tool as template
- Registry integration tests

## Phase 1: Basic File Operations (IN PROGRESS)

### 1. Move/Rename File Tool
**Status**: To be implemented
**Files to Create**:
- `tools/file/move.go` - Implementation
- `tools/schemas/move_file.go` - Input schema
- `tools/file/move_test.go` - Comprehensive tests

**Test Coverage Required**:
- Move file within same directory
- Move file to different directory
- Rename file (same directory)
- Path traversal protection (source and destination)
- Source file doesn't exist
- Destination already exists (should error or overwrite with flag)
- Moving directories
- Special characters in paths
- Large file moves

**Implementation Notes**:
- Use `os.Rename()` for same-filesystem moves
- Consider copy + delete for cross-filesystem moves
- Validate both source and destination paths
- Atomic operation where possible

### 2. Delete File Tool
**Status**: To be implemented
**Files to Create**:
- `tools/file/delete.go` - Implementation
- `tools/schemas/delete_file.go` - Input schema
- `tools/file/delete_test.go` - Comprehensive tests

**Test Coverage Required**:
- Delete existing file
- Delete non-existent file (should error gracefully)
- Path traversal protection
- Delete directory (should require recursive flag or error)
- Verify file is actually deleted
- Cannot delete outside working directory
- Handle permission errors

**Implementation Notes**:
- Use `os.Remove()` for files
- Use `os.RemoveAll()` for directories with recursive flag
- Consider adding confirmation for destructive operations
- Log deletions for audit trail

### 3. Find Files Tool
**Status**: To be implemented
**Files to Create**:
- `tools/file/find.go` - Implementation
- `tools/schemas/find_files.go` - Input schema
- `tools/file/find_test.go` - Comprehensive tests

**Test Coverage Required**:
- Find by name pattern (glob: *.go, test*.js)
- Find by content (regex search)
- Find by size (larger/smaller than X)
- Find by modification time
- Recursive vs non-recursive search
- Path traversal protection
- Performance with many files (limit results)
- Case-sensitive vs case-insensitive

**Implementation Notes**:
- Use `filepath.Walk()` or `filepath.WalkDir()` for traversal
- Support multiple search criteria (AND logic)
- Limit results to prevent overwhelming output
- Return relative paths from working directory
- Consider using existing `search_files` tool as foundation

---

## Phase 2: Git Integration Tools

### 1. Git Status Tool ⭐ START HERE
**Priority**: HIGH - Foundation for all git operations
**Files to Create**:
- `tools/git/status.go` - Implementation
- `tools/schemas/git_status.go` - Input schema
- `tools/git/status_test.go` - Comprehensive tests

**Test Coverage Required**:
- Show status in valid git repository
- Handle non-git directory gracefully
- Show staged changes
- Show unstaged changes
- Show untracked files
- Empty repository (no commits)
- Detached HEAD state
- Parse git status output correctly

**Implementation Notes**:
- Execute `git status --porcelain` for machine-readable output
- Parse output into structured format
- Handle git not installed
- Handle corrupted git repositories
- Return human-readable summary

**Test Setup**:
```go
// Use TestWorkspace.CreateGitRepo() to set up test environment
workspace := testutil.NewTestWorkspace(t)
workspace.CreateGitRepo()
workspace.CreateFile("test.txt", "content")
// Execute git commands to create test scenarios
```

### 2. Git Diff Tool
**Priority**: HIGH - Essential for code review
**Files to Create**:
- `tools/git/diff.go` - Implementation
- `tools/schemas/git_diff.go` - Input schema
- `tools/git/diff_test.go` - Comprehensive tests

**Test Coverage Required**:
- Diff unstaged changes
- Diff staged changes
- Diff specific file
- Diff between commits
- Diff between branches
- No changes scenario
- Binary file diffs
- Large diffs (truncation handling)

**Implementation Notes**:
- Execute `git diff` with appropriate flags
- Support different diff modes (staged, unstaged, commit range)
- Format output for readability
- Handle binary files
- Limit output size to prevent context overflow

### 3. Git Branch Tool
**Priority**: MEDIUM - Useful for workflow management
**Files to Create**:
- `tools/git/branch.go` - Implementation
- `tools/schemas/git_branch.go` - Input schema
- `tools/git/branch_test.go` - Comprehensive tests

**Test Coverage Required**:
- List all branches
- Create new branch
- Switch to existing branch
- Delete branch
- Current branch indicator
- Remote branches
- Branch doesn't exist
- Cannot delete current branch

**Implementation Notes**:
- Support multiple operations (list, create, switch, delete)
- Use `git branch` and `git checkout` commands
- Parse branch list output
- Handle remote branches separately
- Validate branch names

### 4. Git Log Tool
**Priority**: MEDIUM - Useful for history review
**Files to Create**:
- `tools/git/log.go` - Implementation
- `tools/schemas/git_log.go` - Input schema
- `tools/git/log_test.go` - Comprehensive tests

**Test Coverage Required**:
- Show recent commits (default 10)
- Show commits for specific file
- Show commits by author
- Show commits in date range
- Format output (oneline, full, custom)
- Empty repository
- Limit number of commits

**Implementation Notes**:
- Execute `git log` with formatting options
- Parse commit information (hash, author, date, message)
- Support filtering options
- Limit output to prevent context overflow
- Return structured data

### 5. Git Commit Tool
**Priority**: LOW - Requires careful validation
**Files to Create**:
- `tools/git/commit.go` - Implementation
- `tools/schemas/git_commit.go` - Input schema
- `tools/git/commit_test.go` - Comprehensive tests

**Test Coverage Required**:
- Commit staged changes
- Commit with message
- Nothing to commit
- Validate commit message
- Amend previous commit
- Sign commits (if configured)
- Pre-commit hooks

**Implementation Notes**:
- Require explicit commit message
- Validate message format
- Check for staged changes first
- Handle pre-commit hook failures
- Consider requiring user approval
- Log all commits for audit

---

## Phase 3: Code Intelligence & Project Management

### 1. Code Formatting Tool
**Priority**: HIGH - Improves code quality
**Files to Create**:
- `tools/code/format.go` - Implementation
- `tools/schemas/format_code.go` - Input schema
- `tools/code/format_test.go` - Comprehensive tests

**Test Coverage Required**:
- Format Go code (gofmt)
- Format JavaScript (prettier)
- Format Python (black)
- Format TypeScript (prettier)
- Handle syntax errors gracefully
- Preserve file if formatter fails
- Formatter not installed
- Multiple files at once
- Dry-run mode

**Implementation Notes**:
- Detect language from file extension
- Use appropriate formatter for each language
- Check if formatter is installed
- Capture formatter output/errors
- Support configuration files (.prettierrc, etc.)
- Consider in-place vs return formatted content

**Formatter Commands**:
```bash
# Go
gofmt -w file.go

# JavaScript/TypeScript
prettier --write file.js

# Python
black file.py

# Rust
rustfmt file.rs
```

### 2. Run Tests Tool
**Priority**: HIGH - Critical for TDD workflow
**Files to Create**:
- `tools/code/run_tests.go` - Implementation
- `tools/schemas/run_tests.go` - Input schema
- `tools/code/run_tests_test.go` - Comprehensive tests

**Test Coverage Required**:
- Run Go tests (`go test`)
- Run JavaScript tests (jest, mocha, npm test)
- Run Python tests (pytest, unittest)
- Run specific test file
- Run specific test function
- Parse test output (pass/fail counts)
- Handle test failures
- Timeout handling
- Verbose vs quiet output

**Implementation Notes**:
- Detect test framework from project files
- Execute appropriate test command
- Parse test output for structured results
- Set reasonable timeouts
- Capture stdout and stderr
- Return summary (X passed, Y failed)

**Test Commands by Language**:
```bash
# Go
go test ./... -v
go test -run TestName

# JavaScript
npm test
jest path/to/test.js

# Python
pytest
python -m unittest test_file.py
```

### 3. Lint Code Tool
**Priority**: MEDIUM - Code quality improvement
**Files to Create**:
- `tools/code/lint.go` - Implementation
- `tools/schemas/lint_code.go` - Input schema
- `tools/code/lint_test.go` - Comprehensive tests

**Test Coverage Required**:
- Lint Go code (golint, staticcheck)
- Lint JavaScript (eslint)
- Lint Python (pylint, flake8)
- Parse linter output
- No issues found
- Multiple issues
- Linter not installed
- Configuration file support

**Implementation Notes**:
- Detect language and appropriate linter
- Check if linter is installed
- Parse linter output into structured format
- Support different severity levels
- Return actionable error messages

### 4. Analyze Code Structure Tool
**Priority**: MEDIUM - Helps understand codebases
**Files to Create**:
- `tools/code/analyze.go` - Implementation
- `tools/schemas/analyze_code.go` - Input schema
- `tools/code/analyze_test.go` - Comprehensive tests

**Test Coverage Required**:
- Extract functions/methods from file
- Extract imports/dependencies
- Calculate code metrics (LOC, complexity)
- Identify code patterns
- Multiple languages
- Syntax errors
- Empty files

**Implementation Notes**:
- Use language-specific parsers (go/parser for Go)
- Extract AST information
- Calculate cyclomatic complexity
- Identify code smells
- Return structured data

### 5. Find Function Definition Tool
**Priority**: MEDIUM - Code navigation
**Files to Create**:
- `tools/code/find_definition.go` - Implementation
- `tools/schemas/find_definition.go` - Input schema
- `tools/code/find_definition_test.go` - Comprehensive tests

**Test Coverage Required**:
- Find function definition by name
- Find method definition
- Find class definition
- Multiple matches
- No matches found
- Search across multiple files
- Language-specific handling

**Implementation Notes**:
- Use regex search as foundation
- Language-specific patterns for different languages
- Return file path and line number
- Include surrounding context
- Handle overloaded functions

### 6. Compress/Extract Files Tool
**Priority**: LOW - Utility function
**Files to Create**:
- `tools/file/compress.go` - Implementation
- `tools/schemas/compress_files.go` - Input schema
- `tools/file/compress_test.go` - Comprehensive tests

**Test Coverage Required**:
- Create zip archive
- Create tar.gz archive
- Extract zip archive
- Extract tar.gz archive
- Handle nested directories
- Path traversal in extraction (SECURITY CRITICAL!)
- Corrupted archive handling
- Large archives

**Implementation Notes**:
- Use `archive/zip` and `archive/tar` packages
- Validate all paths during extraction
- Prevent zip slip vulnerability
- Support compression levels
- Handle symbolic links carefully

---

## Implementation Workflow for Each Tool

### Step 1: Write Tests First (TDD)
```bash
# Create test file
touch tools/[category]/[tool_name]_test.go

# Write comprehensive tests following read_test.go template
# Run tests (they should fail)
go test ./tools/[category] -v -run Test[ToolName]
```

### Step 2: Create Schema
```bash
# Create schema file
touch tools/schemas/[tool_name].go

# Define input struct with json tags and descriptions
# Generate schema using utils.GenerateSchema
```

### Step 3: Implement Tool
```bash
# Create implementation file
touch tools/[category]/[tool_name].go

# Implement Tool interface:
# - Name() string
# - Description() string
# - InputSchema() anthropic.ToolInputSchemaParam
# - Execute(ctx, agent, input) (string, error)
```

### Step 4: Register Tool
```go
// In main.go RegisterTools() method
a.toolRegistry.Register([category].New[ToolName]Tool())
```

### Step 5: Update System Prompt
```bash
# Edit system_prompt.txt
# Add tool to "Your Capabilities" section
# Include tool name, description, and usage instructions
```

### Step 6: Verify Tests Pass
```bash
# Run all tests
go test ./tools/... -v

# Run with coverage
go test ./tools/... -cover

# Run specific tool tests
go test ./tools/[category] -v -run Test[ToolName]
```

---

## Security Checklist for Each Tool

Before implementing any tool, verify:

- [ ] Path traversal protection (no `..` in paths)
- [ ] All paths resolved through `agent.ResolveFilePath()`
- [ ] Operations restricted to working directory
- [ ] Input validation (required fields, types)
- [ ] Error messages don't leak sensitive information
- [ ] Destructive operations logged
- [ ] Consider user approval for dangerous operations
- [ ] Handle symbolic links safely
- [ ] Validate file permissions
- [ ] Timeout handling for long operations

---

## Test Template Structure

Every tool test should include:

```go
package [category]

import (
    "context"
    "encoding/json"
    "testing"
    
    "anthropic-chat/tools/schemas"
    "anthropic-chat/tools/testutil"
)

func Test[ToolName]Tool_Interface(t *testing.T) {
    tool := New[ToolName]Tool()
    wrappedTool := testutil.WrapMainTool(tool)
    testutil.ValidateToolInterface(t, wrappedTool)
    
    // Verify name and description
}

func Test[ToolName]Tool_Schema(t *testing.T) {
    tool := New[ToolName]Tool()
    schema := tool.InputSchema()
    testutil.ValidateSchema(t, schema, "[ToolName]Tool")
    
    // Verify schema structure
}

func Test[ToolName]Tool_Execute_Success(t *testing.T) {
    tool := New[ToolName]Tool()
    wrappedTool := testutil.WrapMainTool(tool)
    workspace := testutil.NewTestWorkspace(t)
    agent := workspace.GetMockContext()
    
    // Create test scenario
    // Execute tool
    // Verify results
}

func Test[ToolName]Tool_Execute_Errors(t *testing.T) {
    // Test all error conditions
}

func Test[ToolName]Tool_Execute_PathTraversal(t *testing.T) {
    // Test security (if applicable)
}

func Test[ToolName]Tool_Execute_EdgeCases(t *testing.T) {
    // Test edge cases specific to this tool
}
```

---

## Phase 2: Git Integration (FUTURE)

**Estimated Timeline**: 1-2 weeks after Phase 1
**Prerequisites**: 
- Phase 1 complete
- Git test utilities created
- Git command execution patterns established

**Tools to Implement** (in order):
1. Git Status (foundation)
2. Git Diff (builds on status)
3. Git Branch (workflow management)
4. Git Log (history review)
5. Git Commit (requires approval system)

**Additional Test Utilities Needed**:
```go
// tools/testutil/git_fixtures.go
- InitGitRepo() - Initialize git repository
- CreateCommit(message, files) - Create test commits
- CreateBranch(name) - Create test branches
- StageFiles(paths) - Stage files for commit
- MockGitConfig() - Set up git user config
```

---

## Phase 3: Code Intelligence (FUTURE)

**Estimated Timeline**: 2-3 weeks after Phase 2
**Prerequisites**:
- Phase 1 and 2 complete
- Language parser integration
- Code analysis patterns established

**Tools to Implement** (in order):
1. Code Formatting (high value, low risk)
2. Run Tests (critical for TDD)
3. Lint Code (code quality)
4. Analyze Code Structure (understanding codebases)
5. Find Function Definition (code navigation)
6. Compress/Extract Files (utility)

**Additional Test Utilities Needed**:
```go
// tools/testutil/code_samples.go
- SampleGoCodeWithErrors - Code with syntax errors
- SampleJavaScriptCodeWithErrors - JS with errors
- SamplePythonCodeWithErrors - Python with errors
- ComplexCodeSample - For complexity analysis
- LargeCodebase - For performance testing
```

---

## Success Metrics

For each tool implementation, verify:

1. **Test Coverage**: Minimum 80% code coverage
2. **Security**: All security tests pass
3. **Performance**: Handles large inputs efficiently
4. **Error Handling**: All error paths tested
5. **Documentation**: Clear descriptions in system prompt
6. **Integration**: Works with existing tools
7. **User Experience**: Clear, actionable output

---

## Notes for Future Implementation

### Common Patterns Identified

1. **File Operations**: Always use `agent.ResolveFilePath()` for security
2. **Command Execution**: Use existing `execute_command` tool as reference
3. **Error Messages**: Be specific but don't leak sensitive paths
4. **Output Format**: Return structured data when possible
5. **Validation**: Validate all inputs before execution

### Lessons Learned from Phase 1

- Test infrastructure prevents bugs early
- TDD approach clarifies requirements
- Security tests are critical
- Edge cases reveal design issues
- Comprehensive tests enable confident refactoring

### Future Enhancements

- Add integration tests that combine multiple tools
- Create performance benchmarks
- Add fuzzing tests for security-critical tools
- Create tool usage examples/documentation
- Build tool dependency graph
- Add telemetry for tool usage patterns

---

## Quick Reference: Running Tests

```bash
# Run all tests
go test ./tools/... -v

# Run specific package
go test ./tools/file -v

# Run specific test
go test ./tools/file -v -run TestReadFileTool_Execute_Success

# Run with coverage
go test ./tools/... -cover

# Generate coverage report
go test ./tools/... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run with race detection
go test ./tools/... -race

# Run tests in parallel
go test ./tools/... -v -parallel 4
```

---

## Contact & Questions

If you have questions about implementing any of these tools:
1. Review the `read_file` tool implementation as reference
2. Check the test template in `tools/file/read_test.go`
3. Use the test utilities in `tools/testutil/`
4. Follow the TDD workflow outlined above

Remember: **Write tests first, implement second, verify third.**
