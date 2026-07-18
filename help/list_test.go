package help_test

import (
	"bytes"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/gechr/clib/help"
	"github.com/gechr/clib/theme"
	"github.com/stretchr/testify/require"
)

// renderDesc renders a single Description blurb and returns the ANSI-stripped
// output, dropping the "Description" heading and its blank line so tests can
// assert on the list body alone.
func renderDesc(t *testing.T, desc string, opts ...help.RendererOption) string {
	t.Helper()
	return stripDesc(renderDescRaw(t, testTheme(), desc, opts...))
}

// renderDescRaw renders a Description with the given theme and returns the full
// output including ANSI styling, so styling can be asserted exactly.
func renderDescRaw(
	t *testing.T, th *theme.Theme, desc string, opts ...help.RendererOption,
) string {
	t.Helper()
	r := help.NewRenderer(th, opts...)
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Description", Content: []help.Content{help.Description(desc)}},
	}
	require.NoError(t, r.Render(&buf, sections))
	return buf.String()
}

func stripDesc(raw string) string {
	return strings.TrimPrefix(ansi.Strip(raw), "Description\n\n")
}

// --- Ordered lists -------------------------------------------------------

func TestRender_NumberedList_AutoNumbered(t *testing.T) {
	// Authors may write "1." on every line; output renumbers sequentially.
	out := renderDesc(t, "1. first\n1. second\n1. third", help.WithDescriptionWidth(0))
	require.Equal(t, "      1. first\n      2. second\n      3. third\n", out)
}

func TestRender_NumberedList_SequentialPreserved(t *testing.T) {
	out := renderDesc(t, "1. first\n2. second\n3. third", help.WithDescriptionWidth(0))
	require.Equal(t, "      1. first\n      2. second\n      3. third\n", out)
}

func TestRender_NumberedList_ArbitraryStartRenumbered(t *testing.T) {
	// Even non-1 starts are renumbered from 1.
	out := renderDesc(t, "5. a\n9. b\n2. c", help.WithDescriptionWidth(0))
	require.Equal(t, "      1. a\n      2. b\n      3. c\n", out)
}

func TestRender_NumberedList_RightAlignedMarkers(t *testing.T) {
	// A run reaching double digits pads single-digit markers so the delimiters
	// line up (" 1." over "10."). The pad width is derived from the final
	// auto-incremented count, not the author's digits.
	out := renderDesc(t, strings.Repeat("1. item\n", 11), help.WithDescriptionWidth(0))
	require.Equal(
		t,
		"       1. item\n"+
			"       2. item\n"+
			"       3. item\n"+
			"       4. item\n"+
			"       5. item\n"+
			"       6. item\n"+
			"       7. item\n"+
			"       8. item\n"+
			"       9. item\n"+
			"      10. item\n"+
			"      11. item\n",
		out,
	)
}

func TestRender_NumberedList_PaddingResetsPerGroup(t *testing.T) {
	// Two paragraph-separated groups: the first has 2 items (no padding), the
	// second reaches 10 (padded). Each renumbers and sizes independently.
	out := renderDesc(
		t,
		"1. a\n1. a\n\n"+strings.Repeat("1. b\n", 10),
		help.WithDescriptionWidth(0),
	)
	require.Equal(
		t,
		"      1. a\n      2. a\n"+
			"\n"+
			"       1. b\n       2. b\n       3. b\n       4. b\n       5. b\n"+
			"       6. b\n       7. b\n       8. b\n       9. b\n      10. b\n",
		out,
	)
}

func TestRender_NumberedList_DelimiterPreserved(t *testing.T) {
	out := renderDesc(t, "1) first\n2) second", help.WithDescriptionWidth(0))
	require.Equal(t, "      1) first\n      2) second\n", out)
}

func TestRender_NumberedList_ResetsAfterText(t *testing.T) {
	out := renderDesc(t, "1. one\n1. two\nbreak\n1. three\n1. four", help.WithDescriptionWidth(0))
	require.Equal(
		t,
		"      1. one\n      2. two\n\n    break\n\n      1. three\n      2. four\n",
		out,
	)
}

// --- Unordered lists -----------------------------------------------------

func TestRender_UnorderedList_NormalisedGlyphAndIndent(t *testing.T) {
	// Mixed author bullets with varied indentation all normalise to the default
	// disc glyph at the list indent (no explicit nesting here).
	out := renderDesc(t, "- alpha\n* beta\n+ gamma", help.WithDescriptionWidth(0))
	require.Equal(t, "      • alpha\n      • beta\n      • gamma\n", out)
}

