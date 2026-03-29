package complete

import (
	"strconv"
	"strings"
)

// ApplyActionArgs supplements action with completion flags parsed from args.
// Existing action fields are only updated when the corresponding flag is present.
func ApplyActionArgs(action *Action, args []string) {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			break
		}
		if !strings.HasPrefix(arg, "--") {
			continue
		}

		name, value, hasValue := strings.Cut(strings.TrimPrefix(arg, "--"), "=")
		switch name {
		case FlagComplete:
			if !hasValue && i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				value = args[i+1]
				i++
			}
			action.Complete = value
		case FlagShell:
			if !hasValue && i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				value = args[i+1]
				i++
			}
			action.Shell = value
		case FlagInstallCompletion:
			if parsed, ok := parseBoolArg(value, hasValue); ok {
				action.InstallCompletion = parsed
			}
		case FlagUninstallCompletion:
			if parsed, ok := parseBoolArg(value, hasValue); ok {
				action.UninstallCompletion = parsed
			}
		case FlagPrintCompletion:
			if parsed, ok := parseBoolArg(value, hasValue); ok {
				action.PrintCompletion = parsed
			}
		}
	}
}

func parseBoolArg(value string, hasValue bool) (bool, bool) {
	if !hasValue {
		return true, true
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, false
	}
	return parsed, true
}
