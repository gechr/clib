package kong_test

import (
	"testing"

	"github.com/gechr/clib/cli/kong"
	"github.com/gechr/clib/complete"
	"github.com/stretchr/testify/require"
)

type testCLI struct {
	Name    string       `name:"name"        help:"Your name"             short:"n" clib:"terse='Name'"`
	Verbose bool         `name:"verbose"     help:"Enable verbose output" short:"v"`
	Draft   *bool        `name:"draft"       help:"Filter by draft"       short:"D" clib:"terse='Draft'"                               negatable:""`
	Output  string       `name:"output"      help:"Output format"         short:"o" clib:"terse='Format',highlight='j,y,t'"            enum:"json,yaml,text"`
	Hidden  string       `name:"hidden-flag" help:"A hidden flag"         hidden:""`
	Limit   *int         `name:"limit"       help:"Maximum results"       short:"L"`
	Authors kong.CSVFlag `name:"authors"     help:"Filter by authors"     short:"a" clib:"terse='Authors',complete='predictor=author'"`
	Query   []string     `name:"query"       help:"Search query"          arg:""`
}

func findFlagByName(flags []complete.FlagMeta, name string) *complete.FlagMeta {
	for i := range flags {
		if flags[i].Name == name {
			return &flags[i]
		}
	}
	return nil
}

func TestReflect_Basic(t *testing.T) {
	flags, err := kong.Reflect(&testCLI{})
	require.NoError(t, err)

	// Should have 8 flags (Name, Verbose, Draft, Output, Hidden, Limit, Authors, Query).
	require.Len(t, flags, 8)
}

func TestReflect_NameFlag(t *testing.T) {
	flags, err := kong.Reflect(&testCLI{})
	require.NoError(t, err)
	f := findFlagByName(flags, "name")
	require.NotNil(t, f)

	require.Equal(t, "n", f.Short)
	require.Equal(t, "Your name", f.Help)
	require.Equal(t, "Name", f.Terse)
	require.Equal(t, "Name", f.Desc())
	require.True(t, f.HasArg)
	require.False(t, f.Hidden)
	require.False(t, f.IsArg)
}

func TestReflect_VerboseFlag(t *testing.T) {
	flags, err := kong.Reflect(&testCLI{})
	require.NoError(t, err)
	f := findFlagByName(flags, "verbose")
	require.NotNil(t, f)

	require.Equal(t, "v", f.Short)
	require.False(t, f.HasArg, "bool flags should not have args")
	require.False(t, f.Negatable, "plain bool should not be negatable")
}

func TestReflect_NegatableFlag(t *testing.T) {
	flags, err := kong.Reflect(&testCLI{})
	require.NoError(t, err)
	f := findFlagByName(flags, "draft")
	require.NotNil(t, f)

	require.True(t, f.Negatable)
	require.Equal(t, "D", f.Short)
	require.False(t, f.HasArg, "*bool flags should not have args")
}

func TestReflect_EnumFlag(t *testing.T) {
	flags, err := kong.Reflect(&testCLI{})
	require.NoError(t, err)
	f := findFlagByName(flags, "output")
	require.NotNil(t, f)

	require.Equal(t, []string{"json", "yaml", "text"}, f.Enum)
	require.Equal(t, "Format", f.Terse)
}

func TestReflect_EnumHighlight(t *testing.T) {
	flags, err := kong.Reflect(&testCLI{})
	require.NoError(t, err)
	f := findFlagByName(flags, "output")
	require.NotNil(t, f)

	require.Equal(t, []string{"j", "y", "t"}, f.EnumHighlight)
}

func TestReflect_HiddenFlag(t *testing.T) {
	flags, err := kong.Reflect(&testCLI{})
	require.NoError(t, err)
	f := findFlagByName(flags, "hidden-flag")
	require.NotNil(t, f)

	require.True(t, f.Hidden)
}

func TestReflect_PointerFlag(t *testing.T) {
	flags, err := kong.Reflect(&testCLI{})
	require.NoError(t, err)
	f := findFlagByName(flags, "limit")
	require.NotNil(t, f)

	require.True(t, f.HasArg, "*int should have arg")
}

func TestReflect_CSVFlag(t *testing.T) {
	flags, err := kong.Reflect(&testCLI{})
	require.NoError(t, err)
	f := findFlagByName(flags, "authors")
	require.NotNil(t, f)

	require.True(t, f.IsCSV)
	require.True(t, f.HasArg)
	require.Equal(t, "predictor=author", f.Complete)
}

