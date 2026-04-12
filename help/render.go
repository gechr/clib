package help

import (
	"fmt"
	"io"
	"os"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	"github.com/gechr/clib/terminal"
	"github.com/gechr/clib/theme"
)

// Renderer renders styled help output.
type Renderer struct {
	Theme *theme.Theme

	argPad       int       // padding between arg and description
	cmdAlign     Alignment // command name alignment
	cmdAlignMode AlignMode // per-section vs global command alignment
	cmdPad       int       // padding between command and description
	flagAlign    Alignment // flag name alignment
	flagPad      int       // padding between flag and description
	maxWidth     int       // max output width (0 = no wrapping)
	wrapStyle    WrapStyle // continuation line indent style
}

// NewRenderer creates a Renderer.
func NewRenderer(th *theme.Theme, opts ...RendererOption) *Renderer {
	if th == nil {
		th = theme.Default()
	}

	r := &Renderer{
		Theme:    th.Init(),
		argPad:   defaultArgPad,
		cmdAlign: AlignLeft,
		cmdPad:   defaultCmdPad,
		flagPad:  defaultFlagPad,
		maxWidth: autoMaxWidth,
	}
	for _, o := range opts {
		o(r)
	}
	return r
}

const (
	indent     = 2
	nestIndent = indent * 2 // additional indent per nesting depth

	defaultFlagPad = 2 // padding between flag and description
	defaultArgPad  = 2 // padding between arg and description
	defaultCmdPad  = 2 // padding between command and description

	overflowPad = 2 // minimum gap when name overflows past descCol

	autoMaxWidth = -1
)

// visibleWidth computes the visible width of a string, ignoring ANSI escapes.
func visibleWidth(s string) int {
	return lipgloss.Width(s)
}

