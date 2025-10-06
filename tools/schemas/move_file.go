package schemas

import (
	"anthropic-chat/utils"
)

// MoveFileInput represents the input schema for the move_file tool
type MoveFileInput struct {
	Source      string `json:"source" jsonschema_description:"Relative path of the file or directory to move."`
	Destination string `json:"destination" jsonschema_description:"Relative path where the file or directory should be moved to."`
}

// MoveFileInputSchema is the cached schema for MoveFileInput
var MoveFileInputSchema = utils.GenerateSchema[MoveFileInput]()
