package help_test

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/gechr/clib/help"
	"github.com/gechr/clib/theme"
	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update", false, "update golden files")

func goldenSections() map[string][]help.Section {
	return map[string][]help.Section{
		// Basic flag rendering.
		"short_and_long": {{Title: "Options", Content: []help.Content{
			help.FlagGroup{
				{Short: "v", Long: "verbose", Desc: "Verbose output"},
				{Short: "o", Long: "output", Placeholder: "fmt", Desc: "Output format"},
			},
		}}},

		// Long-only flags (no extra indent needed).
		"long_only": {{Title: "Options", Content: []help.Content{
			help.FlagGroup{
				{Long: "approve", Desc: "Approve PRs"},
				{Long: "close", Desc: "Close PRs"},
			},
		}}},

		// Long-only flags in a section with short flags get indented.
		"mixed_indent": {{Title: "Options", Content: []help.Content{
			help.FlagGroup{
				{Short: "a", Long: "author", Placeholder: "user", Desc: "Filter by author"},
				{Long: "organization", Placeholder: "org", Desc: "Limit to org"},
			},
		}}},

		// clib:hide-long — suppress long form, show short only.
		"hide_long": {{Title: "Filters", Content: []help.Content{
			help.FlagGroup{
				{Short: "i", Placeholder: "regex", Desc: "Filter by regex"},
				{Long: "include", Placeholder: "name", Desc: "Include by exact name"},
			},
		}}},

		// clib:hide-short — suppress short form, show long only.
		"hide_short": {{Title: "Options", Content: []help.Content{
			help.FlagGroup{
				{Long: "verbose", Desc: "Verbose output"},
				{Short: "q", Long: "quiet", Desc: "Suppress output"},
			},
		}}},

		// clib:no-indent — long-only flag at same indent as short flags.
		"no_indent": {{Title: "Filters", Content: []help.Content{
			help.FlagGroup{
				{Short: "i", Placeholder: "regex", Desc: "Filter by regex"},
				{
					Long:        "include",
					NoIndent:    true,
					Placeholder: "name",
					Desc:        "Include by exact name",
				},
			},
		}}},

		// Multiple flag groups (blank line separator).
		"groups": {{Title: "Filters", Content: []help.Content{
			help.FlagGroup{
				{Short: "O", Long: "owner", Placeholder: "owner", Desc: "Owner/organization"},
			},
			help.FlagGroup{
				{Long: "archived", Desc: "Include archived"},
				{Long: "forked", Desc: "Include forked"},
				{Short: "l", Long: "language", Placeholder: "lang", Desc: "Filter by language"},
			},
		}}},

		// Enum values appended after description.
		"enum": {{Title: "Options", Content: []help.Content{
			help.FlagGroup{
				{
					Short: "s", Long: "state", Placeholder: "state",
					Desc: "Filter by state",
					Enum: []string{"open", "closed", "merged", "all"},
				},
			},
		}}},

		// Enum with default annotation.
		"enum_default": {{Title: "Options", Content: []help.Content{
			help.FlagGroup{
				{
					Long: "color", Placeholder: "when",
					Desc:        "When to use color",
					Enum:        []string{"auto", "always", "never"},
					EnumDefault: "auto",
				},
			},
		}}},

		// Repeatable flag with ellipsis.
		"repeatable": {{Title: "Options", Content: []help.Content{
			help.FlagGroup{
				{Long: "label", Placeholder: "tag", Repeatable: true, Desc: "Add labels"},
			},
		}}},

		// Literal placeholder (no angle brackets).
		"placeholder_literal": {{Title: "Options", Content: []help.Content{
			help.FlagGroup{
				{Long: "to", Placeholder: "1.2.3", PlaceholderLiteral: true, Desc: "Set version"},
			},
		}}},

		// Description with bracketed note.
		"desc_note": {{Title: "Options", Content: []help.Content{
			help.FlagGroup{
				{Long: "mirror", Desc: "Create a mirror clone (git only)"},
				{Long: "depth", Placeholder: "n", Desc: "Shallow clone [default: full]"},
			},
		}}},

		// Full multi-section output exercising all features.
		"full": {
			{Title: "Usage", Content: []help.Content{
				help.Usage{Command: "acme", ShowOptions: true, Args: []help.Arg{
					{Name: "target"},
				}},
			}},
			{Title: "Arguments", Content: []help.Content{
				help.Args{
					{Name: "target", Desc: "Target to operate on"},
				},
			}},
			{Title: "Filters", Content: []help.Content{
				help.FlagGroup{
					{Short: "O", Long: "owner", Placeholder: "owner", Desc: "Owner/organization"},
				},
				help.FlagGroup{
					{Long: "archived", Desc: "Include archived"},
					{Short: "l", Long: "language", Placeholder: "lang", Desc: "Filter by language"},
				},
				help.FlagGroup{
					{Short: "i", Placeholder: "regex", Desc: "Filter by regex"},
					{
						Long:        "include",
						NoIndent:    true,
						Placeholder: "name",
						Desc:        "Include by exact name",
					},
				},
			}},
			{Title: "Options", Content: []help.Content{
				help.FlagGroup{
					{Short: "b", Long: "branch", Placeholder: "name", Desc: "Branch to use"},
					{Short: "Q", Long: "quick", Desc: "Quick mode"},
					{
						Long: "method", Placeholder: "type", Desc: "Method",
						Enum: []string{"ssh", "https"}, EnumDefault: "ssh",
					},
				},
				help.FlagGroup{
					{Short: "n", Long: "dry", Desc: "Dry run"},
					{Short: "q", Long: "quiet", Desc: "Suppress output"},
					{Short: "v", Long: "verbose", Desc: "Verbose output"},
				},
			}},
			{Title: "Examples", Content: []help.Content{
				help.Examples{
					{Comment: "Basic usage", Command: "acme target"},
					{Comment: "With filters", Command: "acme --language=Go all"},
				},
			}},
		},
	}
}

func TestGolden(t *testing.T) {
	th := theme.Default()
	r := help.NewRenderer(th)

	for name, sections := range goldenSections() {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			err := r.Render(&buf, sections)
			require.NoError(t, err)

			got := ansi.Strip(buf.String())
			goldenFile := filepath.Join("testdata", name+".golden")

			if *update {
				require.NoError(t, os.WriteFile(goldenFile, []byte(got), 0o644))
				return
			}

			want, err := os.ReadFile(goldenFile)
			require.NoError(t, err, "golden file missing; run with -update to create")
			require.Equal(t, string(want), got)
		})
	}
}
