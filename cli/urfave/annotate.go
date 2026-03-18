package urfave

import (
	"reflect"
	"strings"
	"sync"

	clilib "github.com/urfave/cli/v3"
)

// FlagExtra holds clib-specific metadata for a urfave Flag.
type FlagExtra struct {
	Complete      string   // completion directive (e.g. "predictor=repo")
	Enum          []string // enum values
	EnumDefault   string   // default enum value (highlighted by EnumStyleHighlightDefault)
	EnumHighlight []string // highlight hints for enum values
	Extension     string   // file extension filter for completion (e.g. "yaml" or "yaml,yml")
	Group         string   // help section group
	HideLong      bool     // hide the long flag from help output
	HideShort     bool     // hide the short flag from help output
	Hint          string   // value type hint for completion (file, dir, command, user, host, url, email)
	NegativeDesc  string   // description for --no- variant (negatable flags)
	Placeholder   string   // value placeholder (e.g. "repo")
	PositiveDesc  string   // description for positive variant (negatable flags)
	Terse         string   // very short description for completions
}

var (
	extras   = make(map[flagKey]*FlagExtra)
	extrasMu sync.RWMutex
)

const (
	flagExtraMetadataKey    = "clib.flag-extra"
	commandExtraMetadataKey = "clib.command-extra"
)

type flagKey struct {
	typ   reflect.Type
	ptr   uintptr
	names string
}

// Extend attaches clib metadata to a urfave Flag.
//
//	cliurfave.Extend(repoFlag, cliurfave.FlagExtra{
//		Group:       "Filters",
//		Placeholder: "repo",
//		Complete:    "predictor=repo",
//	})
//
// Metadata is bound onto command-local metadata the first time clib walks a
// command tree. Pointer flags are keyed by identity; value flags fall back to
// type+name matching so custom non-comparable value flags still work.
func Extend(flag clilib.Flag, extra FlagExtra) {
	if flag == nil {
		return
	}
	key, ok := newFlagKey(flag)
	if !ok {
		return
	}
	extrasMu.Lock()
	extras[key] = &extra
	extrasMu.Unlock()
}

func prepareFlagExtras(cmd *clilib.Command) {
	if cmd == nil {
		return
	}

	root := cmd.Root()
	if root == nil {
		root = cmd
	}

	extrasMu.Lock()
	defer extrasMu.Unlock()

	used := make(map[flagKey]struct{})
	bindFlagExtrasLocked(root, used)
	for key := range used {
		delete(extras, key)
	}
}

func bindFlagExtrasLocked(cmd *clilib.Command, used map[flagKey]struct{}) {
	if cmd == nil {
		return
	}

	local := flagExtrasFromMetadata(cmd)
	for _, flag := range cmd.Flags {
		key, ok := newFlagKey(flag)
		if !ok {
			continue
		}
		extra, ok := extras[key]
		if !ok {
			continue
		}
		if local == nil {
			local = make(map[flagKey]FlagExtra)
		}
		local[key] = *extra
		used[key] = struct{}{}
	}
	if len(local) > 0 {
		if cmd.Metadata == nil {
			cmd.Metadata = map[string]any{}
		}
		cmd.Metadata[flagExtraMetadataKey] = local
	}
	for _, child := range cmd.Commands {
		bindFlagExtrasLocked(child, used)
	}
}

func getFlagExtra(cmd *clilib.Command, flag clilib.Flag) *FlagExtra {
	if flag == nil {
		return nil
	}
	key, ok := newFlagKey(flag)
	if !ok {
		return nil
	}

	if local := flagExtrasFromMetadata(cmd); local != nil {
		if extra, ok := local[key]; ok {
			return new(extra)
		}
	}

	extrasMu.RLock()
	extra := extras[key]
	extrasMu.RUnlock()
	return extra
}

func flagExtrasFromMetadata(cmd *clilib.Command) map[flagKey]FlagExtra {
	if cmd == nil || cmd.Metadata == nil {
		return nil
	}
	switch extras := cmd.Metadata[flagExtraMetadataKey].(type) {
	case map[flagKey]FlagExtra:
		return extras
	default:
		return nil
	}
}

// CommandExtra holds clib-specific metadata for a urfave Command.
type CommandExtra struct {
	PathArgs bool // enable file completion for positional args
}

var cmdExtrasMu sync.RWMutex

// ExtendCommand attaches clib metadata to a urfave Command.
func ExtendCommand(cmd *clilib.Command, extra CommandExtra) {
	if cmd == nil {
		return
	}
	cmdExtrasMu.Lock()
	if cmd.Metadata == nil {
		cmd.Metadata = map[string]any{}
	}
	cmd.Metadata[commandExtraMetadataKey] = extra
	cmdExtrasMu.Unlock()
}

func getCommandExtra(cmd *clilib.Command) *CommandExtra {
	if cmd == nil {
		return nil
	}
	cmdExtrasMu.RLock()
	defer cmdExtrasMu.RUnlock()

	switch value := cmd.Metadata[commandExtraMetadataKey].(type) {
	case CommandExtra:
		return new(value)
	case *CommandExtra:
		return value
	default:
		return nil
	}
}

// resetExtras clears all registered flag and command extras. For testing only.
func resetExtras() {
	extrasMu.Lock()
	clear(extras)
	extrasMu.Unlock()
}

func newFlagKey(flag clilib.Flag) (flagKey, bool) {
	v := reflect.ValueOf(flag)
	if !v.IsValid() {
		return flagKey{}, false
	}
	key := flagKey{typ: v.Type()}
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return flagKey{}, false
		}
		key.ptr = v.Pointer()
		return key, true
	}
	names := flag.Names()
	if len(names) == 0 {
		return flagKey{}, false
	}
	key.names = strings.Join(names, "\x00")
	return key, true
}
