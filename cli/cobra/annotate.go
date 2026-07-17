package cobra

import (
	"encoding/json"

	"github.com/gechr/clib/complete"
	cobralib "github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// FlagExtra holds clib-specific metadata for a pflag.Flag.
type FlagExtra struct {
	Aliases        []string       `json:"aliases"`        // additional long-flag aliases
	Complete       string         `json:"complete"`       // completion directive (e.g. "predictor=repo")
	CompleteHidden bool           `json:"completeHidden"` // still offer this flag in completions even when hidden from help
	Enum           []string       `json:"enum"`           // enum values
	EnumDefault    string         `json:"enumDefault"`    // default enum value (highlighted by EnumStyleHighlightDefault)
	EnumHighlight  []string       `json:"enumHighlight"`  // highlight hints for enum values
	EnumTerse      []string       `json:"enumTerse"`      // short descriptions for enum values (parallel to Enum)
	Extension      string         `json:"extension"`      // file extension filter for completion (e.g. "yaml" or "yaml,yml")
	Group          string         `json:"group"`          // help section group
	HideLong       bool           `json:"hideLong"`       // hide the long flag from help output
	HideShort      bool           `json:"hideShort"`      // hide the short flag from help output
	Hint           string         `json:"hint"`           // value type hint for completion (file, dir, command, user, host, url, email)
	NoIndent       bool           `json:"noIndent"`       // suppress short-flag alignment indent in help
	Negatable      bool           `json:"negatable"`      // supports --no- prefix
	NegativeDesc   string         `json:"negativeDesc"`   // description for --no- variant (negatable flags)
	NegativeOnly   bool           `json:"negativeOnly"`   // advertise only the --no- variant in help (negatable flags)
	Order          complete.Order `json:"order"`          // completion ordering mode
	Placeholder    string         `json:"placeholder"`    // value placeholder (e.g. "repo")
	PositiveDesc   string         `json:"positiveDesc"`   // description for positive variant (negatable flags)
	PositiveOnly   bool           `json:"positiveOnly"`   // advertise only the positive variant in help (negatable flags)
	Terse          string         `json:"terse"`          // very short description for completions
}

const (
	extraAnnotationKey        = "clib.extra"
	commandExtraAnnotationKey = "clib.command-extra"
)

// CommandExtra holds clib-specific metadata for a cobra command.
type CommandExtra struct {
	Alias string `json:"alias"` // command invoked by this alias command
}

// ExtendCommand attaches clib metadata to a cobra command.
func ExtendCommand(cmd *cobralib.Command, extra CommandExtra) {
	if cmd == nil {
		return
	}
	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}
	data, err := json.Marshal(extra)
	if err != nil {
		return
	}
	cmd.Annotations[commandExtraAnnotationKey] = string(data)
}

func getCommandExtra(cmd *cobralib.Command) *CommandExtra {
	if cmd == nil || cmd.Annotations == nil {
		return nil
	}
	data := cmd.Annotations[commandExtraAnnotationKey]
	if data == "" {
		return nil
	}
	var extra CommandExtra
	if err := json.Unmarshal([]byte(data), &extra); err != nil {
		return nil
	}
	return &extra
}

// Extend attaches clib metadata to a pflag.Flag.
//
//	cobracli.Extend(f.Lookup("repo"), cobracli.FlagExtra{
//		Group:       "Filters",
//		Placeholder: "repo",
//		Complete:    "predictor=repo",
//	})
func Extend(flag *pflag.Flag, extra FlagExtra) {
	if flag == nil {
		return
	}
	if flag.Annotations == nil {
		flag.Annotations = map[string][]string{}
	}
	data, err := json.Marshal(extra)
	if err != nil {
		return
	}
	flag.Annotations[extraAnnotationKey] = []string{string(data)}
}

func getExtra(f *pflag.Flag) *FlagExtra {
	if f == nil || len(f.Annotations[extraAnnotationKey]) == 0 {
		return nil
	}

	var extra FlagExtra
	if err := json.Unmarshal([]byte(f.Annotations[extraAnnotationKey][0]), &extra); err != nil {
		return nil
	}
	return &extra
}

// resetExtras is a no-op now that extras are stored on flags directly.
func resetExtras() {}
