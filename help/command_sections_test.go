package help_test

import (
	"bytes"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/gechr/clib/help"
	"github.com/stretchr/testify/require"
)

func TestBuildCommandSections(t *testing.T) {
	tests := []struct {
		name     string
		commands []help.ClassifiedCommand
		opts     []help.CommandSectionsOption
		want     []help.Section
	}{
		{name: "empty"},
		{
			name: "flat preserves input order",
			commands: []help.ClassifiedCommand{
				{Command: help.Command{Name: "bravo"}},
				{Command: help.Command{Name: "alpha"}},
			},
			want: []help.Section{{Title: "Commands", Content: []help.Content{
				help.CommandGroup{{Name: "bravo"}, {Name: "alpha"}},
			}}},
		},
		{
			name: "numeric blocks surround plain commands",
			commands: []help.ClassifiedCommand{
				{Command: help.Command{Name: "last"}, Group: "Agents/-1"},
				{Command: help.Command{Name: "second"}, Group: "Agents/2"},
				{Command: help.Command{Name: "plain"}, Group: "Agents"},
				{Command: help.Command{Name: "first"}, Group: "Agents/0"},
				{Command: help.Command{Name: "penultimate"}, Group: "Agents/-2"},
				{Command: help.Command{Name: "also-first"}, Group: "Agents/0"},
			},
			want: []help.Section{{Title: "Agents", Content: []help.Content{
				help.CommandGroup{{Name: "first"}, {Name: "also-first"}},
				help.CommandGroup{{Name: "second"}},
				help.CommandGroup{{Name: "plain"}},
				help.CommandGroup{{Name: "penultimate"}},
				help.CommandGroup{{Name: "last"}},
			}}},
		},
		{
			name: "named blocks and sections retain first-seen order",
			commands: []help.ClassifiedCommand{
				{Command: help.Command{Name: "zulu"}, Group: "Tools/z"},
				{Command: help.Command{Name: "login"}, Group: "Account"},
				{Command: help.Command{Name: "alpha"}, Group: "Tools/a"},
				{Command: help.Command{Name: "zulu-again"}, Group: "Tools/z"},
			},
			want: []help.Section{
				{Title: "Tools", Content: []help.Content{
					help.CommandGroup{{Name: "zulu"}, {Name: "zulu-again"}},
					help.CommandGroup{{Name: "alpha"}},
				}},
				{Title: "Account", Content: []help.Content{help.CommandGroup{{Name: "login"}}}},
			},
		},
		{
			name: "native inheritance override and literal slash",
			commands: []help.ClassifiedCommand{
				{Command: help.Command{Name: "run"}, NativeGroup: "Tools/Local"},
				{Command: help.Command{Name: "alpha"}, NativeGroup: "Tools/Local", Group: "/0"},
				{
					Command:     help.Command{Name: "login"},
					NativeGroup: "Tools/Local",
					Group:       "Account",
				},
			},
			want: []help.Section{
				{Title: "Tools/Local", Content: []help.Content{
					help.CommandGroup{{Name: "alpha"}},
					help.CommandGroup{{Name: "run"}},
				}},
				{Title: "Account", Content: []help.Content{help.CommandGroup{{Name: "login"}}}},
			},
		},
		{
			name: "native order omits empty and duplicate declarations",
			commands: []help.ClassifiedCommand{
				{Command: help.Command{Name: "version"}},
				{Command: help.Command{Name: "login"}, NativeGroup: "Account"},
				{Command: help.Command{Name: "run"}, Group: "Tools"},
			},
			opts: []help.CommandSectionsOption{
				help.WithCommandGroupOrder("Tools", "Empty", "Tools", "Account"),
			},
			want: []help.Section{
				{Title: "Tools", Content: []help.Content{help.CommandGroup{{Name: "run"}}}},
				{Title: "Account", Content: []help.Content{help.CommandGroup{{Name: "login"}}}},
				{
					Title:   "Additional Commands",
					Content: []help.Content{help.CommandGroup{{Name: "version"}}},
				},
			},
		},
		{
			name: "alphabetical sections keep ungrouped last",
			commands: []help.ClassifiedCommand{
				{Command: help.Command{Name: "version"}},
				{Command: help.Command{Name: "run"}, Group: "Tools"},
				{Command: help.Command{Name: "login"}, Group: "Account"},
			},
			opts: []help.CommandSectionsOption{
				help.WithCommandGroupOrder("Tools", "Account"), help.WithSortedCommandGroupOrder(),
			},
			want: []help.Section{
				{Title: "Account", Content: []help.Content{help.CommandGroup{{Name: "login"}}}},
				{Title: "Tools", Content: []help.Content{help.CommandGroup{{Name: "run"}}}},
				{
					Title:   "Additional Commands",
					Content: []help.Content{help.CommandGroup{{Name: "version"}}},
				},
			},
		},
		{
			name: "relative blocks merge with explicit fallback",
			commands: []help.ClassifiedCommand{
				{Command: help.Command{Name: "late"}, Group: "/-1"},
				{Command: help.Command{Name: "explicit"}, Group: "Additional Commands/-1"},
				{Command: help.Command{Name: "plain"}},
				{Command: help.Command{Name: "early"}, Group: "/0"},
			},
			want: []help.Section{{Title: "Additional Commands", Content: []help.Content{
				help.CommandGroup{{Name: "early"}},
				help.CommandGroup{{Name: "plain"}},
				help.CommandGroup{{Name: "late"}, {Name: "explicit"}},
			}}},
		},
		{
			name: "relative blocks without named groups",
			commands: []help.ClassifiedCommand{
				{Command: help.Command{Name: "plain"}},
				{Command: help.Command{Name: "early"}, Group: "/0"},
			},
			want: []help.Section{{Title: "Commands", Content: []help.Content{
				help.CommandGroup{{Name: "early"}}, help.CommandGroup{{Name: "plain"}},
			}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, help.BuildCommandSections(tt.commands, tt.opts...))
		})
	}
}

