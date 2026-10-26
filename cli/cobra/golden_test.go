package cobra_test

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"

	cobracli "github.com/gechr/clib/cli/cobra"
	"github.com/gechr/clib/help"
	"github.com/gechr/clib/theme"
	cobralib "github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update", false, "update golden files")

func sectionsLowercasePlaceholders() []help.Section {
	cmd := &cobralib.Command{Use: "catalog"}
	cmd.Flags().StringP("channel", "c", "", "Channel `CHANNEL`, name, or alias")
	cmd.Flags().IntP("max-items", "M", 0, "Maximum messages to return")
	cobracli.Extend(cmd.Flags().Lookup("max-items"), cobracli.FlagExtra{
		Placeholder: "N",
	})
	return cobracli.Sections(cmd)
}

func sectionsPreservePlaceholders() []help.Section {
	cmd := &cobralib.Command{Use: "catalog"}
	cmd.Flags().StringP("channel", "c", "", "Channel `CHANNEL`, name, or alias")
	cmd.Flags().IntP("max-items", "M", 0, "Maximum messages to return")
	cobracli.Extend(cmd.Flags().Lookup("max-items"), cobracli.FlagExtra{
		Placeholder: "N",
	})
	return cobracli.SectionsWithOptions(cobracli.WithPreservePlaceholders())(cmd)
}

// sectionsFlagRefsInBackticks exercises backticked content in flag
// descriptions. Backticked flag references (`--foo`) and backticked
// identifiers on bool flags must survive to the renderer as inline code
// markers, never as value placeholders. Non-bool flags continue to honour
// the standard pflag placeholder convention.
func sectionsFlagRefsInBackticks() []help.Section {
	cmd := &cobralib.Command{Use: "run"}
	cmd.Flags().Bool("alpha", false, "Alias for `--foo=0`")
	cmd.Flags().Bool("bravo", false, "Print totals (implies `--alpha`)")
	cmd.Flags().Bool("charlie", false, "Use the `embedded` backend")
	cmd.Flags().String("delta", "", "User `name` to set")
	return cobracli.Sections(cmd)
}

// sectionsLongDescription exercises a command-level long description
// (cmd.Long), surfaced as a help.Description blurb below the Usage line and
// rendered by the shared renderer (paragraph breaks, backtick styling).
func sectionsLongDescription() []help.Section {
	cmd := &cobralib.Command{
		Use:   "deploy <env>",
		Short: "Deploy the app",
		Long: "Deploy the application to the named environment.\n\n" +
			"Pre-flight checks run before cutover; the previous release\n" +
			"is retained so `--rollback` can restore it on failure.",
	}
	cmd.Flags().Bool("rollback", false, "Restore the previous release")
	return cobracli.Sections(cmd)
}

// sectionsEnumRefs exercises cross-referenced enum values: ValidArgs become
// the <provider> positional's enum set, so `github` in the command's long
// description renders in the positional-arg color; a flag enum value (`debug`)
// renders in the flag color. The reference lives in cmd.Long rather than a
// value flag's usage string because pflag would otherwise consume the first
// backticked word as the flag's value placeholder.
func sectionsEnumRefs() []help.Section {
	cmd := &cobralib.Command{
		Use:       "login <provider>",
		Short:     "Authenticate with a provider",
		Long:      "Authenticate with a provider.\n\nThe `github` provider uses an OAuth device flow.",
		ValidArgs: []string{"github", "gitlab", "gitea"},
	}
	cmd.Flags().String("log-level", "", "Logging verbosity")
	cobracli.Extend(cmd.Flags().Lookup("log-level"), cobracli.FlagExtra{
		Enum: []string{"debug", "info", "warn"},
	})
	cmd.Flags().Bool("verbose", false, "Shorthand for `debug` logging")
	return cobracli.Sections(cmd)
}

// sectionsNegatable exercises the three negatable renderings: the default
// bracketed [no-] form, a PositiveOnly extra advertising only --prerelease,
// and a NegativeOnly extra advertising only --no-cache.
func sectionsNegatable() []help.Section {
	cmd := &cobralib.Command{Use: "clover"}
	cmd.Flags().Bool("downgrade", false, "Allow selecting versions older than the current one")
	cmd.Flags().Bool("prerelease", false, "Allow selecting prerelease versions")
	cmd.Flags().Bool("cache", true, "Reuse cached HTTP responses across runs")
	cobracli.Extend(cmd.Flags().Lookup("downgrade"), cobracli.FlagExtra{
		Negatable: true,
	})
	cobracli.Extend(cmd.Flags().Lookup("prerelease"), cobracli.FlagExtra{
		Negatable:    true,
		PositiveOnly: true,
	})
	cobracli.Extend(cmd.Flags().Lookup("cache"), cobracli.FlagExtra{
		Negatable:    true,
		NegativeOnly: true,
	})
	return cobracli.Sections(cmd)
}

func sectionsAliases() []help.Section {
	root := &cobralib.Command{Use: "app"}
	run := &cobralib.Command{
		Use:   "run",
		Short: "Run the app",
		Run:   func(*cobralib.Command, []string) {},
	}
	publish := &cobralib.Command{Use: "publish", Run: func(*cobralib.Command, []string) {}}
	cobracli.ExtendCommand(publish, cobracli.CommandExtra{Alias: "tool release"})
	root.AddCommand(run, publish)
	return cobracli.Sections(root)
}

