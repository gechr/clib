//nolint:dupword // fish script keywords (e.g. "end\nend") trigger false positives
package fish

import (
	"testing"

	"github.com/gechr/clib/complete"
	"github.com/stretchr/testify/require"
)

// --- Generator helpers ---

func flatGen() *complete.Generator {
	return complete.NewGenerator("clibapp").FromFlags([]complete.FlagMeta{
		{
			Name:     "author",
			Short:    "a",
			Terse:    "Author",
			Help:     "Filter by author",
			HasArg:   true,
			Complete: "predictor=author",
		},
		{
			Name:   "state",
			Short:  "s",
			Terse:  "State",
			Help:   "Filter by state",
			HasArg: true,
			Enum:   []string{"open", "closed", "merged", "all"},
		},
		{Name: "verbose", Short: "v", Terse: "Verbose", Help: "Enable verbose output"},
		{
			Name:     "columns",
			Terse:    "Table columns",
			Help:     "Table columns",
			HasArg:   true,
			Complete: "predictor=columns,comma",
		},
		{Name: "hidden-flag", Help: "A hidden flag", HasArg: true, Hidden: true},
		{Name: "limit", Short: "L", Terse: "Max results", Help: "Max results", HasArg: true},
		{
			Name:     "ci",
			Terse:    "CI status",
			Help:     "Filter by CI status",
			HasArg:   true,
			Complete: "values=success failure pending",
		},
		{Name: "merge", Terse: "Auto-merge", Help: "Enable auto-merge", Negatable: true},
		{Name: "debug", Terse: "Debug", Help: "Enable debug logging", Complete: "-"},
		{Name: "query", Help: "Search query", IsArg: true, HasArg: true, IsSlice: true},
	})
}

func genSubcommands() *complete.Generator {
	return &complete.Generator{
		AppName: "myapp",
		Specs: []complete.Spec{
			{LongFlag: "verbose", ShortFlag: "v", Terse: "Verbose output"},
			{
				LongFlag: "color",
				Terse:    "Color mode",
				HasArg:   true,
				Values:   []string{"auto", "always", "never"},
			},
			{LongFlag: "@secret", HasArg: true, Hidden: true},
		},
		Subs: []complete.SubSpec{
			{
				Name:  "build",
				Terse: "Build the project",
				Specs: []complete.Spec{
					{LongFlag: "output", ShortFlag: "o", Terse: "Output path", HasArg: true},
					{LongFlag: "release", Terse: "Release build"},
				},
			},
			{
				Name:    "test",
				Aliases: []string{"t"},
				Terse:   "Run tests",
				Specs: []complete.Spec{
					{LongFlag: "coverage", Terse: "Enable coverage"},
					{LongFlag: "run", ShortFlag: "r", Terse: "Test pattern", HasArg: true},
				},
			},
		},
	}
}

func genNested() *complete.Generator {
	return &complete.Generator{
		AppName: "myapp",
		Specs: []complete.Spec{
			{LongFlag: "verbose", Terse: "Verbose"},
		},
		Subs: []complete.SubSpec{
			{
				Name:  "auth",
				Terse: "Manage authentication",
				Specs: []complete.Spec{
					{LongFlag: "token", Terse: "Auth token", HasArg: true},
				},
				Subs: []complete.SubSpec{
					{
						Name:  "login",
						Terse: "Log in",
						Specs: []complete.Spec{
							{LongFlag: "browser", Terse: "Open browser"},
						},
					},
					{
						Name:  "logout",
						Terse: "Log out",
					},
				},
			},
			{
				Name:  "run",
				Terse: "Run command",
			},
		},
	}
}

