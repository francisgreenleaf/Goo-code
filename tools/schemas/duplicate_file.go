package schemas

import (
	"anthropic-chat/utils"
)

// DuplicateFileInput represents the input schema for the duplicate_file tool
type DuplicateFileInput struct {
	Path string `json:"path" jsonschema_description:"Relative file path in working directory to duplicate."`
}

// DuplicateFileInputSchema is the cached schema for DuplicateFileInput
var DuplicateFileInputSchema = utils.GenerateSchema[DuplicateFileInput]()
