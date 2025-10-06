package schemas

import (
	"anthropic-chat/utils"
)

// SearchFilesInput represents the input schema for the search_files tool
type SearchFilesInput struct {
	Path        string `json:"path" jsonschema_description:"Directory to search in (relative to working directory)."`
	Regex       string `json:"regex" jsonschema_description:"Rust regex pattern to search for."`
	FilePattern string `json:"file_pattern,omitempty" jsonschema_description:"Glob pattern to filter files (e.g., '*.ts' for TypeScript files)."`
}

// SearchFilesInputSchema is the cached schema for SearchFilesInput
var SearchFilesInputSchema = utils.GenerateSchema[SearchFilesInput]()
