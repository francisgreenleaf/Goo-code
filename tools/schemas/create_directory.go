package schemas

import (
	"anthropic-chat/utils"
)

// CreateDirectoryInput represents the input schema for the create_directory tool
type CreateDirectoryInput struct {
	Path string `json:"path" jsonschema_description:"Directory path to create (relative to working directory)."`
}

// CreateDirectoryInputSchema is the cached schema for CreateDirectoryInput
var CreateDirectoryInputSchema = utils.GenerateSchema[CreateDirectoryInput]()
