package help

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHelpContentMarkers(_ *testing.T) {
	// These are marker methods satisfying the Content interface.
	// Call each directly to ensure coverage of the empty stubs.
	FlagGroup{}.helpContent()
	Args{}.helpContent()
	CommandGroup{}.helpContent()
	Usage{}.helpContent()
	Text("").helpContent()
	Examples{}.helpContent()
	(&Section{}).helpContent()
}

func TestUnclosedBracketCol(t *testing.T) {
	tests := []struct {
		text string
		want int
	}{
		{"no brackets here", -1},
		{"closed [ok]", -1},
		{"open [values", 6},
		{"desc [a, b,", 6},
		{"nested [a, [b], c", 8}, // outer '[' is unclosed
		{"[only bracket", 1},     // col after '['
		{"all closed [a] [b]", -1},
		{"mixed [a] [b, c", 11}, // second '[' is unclosed
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			got := unclosedBracketCol(tt.text)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestTrailingBracketCol(t *testing.T) {
	tests := []struct {
		text string
		want int
	}{
		{"no brackets", -1},
		{"desc [a, b, c]", 5},
		{"desc [default: x] [a, b]", 18}, // trailing pair
		{"[only brackets]", -1},          // entire string is bracket
		{"desc [a, [b], c]", 5},          // nested, trailing ']' matches outer '['
		{"desc [a] suffix", -1},          // ']' not at end
		{"desc text", -1},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			got := trailingBracketCol(tt.text)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestApply(t *testing.T) {
	sections := []Section{
		{Title: "A", Content: []Content{Text("hello")}},
	}
	addB := OptionFunc(func(s []Section) []Section {
		return append(s, Section{Title: "B", Content: []Content{Text("world")}})
	})
	result := Apply(sections, addB)
	require.Len(t, result, 2)
	require.Equal(t, "A", result[0].Title)
	require.Equal(t, "B", result[1].Title)
}

func TestApply_NoOpts(t *testing.T) {
	sections := []Section{{Title: "X"}}
	result := Apply(sections)
	require.Len(t, result, 1)
	require.Equal(t, "X", result[0].Title)
}

func TestSplitHelpFlags_RemovesCombinedHelpFlag(t *testing.T) {
	sections := []Section{
		{Title: "Flags", Content: []Content{
			FlagGroup{
				{Short: "v", Long: "verbose", Desc: "Verbose"},
				{Short: "h", Long: "help", Desc: "Print help"},
			},
		}},
	}

	result := SplitHelpFlags(sections, "Short help", "Long help")

	require.Len(t, result, 1)
	require.Equal(t, "Flags", result[0].Title)

	// Should have 2 content blocks: original group (minus help) + new help group.
	require.Len(t, result[0].Content, 2)

	orig, ok := result[0].Content[0].(FlagGroup)
	require.True(t, ok)
	require.Len(t, orig, 1)
	require.Equal(t, "verbose", orig[0].Long)

	helpGroup, ok := result[0].Content[1].(FlagGroup)
	require.True(t, ok)
	require.Len(t, helpGroup, 2)
	require.Equal(t, "h", helpGroup[0].Short)
	require.Empty(t, helpGroup[0].Long)
	require.Equal(t, "Short help", helpGroup[0].Desc)
	require.Equal(t, "help", helpGroup[1].Long)
	require.Empty(t, helpGroup[1].Short)
	require.Equal(t, "Long help", helpGroup[1].Desc)
}

func TestSplitHelpFlags_RemovesEmptySections(t *testing.T) {
	// Section that only has the help flag - should be removed,
	// and the help group appended to the previous section.
	sections := []Section{
		{Title: "Filters", Content: []Content{
			FlagGroup{{Short: "a", Long: "author", Desc: "Author"}},
		}},
		{Title: "Flags", Content: []Content{
			FlagGroup{{Short: "h", Long: "help", Desc: "Print help"}},
		}},
	}

	result := SplitHelpFlags(sections, "Short", "Long")

	// "Flags" section should be removed (empty after removing help flag).
	// Help group appended to "Filters" (last section with flag content).
	require.Len(t, result, 1)
	require.Equal(t, "Filters", result[0].Title)
	require.Len(t, result[0].Content, 2)
}

func TestSplitHelpFlags_NoExistingHelpFlag(t *testing.T) {
	sections := []Section{
		{Title: "Flags", Content: []Content{
			FlagGroup{{Long: "verbose", Desc: "Verbose"}},
		}},
	}

	result := SplitHelpFlags(sections, "Short", "Long")

	require.Len(t, result, 1)
	require.Len(t, result[0].Content, 2)

	helpGroup2, ok := result[0].Content[1].(FlagGroup)
	require.True(t, ok)
	require.Len(t, helpGroup2, 2)
	require.Equal(t, "h", helpGroup2[0].Short)
	require.Equal(t, "help", helpGroup2[1].Long)
}

func TestSplitHelpFlags_NoFlagSections(t *testing.T) {
	sections := []Section{
		{Title: "Usage", Content: []Content{Text("mycli [options]")}},
	}

	result := SplitHelpFlags(sections, "Short", "Long")

	// Should create a new "Options" section.
	require.Len(t, result, 2)
	require.Equal(t, "Options", result[1].Title)
	helpGroup3, ok := result[1].Content[0].(FlagGroup)
	require.True(t, ok)
	require.Len(t, helpGroup3, 2)
}

func TestIsLongHelp(t *testing.T) {
	require.True(t, IsLongHelp([]string{"cmd", "--help"}))
	require.False(t, IsLongHelp([]string{"cmd", "-h"}))
	require.True(t, IsLongHelp([]string{"cmd", "--verbose", "--help"}))
	require.False(t, IsLongHelp(nil))
	require.False(t, IsLongHelp([]string{}))
	require.False(t, IsLongHelp([]string{"cmd"}))

	// --help after -- should not count.
	require.False(t, IsLongHelp([]string{"cmd", "--", "--help"}))
}

func TestWithFlagDefault(t *testing.T) {
	sections := []Section{
		{Title: "Filters", Content: []Content{
			FlagGroup{
				{Long: "org", Desc: "Limit to organization"},
				{Long: "repo", Desc: "Limit to repo"},
			},
		}},
	}

	opt := WithFlagDefault("org", "myorg")
	result := Apply(sections, opt)

	fg, ok := result[0].Content[0].(FlagGroup)
	require.True(t, ok)
	require.Equal(t, "Limit to organization [default: myorg]", fg[0].Desc)
	require.Equal(t, "Limit to repo", fg[1].Desc)
}

func TestWithFlagDefault_EmptyValue(t *testing.T) {
	sections := []Section{
		{Title: "Filters", Content: []Content{
			FlagGroup{{Long: "org", Desc: "Limit to organization"}},
		}},
	}

	opt := WithFlagDefault("org", "")
	result := Apply(sections, opt)

	fg, ok := result[0].Content[0].(FlagGroup)
	require.True(t, ok)
	require.Equal(t, "Limit to organization", fg[0].Desc)
}

func TestWithFlagDefault_NotFound(t *testing.T) {
	sections := []Section{
		{Title: "Filters", Content: []Content{
			FlagGroup{{Long: "org", Desc: "Limit to organization"}},
		}},
	}

	opt := WithFlagDefault("missing", "val")
	result := Apply(sections, opt)

	fg, ok := result[0].Content[0].(FlagGroup)
	require.True(t, ok)
	require.Equal(t, "Limit to organization", fg[0].Desc)
}

func TestWithLongHelp_LongHelp(t *testing.T) {
	extra := Section{Title: "Examples", Content: []Content{Text("example")}}
	opt := WithLongHelp([]string{"cmd", "--help"}, extra)

	sections := []Section{{Title: "Usage"}}
	result := Apply(sections, opt)
	require.Len(t, result, 2)
	require.Equal(t, "Examples", result[1].Title)
}

func TestWithLongHelp_ShortHelp(t *testing.T) {
	extra := Section{Title: "Examples", Content: []Content{Text("example")}}
	opt := WithLongHelp([]string{"cmd", "-h"}, extra)

	sections := []Section{{Title: "Usage"}}
	result := Apply(sections, opt)
	require.Len(t, result, 1)
}
