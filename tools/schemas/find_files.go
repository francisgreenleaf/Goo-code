package schemas

import (
	"anthropic-chat/utils"
)

// FindFilesInput represents the input schema for the find_files tool
type FindFilesInput struct {
	Pattern   string `json:"pattern" jsonschema_description:"Glob pattern to match files (e.g., '*.go', 'test*.js', '*.txt')."`
	Recursive bool   `json:"recursive" jsonschema_description:"If true, search recursively in subdirectories. If false, search only in the top level."`
	Limit     int    `json:"limit" jsonschema_description:"Maximum number of results to return. Defaults to 100 if not specified."`
}

// FindFilesInputSchema is the cached schema for FindFilesInput
var FindFilesInputSchema = utils.GenerateSchema[FindFilesInput]()