func TestRender_UnorderedList_KeepAuthorCharsWhenDisabled(t *testing.T) {
	// An empty char set leaves each author's marker unchanged.
	th := theme.Dark().With(theme.WithHelpDescUnorderedListChars())
	out := stripDesc(renderDescRaw(t, th, "- dash\n* star\n+ plus", help.WithDescriptionWidth(0)))
	require.Equal(t, "      - dash\n      * star\n      + plus\n", out)
}

func TestRender_UnorderedList_CustomChars(t *testing.T) {
	th := theme.Dark().With(theme.WithHelpDescUnorderedListChars(">", ">>"))
	out := stripDesc(
		renderDescRaw(t, th, "- a\n  - b\n    - c\n      - d", help.WithDescriptionWidth(0)),
	)
	// Cycles through the two custom glyphs by depth: >, >>, >, >>.
	require.Equal(t, "      > a\n        >> b\n           > c\n             >> d\n", out)
}

// --- Nesting (GitHub-flavoured Markdown rules) ---------------------------

func TestRender_Nesting_RequiresExplicitIndent(t *testing.T) {
	// Bullets with no indentation after a number are siblings, not children:
	// they stay top-level and end the ordered numbering.
	out := renderDesc(t, "1. abcde\n- bar\n- baz\n1. next", help.WithDescriptionWidth(0))
	require.Equal(
		t,
		"      1. abcde\n      • bar\n      • baz\n      1. next\n",
		out,
	)
}

func TestRender_Nesting_BulletsUnderNumber(t *testing.T) {
	// Indented bullets nest under the number, aligned under its text, and take
	// the depth-1 glyph.
	out := renderDesc(t, "1. abcde\n  - fghthhrth\n  - gerwgewge", help.WithDescriptionWidth(0))
	require.Equal(
		t,
		"      1. abcde\n         ◦ fghthhrth\n         ◦ gerwgewge\n",
		out,
	)
}

func TestRender_Nesting_GithubExampleThreeLevels(t *testing.T) {
	// Reproduces the GitHub rendering: numbered top level, then two nested
	// bullet levels cycling circle then square, each aligned under its parent.
	out := renderDesc(
		t,
		"1. First list item\n   - First nested list item\n      - Second nested list item",
		help.WithDescriptionWidth(0),
	)
	require.Equal(
		t,
		"      1. First list item\n"+
			"         ◦ First nested list item\n"+
			"           ▪ Second nested list item\n",
		out,
	)
}

func TestRender_Nesting_PureBulletCyclesGlyphs(t *testing.T) {
	out := renderDesc(t, "- a\n  - b\n    - c", help.WithDescriptionWidth(0))
	require.Equal(t, "      • a\n        ◦ b\n          ▪ c\n", out)
}

func TestRender_Nesting_GlyphCyclesBackAtFourthLevel(t *testing.T) {
	out := renderDesc(t, "- a\n  - b\n    - c\n      - d", help.WithDescriptionWidth(0))
	// depth 3 wraps back to the first glyph (disc).
	require.Equal(
		t,
		"      • a\n        ◦ b\n          ▪ c\n            • d\n",
		out,
	)
}

func TestRender_Nesting_Dedent(t *testing.T) {
	// Returning to a shallower indent pops back to that level's glyph/column.
	out := renderDesc(t, "- a\n  - b\n- c", help.WithDescriptionWidth(0))
	require.Equal(t, "      • a\n        ◦ b\n      • c\n", out)
}

func TestRender_Nesting_BlankSeparatesGroupsAndResets(t *testing.T) {
	out := renderDesc(t, "- x\n  - y\n\n- p\n  - q", help.WithDescriptionWidth(0))
	require.Equal(
		t,
		"      • x\n        ◦ y\n\n      • p\n        ◦ q\n",
		out,
	)
}

// --- Indent option -------------------------------------------------------

func TestRender_List_CustomIndent(t *testing.T) {
	out := renderDesc(t, "1. a\n- b", help.WithDescriptionWidth(0), help.WithListIndent(4))
	// content indent (2) + description indent (2) + list indent (4) = 8
	require.Equal(t, "        1. a\n        • b\n", out)
}

func TestRender_List_ZeroIndent(t *testing.T) {
	out := renderDesc(t, "1. a", help.WithDescriptionWidth(0), help.WithListIndent(0))
	require.Equal(t, "    1. a\n", out)
}

// --- Block spacing -------------------------------------------------------

