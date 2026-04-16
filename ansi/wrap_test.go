package ansi_test

import (
	"strings"
	"testing"

	xansi "github.com/charmbracelet/x/ansi"
	"github.com/gechr/clib/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// softWrapCases tests the soft wrap algorithm (space-only breakpoints, hard
// fallback for long words). Ported from charmbracelet/x/ansi wrapCases with
// hyphen-break expectations removed - our implementation only breaks on
// unicode spaces by default.
var softWrapCases = []struct {
	name     string
	input    string
	expected string
	width    int
}{
	{"passthrough", "alpha bravo", "alpha bravo", 11},
	{"simple", "alpha bravo charlie", "alpha\nbravo\ncharlie", 7},
	{"long word hard fallback", "abcdefghijklmno", "abcdefg\nhijklmn\no", 7},
	{
		"mixed long and short",
		"the quick brown foxxxxxxxxxxxxxxxx jumped over the lazy dog.",
		"the quick brown\nfoxxxxxxxxxxxxxx\nxx jumped over\nthe lazy dog.",
		16,
	},
	{"exact fit", "alpha", "alpha", 5},
	{"single char width", "VERTICAL", "V\nE\nR\nT\nI\nC\nA\nL", 1},
	{"zero width passthrough", "alpha bravo\n ", "alpha bravo\n ", 0},
	{"explicit newlines", "\nalpha bravo\n\n\ncharlie\n", "\nalpha\nbravo\n\n\ncharlie\n", 7},
	{"trailing space trimmed", "alpha ", "alpha", 5},
	{"double space", "a  bravo charlie", "a  bravo\ncharlie", 8},
	{"tab", "alpha\tbravo", "alpha\nbravo", 5},

	// Whitespace edge cases (ported from upstream).
	{"remove trailing spaces", "foo    \nb   ar   ", "foo\nb\nar", 4},
	{"space trail width", "foo\nb\t a\n bar", "foo\nb\t a\n bar", 4},
	{"explicit trailing newline", "foo bar foo\n", "foo\nbar\nfoo\n", 4},
	{"mixed blank lines", "\nfoo bar\n\n\nfoo\n", "\nfoo\nbar\n\n\nfoo\n", 4},
	{
		"complex mixed",
		" This is a list: \n\n\t* foo\n\t* bar\n\n\n\t* foo  \nbar    ",
		" This\nis a\nlist: \n\n\t* foo\n\t* bar\n\n\n\t* foo\nbar",
		6,
	},

	// Hyphens are NOT breakpoints by default.
	{"hyphen intact", "alpha-bravo", "alpha-bravo", 11},
	{"hyphen intact narrow", "alpha-bravo charlie", "alpha-bravo\ncharlie", 12},
	{"hyphen hard fallback", "alpha-bravo-charlie-delta", "alpha-brav\no-charlie-\ndelta", 10},

	// Unicode spaces.
	{"nbsp not breakpoint", "alpha\u00a0bravo", "alpha\u00a0b\nravo", 7},
	{"narrow nbsp", "0\u202f1\u202f2\u202f3\u202f4", "0\u202f1\u202f2\u202f3\n4", 7},
	{"medium math space", "0\u205f1\u205f2\u205f3\u205f4", "0\u205f1\u205f2\u205f3\n4", 7},
	{"ideographic space", "0\u30001\u30002\u30003\u3000", "0\u30001\u30002\n3\u3000", 7},

	// Multi-byte spaces.
	{
		"multi byte spaces",
		"A\u202fB\u202fC\u202fDA\u205f\u205fB\u205fC\u205fDA\u3000B\u3000C\u3000D",
		"A\u202fB\u202fC\nDA\u205f\u205fB\u205fC\nDA\u3000B\n" + "C\u3000D",
		7,
	},

	// Wide characters.
	{"asian", "こんにち", "こんに\nち", 7},
	{"emoji", "😃👰🏻‍♀️🫧", "😃\n👰🏻‍♀️\n🫧", 2},
}

func TestWrapSoft(t *testing.T) {
	for _, tc := range softWrapCases {
		t.Run(tc.name, func(t *testing.T) {
			got := ansi.WrapSoft(tc.input, tc.width)
			assert.Equal(t, tc.expected, got)
		})
	}
}

// ANSI tests verify properties (visible text, width constraints, ANSI
// preservation) rather than exact byte sequences, since the WrapWriter
// inserts style reset/reapply sequences at line breaks.

func TestWrapSoft_StylePassthrough(t *testing.T) {
	input := "\x1B[38;2;249;38;114malpha\x1B[0m \x1B[38;2;230;219;116mbravo\x1B[0m"
	got := ansi.WrapSoft(input, 11)

	assert.Equal(t, "alpha bravo", xansi.Strip(got))
	assert.NotContains(t, got, "\n")
}

func TestWrapSoft_StyleWrap(t *testing.T) {
	input := "I really \x1B[38;2;249;38;114mlove\x1B[0m Go!"
	got := ansi.WrapSoft(input, 8)

	assert.Equal(t, "I really\nlove Go!", xansi.Strip(got))
	assert.Contains(t, got, "\x1b[")
}

func TestWrapSoft_LongStyle(t *testing.T) {
	input := "\x1B[38;2;249;38;114ma really long string\x1B[0m"
	got := ansi.WrapSoft(input, 10)

	assert.Equal(t, "a really\nlong\nstring", xansi.Strip(got))
	for line := range strings.SplitSeq(got, "\n") {
		assert.LessOrEqual(t, xansi.StringWidth(line), 10)
	}
}

func TestWrapSoft_StyleCodeNotWrapped(t *testing.T) {
	input := "\x1B[38;2;249;38;114m(\x1B[0m\x1B[38;2;248;248;242mjust another test\x1B[38;2;249;38;114m)\x1B[0m"
	got := ansi.WrapSoft(input, 7)

	assert.Equal(t, "(just\nanother\ntest)", xansi.Strip(got))
}

func TestWrapSoft_OSC8Hyperlink(t *testing.T) {
	input := "สวัสดีสวัสดี\x1b]8;;https://example.com\x1b\\ สวัสดีสวัสดี\x1b]8;;\x1b\\"
	got := ansi.WrapSoft(input, 8)

	assert.Contains(t, got, "\n")
	assert.Contains(t, got, "https://example.com")
}

func TestWrapSoft_TrailingSpaceStyle(t *testing.T) {
	input := "\x1b[malpha \x1b[m"
	got := ansi.WrapSoft(input, 5)

	assert.Equal(t, "alpha", xansi.Strip(got))
}

func TestWrapSoft_EmojiInStyledText(t *testing.T) {
	input := "I really \x1B[38;2;249;38;114mlove u🫧\x1B[0m"
	got := ansi.WrapSoft(input, 8)

	assert.Equal(t, "I really\nlove u🫧", xansi.Strip(got))
	assert.Contains(t, got, "\x1b[")
}

func TestWrapSoft_Paragraph(t *testing.T) {
	input := "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."
	got := ansi.WrapSoft(input, 30)

	for i, line := range strings.Split(got, "\n") {
		w := xansi.StringWidth(line)
		assert.LessOrEqual(t, w, 30, "line %d: %q (width=%d)", i, line, w)
	}
	assert.Equal(t, input, strings.ReplaceAll(got, "\n", " "))
}

func TestWrapSoft_ParagraphWithStyles(t *testing.T) {
	input := "Lorem ipsum dolor \x1b[1msit\x1b[m amet, consectetur adipiscing elit."
	got := ansi.WrapSoft(input, 30)

	stripped := xansi.Strip(got)
	for i, line := range strings.Split(stripped, "\n") {
		assert.LessOrEqual(t, len(line), 30, "line %d: %q", i, line)
	}
	assert.Contains(t, got, "\x1b[")
}

func TestWrapSoft_AllLinesWithinWidth(t *testing.T) {
	input := "alpha bravo charlie delta echo foxtrot golf hotel india juliet"
	widths := []int{5, 10, 15, 20, 30}

	for _, width := range widths {
		got := ansi.WrapSoft(input, width)
		for i, line := range strings.Split(got, "\n") {
			w := xansi.StringWidth(line)
			assert.LessOrEqual(
				t,
				w,
				width,
				"line %d exceeds width %d: %q (width=%d)",
				i,
				width,
				line,
				w,
			)
		}
	}
}

// --- Hard wrap ---

func TestWrapHard(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		width    int
	}{
		{"basic", "abcdefghij", "abcde\nfghij", 5},
		{"mid word", "alpha bravo", "alpha b\nravo", 7},
		{"passthrough", "alpha", "alpha", 10},
		{"zero width", "alpha", "alpha", 0},
		{"passthrough short", "\x1b[31malpha\x1b[0m", "\x1b[31malpha\x1b[0m", 10},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ansi.WrapHard(tc.input, tc.width)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestWrapHard_PreservesANSI(t *testing.T) {
	input := "\x1b[31malpha bravo\x1b[0m"
	got := ansi.WrapHard(input, 5)

	assert.Equal(t, "alpha\nbravo", xansi.Strip(got))
	assert.Contains(t, got, "\x1b[")
	assert.Contains(t, got, "\n")
}

// --- Wrapper builder ---

func TestWrapper_DefaultSoftWrap(t *testing.T) {
	w := ansi.NewWrapper(ansi.WithWidth(20))
	got := w.Wrap("alpha bravo charlie delta echo foxtrot")

	require.Contains(t, got, "\n")
	for line := range strings.SplitSeq(got, "\n") {
		assert.LessOrEqual(t, xansi.StringWidth(line), 20)
	}
}

func TestWrapper_HardWrap(t *testing.T) {
	w := ansi.NewWrapper(ansi.WithWidth(10), ansi.WithWrapHard())
	got := w.Wrap("abcdefghijklmnopqrstuvwxyz")

	for line := range strings.SplitSeq(got, "\n") {
		assert.LessOrEqual(t, xansi.StringWidth(line), 10)
	}
}

func TestWrapper_Breakpoints(t *testing.T) {
	w := ansi.NewWrapper(ansi.WithWidth(10), ansi.WithBreakpoints("-"))
	got := w.Wrap("alpha-bravo-charlie")

	// With "-" as breakpoint, should break at hyphens.
	found := false
	for line := range strings.SplitSeq(got, "\n") {
		if strings.HasSuffix(strings.TrimSpace(line), "alpha-") {
			found = true
		}
	}
	assert.True(t, found, "hyphen should act as breakpoint when configured, got: %q", got)
}

func TestWrapper_NoBreakpoints_HyphenIntact(t *testing.T) {
	w := ansi.NewWrapper(ansi.WithWidth(12))
	got := w.Wrap("alpha-bravo charlie")

	for line := range strings.SplitSeq(got, "\n") {
		trimmed := strings.TrimSpace(line)
		assert.NotEqual(t, "alpha-", trimmed, "hyphen should not be a breakpoint by default")
	}
}

func TestWrapper_WidthFunc(t *testing.T) {
	calls := 0
	w := ansi.NewWrapper(ansi.WithWidthFunc(func() int {
		calls++
		return 10
	}))
	got := w.Wrap("alpha bravo charlie")

	require.Contains(t, got, "\n")
	assert.Equal(t, 1, calls)
}

func TestWrapper_WidthFuncOverridesWidth(t *testing.T) {
	w := ansi.NewWrapper(
		ansi.WithWidth(1000),
		ansi.WithWidthFunc(func() int { return 10 }),
	)
	got := w.Wrap("alpha bravo charlie")
	assert.Contains(t, got, "\n")
}

func TestWrapper_PreserveStyleTrue(t *testing.T) {
	w := ansi.NewWrapper(ansi.WithWidth(6))
	got := w.Wrap("\x1b[31malpha bravo\x1b[0m")

	// Should contain ANSI reset/reapply across the break.
	assert.Contains(t, got, "\x1b[")
	assert.Contains(t, got, "\n")
}

func TestWrapper_PreserveStyleFalse(t *testing.T) {
	w := ansi.NewWrapper(ansi.WithWidth(6), ansi.WithPreserveStyle(false))
	got := w.Wrap("\x1b[31malpha bravo\x1b[0m")

	assert.Contains(t, got, "\n")
}

func TestWrapper_PreservesNewlines(t *testing.T) {
	w := ansi.NewWrapper(ansi.WithWidth(80))
	got := w.Wrap("alpha\nbravo")
	assert.Equal(t, "alpha\nbravo", got)
}

func TestWrapper_ZeroWidth(t *testing.T) {
	w := ansi.NewWrapper(ansi.WithWidth(0))
	assert.Equal(t, "alpha bravo", w.Wrap("alpha bravo"))
}
