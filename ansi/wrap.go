package ansi

import (
	"bytes"
	"io"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"

	"charm.land/lipgloss/v2"
	xansi "github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/parser"
	"github.com/gechr/clib/terminal"
)

// wrapMode selects the wrapping algorithm.
type wrapMode int

const (
	wrapSoft wrapMode = iota // break at spaces, hard-wrap fallback for long words
	wrapHard                 // break at exact column width
)

// Wrapper wraps text to a configured width, preserving ANSI escape sequences.
type Wrapper struct {
	width        int
	widthFunc    func() int
	mode         wrapMode
	breakpoints  string
	preserveAnsi bool
}

// NewWrapper creates a [Wrapper] with the given options.
// Defaults: soft wrap, no additional breakpoints, ANSI style preservation enabled.
func NewWrapper(opts ...WrapOption) *Wrapper {
	w := &Wrapper{
		mode:         wrapSoft,
		preserveAnsi: true,
	}
	for _, o := range opts {
		o(w)
	}
	return w
}

// Wrap wraps s according to the configured mode and width.
// Returns s unchanged if the effective width is < 1.
func (w *Wrapper) Wrap(s string) string {
	width := w.effectiveWidth()
	if width < 1 {
		return s
	}

	var wrapped string
	switch w.mode { //nolint:exhaustive // wrapSoft is the default
	case wrapHard:
		wrapped = xansi.Hardwrap(s, width, false)
	default:
		wrapped = softWrap(s, width, w.breakpoints)
	}

	if !w.preserveAnsi {
		return wrapped
	}
	return reapplyStyles(wrapped)
}

// effectiveWidth returns the width to use: widthFunc > explicit width > terminal width.
func (w *Wrapper) effectiveWidth() int {
	if w.widthFunc != nil {
		return w.widthFunc()
	}
	if w.width > 0 {
		return w.width
	}
	return terminal.Width(os.Stdout)
}

// WrapOption configures a [Wrapper].
type WrapOption func(*Wrapper)

// WithWidth sets a static wrap width.
func WithWidth(width int) WrapOption {
	return func(w *Wrapper) { w.width = width }
}

// WithWidthFunc sets a dynamic width function, called on each [Wrapper.Wrap] invocation.
// Takes precedence over [WithWidth].
func WithWidthFunc(fn func() int) WrapOption {
	return func(w *Wrapper) { w.widthFunc = fn }
}

// WithWrapSoft selects soft wrapping: break at space boundaries, with
// hard-wrap fallback for words longer than the width. This is the default.
func WithWrapSoft() WrapOption {
	return func(w *Wrapper) { w.mode = wrapSoft }
}

// WithWrapHard selects hard wrapping: break at the exact column width,
// even mid-word.
func WithWrapHard() WrapOption {
	return func(w *Wrapper) { w.mode = wrapHard }
}

// WithBreakpoints adds characters (beyond spaces) that are treated as
// word break opportunities during soft wrapping. Has no effect in hard
// wrap mode.
func WithBreakpoints(chars string) WrapOption {
	return func(w *Wrapper) { w.breakpoints = chars }
}

// WithPreserveStyle controls whether ANSI styles and hyperlinks are
// reset and reapplied across line breaks. Default: true.
func WithPreserveStyle(preserve bool) WrapOption {
	return func(w *Wrapper) { w.preserveAnsi = preserve }
}

// WrapSoft wraps s to fit within width columns, breaking at space boundaries.
// Words longer than width are hard-wrapped. ANSI styles are preserved.
func WrapSoft(s string, width int) string {
	return NewWrapper(WithWidth(width), WithWrapSoft()).Wrap(s)
}

// WrapHard wraps s at exactly width columns, breaking mid-word if needed.
// ANSI styles are preserved.
func WrapHard(s string, width int) string {
	return NewWrapper(WithWidth(width), WithWrapHard()).Wrap(s)
}

