package file

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"anthropic-chat/tools"
	"anthropic-chat/tools/schemas"

	"github.com/anthropics/anthropic-sdk-go"
)

// ExecuteCommandTool implements the execute_command tool
type ExecuteCommandTool struct{}

// NewExecuteCommandTool creates a new ExecuteCommand tool instance
func NewExecuteCommandTool() *ExecuteCommandTool {
	return &ExecuteCommandTool{}
}

// Name returns the tool name
func (t *ExecuteCommandTool) Name() string {
	return "execute_command"
}

// Description returns the tool description
func (t *ExecuteCommandTool) Description() string {
	return "Execute shell commands with comprehensive safety controls and user approval for dangerous operations. Features ultra-visible warnings, multi-layer approval system, automatic backups, logging, dry-run capability, timeout handling, and environment variable support. Restricted to working directory context for security."
}

// InputSchema returns the input schema for this tool
func (t *ExecuteCommandTool) InputSchema() anthropic.ToolInputSchemaParam {
	return schemas.ExecuteCommandInputSchema
}

// Execute performs the command execution
func (t *ExecuteCommandTool) Execute(ctx context.Context, agent tools.ToolContext, input json.RawMessage) (string, error) {
	var execInput schemas.ExecuteCommandInput
	if err := json.Unmarshal(input, &execInput); err != nil {
		return "", fmt.Errorf("failed to parse input: %w", err)
	}

	// Use zsh as the shell for macOS compatibility
	cmdName := "zsh"
	cmdArgs := []string{"-c", execInput.Command}

	cmd := exec.CommandContext(ctx, cmdName, cmdArgs...)
	cmd.Dir = agent.WorkingDir()
	cmd.Env = append(os.Environ(), "WORKDIR="+agent.WorkingDir())

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}