func hintsGen() *complete.Generator {
	return &complete.Generator{
		AppName: "myapp",
		Specs: []complete.Spec{
			{
				LongFlag:  "config",
				ShortFlag: "c",
				Terse:     "Config file",
				HasArg:    true,
				Extension: "yaml,yml",
			},
			{
				LongFlag:  "output",
				ShortFlag: "o",
				Terse:     "Output path",
				HasArg:    true,
				ValueHint: complete.HintFile,
			},
			{
				LongFlag:  "dir",
				ShortFlag: "d",
				Terse:     "Directory",
				HasArg:    true,
				ValueHint: complete.HintDir,
			},
			{
				LongFlag:  "shell",
				Terse:     "Shell command",
				HasArg:    true,
				ValueHint: complete.HintCommand,
			},
			{LongFlag: "user", Terse: "User name", HasArg: true, ValueHint: complete.HintUser},
			{LongFlag: "host", Terse: "Host name", HasArg: true, ValueHint: complete.HintHost},
			{LongFlag: "verbose", ShortFlag: "v", Terse: "Verbose output"},
		},
	}
}

func valuesGen() *complete.Generator {
	return &complete.Generator{
		AppName: "myapp",
		Specs: []complete.Spec{
			{
				LongFlag:  "format",
				ShortFlag: "f",
				Terse:     "Output format",
				HasArg:    true,
				ValueDescs: []complete.ValueDesc{
					{Value: "json", Desc: "JSON output"},
					{Value: "yaml", Desc: "YAML output"},
					{Value: "text", Desc: "Plain text"},
				},
			},
			{
				LongFlag:  "tags",
				ShortFlag: "t",
				Terse:     "Filter tags",
				HasArg:    true,
				CommaList: true,
				Values:    []string{"bug", "feature", "docs"},
			},
			{LongFlag: "verbose", ShortFlag: "v", Terse: "Verbose output"},
		},
	}
}

func commaGen() *complete.Generator {
	return &complete.Generator{
		AppName: "myapp",
		Specs: []complete.Spec{
			{
				LongFlag:  "tags",
				ShortFlag: "t",
				Terse:     "Filter tags",
				HasArg:    true,
				CommaList: true,
				Values:    []string{"bug", "feature", "docs"},
			},
			{
				LongFlag:  "labels",
				Terse:     "Labels",
				HasArg:    true,
				CommaList: true,
				Dynamic:   "labels",
			},
			{LongFlag: "verbose", ShortFlag: "v", Terse: "Verbose output"},
		},
	}
}

func pathArgsGen() *complete.Generator {
	return &complete.Generator{
		AppName: "myapp",
		Specs: []complete.Spec{
			{LongFlag: "verbose", ShortFlag: "v", Terse: "Verbose output"},
		},
		Subs: []complete.SubSpec{
			{
				Name:     "edit",
				Terse:    "Edit files",
				PathArgs: true,
				Specs: []complete.Spec{
					{LongFlag: "editor", Terse: "Editor command", HasArg: true},
				},
			},
			{
				Name:  "list",
				Terse: "List items",
			},
		},
	}
}

func dynamicArgsGen() *complete.Generator {
	return &complete.Generator{
		AppName:     "myapp",
		DynamicArgs: []string{"items", "subitems"},
		Specs: []complete.Spec{
			{LongFlag: "verbose", ShortFlag: "v", Terse: "Verbose output"},
		},
	}
}

func hyphenatedGen() *complete.Generator {
	return &complete.Generator{
		AppName: "my-app",
		Specs: []complete.Spec{
			{LongFlag: "verbose", ShortFlag: "v", Terse: "Verbose output"},
		},
		Subs: []complete.SubSpec{
			{
				Name:  "build",
				Terse: "Build the project",
				Specs: []complete.Spec{
					{LongFlag: "output", ShortFlag: "o", Terse: "Output path", HasArg: true},
				},
			},
			{
				Name:  "test",
				Terse: "Run tests",
			},
		},
	}
}

// --- Generate tests ---

