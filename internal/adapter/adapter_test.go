package adapter_test

import (
	"testing"

	"github.com/gechr/clib/help"
	"github.com/gechr/clib/internal/adapter"
	"github.com/stretchr/testify/require"
)

func TestApplyLongHelp(t *testing.T) {
	t.Parallel()

	sections := []help.Section{
		{
			Title: "Examples",
			Content: []help.Content{help.Examples{{
				Comment: "List items",
				Command: "app list",
			}}},
		},
		{
			Title: "Usage",
			Content: []help.Content{
				help.Usage{Command: "app"},
				help.Description("A longer description."),
			},
		},
		{
			Title:   "Options",
			Content: []help.Content{help.FlagGroup{{Long: "verbose"}}},
		},
	}

	t.Run("short help", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, []help.Section{
			{
				Title:   "Usage",
				Content: []help.Content{help.Usage{Command: "app"}},
			},
			{
				Title:   "Options",
				Content: []help.Content{help.FlagGroup{{Long: "verbose"}}},
			},
		}, adapter.ApplyLongHelp(sections, []string{"app", "-h"}))
	})

	t.Run("long help", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, []help.Section{
			sections[1],
			sections[2],
			sections[0],
		}, adapter.ApplyLongHelp(sections, []string{"app", "--help"}))
	})

	t.Run("description override", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, []help.Section{
			sections[1],
			sections[2],
		}, adapter.ApplyLongHelp(
			sections,
			[]string{"app", "-h"},
			help.WithAlwaysShowDescription(),
		))
	})

	t.Run("examples override", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, []help.Section{
			sections[0],
			{
				Title:   "Usage",
				Content: []help.Content{help.Usage{Command: "app"}},
			},
			sections[2],
		}, adapter.ApplyLongHelp(
			sections,
			[]string{"app", "-h"},
			help.WithAlwaysShowExamples(),
		))
	})

	t.Run("both overrides", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, sections, adapter.ApplyLongHelp(
			sections,
			[]string{"app", "-h"},
			help.WithAlwaysShowDescription(),
			help.WithAlwaysShowExamples(),
		))
	})
}

func TestNegatableLong(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name         string
		prefix       string
		positiveOnly bool
		negativeOnly bool
		want         string
	}{
		{name: "color", prefix: "no-", want: "[no-]color"},
		{name: "color", prefix: "no-", positiveOnly: true, want: "color"},
		{name: "color", prefix: "no-", negativeOnly: true, want: "no-color"},
		{
			name:         "color",
			prefix:       "no-",
			positiveOnly: true,
			negativeOnly: true,
			want:         "color",
		},
		{name: "feature", prefix: "disable-", want: "[disable-]feature"},
	} {
		require.Equal(t, tt.want, adapter.NegatableLong(
			tt.name,
			tt.prefix,
			tt.positiveOnly,
			tt.negativeOnly,
		))
	}
}

func TestApplyFlagVisibility(t *testing.T) {
	t.Parallel()

	t.Run("applies extras", func(t *testing.T) {
		t.Parallel()

		flag := help.Flag{Long: "verbose", Short: "v"}
		adapter.ApplyFlagVisibility(&flag, true, true, true)

		require.Equal(t, help.Flag{NoIndent: true}, flag)
	})

	t.Run("preserves defaults", func(t *testing.T) {
		t.Parallel()

		flag := help.Flag{Long: "verbose", Short: "v"}
		adapter.ApplyFlagVisibility(&flag, false, false, false)

		require.Equal(t, help.Flag{Long: "verbose", Short: "v"}, flag)
	})
}

func TestNormalizePlaceholder(t *testing.T) {
	t.Parallel()

	require.Equal(t, "VALUE", adapter.NormalizePlaceholder("VALUE", false))
	require.Equal(t, "value", adapter.NormalizePlaceholder("VALUE", true))
}
