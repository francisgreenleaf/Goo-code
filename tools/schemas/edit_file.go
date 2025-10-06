package schemas

import (
	"anthropic-chat/utils"
)

// EditOperation represents a single edit operation
type EditOperation struct {
	Type    string `json:"type" jsonschema_description:"Type of operation: 'insert', 'delete', 'replace', or 'append'"`
	Line    *int   `json:"line,omitempty" jsonschema_description:"Line number for insert/delete/replace operations (1-based, optional for append)"`
	EndLine *int   `json:"end_line,omitempty" jsonschema_description:"End line number for delete/replace range operations (1-based, inclusive)"`
	Content string `json:"content,omitempty" jsonschema_description:"Content to insert/replace/append"`
}

// EditFileInput represents the input schema for the edit_file tool
type EditFileInput struct {
	Path       string          `json:"path" jsonschema_description:"Relative file path in working directory"`
	Operations []EditOperation `json:"operations" jsonschema_description:"List of edit operations to perform in order"`
}

// EditFileInputSchema is the cached schema for EditFileInput
var EditFileInputSchema = utils.GenerateSchema[EditFileInput]()