// reapplyStyles pipes wrapped text through lipgloss's [WrapWriter] to
// reset and reapply ANSI styles/links at each line break.
func reapplyStyles(s string) string {
	var buf bytes.Buffer
	w := lipgloss.NewWrapWriter(&buf)
	_, _ = io.WriteString(w, s)
	_ = w.Close()
	return buf.String()
}

// softWrap wraps text at word boundaries (unicode spaces), with optional
// additional breakpoints. Words longer than limit are hard-wrapped.
// ANSI escape sequences are skipped for width calculation but preserved.
func softWrap(s string, limit int, breakpoints string) string {
	var (
		buf        bytes.Buffer
		word       bytes.Buffer
		space      bytes.Buffer
		spaceWidth int
		curWidth   int
		wordLen    int
		pstate     = parser.GroundState
	)

	addSpace := func() {
		if spaceWidth == 0 && space.Len() == 0 {
			return
		}
		curWidth += spaceWidth
		buf.Write(space.Bytes())
		space.Reset()
		spaceWidth = 0
	}

	addWord := func() {
		if word.Len() == 0 {
			return
		}
		addSpace()
		curWidth += wordLen
		buf.Write(word.Bytes())
		word.Reset()
		wordLen = 0
	}

	addNewline := func() {
		buf.WriteByte('\n')
		curWidth = 0
		space.Reset()
		spaceWidth = 0
	}

	i := 0
	for i < len(s) {
		state, action := parser.Table.Transition(pstate, s[i])

		if state == parser.Utf8State { //nolint:nestif // required
			cluster, width := xansi.FirstGraphemeCluster(s[i:], xansi.GraphemeWidth)
			i += len(cluster)

			r, _ := utf8.DecodeRuneInString(cluster)
			switch {
			case r != utf8.RuneError && unicode.IsSpace(r) && r != '\u00a0':
				addWord()
				space.WriteRune(r)
				spaceWidth += width
			case breakpoints != "" && strings.ContainsAny(cluster, breakpoints):
				addSpace()
				if curWidth+wordLen+width > limit {
					word.WriteString(cluster)
					wordLen += width
				} else {
					addWord()
					buf.WriteString(cluster)
					curWidth += width
				}
			default:
				if wordLen+width > limit {
					addWord()
				}
				word.WriteString(cluster)
				wordLen += width
				if curWidth+wordLen+spaceWidth > limit {
					addNewline()
				}
				if wordLen == limit {
					addWord()
				}
			}

			pstate = parser.GroundState
			continue
		}

		switch action {
		case parser.PrintAction, parser.ExecuteAction:
			r := rune(s[i])
			switch {
			case r == '\n':
				if wordLen == 0 {
					if curWidth+spaceWidth > limit {
						curWidth = 0
					} else {
						buf.Write(space.Bytes())
					}
					space.Reset()
					spaceWidth = 0
				}
				addWord()
				addNewline()
			case unicode.IsSpace(r):
				addWord()
				space.WriteRune(r)
				spaceWidth++
			case breakpoints != "" && strings.ContainsRune(breakpoints, r):
				addSpace()
				if curWidth+wordLen >= limit {
					word.WriteRune(r)
					wordLen++
				} else {
					addWord()
					buf.WriteRune(r)
					curWidth++
				}
			default:
				if curWidth == limit {
					addNewline()
				}
				word.WriteRune(r)
				wordLen++
				if wordLen == limit {
					addWord()
				}
				if curWidth+wordLen+spaceWidth > limit {
					addNewline()
				}
			}
		default:
			word.WriteByte(s[i])
		}

		if pstate != parser.Utf8State {
			pstate = state
		}
		i++
	}

	if wordLen == 0 {
		if curWidth+spaceWidth > limit {
			curWidth = 0
		} else {
			buf.Write(space.Bytes())
		}
		space.Reset()
		spaceWidth = 0
	}

	addWord()
	return buf.String()
}
