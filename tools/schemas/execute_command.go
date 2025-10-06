package schemas

import (
	"anthropic-chat/utils"
)

// ExecuteCommandInput represents the input schema for the execute_command tool
type ExecuteCommandInput struct {
	Command          string `json:"command" jsonschema_description:"Shell command to execute."`
	RequiresApproval bool   `json:"requires_approval" jsonschema_description:"Whether this command requires explicit user approval before execution."`
}

// ExecuteCommandInputSchema is the cached schema for ExecuteCommandInput
var ExecuteCommandInputSchema = utils.GenerateSchema[ExecuteCommandInput]()
