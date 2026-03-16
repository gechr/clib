package complete_test

import (
	"testing"

	"github.com/gechr/clib/complete"
	"github.com/stretchr/testify/require"
)

func TestValidateShellSafe(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"simple name", "myapp", false},
		{"with hyphens", "my-app", false},
		{"with underscores", "my_app", false},
		{"with dots", "my.app", false},
		{"with digits", "app2", false},
		{"with at sign", "@complete", false},
		{"mixed safe chars", "my-app_v2.0", false},
		{"empty", "", true},
		{"space", "my app", true},
		{"semicolon", "app;echo", true},
		{"backtick", "app`id`", true},
		{"dollar", "app$HOME", true},
		{"single quote", "app'", true},
		{"double quote", `app"`, true},
		{"pipe", "app|cat", true},
		{"ampersand", "app&", true},
		{"newline", "app\necho", true},
		{"parenthesis", "$(cmd)", true},
		{"slash", "path/to/app", true},
		{"backslash", `app\n`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := complete.ValidateShellSafe(tt.input, "test")
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateShellSafe_LabelInError(t *testing.T) {
	err := complete.ValidateShellSafe("bad;input", "AppName")
	require.EqualError(t, err, `AppName contains unsafe characters: "bad;input"`)
}

func TestValidateShellSafe_EmptyLabel(t *testing.T) {
	err := complete.ValidateShellSafe("", "Dynamic")
	require.EqualError(t, err, "Dynamic must not be empty")
}

func TestValidateGenerator_SubcommandIdentifiers(t *testing.T) {
	t.Run("subcommand name", func(t *testing.T) {
		g := &complete.Generator{
			AppName: "app",
			Subs: []complete.SubSpec{
				{Name: "bad;name"},
			},
		}

		err := complete.ValidateGenerator(g)
		require.EqualError(t, err, `SubcommandName contains unsafe characters: "bad;name"`)
	})

	t.Run("subcommand alias", func(t *testing.T) {
		g := &complete.Generator{
			AppName: "app",
			Subs: []complete.SubSpec{
				{Name: "good", Aliases: []string{"bad alias"}},
			},
		}

		err := complete.ValidateGenerator(g)
		require.EqualError(t, err, `SubcommandAlias contains unsafe characters: "bad alias"`)
	})
}

func TestValidateGenerator_SpecIdentifiers(t *testing.T) {
	t.Run("long flag", func(t *testing.T) {
		g := &complete.Generator{
			AppName: "app",
			Specs: []complete.Spec{
				{LongFlag: "bad;flag"},
			},
		}

		err := complete.ValidateGenerator(g)
		require.EqualError(t, err, `LongFlag contains unsafe characters: "bad;flag"`)
	})

	t.Run("short flag", func(t *testing.T) {
		g := &complete.Generator{
			AppName: "app",
			Specs: []complete.Spec{
				{ShortFlag: "$"},
			},
		}

		err := complete.ValidateGenerator(g)
		require.EqualError(t, err, `ShortFlag contains unsafe characters: "$"`)
	})

	t.Run("extension", func(t *testing.T) {
		g := &complete.Generator{
			AppName: "app",
			Specs: []complete.Spec{
				{Extension: "yaml,$bad"},
			},
		}

		err := complete.ValidateGenerator(g)
		require.EqualError(t, err, `Extension contains unsafe characters: "$bad"`)
	})
}