func TestGenerate_Flat(t *testing.T) {
	out, err := Generate(flatGen())
	require.NoError(t, err)

	expected := `complete -c clibapp -f

# Comma-separated columns completion
function __clibapp_complete_columns
    set -l value (string replace -r '^--columns=' '' -- (commandline -ct))
    set -l _flag_columns (clibapp --@complete=columns)
    if string match -qr '^(?<prefix>.*,)' -- $value
        set -l selected (string split ',' -- $prefix)
        for col in $_flag_columns
            if not contains -- $col $selected
                printf '%s\n' "$prefix$col"
            end
        end
    else
        printf '%s\n' $_flag_columns
    end
end

complete -c clibapp -s a -l author -x -a "(clibapp --@complete=author)" -d "Author"
complete -c clibapp -l ci -x -a "success failure pending" -d "CI status"
complete -c clibapp -l columns -x -kra "(__clibapp_complete_columns)" -d "Table columns"
complete -c clibapp -s L -l limit -r -d "Max results"
complete -c clibapp -l merge -d "Auto-merge"
complete -c clibapp -l no-merge -d "Disable auto-merge"
complete -c clibapp -s s -l state -x -a "open closed merged all" -d "State"
complete -c clibapp -s v -l verbose -d "Verbose"
`
	require.Equal(t, expected, out)
}

func TestGenerate_Subcommands(t *testing.T) {
	out, err := Generate(genSubcommands())
	require.NoError(t, err)

	expected := `complete -c myapp -f

function __myapp_global_optspecs
    string join \n -- 'v/verbose' 'color='
end

function __myapp_needs_command
    set -l cmd (commandline -xpc)
    set -e cmd[1]
    argparse -s (__myapp_global_optspecs) -- $cmd 2>/dev/null
    or return
    if set -q argv[1]
        printf '%s\n' $argv
        return 1
    end
    return 0
end

function __myapp_using_subcommand
    set -l cmd (__myapp_needs_command)
    test -z "$cmd"
    and return 1
    for arg in $argv
        if contains -- $arg $cmd
            return 0
        end
    end
    return 1
end

complete -c myapp -n '__myapp_needs_command' -a build -d "Build the project"
complete -c myapp -n '__myapp_needs_command' -a test -d "Run tests"

complete -c myapp -n '__myapp_needs_command' -l color -x -a "auto always never" -d "Color mode"
complete -c myapp -n '__myapp_needs_command' -s v -l verbose -d "Verbose output"

complete -c myapp -n '__myapp_using_subcommand build' -s o -l output -r -d "Output path"
complete -c myapp -n '__myapp_using_subcommand build' -l release -d "Release build"

complete -c myapp -n '__myapp_using_subcommand test t' -l coverage -d "Enable coverage"
complete -c myapp -n '__myapp_using_subcommand test t' -s r -l run -r -d "Test pattern"
`
	require.Equal(t, expected, out)
}

func TestGenerate_Nested(t *testing.T) {
	out, err := Generate(genNested())
	require.NoError(t, err)

	expected := `complete -c myapp -f

function __myapp_global_optspecs
    string join \n -- 'verbose'
end

function __myapp_needs_command
    set -l cmd (commandline -xpc)
    set -e cmd[1]
    argparse -s (__myapp_global_optspecs) -- $cmd 2>/dev/null
    or return
    if set -q argv[1]
        printf '%s\n' $argv
        return 1
    end
    return 0
end

function __myapp_using_subcommand
    set -l cmd (__myapp_needs_command)
    test -z "$cmd"
    and return 1
    for arg in $argv
        if contains -- $arg $cmd
            return 0
        end
    end
    return 1
end

complete -c myapp -n '__myapp_needs_command' -a auth -d "Manage authentication"
complete -c myapp -n '__myapp_needs_command' -a run -d "Run command"

complete -c myapp -n '__myapp_needs_command' -l verbose -d "Verbose"

complete -c myapp -n '__myapp_using_subcommand auth; and not __myapp_using_subcommand login logout' -a login -d "Log in"
complete -c myapp -n '__myapp_using_subcommand auth; and not __myapp_using_subcommand login logout' -a logout -d "Log out"
complete -c myapp -n '__myapp_using_subcommand auth; and not __myapp_using_subcommand login logout' -l token -r -d "Auth token"

complete -c myapp -n '__myapp_using_subcommand auth; and __myapp_using_subcommand login' -l browser -d "Open browser"
`
	require.Equal(t, expected, out)
}

