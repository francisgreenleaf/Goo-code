package schemas

import (
	"anthropic-chat/utils"
)

// WriteFileInput represents the input schema for the write_file tool
type WriteFileInput struct {
	Path    string `json:"path" jsonschema_description:"Relative file path in working directory to write to."`
	Content string `json:"content" jsonschema_description:"Complete content to write to the file."`
}

// WriteFileInputSchema is the cached schema for WriteFileInput
var WriteFileInputSchema = utils.GenerateSchema[WriteFileInput]()