func TestBuildCommandSections_MixedSubkeys(t *testing.T) {
	commands := []help.ClassifiedCommand{
		{Command: help.Command{Name: "three"}, Group: "Tools/3"},
		{Command: help.Command{Name: "named"}, Group: "Tools/z"},
		{Command: help.Command{Name: "two"}, Group: "Tools/2"},
		{Command: help.Command{Name: "one"}, Group: "Tools/1"},
	}
	sections := help.BuildCommandSections(commands)
	require.Equal(t, []help.Content{
		help.CommandGroup{{Name: "one"}},
		help.CommandGroup{{Name: "named"}},
		help.CommandGroup{{Name: "two"}},
		help.CommandGroup{{Name: "three"}},
	}, sections[0].Content)
}

func TestRender_CommandBlocksAlignAcrossSection(t *testing.T) {
	sections := []help.Section{
		{Title: "Agents", Content: []help.Content{
			help.CommandGroup{{Name: "charlie", Desc: "Start charlie"}},
			help.CommandGroup{{Name: "run", Desc: "Run a tool"}},
		}},
		{
			Title:   "Account",
			Content: []help.Content{help.CommandGroup{{Name: "login", Desc: "Sign in"}}},
		},
	}
	var buf bytes.Buffer
	require.NoError(t, help.NewRenderer(nil).Render(&buf, sections))
	require.Equal(
		t,
		"Agents\n\n  charlie  Start charlie\n\n  run      Run a tool\n\nAccount\n\n  login  Sign in\n",
		ansi.Strip(buf.String()),
	)
	buf.Reset()
	r := help.NewRenderer(nil, help.WithCommandAlignMode(help.AlignModeGlobal))
	require.NoError(t, r.Render(&buf, sections))
	require.Equal(t, `Agents

  charlie  Start charlie

  run      Run a tool

Account

  login    Sign in
`, ansi.Strip(buf.String()))
}

func TestRender_CommandAndAliasBlocksAlign(t *testing.T) {
	sections := []help.Section{{Title: "Tools", Content: []help.Content{
		help.CommandGroup{{Name: "run", Desc: "Run a tool"}},
		help.AliasGroup{{Name: "publish", Target: "release"}},
	}}}
	var buf bytes.Buffer
	require.NoError(t, help.NewRenderer(nil).Render(&buf, sections))
	require.Equal(
		t,
		"Tools\n\n  run      Run a tool\n\n  publish  Alias for release\n",
		ansi.Strip(buf.String()),
	)
}

func TestRender_NestedCommandSectionsAlignIndependently(t *testing.T) {
	sections := []help.Section{{Title: "Top", Content: []help.Content{
		help.CommandGroup{{Name: "a", Desc: "short"}},
		&help.Section{Title: "Sub", Content: []help.Content{
			help.CommandGroup{{Name: "verylongname", Desc: "long"}},
			help.CommandGroup{{Name: "b", Desc: "nested"}},
		}},
	}}}
	var buf bytes.Buffer
	require.NoError(t, help.NewRenderer(nil).Render(&buf, sections))
	require.Equal(
		t,
		"Top\n\n  a  short\n\n    Sub\n\n      verylongname  long\n\n      b             nested\n",
		ansi.Strip(buf.String()),
	)
}