func TestGenerate_Hints(t *testing.T) {
	out, err := Generate(hintsGen())
	require.NoError(t, err)

	expected := `complete -c myapp -f

complete -c myapp -s c -l config -k -xa "(__fish_complete_suffix .yaml .yml)" -d "Config file"
complete -c myapp -s d -l dir -r -f -a "(__fish_complete_directories)" -d "Directory"
complete -c myapp -l host -r -f -a "(__fish_print_hostnames)" -d "Host name"
complete -c myapp -s o -l output -r -F -d "Output path"
complete -c myapp -l shell -r -f -a "(__fish_complete_command)" -d "Shell command"
complete -c myapp -l user -r -f -a "(__fish_complete_users)" -d "User name"
complete -c myapp -s v -l verbose -d "Verbose output"
`
	require.Equal(t, expected, out)
}

func TestGenerate_Values(t *testing.T) {
	out, err := Generate(valuesGen())
	require.NoError(t, err)

	expected := `complete -c myapp -f

# format value completion
function __myapp_complete_format
    set -l _flag_format "json" "yaml" "text"
    set -l _flag_format_desc "JSON output" "YAML output" "Plain text"
    for i in (seq (count $_flag_format))
        printf '%s\t%s\n' $_flag_format[$i] $_flag_format_desc[$i]
    end
end

# Comma-separated tags completion
function __myapp_complete_tags
    set -l value (string replace -r '^--tags=' '' -- (commandline -ct))
    set -l _flag_tags "bug" "feature" "docs"
    if string match -qr '^(?<prefix>.*,)' -- $value
        set -l selected (string split ',' -- $prefix)
        for col in $_flag_tags
            if not contains -- $col $selected
                printf '%s\n' "$prefix$col"
            end
        end
    else
        printf '%s\n' $_flag_tags
    end
end

complete -c myapp -s f -l format -x -ra "(__myapp_complete_format)" -d "Output format"
complete -c myapp -s t -l tags -x -kra "(__myapp_complete_tags)" -d "Filter tags"
complete -c myapp -s v -l verbose -d "Verbose output"
`
	require.Equal(t, expected, out)
}

func TestGenerate_CommaList(t *testing.T) {
	out, err := Generate(commaGen())
	require.NoError(t, err)

	expected := `complete -c myapp -f

# Comma-separated labels completion
function __myapp_complete_labels
    set -l value (string replace -r '^--labels=' '' -- (commandline -ct))
    set -l _flag_labels (myapp --@complete=labels)
    if string match -qr '^(?<prefix>.*,)' -- $value
        set -l selected (string split ',' -- $prefix)
        for col in $_flag_labels
            if not contains -- $col $selected
                printf '%s\n' "$prefix$col"
            end
        end
    else
        printf '%s\n' $_flag_labels
    end
end

# Comma-separated tags completion
function __myapp_complete_tags
    set -l value (string replace -r '^--tags=' '' -- (commandline -ct))
    set -l _flag_tags "bug" "feature" "docs"
    if string match -qr '^(?<prefix>.*,)' -- $value
        set -l selected (string split ',' -- $prefix)
        for col in $_flag_tags
            if not contains -- $col $selected
                printf '%s\n' "$prefix$col"
            end
        end
    else
        printf '%s\n' $_flag_tags
    end
end

complete -c myapp -l labels -x -kra "(__myapp_complete_labels)" -d "Labels"
complete -c myapp -s t -l tags -x -kra "(__myapp_complete_tags)" -d "Filter tags"
complete -c myapp -s v -l verbose -d "Verbose output"
`
	require.Equal(t, expected, out)
}

