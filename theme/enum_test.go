package theme_test

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/gechr/clib/theme"
	"github.com/stretchr/testify/require"
)

func TestFmtEnum_PlainValues(t *testing.T) {
	th := theme.Default()
	got := th.FmtEnum([]theme.EnumValue{
		{Name: "open"},
		{Name: "closed"},
	})
	plain := ansi.Strip(got)
	require.Equal(t, "[open, closed]", plain)
}

func TestFmtEnum_WithBoldHighlights(t *testing.T) {
	th := theme.Default()
	got := th.FmtEnum([]theme.EnumValue{
		{Name: "open", Bold: "o"},
		{Name: "closed", Bold: "c"},
		{Name: "merged", Bold: "m"},
	})
	plain := ansi.Strip(got)
	require.Equal(t, "[open, closed, merged]", plain)
	// Should contain styled output (not equal to plain).
	require.NotEqual(t, plain, got)
}

func TestFmtEnum_SingleValue(t *testing.T) {
	th := theme.Default()
	got := th.FmtEnum([]theme.EnumValue{
		{Name: "all"},
	})
	plain := ansi.Strip(got)
	require.Equal(t, "[all]", plain)
}

func TestFmtEnum_DuplicateBold_Renders(t *testing.T) {
	th := theme.Default()
	got := th.FmtEnum([]theme.EnumValue{
		{Name: "open", Bold: "o"},
		{Name: "other", Bold: "o"},
	})
	require.Equal(t, "[open, other]", ansi.Strip(got))
}

func TestFmtEnum_BoldNotFoundInName(t *testing.T) {
	th := theme.Default()
	// Bold substring doesn't appear in Name - should render as plain dim.
	got := th.FmtEnum([]theme.EnumValue{
		{Name: "open", Bold: "z"},
	})
	plain := ansi.Strip(got)
	require.Equal(t, "[open]", plain)
}

func TestFmtEnum_BoldAtStart(t *testing.T) {
	th := theme.Default()
	got := th.FmtEnum([]theme.EnumValue{
		{Name: "open", Bold: "op"},
	})
	plain := ansi.Strip(got)
	require.Equal(t, "[open]", plain)
	require.NotEqual(t, plain, got)
}

func TestFmtEnum_BoldAtEnd(t *testing.T) {
	th := theme.Default()
	got := th.FmtEnum([]theme.EnumValue{
		{Name: "open", Bold: "en"},
	})
	plain := ansi.Strip(got)
	require.Equal(t, "[open]", plain)
	require.NotEqual(t, plain, got)
}

func TestFmtEnum_BoldInMiddle(t *testing.T) {
	th := theme.Default()
	got := th.FmtEnum([]theme.EnumValue{
		{Name: "closed", Bold: "los"},
	})
	plain := ansi.Strip(got)
	require.Equal(t, "[closed]", plain)
	require.NotEqual(t, plain, got)
}

func TestFmtEnum_MixedBoldAndPlain(t *testing.T) {
	th := theme.Default()
	got := th.FmtEnum([]theme.EnumValue{
		{Name: "open", Bold: "o"},
		{Name: "all"},
	})
	plain := ansi.Strip(got)
	require.Equal(t, "[open, all]", plain)
}

func TestFmtEnumDefault(t *testing.T) {
	th := theme.Default()
	got := th.FmtEnumDefault("open", []theme.EnumValue{
		{Name: "open", Bold: "o"},
		{Name: "closed", Bold: "c"},
	})
	plain := ansi.Strip(got)
	require.Equal(t, "[open, closed] (default: open)", plain)
}

func TestDimDefault_Empty(t *testing.T) {
	th := theme.Default()
	got := th.DimDefault("")
	require.Empty(t, got)
}

func TestDimDefault_NonEmpty(t *testing.T) {
	th := theme.Default()
	got := th.DimDefault("foo")
	plain := ansi.Strip(got)
	require.Equal(t, " (default: foo)", plain)
}

func TestDimNote(t *testing.T) {
	th := theme.Default()
	got := th.DimNote("required")
	plain := ansi.Strip(got)
	require.Equal(t, "(required)", plain)
	// Should be styled.
	require.NotEqual(t, plain, got)
}