func TestReflect_ArgFlag(t *testing.T) {
	flags, err := kong.Reflect(&testCLI{})
	require.NoError(t, err)

	var argFlag *complete.FlagMeta
	for i := range flags {
		if flags[i].IsArg {
			argFlag = &flags[i]
			break
		}
	}
	require.NotNil(t, argFlag)
	require.True(t, argFlag.IsArg)
}

func TestFlagMeta_Desc_FallbackToHelp(t *testing.T) {
	f := complete.FlagMeta{Help: "Long help text"}
	require.Equal(t, "Long help text", f.Desc())
}

func TestFlagMeta_Desc_ExplicitTerse(t *testing.T) {
	f := complete.FlagMeta{Help: "Long help text", Terse: "Short"}
	require.Equal(t, "Short", f.Desc())
}

func TestReflect_NonStruct(t *testing.T) {
	s := "not a struct"
	flags, err := kong.Reflect(&s)
	require.NoError(t, err)
	require.Nil(t, flags)
}

func TestReflect_NilPointer(t *testing.T) {
	// Passing a typed nil pointer should return nil, not panic.
	var cli *testCLI
	flags, err := kong.Reflect(cli)
	require.NoError(t, err)
	require.Nil(t, flags)
}

type TestEmbeddedBase struct {
	Debug bool `name:"debug" help:"Debug mode"`
}

type testEmbeddedCLI struct {
	TestEmbeddedBase

	Name string `name:"name" help:"Name"`
}

func TestReflect_EmbeddedStruct(t *testing.T) {
	flags, err := kong.Reflect(&testEmbeddedCLI{})
	require.NoError(t, err)
	require.Len(t, flags, 2)

	f := findFlagByName(flags, "debug")
	require.NotNil(t, f)
	require.False(t, f.HasArg)
}

// testPrlLikeCLI mirrors prl's CLI struct to exercise all tag varieties.
type testPrlLikeCLI struct {
	kong.CompletionFlags

	Query   []string      `help:"Search query" arg:""                   optional:""`
	Org     kong.CSVFlag  `name:"org"          help:"GitHub orgs"       aliases:"organization,owner"                                    clib:"terse='Organization'"`
	Repo    string        `name:"repo"         help:"Limit to repo"     short:"R"                                                       clib:"terse='Repository',complete='predictor=repo'"`
	Author  *kong.CSVFlag `name:"author"       help:"Filter by authors" short:"a"                                                       clib:"terse='Author',complete='predictor=author'"`
	Columns kong.CSVFlag  `name:"columns"      help:"Table columns"     clib:"terse='Table columns',complete='predictor=columns,comma'"`
	State   string        `name:"state"        help:"PR state"          short:"s"                                                       clib:"terse='State',complete='values=open closed merged all'"`
	Match   string        `name:"match"        help:"Restrict search"   clib:"terse='Search field'"                                     placeholder:"field"`
	Verbose bool          `name:"verbose"      help:"Verbose output"`
	Limit   *int          `name:"limit"        help:"Max results"       short:"L"                                                       clib:"group='output'"`
	Filter  []string      `name:"filter"       help:"Search qualifier"  short:"f"`
	Debug   bool          `name:"debug"        help:"Debug mode"`
	NoName  string        // no name tag -> auto-derived as "no-name"

	Sub struct{} `help:"A subcommand" cmd:""` // cmd fields should be skipped
}

func TestReflect_PrlLikeCLI(t *testing.T) {
	flags, err := kong.Reflect(&testPrlLikeCLI{})
	require.NoError(t, err)

	// 1 arg (Query) + 10 named + 1 auto-derived (NoName) = 12 total
	// (CompletionFlags excluded, Sub cmd excluded).
	require.Len(t, flags, 12)
}

func TestReflect_Aliases(t *testing.T) {
	flags, err := kong.Reflect(&testPrlLikeCLI{})
	require.NoError(t, err)
	f := findFlagByName(flags, "org")
	require.NotNil(t, f)

	require.Equal(t, []string{"organization", "owner"}, f.Aliases)
}

func TestReflect_Optional(t *testing.T) {
	flags, err := kong.Reflect(&testPrlLikeCLI{})
	require.NoError(t, err)

	var arg *complete.FlagMeta
	for i := range flags {
		if flags[i].IsArg {
			arg = &flags[i]
			break
		}
	}
	require.NotNil(t, arg)
	require.True(t, arg.Optional)
}

func TestReflect_Placeholder(t *testing.T) {
	flags, err := kong.Reflect(&testPrlLikeCLI{})
	require.NoError(t, err)
	f := findFlagByName(flags, "match")
	require.NotNil(t, f)

	require.Equal(t, "field", f.Placeholder)
}

