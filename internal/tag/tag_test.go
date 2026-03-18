package tag_test

import (
	"testing"

	"github.com/gechr/clib/internal/tag"
	"github.com/stretchr/testify/require"
)

func TestParse_QuotedValue(t *testing.T) {
	val, ok, err := tag.Parse("group='Filters',placeholder='repo'", "group")
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "Filters", val)
}

func TestParse_SecondKey(t *testing.T) {
	val, ok, err := tag.Parse("group='Filters',placeholder='repo'", "placeholder")
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "repo", val)
}

func TestParse_BareKey(t *testing.T) {
	val, ok, err := tag.Parse("negatable,group='Filters'", "negatable")
	require.NoError(t, err)
	require.True(t, ok)
	require.Empty(t, val)
}

func TestParse_NotFound(t *testing.T) {
	val, ok, err := tag.Parse("group='Filters'", "missing")
	require.NoError(t, err)
	require.False(t, ok)
	require.Empty(t, val)
}

func TestParse_UnquotedValue(t *testing.T) {
	val, ok, err := tag.Parse("group=Filters", "group")
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "Filters", val)
}

func TestParse_Empty(t *testing.T) {
	val, ok, err := tag.Parse("", "group")
	require.NoError(t, err)
	require.False(t, ok)
	require.Empty(t, val)
}

func TestParse_QuotedValueWithComma(t *testing.T) {
	val, ok, err := tag.Parse("terse='hello, world',group='G'", "terse")
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "hello, world", val)
}

func TestParse_TrimmedEntries(t *testing.T) {
	val, ok, err := tag.Parse("group='Filters', terse='Author'", "terse")
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "Author", val)
}

func TestParse_SingleEntry(t *testing.T) {
	val, ok, err := tag.Parse("group='Output'", "group")
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "Output", val)
}

func TestSplitCSV(t *testing.T) {
	result := tag.SplitCSV("a,b,c")
	require.Equal(t, []string{"a", "b", "c"}, result)
}

func TestSplitCSV_WithSpaces(t *testing.T) {
	result := tag.SplitCSV("a , b , c")
	require.Equal(t, []string{"a", "b", "c"}, result)
}

func TestSplitCSV_Empty(t *testing.T) {
	result := tag.SplitCSV("")
	require.Nil(t, result)
}

func TestSplitCSV_SingleValue(t *testing.T) {
	result := tag.SplitCSV("only")
	require.Equal(t, []string{"only"}, result)
}
