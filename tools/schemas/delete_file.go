package schemas

import (
	"anthropic-chat/utils"
)

// DeleteFileInput represents the input schema for the delete_file tool
type DeleteFileInput struct {
	Path      string `json:"path" jsonschema_description:"Relative path of the file or directory to delete."`
	Recursive bool   `json:"recursive" jsonschema_description:"If true, delete directories and their contents recursively. Required for non-empty directories."`
}

// DeleteFileInputSchema is the cached schema for DeleteFileInput
var DeleteFileInputSchema = utils.GenerateSchema[DeleteFileInput]()