func TestReflect_PlaceholderOverride_NativeTag(t *testing.T) {
	flags, err := kong.Reflect(&testPrlLikeCLI{})
	require.NoError(t, err)
	f := findFlagByName(flags, "match")
	require.NotNil(t, f)

	require.True(t, f.PlaceholderOverride, "native placeholder tag should set PlaceholderOverride")
}

func TestReflect_PlaceholderOverride_ClibTag(t *testing.T) {
	type CLI struct {
		Output string `name:"output" help:"Output" clib:"placeholder='path'"`
	}
	flags, err := kong.Reflect(&CLI{})
	require.NoError(t, err)
	f := findFlagByName(flags, "output")
	require.NotNil(t, f)

	require.Equal(t, "path", f.Placeholder)
	require.True(t, f.PlaceholderOverride, "clib placeholder tag should set PlaceholderOverride")
}

func TestReflect_PlaceholderOverride_NotSetWhenEmpty(t *testing.T) {
	type CLI struct {
		Name string `name:"name" help:"Name"`
	}
	flags, err := kong.Reflect(&CLI{})
	require.NoError(t, err)
	f := findFlagByName(flags, "name")
	require.NotNil(t, f)

	require.False(t, f.PlaceholderOverride, "empty placeholder should not set PlaceholderOverride")
}

func TestReflect_Group(t *testing.T) {
	flags, err := kong.Reflect(&testPrlLikeCLI{})
	require.NoError(t, err)
	f := findFlagByName(flags, "limit")
	require.NotNil(t, f)

	require.Equal(t, "output", f.Group)
}

func TestReflect_PointerCSVFlag(t *testing.T) {
	flags, err := kong.Reflect(&testPrlLikeCLI{})
	require.NoError(t, err)
	f := findFlagByName(flags, "author")
	require.NotNil(t, f)

	require.True(t, f.IsCSV, "*CSVFlag should be detected as CSV")
	require.True(t, f.HasArg)
	require.Equal(t, "predictor=author", f.Complete)
}

func TestReflect_IsSlice(t *testing.T) {
	flags, err := kong.Reflect(&testPrlLikeCLI{})
	require.NoError(t, err)
	f := findFlagByName(flags, "filter")
	require.NotNil(t, f)

	require.True(t, f.IsSlice)
	require.True(t, f.HasArg)
}

func TestReflect_CompleteTag(t *testing.T) {
	flags, err := kong.Reflect(&testPrlLikeCLI{})
	require.NoError(t, err)

	f := findFlagByName(flags, "repo")
	require.NotNil(t, f)
	require.Equal(t, "predictor=repo", f.Complete)

	f = findFlagByName(flags, "columns")
	require.NotNil(t, f)
	require.Equal(t, "predictor=columns,comma", f.Complete)

	f = findFlagByName(flags, "state")
	require.NotNil(t, f)
	require.Equal(t, "values=open closed merged all", f.Complete)
}

func TestReflect_AutoDeriveName(t *testing.T) {
	flags, err := kong.Reflect(&testPrlLikeCLI{})
	require.NoError(t, err)
	f := findFlagByName(flags, "no-name")
	require.NotNil(t, f, "field without name tag should auto-derive kebab-case name")
	require.Equal(t, "NoName", f.Origin)
	require.True(t, f.HasArg, "string field should have arg")
}

func TestReflect_SkipCmdFields(t *testing.T) {
	flags, err := kong.Reflect(&testPrlLikeCLI{})
	require.NoError(t, err)
	for _, f := range flags {
		if f.Origin == "Sub" {
			t.Fatal("cmd field should be skipped")
		}
	}
}

func TestReflect_ArgOrigin(t *testing.T) {
	flags, err := kong.Reflect(&testPrlLikeCLI{})
	require.NoError(t, err)

	var arg *complete.FlagMeta
	for i := range flags {
		if flags[i].IsArg {
			arg = &flags[i]
			break
		}
	}
	require.NotNil(t, arg)
	require.Equal(t, "Query", arg.Origin)
	require.True(t, arg.IsSlice)
	require.True(t, arg.HasArg)
}

type testEmbeddedPtrCLI struct {
	*TestEmbeddedBase

	Name string `name:"name" help:"Name"`
}

func TestReflect_EmbeddedPointerStruct(t *testing.T) {
	flags, err := kong.Reflect(&testEmbeddedPtrCLI{})
	require.NoError(t, err)
	require.Len(t, flags, 2)

	f := findFlagByName(flags, "debug")
	require.NotNil(t, f)
	require.False(t, f.HasArg)
}