func TestGenerate_PathArgs(t *testing.T) {
	out, err := Generate(pathArgsGen())
	require.NoError(t, err)

	expected := `complete -c myapp -f

function __myapp_global_optspecs
    string join \n -- 'v/verbose'
end

function __myapp_needs_command
    set -l cmd (commandline -xpc)
    set -e cmd[1]
    argparse -s (__myapp_global_optspecs) -- $cmd 2>/dev/null
    or return
    if set -q argv[1]
        printf '%s\n' $argv
        return 1
    end
    return 0
end

function __myapp_using_subcommand
    set -l cmd (__myapp_needs_command)
    test -z "$cmd"
    and return 1
    for arg in $argv
        if contains -- $arg $cmd
            return 0
        end
    end
    return 1
end

complete -c myapp -n '__myapp_needs_command' -a edit -d "Edit files"
complete -c myapp -n '__myapp_needs_command' -a list -d "List items"

complete -c myapp -n '__myapp_needs_command' -s v -l verbose -d "Verbose output"

complete -c myapp -n '__myapp_using_subcommand edit' -l editor -r -d "Editor command"
complete -c myapp -n '__myapp_using_subcommand edit' -F
`
	require.Equal(t, expected, out)
}

func TestGenerate_DynamicArgs(t *testing.T) {
	out, err := Generate(dynamicArgsGen())
	require.NoError(t, err)

	expected := `complete -c myapp -f

complete -c myapp -s v -l verbose -d "Verbose output"

function __myapp_dynamic_args
    set -l tokens (commandline -xpc)
    set -e tokens[1]
    set -l positional
    set -l skip_next 0
    set -l dashdash 0
    for t in $tokens
        if test $dashdash -eq 1
            set -a positional $t
        else if test $skip_next -eq 1
            set skip_next 0
        else if test "$t" = --
            set dashdash 1
        else if not string match -q -- '-*' $t
            set -a positional $t
        end
    end
    set -l nargs (count $positional)
    switch $nargs
        case 0
            myapp --@complete=items
        case 1
            myapp --@complete=subitems -- $positional
        case '*'
            myapp --@complete=subitems -- $positional
    end
end

complete -c myapp -a "(__myapp_dynamic_args)" -f
`
	require.Equal(t, expected, out)
}

func TestGenerate_Hyphenated(t *testing.T) {
	out, err := Generate(hyphenatedGen())
	require.NoError(t, err)

	expected := `complete -c my-app -f

function __my_app_global_optspecs
    string join \n -- 'v/verbose'
end

function __my_app_needs_command
    set -l cmd (commandline -xpc)
    set -e cmd[1]
    argparse -s (__my_app_global_optspecs) -- $cmd 2>/dev/null
    or return
    if set -q argv[1]
        printf '%s\n' $argv
        return 1
    end
    return 0
end

function __my_app_using_subcommand
    set -l cmd (__my_app_needs_command)
    test -z "$cmd"
    and return 1
    for arg in $argv
        if contains -- $arg $cmd
            return 0
        end
    end
    return 1
end

complete -c my-app -n '__my_app_needs_command' -a build -d "Build the project"
complete -c my-app -n '__my_app_needs_command' -a test -d "Run tests"

complete -c my-app -n '__my_app_needs_command' -s v -l verbose -d "Verbose output"

complete -c my-app -n '__my_app_using_subcommand build' -s o -l output -r -d "Output path"
`
	require.Equal(t, expected, out)
}

func TestGenerate_ErrorOnUnsafeAppName(t *testing.T) {
	_, err := Generate(&complete.Generator{AppName: "bad;name"})
	require.EqualError(t, err, `AppName contains unsafe characters: "bad;name"`)
}

func TestGenerate_ErrorOnUnsafeDynamic(t *testing.T) {
	_, err := Generate(&complete.Generator{AppName: "app", DynamicArgs: []string{"bad;arg"}})
	require.EqualError(t, err, `DynamicArgs contains unsafe characters: "bad;arg"`)
}
