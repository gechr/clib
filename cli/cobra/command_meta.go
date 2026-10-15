package cobra

import (
	"github.com/gechr/clib/complete"
	cobralib "github.com/spf13/cobra"
)

// CommandMetadata contains completion annotations for one command.
type CommandMetadata struct {
	// DynamicArgs maps positional indices to predictors. The final entry repeats.
	DynamicArgs []string
	PathArgs    bool
}

// CommandMeta reads a command's clib annotations, including the root command.
// A nil command returns empty metadata. Malformed annotations panic, as in
// Subcommands and the other annotation readers.
func CommandMeta(cmd *cobralib.Command) CommandMetadata {
	var sub complete.SubSpec
	applyCommandAnnotations(&sub, cmd)
	return CommandMetadata{DynamicArgs: sub.DynamicArgs, PathArgs: sub.PathArgs}
}