// Render writes help sections to w.
func (r *Renderer) Render(w io.Writer, sections []Section) error {
	rr := *r
	rr.Theme = rr.Theme.Init()
	rr.maxWidth = rr.resolveMaxWidth(w)

	descCol := rr.computeDescCol(sections)
	cmdDescCol := rr.computeCmdDescCol(sections)

	for i, sec := range sections {
		if err := rr.renderSection(w, &sec, descCol, cmdDescCol, 0); err != nil {
			return err
		}
		if i < len(sections)-1 {
			if _, err := io.WriteString(w, "\n"); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *Renderer) renderSection(
	w io.Writer,
	sec *Section,
	descCol, cmdDescCol, depth int,
) error {
	titleInd := nestIndent * depth
	contentInd := titleInd + indent

	titleStyle := r.Theme.HelpSection
	if depth > 0 {
		titleStyle = r.Theme.HelpCommand
	}

	ind := ""
	if titleInd > 0 {
		ind = strings.Repeat(" ", titleInd)
	}
	if _, err := fmt.Fprintf(w, "%s%s\n\n", ind, titleStyle.Render(sec.Title)); err != nil {
		return err
	}

	hasShort := sectionHasShort(sec)
	for i, content := range sec.Content {
		if i > 0 {
			if _, err := io.WriteString(w, "\n"); err != nil {
				return err
			}
		}
		if err := r.renderContent(
			w,
			content,
			hasShort,
			descCol,
			cmdDescCol,
			contentInd,
			depth,
		); err != nil {
			return err
		}
	}
	return nil
}

func (r *Renderer) renderContent(
	w io.Writer, content Content, hasShort bool, descCol, cmdDescCol, ind, depth int,
) error {
	switch c := content.(type) {
	case FlagGroup:
		return r.renderFlags(w, c, hasShort, descCol, ind)
	case Args:
		return r.renderArgs(w, c, descCol, ind)
	case CommandGroup:
		return r.renderCommands(w, c, cmdDescCol, ind)
	case Usage:
		return r.renderUsage(w, c, ind)
	case Text:
		_, err := fmt.Fprintf(w, "%s%s\n", strings.Repeat(" ", ind), string(c))
		return err
	case Examples:
		return r.renderExamples(w, c, ind)
	case *Section:
		return r.renderSection(w, c, descCol, cmdDescCol, depth+1)
	}
	return nil
}

func (r *Renderer) renderFlags(
	w io.Writer,
	flags FlagGroup,
	hasShort bool,
	descCol, ind int,
) error {
	for _, f := range flags {
		line, err := r.formatFlag(f, hasShort, descCol, ind)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}
	return nil
}

func (r *Renderer) argStyle(a Arg) *lipgloss.Style {
	if a.Required {
		return r.Theme.HelpArgRequired
	}
	return r.Theme.HelpArg
}

func (r *Renderer) renderArgs(w io.Writer, args Args, _, ind int) error {
	// Args compute their own description column, independent of flags.
	argDescCol := 0
	for _, a := range args {
		argWidth := ind + visibleWidth(r.argStyle(a).Render(BracketArg(a)))
		if argWidth > argDescCol {
			argDescCol = argWidth
		}
	}
	argDescCol += r.argPad

	for _, a := range args {
		if _, err := fmt.Fprintln(w, r.formatArg(a, argDescCol, ind)); err != nil {
			return err
		}
	}
	return nil
}

func (r *Renderer) renderCommands(
	w io.Writer,
	cmds CommandGroup,
	globalCmdDescCol, ind int,
) error {
	descCol := globalCmdDescCol
	if r.cmdAlignMode == AlignModeSection {
		// Compute description column from this section's commands only.
		descCol = 0
		for _, c := range cmds {
			cw := ind + visibleWidth(r.Theme.HelpSubcommand.Render(c.Name))
			if cw > descCol {
				descCol = cw
			}
		}
		descCol += r.cmdPad
	}

	for _, c := range cmds {
		if _, err := fmt.Fprintln(w, r.formatCommand(c, descCol, ind)); err != nil {
			return err
		}
	}
	return nil
}

func (r *Renderer) formatCommand(c Command, descCol, ind int) string {
	var sb strings.Builder

	namePart := r.Theme.HelpSubcommand.Render(c.Name)
	nameWidth := visibleWidth(namePart)

	if r.cmdAlign == AlignRight {
		// Right-align: pad before the name so it ends at descCol - cmdPad.
		rightEdge := descCol - r.cmdPad
		pad := max(rightEdge-nameWidth, ind)
		sb.WriteString(strings.Repeat(" ", pad))
		sb.WriteString(namePart)
	} else {
		sb.WriteString(strings.Repeat(" ", ind))
		sb.WriteString(namePart)
	}

	if c.Desc != "" {
		currentWidth := ind + nameWidth
		if r.cmdAlign == AlignRight {
			currentWidth = descCol - r.cmdPad
		}
		actualDescCol := descCol
		if currentWidth < descCol {
			sb.WriteString(strings.Repeat(" ", descCol-currentWidth))
		} else {
			sb.WriteString("  ")
			actualDescCol = currentWidth + overflowPad
		}
		sb.WriteString(r.wrapDesc(r.renderDesc(c.Desc), actualDescCol))
	}

	return sb.String()
}

func (r *Renderer) renderUsage(w io.Writer, u Usage, ind int) error {
	var parts []string
	parts = append(parts, r.Theme.HelpCommand.Render(u.Command))
	// Subcommand args come before [options].
	for _, a := range u.Args {
		if a.IsSubcommand {
			parts = append(parts, r.argStyle(a).Render(BracketArg(a)))
		}
	}
	if u.ShowOptions {
		parts = append(parts, r.Theme.HelpFlag.Render("[options]"))
	}
	for _, a := range u.Args {
		if !a.IsSubcommand {
			parts = append(parts, r.argStyle(a).Render(BracketArg(a)))
		}
	}
	_, err := fmt.Fprintf(w, "%s%s\n", strings.Repeat(" ", ind), strings.Join(parts, " "))
	return err
}

func (r *Renderer) renderExamples(w io.Writer, examples Examples, ind int) error {
	pad := strings.Repeat(" ", ind)
	for i, ex := range examples {
		if i > 0 {
			if _, err := io.WriteString(w, "\n"); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(
			w,
			"%s%s\n",
			pad,
			r.Theme.HelpDim.Render("# "+ex.Comment),
		); err != nil {
			return err
		}
		ue := r.Theme.HelpUsageExample
		if _, err := fmt.Fprintf(
			w,
			"%s%s %s\n",
			pad,
			ue.PromptStyle.Render(ue.Prompt),
			ue.CommandStyle.Render(ex.Command),
		); err != nil {
			return err
		}
	}
	return nil
}

func (r *Renderer) formatFlag(f Flag, hasShort bool, descCol, ind int) (string, error) {
	var sb strings.Builder

	flagPart := r.buildFlagPart(f, hasShort)
	flagWidth := visibleWidth(flagPart)

	if r.flagAlign == AlignRight {
		// Right-align: pad before the flag so it ends at descCol - flagPad.
		rightEdge := descCol - r.flagPad
		pad := max(rightEdge-flagWidth, ind)
		sb.WriteString(strings.Repeat(" ", pad))
		sb.WriteString(flagPart)
	} else {
		sb.WriteString(strings.Repeat(" ", ind))
		sb.WriteString(flagPart)
	}

	descPart, err := r.buildDescPart(f)
	if err != nil {
		return "", err
	}
	if descPart != "" {
		currentWidth := ind + flagWidth
		if r.flagAlign == AlignRight {
			currentWidth = descCol - r.flagPad
		}
		actualDescCol := descCol
		if currentWidth < descCol {
			sb.WriteString(strings.Repeat(" ", descCol-currentWidth))
		} else {
			sb.WriteString("  ")
			actualDescCol = currentWidth + overflowPad
		}
		sb.WriteString(r.wrapDesc(descPart, actualDescCol))
	}

	return sb.String(), nil
}

// buildDescPart combines the rendered description and enum suffix for a flag.
func (r *Renderer) buildDescPart(f Flag) (string, error) {
	var parts []string
	if f.Desc != "" {
		parts = append(parts, r.renderDesc(f.Desc))
	}
	enumStr, err := r.renderEnum(f)
	if err != nil {
		return "", err
	}
	if enumStr != "" {
		parts = append(parts, enumStr)
	}
	return strings.Join(parts, " "), nil
}

// renderEnum builds a styled enum suffix from f.Enum/EnumHighlight,
// respecting Theme.EnumStyle.
func (r *Renderer) renderEnum(f Flag) (string, error) {
	if len(f.Enum) == 0 {
		return "", nil
	}
	if len(f.EnumHighlight) > 0 && len(f.EnumHighlight) != len(f.Enum) {
		return "", fmt.Errorf("help: EnumHighlight length must match Enum length")
	}
	values := make([]theme.EnumValue, len(f.Enum))
	for i, v := range f.Enum {
		ev := theme.EnumValue{Name: v}
		switch r.Theme.EnumStyle {
		case theme.EnumStylePlain:
			// No highlighting.
			break
		case theme.EnumStyleHighlightPrefix:
			if i < len(f.EnumHighlight) {
				ev.Bold = f.EnumHighlight[i]
			}
		case theme.EnumStyleHighlightDefault:
			ev.IsDefault = f.EnumDefault != "" && v == f.EnumDefault
		case theme.EnumStyleHighlightBoth:
			if i < len(f.EnumHighlight) {
				ev.Bold = f.EnumHighlight[i]
			}
			ev.IsDefault = f.EnumDefault != "" && v == f.EnumDefault
		}
		values[i] = ev
	}
	hasDefaultHighlight := r.Theme.EnumStyle == theme.EnumStyleHighlightDefault ||
		r.Theme.EnumStyle == theme.EnumStyleHighlightBoth
	if f.EnumDefault != "" && !hasDefaultHighlight {
		return r.Theme.FmtEnumDefault(f.EnumDefault, values), nil
	}
	return r.Theme.FmtEnum(values), nil
}

func (r *Renderer) formatArg(a Arg, descCol, ind int) string {
	var sb strings.Builder

	sb.WriteString(strings.Repeat(" ", ind))

	argPart := r.argStyle(a).Render(BracketArg(a))
	sb.WriteString(argPart)

	if a.Desc != "" {
		argWidth := ind + visibleWidth(argPart)
		actualDescCol := descCol
		if argWidth < descCol {
			sb.WriteString(strings.Repeat(" ", descCol-argWidth))
		} else {
			sb.WriteString("  ")
			actualDescCol = argWidth + overflowPad
		}
		sb.WriteString(r.wrapDesc(a.Desc, actualDescCol))
	}

	return sb.String()
}

// renderDesc renders a description, styling backtick-enclosed text and
// dimming trailing bracketed or parenthesized notes.
//
// Backtick processing: `text` -> styled with HelpDescBacktick (backticks removed).
//
// Suffix patterns:
//   - Trailing "(note)" -> HelpFlagNote style
//   - Trailing "[default: ...]" -> HelpFlagDefault style
//   - Trailing "[example: ...]" -> HelpFlagExample style
//   - Trailing "[note]" -> HelpDim style (fallback)
func (r *Renderer) renderDesc(desc string) string {
	desc = r.renderBackticks(desc)

	// Try parenthesized note first.
	if r.Theme.HelpFlagNote != nil {
		if styled, ok := r.styledSuffix(desc, NoteOpen, NoteClose, *r.Theme.HelpFlagNote); ok {
			return styled
		}
	}

	// Try bracket patterns with specific prefix matching.
	if open := strings.LastIndex(desc, OptOpen); open >= 0 && strings.HasSuffix(desc, OptClose) {
		prefix := desc[:open]
		if prefix == "" {
			return desc
		}
		note := desc[open:]
		inner := note[len(OptOpen) : len(note)-len(OptClose)]

		// Pick style based on bracket content prefix.
		style := r.Theme.HelpDim
		switch {
		case strings.HasPrefix(inner, "default: "):
			if r.Theme.HelpFlagDefault != nil {
				style = r.Theme.HelpFlagDefault
			}
		case strings.HasPrefix(inner, "example: "):
			if r.Theme.HelpFlagExample != nil {
				style = r.Theme.HelpFlagExample
			}
		}
		return strings.TrimRight(prefix, " ") + " " + style.Render(note)
	}

	return desc
}

// renderBackticks replaces `text` and 'text' with styled text (delimiters removed).
// Single-quoted strings are only matched when not preceded/followed by a letter,
// so contractions like "don't" are left intact.
// When HelpDescBacktick is nil, delimiters are left intact.
func (r *Renderer) renderBackticks(s string) string {
	if r.Theme.HelpDescBacktick == nil {
		return s
	}
	var sb strings.Builder
	for i := 0; i < len(s); {
		switch {
		case s[i] == '`':
			end := strings.IndexByte(s[i+1:], '`')
			if end < 0 {
				sb.WriteString(s[i:])
				return sb.String()
			}
			end += i + 1
			sb.WriteString(r.Theme.HelpDescBacktick.Render(s[i+1 : end]))
			i = end + 1

		case s[i] == '\'' && !isLetterAt(s, i-1):
			end := strings.IndexByte(s[i+1:], '\'')
			if end < 0 || isLetterAt(s, i+1+end+1) {
				sb.WriteByte(s[i])
				i++
				continue
			}
			end += i + 1
			sb.WriteString(r.Theme.HelpDescBacktick.Render(s[i+1 : end]))
			i = end + 1

		default:
			sb.WriteByte(s[i])
			i++
		}
	}
	return sb.String()
}

// isLetterAt reports whether s[i] is an ASCII letter. Returns false if i is out of bounds.
func isLetterAt(s string, i int) bool {
	if i < 0 || i >= len(s) {
		return false
	}
	c := s[i]
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func (r *Renderer) styledSuffix(
	desc, openTok, closeTok string,
	style lipgloss.Style,
) (string, bool) {
	idx := strings.LastIndex(desc, openTok)
	if idx < 0 || !strings.HasSuffix(desc, closeTok) {
		return "", false
	}
	prefix := desc[:idx]
	if prefix == "" {
		return "", false
	}
	// Only match when a space precedes the opening token - avoids
	// splitting embedded parens like "context(s)" or "foo(bar)".
	if prefix[len(prefix)-1] != ' ' {
		return "", false
	}
	note := desc[idx:]
	return strings.TrimRight(prefix, " ") + " " + style.Render(note), true
}

func (r *Renderer) buildFlagPart(f Flag, hasShort bool) string {
	var parts []string

	shortStr := ""
	if f.Short != "" {
		shortStr = "-" + f.Short
	}
	longStr := ""
	if f.Long != "" {
		longStr = "--" + f.Long
	}
	placeholderStr := ""
	repeatSuffix := ""
	if f.Placeholder != "" {
		if f.PlaceholderLiteral {
			placeholderStr = f.Placeholder
		} else {
			placeholderStr = ArgOpen + f.Placeholder + ArgClose
		}
		if f.Repeatable && r.Theme.HelpRepeatEllipsisEnabled {
			repeatSuffix = r.Theme.HelpRepeatEllipsis.Render(EllipsisShort)
		}
	}

	renderPlaceholder := func() string {
		if placeholderStr == "" {
			return ""
		}
		sep := string(r.Theme.HelpKeyValueSeparator)
		if r.Theme.HelpKeyValueSeparatorStyle != nil {
			sep = r.Theme.HelpKeyValueSeparatorStyle.Render(sep)
		}
		return sep + r.Theme.HelpValuePlaceholder.Render(placeholderStr) + repeatSuffix
	}

	switch {
	case shortStr != "" && longStr != "":
		parts = append(parts, r.Theme.HelpFlag.Render(shortStr)+",")
		parts = append(parts, r.Theme.HelpFlag.Render(longStr)+renderPlaceholder())
	case shortStr != "":
		parts = append(parts, r.Theme.HelpFlag.Render(shortStr)+renderPlaceholder())
	case longStr != "":
		longPart := r.Theme.HelpFlag.Render(longStr) + renderPlaceholder()
		if hasShort && !f.NoIndent {
			// Mixed section: extra indent to align -- with short-flag entries.
			parts = append(parts, "    "+longPart)
		} else {
			parts = append(parts, longPart)
		}
	}

	return strings.Join(parts, " ")
}

func (r *Renderer) computeDescCol(sections []Section) int {
	descCol := 0
	r.walkFlags(sections, 0, func(f Flag, hasShort bool, ind int) {
		w := ind + visibleWidth(r.buildFlagPart(f, hasShort))
		if w > descCol {
			descCol = w
		}
	})
	return descCol + r.flagPad
}

// computeCmdDescCol computes the global description column across all
// CommandGroup content in sections. Only used when cmdAlignMode is AlignModeGlobal.
func (r *Renderer) computeCmdDescCol(sections []Section) int {
	if r.cmdAlignMode != AlignModeGlobal {
		return 0
	}
	col := 0
	r.walkCommands(sections, 0, func(c Command, ind int) {
		w := ind + visibleWidth(r.Theme.HelpSubcommand.Render(c.Name))
		if w > col {
			col = w
		}
	})
	return col + r.cmdPad
}

// sectionHasShort reports whether any FlagGroup in a section contains a flag
// with a short name. Used to determine whether long-only flags need extra
// indent to align their "--" with the "--" in "-X, --long" entries.
func sectionHasShort(sec *Section) bool {
	for _, content := range sec.Content {
		fg, ok := content.(FlagGroup)
		if !ok {
			continue
		}
		for _, f := range fg {
			if f.Short != "" {
				return true
			}
		}
	}
	return false
}

func (r *Renderer) walkCommands(
	sections []Section, depth int,
	fn func(Command, int),
) {
	ind := nestIndent*depth + indent
	for _, sec := range sections {
		for _, content := range sec.Content {
			switch c := content.(type) {
			case CommandGroup:
				for _, cmd := range c {
					fn(cmd, ind)
				}
			case *Section:
				r.walkCommands([]Section{*c}, depth+1, fn)
			}
		}
	}
}

// wrapDesc wraps a styled description string to fit within maxWidth,
// indenting continuation lines according to the configured [WrapStyle].
//
// Returns the input unchanged when maxWidth is 0 or the text fits.
func (r *Renderer) wrapDesc(desc string, descCol int) string {
	if r.maxWidth <= 0 || descCol >= r.maxWidth {
		return desc
	}
	avail := r.maxWidth - descCol
	if visibleWidth(desc) <= avail {
		return desc
	}

	// WrapBracketBelow: break before '[' and place bracket content on
	// the next line at descCol, with continuation at descCol+1.
	if r.wrapStyle == WrapBracketBelow {
		if result, ok := r.wrapBracketBelow(desc, descCol, avail); ok {
			return result
		}
	}

	wrapped := ansi.Wordwrap(desc, avail, " ")
	lines := strings.Split(wrapped, "\n")
	if len(lines) <= 1 {
		return desc
	}

	// WrapBracketAlign: align continuation lines after the unclosed '['.
	padCol := descCol
	if r.wrapStyle == WrapBracketAlign {
		if bc := unclosedBracketCol(lines[0]); bc >= 0 {
			candidate := descCol + bc
			if candidate < r.maxWidth {
				contAvail := r.maxWidth - candidate
				contText := strings.Join(lines[1:], " ")
				rewrapped := ansi.Wordwrap(contText, contAvail, " ")
				lines = append(lines[:1], strings.Split(rewrapped, "\n")...)
				padCol = candidate
			}
		}
	}

	pad := strings.Repeat(" ", padCol)
	for i := 1; i < len(lines); i++ {
		lines[i] = pad + lines[i]
	}
	return strings.Join(lines, "\n")
}

// wrapBracketBelow splits desc before a trailing bracketed list '[...]',
// wraps each part independently, and returns the assembled result with
// bracket content indented below the description text.
func (r *Renderer) wrapBracketBelow(desc string, descCol, avail int) (string, bool) {
	col := trailingBracketCol(desc)
	if col <= 0 {
		return "", false
	}

	// Split styled text at the bracket position using ANSI-aware utilities.
	prefix := strings.TrimRight(ansi.Truncate(desc, col, ""), " ")
	bracket := ansi.Cut(desc, col, visibleWidth(desc))
	if visibleWidth(prefix) == 0 {
		return "", false
	}

	// Wrap prefix at descCol if needed.
	var lines []string
	if visibleWidth(prefix) > avail {
		wrapped := ansi.Wordwrap(prefix, avail, " ")
		lines = strings.Split(wrapped, "\n")
	} else {
		lines = []string{prefix}
	}
	descPad := strings.Repeat(" ", descCol)
	for i := 1; i < len(lines); i++ {
		lines[i] = descPad + lines[i]
	}

	// Wrap bracket content. First line starts at descCol (full avail width);
	// continuation lines start at descCol+1 (one narrower, after '[').
	bracketWrapped := ansi.Wordwrap(bracket, avail, " ")
	bracketLines := strings.Split(bracketWrapped, "\n")
	if len(bracketLines) > 1 {
		contAvail := avail - 1
		if contAvail > 0 {
			contText := strings.Join(bracketLines[1:], " ")
			rewrapped := ansi.Wordwrap(contText, contAvail, " ")
			bracketLines = append(bracketLines[:1], strings.Split(rewrapped, "\n")...)
		}
	}

	bracketPad := strings.Repeat(" ", descCol+1)
	lines = append(lines, descPad+bracketLines[0])
	for _, bl := range bracketLines[1:] {
		lines = append(lines, bracketPad+bl)
	}

	return strings.Join(lines, "\n"), true
}

// unclosedBracketCol returns the visible-width column of the character after
// the last unclosed '[' in text, or -1 if all brackets are closed. It tracks
// bracket depth with a stack so nested pairs (e.g. "[default: [a]]") are
// handled correctly, and uses per-rune display widths so East-Asian wide
// characters are measured accurately.
func unclosedBracketCol(text string) int {
	stripped := ansi.Strip(text)
	var openStack []int // visible-width positions of unmatched '['
	col := 0
	for _, c := range stripped {
		switch c {
		case '[':
			openStack = append(openStack, col)
		case ']':
			if len(openStack) > 0 {
				openStack = openStack[:len(openStack)-1]
			}
		}
		col += lipgloss.Width(string(c))
	}
	if len(openStack) > 0 {
		return openStack[len(openStack)-1] + 1
	}
	return -1
}

// trailingBracketCol returns the visible-width column of a '[' whose matching
// ']' ends the string. This identifies trailing bracketed lists like enum
// values "[a, b, c]". Returns -1 when no such bracket exists. Operates on
// stripped (ANSI-free) text for simplicity.
func trailingBracketCol(s string) int {
	stripped := ansi.Strip(s)
	if !strings.HasSuffix(stripped, "]") {
		return -1
	}

	// Pair brackets and find the '[' that matches the trailing ']'.
	var openStack []int
	type pair struct{ open, close int }
	var pairs []pair
	for i, c := range stripped {
		switch c {
		case '[':
			openStack = append(openStack, i)
		case ']':
			if len(openStack) > 0 {
				pairs = append(pairs, pair{openStack[len(openStack)-1], i})
				openStack = openStack[:len(openStack)-1]
			}
		}
	}

	lastClose := len(stripped) - 1
	for _, p := range pairs {
		if p.close == lastClose {
			if p.open == 0 {
				return -1 // entire string is bracketed, no prefix to split
			}
			return p.open
		}
	}
	return -1
}

func (r *Renderer) resolveMaxWidth(w io.Writer) int {
	if r.maxWidth != autoMaxWidth {
		return r.maxWidth
	}
	return writerWidth(w)
}

func writerWidth(w io.Writer) int {
	switch wt := w.(type) {
	case *os.File:
		return terminal.Width(wt)
	case *colorprofile.Writer:
		return writerWidth(wt.Forward)
	default:
		return 0
	}
}

func (r *Renderer) walkFlags(
	sections []Section, depth int,
	fn func(Flag, bool, int),
) {
	ind := nestIndent*depth + indent
	for _, sec := range sections {
		hasShort := sectionHasShort(&sec)
		for _, content := range sec.Content {
			switch c := content.(type) {
			case FlagGroup:
				for _, f := range c {
					fn(f, hasShort, ind)
				}
			case *Section:
				r.walkFlags([]Section{*c}, depth+1, fn)
			}
		}
	}
}
