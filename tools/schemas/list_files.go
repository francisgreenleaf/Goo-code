package schemas

import (
	"anthropic-chat/utils"
)

// ListFilesInput represents the input schema for the list_files tool
type ListFilesInput struct {
	Path      string `json:"path,omitempty" jsonschema_description:"Optional relative path (defaults to current directory)."`
	Recursive bool   `json:"recursive,omitempty" jsonschema_description:"List files recursively (defaults to false for top-level only)."`
	Limit     int    `json:"limit,omitempty" jsonschema_description:"Maximum number of files to return (defaults to 100)."`
	Pattern   string `json:"pattern,omitempty" jsonschema_description:"Optional glob pattern to filter files (e.g., '*.go', '*.js')."`
}

// ListFilesInputSchema is the cached schema for ListFilesInput
var ListFilesInputSchema = utils.GenerateSchema[ListFilesInput]()