func TestReflect_EnumDefaultFromNativeTag(t *testing.T) {
	type CLI struct {
		Color string `name:"color" help:"Color mode" default:"auto" enum:"auto,always,never"`
	}

	flags, err := kong.Reflect(&CLI{})
	require.NoError(t, err)
	f := findFlagByName(flags, "color")
	require.NotNil(t, f)

	require.Equal(t, []string{"auto", "always", "never"}, f.Enum)
	require.Equal(t, "auto", f.EnumDefault, "should read default from native tag")
}

func TestReflect_EnumDefaultClibOverridesNative(t *testing.T) {
	type CLI struct {
		State string `name:"state" help:"State" clib:"default='open'" default:"closed" enum:"open,closed"`
	}

	flags, err := kong.Reflect(&CLI{})
	require.NoError(t, err)
	f := findFlagByName(flags, "state")
	require.NotNil(t, f)

	require.Equal(t, "open", f.EnumDefault, "clib default should take precedence")
}

func TestReflect_EnumDefaultNotSetWithoutEnum(t *testing.T) {
	type CLI struct {
		Name string `name:"name" help:"Name" default:"foo"`
	}

	flags, err := kong.Reflect(&CLI{})
	require.NoError(t, err)
	f := findFlagByName(flags, "name")
	require.NotNil(t, f)

	require.Empty(t, f.EnumDefault, "should not set EnumDefault when no enum")
}

func TestReflect_FieldNameToFlag(t *testing.T) {
	type CLI struct {
		Config     string `help:"Config file"`
		NoConfig   bool   `help:"Disable config"`
		DryRun     bool   `help:"Dry run"`
		Verbose    bool   `help:"Verbose"`
		Color      string `help:"Color mode"     enum:"auto,always,never"`
		MaxMajor   string `help:"Max major"`
		S3         bool   `help:"S3"`
		HTMLParser string `help:"HTML parser"`
	}

	flags, err := kong.Reflect(&CLI{})
	require.NoError(t, err)
	for _, tc := range []struct{ field, want string }{
		{"Config", "config"},
		{"NoConfig", "no-config"},
		{"DryRun", "dry-run"},
		{"Verbose", "verbose"},
		{"Color", "color"},
		{"MaxMajor", "max-major"},
		{"S3", "s3"},
		{"HTMLParser", "html-parser"},
	} {
		var found *complete.FlagMeta
		for i := range flags {
			if flags[i].Origin == tc.field {
				found = &flags[i]
				break
			}
		}
		require.NotNil(t, found, "field %s not found", tc.field)
		require.Equal(t, tc.want, found.Name, "field %s", tc.field)
	}
}

func TestReflect_ExcludesCompletionFlags(t *testing.T) {
	flags, err := kong.Reflect(&testPrlLikeCLI{})
	require.NoError(t, err)

	completionNames := []string{
		complete.FlagComplete, complete.FlagShell, "install-completion",
		"uninstall-completion", "print-completion",
	}
	for _, name := range completionNames {
		f := findFlagByName(flags, name)
		require.Nil(t, f, "CompletionFlags field %q should be excluded", name)
	}
}

func TestReflect_ValueHint(t *testing.T) {
	type CLI struct {
		Output string `name:"output" help:"Output path" clib:"hint='file'"`
	}
	flags, err := kong.Reflect(&CLI{})
	require.NoError(t, err)
	f := findFlagByName(flags, "output")
	require.NotNil(t, f)
	require.Equal(t, "file", f.ValueHint)
}

func TestReflect_PositiveNegativeDesc(t *testing.T) {
	type CLI struct {
		Draft bool `name:"draft" help:"Filter by draft" clib:"positive='Include drafts',negative='Exclude drafts'" negatable:""`
	}
	flags, err := kong.Reflect(&CLI{})
	require.NoError(t, err)
	f := findFlagByName(flags, "draft")
	require.NotNil(t, f)
	require.True(t, f.Negatable)
	require.Equal(t, "Include drafts", f.PositiveDesc)
	require.Equal(t, "Exclude drafts", f.NegativeDesc)
}

func TestReflect_ExtTag(t *testing.T) {
	type CLI struct {
		Config string `name:"config" help:"Config file" clib:"ext='yaml'"`
	}

	flags, err := kong.Reflect(&CLI{})
	require.NoError(t, err)
	f := findFlagByName(flags, "config")
	require.NotNil(t, f)
	require.Equal(t, "yaml", f.Extension)
}