func commandGroupsRoot() *cobralib.Command {
	root := &cobralib.Command{Use: "app"}
	root.AddGroup(
		&cobralib.Group{ID: "agents", Title: "Agents"},
		&cobralib.Group{ID: "account", Title: "Account"},
	)
	for _, entry := range []struct{ name, desc, native, group string }{
		{"alpha", "Start alpha on a model", "agents", "/0"},
		{"bravo", "Start bravo inside a sandbox", "agents", "/0"},
		{"charlie", "Start charlie on a model", "agents", "/0"},
		{"run", "Run any tool with the shared settings", "agents", "/-1"},
		{"login", "Sign in", "agents", "Account"},
		{"logout", "Forget the stored key", "account", ""},
	} {
		command := &cobralib.Command{
			Use:     entry.name,
			Short:   entry.desc,
			GroupID: entry.native,
			Run:     func(*cobralib.Command, []string) {},
		}
		cobracli.ExtendCommand(command, cobracli.CommandExtra{Group: entry.group})
		root.AddCommand(command)
	}
	hidden := &cobralib.Command{
		Use:    "hidden",
		Hidden: true,
		Run:    func(*cobralib.Command, []string) {},
	}
	cobracli.ExtendCommand(hidden, cobracli.CommandExtra{Group: "Hidden"})
	root.AddCommand(hidden, &cobralib.Command{Use: "unavailable"})
	return root
}

func sectionsCommandGroups() []help.Section {
	return cobracli.Sections(commandGroupsRoot())
}

func TestGolden(t *testing.T) {
	r := help.NewRenderer(theme.Dark())

	scenarios := map[string][]help.Section{
		"command_groups":         sectionsCommandGroups(),
		"aliases":                sectionsAliases(),
		"lowercase_placeholders": sectionsLowercasePlaceholders(),
		"preserve_placeholders":  sectionsPreservePlaceholders(),
		"flag_refs_in_backticks": sectionsFlagRefsInBackticks(),
		"long_description":       sectionsLongDescription(),
		"enum_refs":              sectionsEnumRefs(),
		"negatable":              sectionsNegatable(),
	}

	for name, sections := range scenarios {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			require.NoError(t, r.Render(&buf, sections))

			got := buf.String()
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

func TestCommandGroupOrder(t *testing.T) {
	root := commandGroupsRoot()
	sections := cobracli.SectionsWithOptions(cobracli.WithSortedGroupOrder())(root)
	require.Equal(t, "Account", sections[1].Title)
	require.Equal(t, "Agents", sections[2].Title)
	sections = cobracli.SectionsWithOptions(
		cobracli.WithSortedGroupOrder(),
		cobracli.WithKeepGroupOrder(),
	)(
		root,
	)
	require.Equal(t, "Agents", sections[1].Title)
	require.Equal(t, "Account", sections[2].Title)
}

func TestCommandGroupsGoldenParity(t *testing.T) {
	want, err := os.ReadFile("testdata/command_groups.golden")
	require.NoError(t, err)
	for _, adapter := range []string{"kong", "urfave"} {
		got, readErr := os.ReadFile(
			filepath.Join("..", adapter, "testdata", "command_groups.golden"),
		)
		require.NoError(t, readErr)
		require.Equal(t, string(want), string(got), adapter)
	}
}

func TestCommandGroupDeclarationOrder(t *testing.T) {
	root := &cobralib.Command{Use: "app"}
	root.AddGroup(
		&cobralib.Group{ID: "account", Title: "Account"},
		&cobralib.Group{ID: "agents", Title: "Agents"},
	)
	root.AddCommand(
		&cobralib.Command{
			Use:     "alpha",
			GroupID: "agents",
			Run:     func(*cobralib.Command, []string) {},
		},
		&cobralib.Command{
			Use:     "login",
			GroupID: "account",
			Run:     func(*cobralib.Command, []string) {},
		},
	)
	sections := cobracli.Sections(root)
	require.Equal(t, "Account", sections[1].Title)
	require.Equal(t, "Agents", sections[2].Title)
}

func TestCommandGroupsAliases(t *testing.T) {
	root := &cobralib.Command{Use: "app"}
	run := &cobralib.Command{Use: "run", Run: func(*cobralib.Command, []string) {}}
	alias := &cobralib.Command{Use: "publish", Run: func(*cobralib.Command, []string) {}}
	cobracli.ExtendCommand(run, cobracli.CommandExtra{Group: "Tools"})
	cobracli.ExtendCommand(alias, cobracli.CommandExtra{Alias: "release", Group: "Tools/0"})
	root.AddCommand(
		run,
		alias,
		&cobralib.Command{Use: "version", Run: func(*cobralib.Command, []string) {}},
	)
	for _, tt := range []struct {
		name           string
		opts           []cobracli.SectionsOption
		inline, hidden bool
	}{
		{name: "separate"},
		{name: "inline", opts: []cobracli.SectionsOption{cobracli.WithInlineCommandAliases()}, inline: true},
		{name: "hidden", opts: []cobracli.SectionsOption{cobracli.WithInlineCommandAliases(), cobracli.WithHideCommandAliases()}, hidden: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			sections := cobracli.SectionsWithOptions(tt.opts...)(root)
			content := []help.Content{help.CommandGroup{{Name: "run"}}}
			if tt.inline {
				content = append(
					[]help.Content{
						help.CommandGroup{{Name: "publish", Desc: "Alias for `release`"}},
					},
					content...)
			}
			want := []help.Section{
				{Title: "Tools", Content: content},
				{
					Title:   "Additional Commands",
					Content: []help.Content{help.CommandGroup{{Name: "version"}}},
				},
			}
			if !tt.inline && !tt.hidden {
				want = append(
					want,
					help.Section{
						Title: "Aliases",
						Content: []help.Content{
							help.AliasGroup{{Name: "publish", Target: "release"}},
						},
					},
				)
			}
			require.Equal(t, want, sections[1:])
		})
	}
}