func TestRender_List_BlockSpacing(t *testing.T) {
	out := renderDesc(
		t,
		"Intro paragraph.\n1. one\n2. two\nOutro paragraph.",
		help.WithDescriptionWidth(0),
	)
	require.Equal(
		t,
		"    Intro paragraph.\n\n      1. one\n      2. two\n\n    Outro paragraph.\n",
		out,
	)
}

func TestRender_List_NoBlankBetweenItems(t *testing.T) {
	out := renderDesc(t, "1. one\n2. two\n3. three", help.WithDescriptionWidth(0))
	//nolint:gocritic // fragment check against generated/styled output; not worth pinning as an exact literal
	require.NotContains(t, out, "\n\n")
}

func TestRender_List_ChildBulletsNoBlankFromParent(t *testing.T) {
	// A number and its nested children form one block: no blank between them.
	out := renderDesc(t, "1. one\n  - child", help.WithDescriptionWidth(0))
	require.Equal(t, "      1. one\n         ◦ child\n", out)
}

// --- Wrapping ------------------------------------------------------------

func TestRender_List_WrapExact_ContinuationUnderText(t *testing.T) {
	// The wrapped continuation begins at the item's text column (under "foo"),
	// not the marker column.
	out := renderDesc(t, "1. foo bar baz qux quux", help.WithMaxWidth(24), help.WithListIndent(0))
	require.Equal(t, "    1. foo bar baz qux\n       quux\n", out)
}

func TestRender_List_WrapNestedBullet(t *testing.T) {
	// A wrapped nested bullet's continuation aligns under its own text column.
	// With list indent 0 the top bullet sits at the base pad (4); the nested
	// "◦" at column 6, its text at column 8, so continuations indent 8.
	out := renderDesc(
		t,
		"- a\n  - bbb ccc ddd eee fff",
		help.WithMaxWidth(20),
		help.WithListIndent(0),
	)
	require.Equal(t, "    • a\n      ◦ bbb ccc ddd\n        eee fff\n", out)
}

// --- Non-list / edge cases ----------------------------------------------

func TestRender_List_NumericProseNotDetected(t *testing.T) {
	out := renderDesc(t, "3.14 is pi and 1st is first", help.WithDescriptionWidth(0))
	require.Equal(t, "    3.14 is pi and 1st is first\n", out)
}

func TestRender_List_MarkerWithoutSpaceIsProse(t *testing.T) {
	out := renderDesc(t, "-dash and 1.no-space stay text", help.WithDescriptionWidth(0))
	require.Equal(t, "    -dash and 1.no-space stay text\n", out)
}

func TestRender_List_SingleItem(t *testing.T) {
	out := renderDesc(t, "1. only", help.WithDescriptionWidth(0))
	require.Equal(t, "      1. only\n", out)
}

// --- Styling & fallback --------------------------------------------------

func TestRender_NumberedList_MarkerStyledExact(t *testing.T) {
	// Numbered markers carry HelpDescNumberedList (bold by default); the default
	// unordered marker resolves to no style (both specific and fallback nil).
	th := theme.Dark()
	raw := renderDescRaw(t, th, "1. a\n\n- b", help.WithDescriptionWidth(0))
	require.Equal(
		t,
		th.HelpSection.Render("Description")+"\n\n"+
			"      "+th.HelpDescNumberedList.Render("1.")+" a\n"+
			"\n"+
			"      • b\n",
		raw,
	)
}

func TestRender_List_NumberedFallsBackToListStyle(t *testing.T) {
	// With no numbered-specific style but a list fallback set, numbered markers
	// use the fallback.
	ls := lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	th := &theme.Theme{HelpDescList: &ls}
	raw := renderDescRaw(t, th, "1. a", help.WithDescriptionWidth(0))
	require.Equal(t, "Description\n\n      "+ls.Render("1.")+" a\n", raw)
}

func TestRender_List_UnorderedFallsBackToListStyle(t *testing.T) {
	ls := lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	th := &theme.Theme{HelpDescList: &ls, HelpDescUnorderedListChars: []string{"•"}}
	raw := renderDescRaw(t, th, "- a", help.WithDescriptionWidth(0))
	require.Equal(t, "Description\n\n      "+ls.Render("•")+" a\n", raw)
}

func TestRender_List_SpecificStyleOverridesFallback(t *testing.T) {
	num := lipgloss.NewStyle().Bold(true)
	ls := lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	th := &theme.Theme{HelpDescNumberedList: &num, HelpDescList: &ls}
	raw := renderDescRaw(t, th, "1. a", help.WithDescriptionWidth(0))
	require.Equal(t, "Description\n\n      "+num.Render("1.")+" a\n", raw)
}
