package complete

// IsRequest reports whether args (excluding argv[0]) is a native completion
// invocation. Words following the request command are data, not action flags.
func IsRequest(args []string) bool {
	return len(args) > 0 && (args[0] == "__complete" || args[0] == "__completeNoDesc")
}
