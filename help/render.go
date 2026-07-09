package help

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	"github.com/gechr/clib/theme"
	gansi "github.com/gechr/x/ansi"
	xstrings "github.com/gechr/x/strings"
	"github.com/gechr/x/terminal"
)

// Renderer renders styled help output.
type Renderer struct {
	Theme *theme.Theme

	argPad              int           // padding between arg and description
	cmdAlign            Alignment     // command name alignment
	cmdAlignMode        AlignMode     // per-section vs global command alignment
	cmdPad              int           // padding between command and description
	descRefs            descRefs      // per-render index of names referenced in Description backticks
	plain               bool          // per-render: true when the writer can't display styling, so leave backtick delimiters intact
	descriptionIndent   int           // extra indent (cols) for Description beyond the section content indent
	descriptionWidth    int           // wrap width for Description content (autoDescriptionWidth = inherit maxWidth, 0 = no wrap)
	descriptionWidthMin int           // lower bound of the flexible wrap-width range (0 = no range)
	descriptionWidthMax int           // upper bound of the flexible wrap-width range (0 = no range)
	backtickStyle       BacktickStyle // how backticked tokens in descriptions are styled
	flagAlign           Alignment     // flag name alignment
	flagPad             int           // padding between flag and description
	hideDefaults        bool          // suppress " (default: X)" annotations globally
	listIndent          int           // leading indent (cols) for detected list items in Descriptions
	maxWidth            int           // max output width (0 = no wrapping)
	wrapStyle           WrapStyle     // continuation line indent style
}

// descRefs indexes the names a Description blurb might reference back to:
// positional arguments (matched as `name` or `<name>`), flags (`--name`,
// `-x`), and subcommands (`name`). Built by collectDescRefs at Render time so
// backtick styling can pick the same style the renderer would use elsewhere
// for the same name.
type descRefs struct {
	args      map[string]Arg
	binary    string // first token of the rendered Usage.Command (e.g. "mycli")
	commands  map[string]struct{}
	flags     map[string]struct{} // rendered flag names, e.g. "--verbose" and "-v"
	argEnums  map[string]Arg      // enum value (e.g. "github") -> the arg that owns it, so the value styles like its arg
	flagEnums map[string]struct{} // enum value owned by a flag -> styled with the flag color (HelpFlag)
}

// NewRenderer creates a Renderer.
func NewRenderer(th *theme.Theme, opts ...RendererOption) *Renderer {
	if th == nil {
		th = theme.Auto()
	}

	r := &Renderer{
		Theme:               th.Init(),
		argPad:              defaultArgPad,
		cmdAlign:            AlignLeft,
		cmdPad:              defaultCmdPad,
		descriptionIndent:   defaultDescriptionIndent,
		descriptionWidth:    autoDescriptionWidth,
		descriptionWidthMin: defaultDescriptionWidthMin,
		descriptionWidthMax: defaultDescriptionWidthMax,
		flagPad:             defaultFlagPad,
		listIndent:          defaultListIndent,
		maxWidth:            autoMaxWidth,
	}
	for _, o := range opts {
		o(r)
	}
	return r
}

const (
	indent     = 2
	nestIndent = indent * 2 // additional indent per nesting depth

	defaultFlagPad           = 2 // padding between flag and description
	defaultArgPad            = 2 // padding between arg and description
	defaultCmdPad            = 2 // padding between command and description
	defaultDescriptionIndent = 2 // extra indent for Description content beyond section content indent
	defaultListIndent        = 2 // leading indent for detected list items in Descriptions

	// defaultDescriptionWidthMin/Max are the flexible wrap-width range applied
	// to Description content when the caller sets neither [WithDescriptionWidth]
	// nor [WithDescriptionWidthRange]. The range lets the renderer even out a
	// paragraph's right edge rather than wrapping at one hard column; both
	// bounds are still capped at [WithMaxWidth] when that is set.
	defaultDescriptionWidthMin = 70
	defaultDescriptionWidthMax = 100

	overflowPad = 2 // minimum gap when name overflows past descCol

	autoMaxWidth         = -1
	autoDescriptionWidth = -1

	// orphanPenaltyDivisor scales down the raggedness penalty for a wrapped
	// paragraph ending in a single orphaned word, so a lone trailing word
	// counts less than a mid-paragraph short line but still discourages the
	// wrap that produced it.
	orphanPenaltyDivisor = 4
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
	rr.plain = writerIsPlain(w)
	rr.descRefs = collectDescRefs(sections)

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
		text := r.renderBackticks(string(c), nil)
		_, err := fmt.Fprintf(w, "%s\n", xstrings.Indent(text, strings.Repeat(" ", ind)))
		return err
	case Description:
		return r.renderDescription(w, c, ind+r.descriptionIndent)
	case Aliases:
		style := r.Theme.HelpAlias
		if style == nil {
			style = r.Theme.HelpCommand
		}
		parts := make([]string, len(c))
		for i, a := range c {
			parts[i] = style.Render(a)
		}
		_, err := fmt.Fprintf(w, "%s%s\n", strings.Repeat(" ", ind), strings.Join(parts, ", "))
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

func (r *Renderer) argStyle(a Arg, all Args) *lipgloss.Style {
	if a.Required {
		return r.Theme.HelpArg
	}
	// Use HelpArgOptional only when optional args coexist with required args,
	// so they are visually distinct. Otherwise fall back to HelpArg.
	for _, o := range all {
		if o.Required {
			return r.Theme.HelpArgOptional
		}
	}
	return r.Theme.HelpArg
}

func (r *Renderer) renderArgs(w io.Writer, args Args, _, ind int) error {
	// Args compute their own description column, independent of flags.
	argDescCol := 0
	for _, a := range args {
		argWidth := ind + visibleWidth(r.argStyle(a, args).Render(BracketArg(a)))
		if argWidth > argDescCol {
			argDescCol = argWidth
		}
	}
	argDescCol += r.argPad

	for _, a := range args {
		if _, err := fmt.Fprintln(w, r.formatArg(a, args, argDescCol, ind)); err != nil {
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
	if u.Raw != "" {
		parts = append(parts, u.Raw)
		_, err := fmt.Fprintf(w, "%s%s\n", strings.Repeat(" ", ind), strings.Join(parts, " "))
		return err
	}
	// Subcommand args come before [options].
	for _, a := range u.Args {
		if a.IsSubcommand {
			s := r.Theme.HelpSubcommand.Bold(false)
			parts = append(parts, s.Render(BracketArg(a)))
		}
	}
	if u.ShowOptions {
		parts = append(parts, r.Theme.HelpFlag.Render("[options]"))
	}
	for _, a := range u.Args {
		if !a.IsSubcommand {
			parts = append(parts, r.argStyle(a, u.Args).Render(BracketArg(a)))
		}
	}
	_, err := fmt.Fprintf(w, "%s%s\n", strings.Repeat(" ", ind), strings.Join(parts, " "))
	return err
}

// renderDescription writes a Description blurb at the given indent. The wrap
// width is taken from [WithDescriptionWidth] when set, falling back to
// [WithMaxWidth] so descriptions wrap to the same column as flag descriptions
// by default. A width of 0 disables wrapping for descriptions specifically.
//
// Whitespace normalisation: leading/trailing blank lines are trimmed, runs
// of consecutive blank lines collapse to a single blank line, and the indent
// pad is suppressed on blank lines so output never contains trailing
// whitespace. Author-supplied non-blank newlines are preserved as paragraph
// breaks.
func (r *Renderer) renderDescription(w io.Writer, d Description, ind int) error {
	text := strings.Trim(string(d), "\n")
	if text == "" {
		return nil
	}
	pad := strings.Repeat(" ", ind)
	lines := r.parseDescLines(strings.Split(text, "\n"))

	styled := make([]string, len(lines))
	for i, dl := range lines {
		styled[i] = dl.styled
	}
	avail, wrap := r.descriptionWrapAvail(styled, ind)

	// Emit lines, managing blank separators. Consecutive author-supplied blanks
	// collapse to one, and a single blank line is enforced at every boundary
	// between a list block and adjacent paragraph text, mirroring how Markdown
	// separates list blocks from surrounding prose.
	pendingBlank := false
	wrote := false
	prevList := false
	for i, dl := range lines {
		if dl.blank {
			if wrote {
				pendingBlank = true
			}
			continue
		}
		if wrote && (pendingBlank || dl.isList != prevList) {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
		pendingBlank = false
		out := []string{styled[i]}
		if wrap {
			if dl.hang > 0 {
				// List item: continuation lines align under the item's text.
				out = wrapParagraphHang(styled[i], avail, dl.hang)
			} else {
				out = wrapDescriptionParagraph(styled[i], avail)
			}
		}
		for _, line := range out {
			if _, err := fmt.Fprintf(w, "%s%s\n", pad, line); err != nil {
				return err
			}
		}
		wrote = true
		prevList = dl.isList
	}
	return nil
}

// descLine is one classified, pre-styled line of a Description blurb.
type descLine struct {
	blank  bool   // an empty source line (paragraph/list separator)
	isList bool   // an unordered or ordered list item
	hang   int    // continuation indent (cols) for a wrapped list item; 0 = derive
	styled string // rendered content, indent + marker + text (empty when blank)
}

// numberedItem is an ordered-list line buffered until its run's marker width
// is known, so all markers in the run can be right-aligned to a common column.
type numberedItem struct {
	idx     int    // index into the descLine slice being built
	ordinal int    // sequential position within the run (1-based)
	delim   byte   // trailing delimiter, '.' or ')'
	text    string // item text following the marker
}

// listFrame is one open nesting level while parsing a Description's lists. It
// records the author indent that opened the level, the column its markers render
// at, and the column its children (one level deeper) align under.
type listFrame struct {
	authorIndent int
	renderCol    int
	contentCol   int
}

// parseDescLines classifies and styles each raw Description line. List items
// (both unordered "-"/"*"/"+" and ordered "1."/"2)", mirroring Markdown) have
// their author-supplied top-level indentation replaced with the configured list
// indent. Ordered items are renumbered sequentially from 1 within each
// contiguous run - so authors can write "1." on every line and still get "1.",
// "2.", "3." - and their markers are right-aligned (left-padded) to the width of
// the run's largest ordinal, so the delimiters line up (" 1." over "10.").
//
// Nesting follows GitHub-flavoured Markdown: it activates only when the author
// indents a child past its parent. A nested item aligns under its parent's text,
// and unordered markers cycle glyphs by depth (HelpDescUnorderedListChars).
func (r *Renderer) parseDescLines(raw []string) []descLine {
	lines := make([]descLine, len(raw))

	var stack []listFrame
	var run []numberedItem
	runCol, runIndent := 0, 0

	// place adjusts the nesting stack for an item at author indent ai and
	// returns the column it renders at and its depth. A deeper indent than the
	// current level opens a child (aligned under the parent's text); an equal
	// indent is a sibling; a shallower indent pops back out. The new frame is
	// left on top so the caller can set its contentCol once the marker width is
	// known.
	place := func(ai int) (int, int) {
		for len(stack) > 0 && ai < stack[len(stack)-1].authorIndent {
			stack = stack[:len(stack)-1]
		}
		var col int
		switch {
		case len(stack) == 0:
			col = r.listIndent
			stack = append(stack, listFrame{authorIndent: ai, renderCol: col, contentCol: col})
		case ai > stack[len(stack)-1].authorIndent:
			col = stack[len(stack)-1].contentCol
			stack = append(stack, listFrame{authorIndent: ai, renderCol: col, contentCol: col})
		default: // sibling at the current level
			top := &stack[len(stack)-1]
			top.authorIndent, col = ai, top.renderCol
		}
		return col, len(stack) - 1
	}

	// flushRun styles the buffered numbered items once the run's width is known.
	// Ordinals increase monotonically, so the last item holds the max.
	flushRun := func() {
		if len(run) == 0 {
			return
		}
		width := len(strconv.Itoa(run[len(run)-1].ordinal)) + 1 // digits + delimiter
		textCol := runCol + width + 1                           // where the item text (and children) align
		style := r.numberedListStyle()
		for _, n := range run {
			marker := xstrings.PadLeft(strconv.Itoa(n.ordinal)+string(n.delim), width)
			if style != nil {
				marker = style.Render(marker)
			}
			lines[n.idx] = descLine{
				isList: true,
				hang:   textCol,
				styled: strings.Repeat(" ", runCol) + marker + " " +
					r.renderBackticks(n.text, nil),
			}
		}
		// Children of this run align under the number's text.
		stack[len(stack)-1].contentCol = textCol
		run = run[:0]
	}

	ordinal := 0
	for i, ln := range raw {
		switch {
		case ln == "":
			// A blank line separates list groups: it ends the current run and
			// closes all nesting so the next group starts afresh.
			flushRun()
			ordinal, stack = 0, stack[:0]
			lines[i] = descLine{blank: true}
		case isListMarkerNumbered(ln):
			ai := leadingSpaces(ln)
			if len(run) > 0 && ai != runIndent {
				flushRun() // an indent change starts a new run at a new level
			}
			if len(run) == 0 {
				runIndent = ai
				runCol, _ = place(ai)
			}
			ordinal++
			marker, rest, _ := splitListMarker(ln)
			run = append(run, numberedItem{
				idx: i, ordinal: ordinal, delim: marker[len(marker)-1], text: rest,
			})
		default:
			marker, rest, ok := splitListMarker(ln)
			if !ok {
				// Paragraph text ends every list context.
				flushRun()
				ordinal, stack = 0, stack[:0]
				lines[i] = descLine{styled: r.renderBackticks(ln, nil)}
				continue
			}
			// Unordered item: flush any pending numbered run first so a nested
			// child can align under the number's text.
			flushRun()
			col, depth := place(leadingSpaces(ln))
			// A top-level unordered item ends any ordered numbering; a nested one
			// leaves the parent's counter intact so it can resume afterwards.
			if depth == 0 {
				ordinal = 0
			}
			glyph := r.unorderedGlyph(marker, depth)
			textCol := col + visibleWidth(glyph) + 1 // where the item text (and children) align
			stack[len(stack)-1].contentCol = textCol
			if style := r.unorderedListStyle(); style != nil {
				glyph = style.Render(glyph)
			}
			lines[i] = descLine{
				isList: true,
				hang:   textCol,
				styled: strings.Repeat(" ", col) + glyph + " " + r.renderBackticks(rest, nil),
			}
		}
	}
	flushRun()
	return lines
}

// unorderedGlyph returns the marker glyph for an unordered item at the given
// nesting depth. When HelpDescUnorderedListChars is set, its glyphs are cycled
// by depth (disc, circle, square, ...); when empty, the author's original marker
// is kept unchanged.
func (r *Renderer) unorderedGlyph(authorMarker string, depth int) string {
	chars := r.Theme.HelpDescUnorderedListChars
	if len(chars) == 0 {
		return authorMarker
	}
	return chars[depth%len(chars)]
}

// leadingSpaces returns the number of leading ASCII spaces in s.
func leadingSpaces(s string) int {
	return len(s) - len(strings.TrimLeft(s, " "))
}

// isListMarkerNumbered reports whether a raw Description line is an ordered-list
// item (after optional leading whitespace).
func isListMarkerNumbered(line string) bool {
	marker, _, ok := splitListMarker(line)
	return ok && isNumberedMarker(marker)
}

// numberedListStyle resolves the style for a numbered-list marker: the specific
// HelpDescNumberedList when set, otherwise the HelpDescList fallback (nil when
// neither is set, meaning the marker is left unstyled).
func (r *Renderer) numberedListStyle() *lipgloss.Style {
	if r.Theme.HelpDescNumberedList != nil {
		return r.Theme.HelpDescNumberedList
	}
	return r.Theme.HelpDescList
}

// unorderedListStyle resolves the style for an unordered-list marker: the
// specific HelpDescUnorderedList when set, otherwise the HelpDescList fallback
// (nil when neither is set).
func (r *Renderer) unorderedListStyle() *lipgloss.Style {
	if r.Theme.HelpDescUnorderedList != nil {
		return r.Theme.HelpDescUnorderedList
	}
	return r.Theme.HelpDescList
}

// descriptionWrapAvail resolves the wrap width available to a Description
// block after its indent. With a width range configured
// ([WithDescriptionWidthRange]), the width in [min, max] whose wrapped
// paragraphs produce the most even right edge wins; the range's upper bound
// is capped at maxWidth so flexible wrapping never overflows the output
// width. Otherwise the fixed descriptionWidth applies, falling back to
// maxWidth. The second return reports whether wrapping is enabled at all.
func (r *Renderer) descriptionWrapAvail(paragraphs []string, ind int) (int, bool) {
	if r.descriptionWidthMax > 0 {
		maxW := r.descriptionWidthMax
		if r.maxWidth > 0 {
			maxW = min(maxW, r.maxWidth)
		}
		minW := min(r.descriptionWidthMin, maxW)
		minAvail := max(minW-ind, 1)
		maxAvail := max(maxW-ind, 1)
		return bestWrapAvail(paragraphs, minAvail, maxAvail), true
	}
	width := r.descriptionWidth
	if width == autoDescriptionWidth {
		width = r.maxWidth
	}
	return max(width-ind, 1), width > 0
}

// bestWrapAvail returns the wrap width in [minAvail, maxAvail] whose greedy
// wrap of paragraphs is least ragged. A single width is chosen for the whole
// block so all paragraphs share one right edge. Ties prefer the lower bound,
// so flexible descriptions stay compact unless a wider width improves the
// edge.
func bestWrapAvail(paragraphs []string, minAvail, maxAvail int) int {
	best, bestScore := minAvail, -1
	for avail := minAvail; avail <= maxAvail; avail++ {
		score := 0
		for _, p := range paragraphs {
			if p == "" {
				continue
			}
			score += raggedness(wrapDescriptionParagraph(p, avail))
		}
		if bestScore < 0 || score < bestScore {
			best, bestScore = avail, score
		}
	}
	return best
}

// raggedness scores how uneven a wrapped paragraph's right edge is: the sum
// of squared gaps between each line and the paragraph's longest line. The
// final line is exempt - a short last line is natural in prose - unless it
// holds a single orphaned word, which reads as a wrap misfire and contributes
// a quarter-weighted penalty.
func raggedness(lines []string) int {
	if len(lines) <= 1 {
		return 0
	}
	widths := make([]int, len(lines))
	longest := 0
	for i, line := range lines {
		widths[i] = visibleWidth(line)
		longest = max(longest, widths[i])
	}
	score := 0
	for _, w := range widths[:len(widths)-1] {
		gap := longest - w
		score += gap * gap
	}
	last := strings.TrimSpace(ansi.Strip(lines[len(lines)-1]))
	if !strings.Contains(last, " ") {
		gap := longest - widths[len(widths)-1]
		score += gap * gap / orphanPenaltyDivisor
	}
	return score
}

// wrapDescriptionParagraph wraps a single paragraph to avail columns. When the
// paragraph starts with literal leading whitespace (e.g. a manually authored
// "  - item" list line), that hanging indent is preserved on wrapped
// continuation lines instead of collapsing to the paragraph's base indent.
func wrapDescriptionParagraph(text string, avail int) []string {
	lines := strings.Split(gansi.WrapSoft(text, avail), "\n")
	if len(lines) <= 1 {
		return lines
	}
	return wrapParagraphHang(text, avail, leadingIndentWidth(lines[0]))
}

// wrapParagraphHang wraps text to avail columns and indents every continuation
// line by hang columns. It is used for list items, whose text column is known
// exactly (so it need not be re-derived from the styled marker, which may use a
// theme glyph the marker detector doesn't recognise). A hang outside (0, avail)
// leaves the greedy wrap unchanged.
func wrapParagraphHang(text string, avail, hang int) []string {
	lines := strings.Split(gansi.WrapSoft(text, avail), "\n")
	if len(lines) <= 1 {
		return lines
	}
	if hang <= 0 || hang >= avail {
		return lines
	}
	contAvail := avail - hang
	contText := strings.Join(lines[1:], " ")
	rewrapped := strings.Split(gansi.WrapSoft(contText, contAvail), "\n")
	hangPad := strings.Repeat(" ", hang)
	for i, line := range rewrapped {
		rewrapped[i] = hangPad + line
	}
	return append(lines[:1], rewrapped...)
}

// listMarkers are the unordered-list bullet characters recognised in
// Description content, mirroring Markdown syntax. Each must be followed by a
// space. (Ordered markers like "1." / "2)" are recognised separately.)
const listMarkers = "-*+•"

const (
	markerTrailingSpace = 1 // the single space that must follow any list marker
	orderedMarkerTrail  = 2 // delimiter + trailing space after the digits, e.g. ". "
)

// leadingIndentWidth returns the visible width of the leading run of spaces
// in line, ignoring ANSI escape sequences. When the leading spaces are
// followed by a recognised list marker (e.g. "- " or "1. "), the marker is
// included so wrapped continuation lines align with the item's text rather
// than its marker.
func leadingIndentWidth(line string) int {
	stripped := ansi.Strip(line)
	trimmed := strings.TrimLeft(stripped, " ")
	leadSpaces := len(stripped) - len(trimmed) // spaces: byte count == visible width
	mw := listMarkerWidth(trimmed)
	if mw == 0 {
		return leadSpaces
	}
	// The marker may include a multi-byte glyph (e.g. "•"), so measure its
	// visible width rather than its byte length for the hanging indent.
	return leadSpaces + visibleWidth(trimmed[:mw])
}

// listMarkerWidth returns the byte width (including the single trailing space)
// of a leading Markdown list marker in s, or 0 when s does not begin with one.
// Recognised markers are unordered bullets ("-", "*", "+", "•") and ordered
// numbers ("1.", "2)"), each followed by a space.
func listMarkerWidth(s string) int {
	if r, size := utf8.DecodeRuneInString(s); size > 0 &&
		strings.ContainsRune(listMarkers, r) &&
		len(s) > size && s[size] == ' ' {
		return size + markerTrailingSpace
	}
	n := 0
	for n < len(s) && s[n] >= '0' && s[n] <= '9' {
		n++
	}
	if n > 0 && n+1 < len(s) && (s[n] == '.' || s[n] == ')') && s[n+1] == ' ' {
		return n + orderedMarkerTrail
	}
	return 0
}

// splitListMarker splits a Description line into its list marker (without the
// trailing space) and the remaining item text, discarding any author-supplied
// leading whitespace so callers can re-indent uniformly. The final bool is
// false when the line does not begin with a recognised list marker.
func splitListMarker(line string) (string, string, bool) {
	s := strings.TrimLeft(line, " ")
	w := listMarkerWidth(s)
	if w == 0 {
		return "", "", false
	}
	return s[:w-1], s[w:], true
}

// isNumberedMarker reports whether marker (a marker returned by
// [splitListMarker], e.g. "1." or "12)") is an ordered-list marker.
func isNumberedMarker(marker string) bool {
	return len(marker) >= 2 && xstrings.IsDigits(marker[:len(marker)-1])
}

func (r *Renderer) renderExamples(w io.Writer, examples Examples, ind int) error {
	pad := strings.Repeat(" ", ind)
	for i, ex := range examples {
		if i > 0 {
			if _, err := io.WriteString(w, "\n"); err != nil {
				return err
			}
		}
		if !xstrings.IsBlank(ex.Comment) {
			if _, err := fmt.Fprintf(
				w,
				"%s%s\n",
				pad,
				r.Theme.HelpDim.Render("# "+ex.Comment),
			); err != nil {
				return err
			}
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
	// Append "(default: X)" annotation for non-enum flags. Enums already
	// surface their default inside the enum-list rendering via EnumDefault
	// and would be redundant.
	if f.Default != "" && !f.HideDefault && !r.hideDefaults && len(f.Enum) == 0 {
		parts = append(parts, strings.TrimSpace(r.Theme.DimDefault(f.Default)))
	}
	return strings.Join(parts, " "), nil
}

// hasEmpty reports whether any element of vals is the empty string.
func hasEmpty(vals []string) bool {
	return slices.Contains(vals, "")
}

// dropEmptyEnum returns the enum and highlight slices with empty enum entries
// removed, keeping the highlight slice index-aligned with the trimmed enum.
func dropEmptyEnum(enum, highlight []string) ([]string, []string) {
	outEnum := make([]string, 0, len(enum))
	var outHL []string
	if len(highlight) > 0 {
		outHL = make([]string, 0, len(highlight))
	}
	for i, v := range enum {
		if v == "" {
			continue
		}
		outEnum = append(outEnum, v)
		if i < len(highlight) {
			outHL = append(outHL, highlight[i])
		}
	}
	return outEnum, outHL
}

// renderEnum builds a styled enum suffix from f.Enum/EnumHighlight,
// respecting Theme.EnumStyle. Empty enum entries are dropped from the display.
func (r *Renderer) renderEnum(f Flag) (string, error) {
	if len(f.Enum) == 0 {
		return "", nil
	}
	if len(f.EnumHighlight) > 0 && len(f.EnumHighlight) != len(f.Enum) {
		return "", fmt.Errorf("help: EnumHighlight length must match Enum length")
	}
	enum := f.Enum
	highlight := f.EnumHighlight
	if hasEmpty(enum) {
		enum, highlight = dropEmptyEnum(enum, highlight)
		if len(enum) == 0 {
			return "", nil
		}
	}
	values := make([]theme.EnumValue, len(enum))
	for i, v := range enum {
		ev := theme.EnumValue{Name: v}
		switch r.Theme.EnumStyle {
		case theme.EnumStylePlain:
			// No highlighting.
			break
		case theme.EnumStyleHighlightPrefix:
			if i < len(highlight) {
				ev.Bold = highlight[i]
			}
		case theme.EnumStyleHighlightDefault:
			ev.IsDefault = f.EnumDefault != "" && v == f.EnumDefault
		case theme.EnumStyleHighlightBoth:
			if i < len(highlight) {
				ev.Bold = highlight[i]
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

func (r *Renderer) formatArg(a Arg, all Args, descCol, ind int) string {
	var sb strings.Builder

	sb.WriteString(strings.Repeat(" ", ind))

	argPart := r.argStyle(a, all).Render(BracketArg(a))
	sb.WriteString(argPart)

	descPart := r.buildArgDescPart(a)
	if descPart != "" {
		argWidth := ind + visibleWidth(argPart)
		actualDescCol := descCol
		if argWidth < descCol {
			sb.WriteString(strings.Repeat(" ", descCol-argWidth))
		} else {
			sb.WriteString("  ")
			actualDescCol = argWidth + overflowPad
		}
		sb.WriteString(r.wrapDesc(descPart, actualDescCol))
	}

	return sb.String()
}

// buildArgDescPart renders a positional arg's description (with backtick/enum
// styling) and appends an auto-derived " (default: X)" annotation when the arg
// has a default, so callers can drop the hand-written suffix from help text.
func (r *Renderer) buildArgDescPart(a Arg) string {
	var parts []string
	if a.Desc != "" {
		parts = append(parts, r.renderDesc(a.Desc))
	}
	if a.Default != "" && !a.HideDefault && !r.hideDefaults {
		parts = append(parts, strings.TrimSpace(r.Theme.DimDefault(a.Default)))
	}
	return strings.Join(parts, " ")
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
	// Try parenthesized note first.
	if r.Theme.HelpFlagNote != nil {
		if styled, ok := r.styledSuffix(desc, NoteOpen, NoteClose, *r.Theme.HelpFlagNote); ok {
			return styled
		}
	}

	// Try bracket patterns with specific prefix matching. The wire format
	// uses literal "[...]" so callers can author descriptions in a stable
	// convention; the rendered output substitutes the theme's configurable
	// bracket characters (HelpDefaultOpen/Close).
	if open, ok := trailingBalancedSuffixStart(desc, OptOpen, OptClose); ok {
		prefix := desc[:open]
		if prefix == "" {
			return desc
		}
		note := desc[open:]
		inner := note[len(OptOpen) : len(note)-len(OptClose)]

		// Pick style and themed brackets based on bracket content prefix.
		// Untagged "[note]" suffixes keep their literal brackets and fall
		// back to HelpDim.
		style := r.Theme.HelpDim
		var openTok, closeTok string
		switch {
		case strings.HasPrefix(inner, "default: "):
			if r.Theme.HelpFlagDefault != nil {
				style = r.Theme.HelpFlagDefault
			}
			openTok = r.Theme.HelpDefaultOpen
			closeTok = r.Theme.HelpDefaultClose
		case strings.HasPrefix(inner, "example: "):
			if r.Theme.HelpFlagExample != nil {
				style = r.Theme.HelpFlagExample
			}
			openTok = r.Theme.HelpExampleOpen
			closeTok = r.Theme.HelpExampleClose
		}
		rendered := note
		if openTok != "" {
			rendered = openTok + inner + closeTok
		}
		return r.renderBackticks(
			strings.TrimRight(prefix, " "),
			nil,
		) + " " + r.renderBackticks(
			rendered,
			style,
		)
	}

	return r.renderBackticks(desc, nil)
}

// renderBackticks replaces `text` and 'text' with styled text (delimiters removed).
// Single-quoted strings are only matched when not preceded/followed by a letter,
// so contractions like "don't" are left intact.
// When HelpDescBacktick is nil, delimiters are left intact.
func (r *Renderer) renderBackticks(s string, base *lipgloss.Style) string {
	if r.plain || r.backtickStyle == BacktickStyleNone {
		return s
	}
	if r.Theme.HelpDescBacktick == nil && base == nil {
		return s
	}
	var sb strings.Builder
	writePlain := func(text string) {
		if text == "" {
			return
		}
		if base != nil {
			sb.WriteString(base.Render(text))
			return
		}
		sb.WriteString(text)
	}
	renderCode := func(text string) string {
		style, hasStyle := r.getBacktickStyle(text)
		if !hasStyle {
			if base != nil {
				return base.Render(text)
			}
			return text
		}
		if base != nil {
			style = style.Inherit(*base)
		}
		return style.Render(text)
	}
	last := 0
	for i := 0; i < len(s); {
		switch {
		case s[i] == '`':
			end := strings.IndexByte(s[i+1:], '`')
			if end < 0 {
				writePlain(s[last:])
				return sb.String()
			}
			writePlain(s[last:i])
			end += i + 1
			sb.WriteString(renderCode(s[i+1 : end]))
			i = end + 1
			last = i

		case s[i] == '\'' && !isLetterAt(s, i-1):
			end := strings.IndexByte(s[i+1:], '\'')
			if end < 0 || isLetterAt(s, i+1+end+1) {
				i++
				continue
			}
			writePlain(s[last:i])
			end += i + 1
			sb.WriteString(renderCode(s[i+1 : end]))
			i = end + 1
			last = i

		default:
			i++
		}
	}
	writePlain(s[last:])
	return sb.String()
}

func (r *Renderer) getBacktickStyle(text string) (lipgloss.Style, bool) {
	if r.descRefs.lookupFlag(text) {
		style, hasStyle := r.flagBacktickBaseStyle()
		if r.Theme.HelpDescBacktick == nil {
			return style, hasStyle
		}
		if !hasStyle {
			return *r.Theme.HelpDescBacktick, true
		}
		return style.Inherit(*r.Theme.HelpDescBacktick), true
	}

	// Context-aware lookup: if the token matches a known positional arg
	// (e.g. `name` or `<name>`) or a known subcommand collected from the
	// rendered sections, prefer that style so descriptions stay consistent
	// with how the same name renders elsewhere in the help screen. Skipped
	// when the renderer is configured for non-smart backtick styling.
	if r.backtickStyle == BacktickStyleSmart {
		if style, ok := r.descRefs.lookup(text, r.Theme); ok {
			if r.Theme.HelpDescBacktick == nil {
				return style, true
			}
			return style.Inherit(*r.Theme.HelpDescBacktick), true
		}
	}

	if r.Theme.HelpDescBacktick == nil {
		return lipgloss.Style{}, false
	}
	return *r.Theme.HelpDescBacktick, true
}

// collectDescRefs walks the rendered sections and indexes positional args
// (from Usage and Args content) and subcommands (from CommandGroup content)
// so backtick styling in Description can resolve them by name.
func collectDescRefs(sections []Section) descRefs {
	refs := descRefs{
		args:      map[string]Arg{},
		commands:  map[string]struct{}{},
		flags:     map[string]struct{}{},
		argEnums:  map[string]Arg{},
		flagEnums: map[string]struct{}{},
	}
	indexArg := func(a Arg) {
		refs.args[a.Name] = a
		// Index each known value so a backtick token like `github` resolves
		// to its owning arg and inherits that arg's color. First writer wins
		// if the same value is shared across args (rare); arg name lookup
		// still takes precedence in lookup().
		for _, v := range a.Enum {
			if v == "" {
				continue
			}
			if _, ok := refs.argEnums[v]; !ok {
				refs.argEnums[v] = a
			}
		}
	}
	indexFlag := func(f Flag) {
		if utf8.RuneCountInString(f.Short) == 1 {
			refs.flags["-"+f.Short] = struct{}{}
		}
		if f.Long != "" {
			refs.flags["--"+f.Long] = struct{}{}
			if name, inverse, ok := splitNegatableLongFlag(f.Long); ok {
				refs.flags["--"+name] = struct{}{}
				refs.flags["--"+inverse] = struct{}{}
			}
		}
		// Flag enum values style with the flag color, so a token like `debug`
		// (a value of --log-level) matches its owning flag. Arg-owned enums
		// take precedence in lookup(), so values shared by an arg keep the arg
		// color.
		for _, v := range f.Enum {
			if v != "" {
				refs.flagEnums[v] = struct{}{}
			}
		}
	}
	var walkSections func([]Section)
	walkSections = func(sections []Section) {
		for _, sec := range sections {
			for _, c := range sec.Content {
				switch v := c.(type) {
				case Usage:
					for _, a := range v.Args {
						indexArg(a)
					}
					// Capture the binary name (first token of the Usage command)
					// so multi-segment command references like "mycli sub cmd" in
					// Description backticks can be styled consistently with how
					// the Usage line renders the same name.
					if refs.binary == "" && v.Command != "" {
						if first, _, _ := strings.Cut(v.Command, " "); first != "" {
							refs.binary = first
						}
					}
				case Args:
					for _, a := range v {
						indexArg(a)
					}
				case FlagGroup:
					for _, f := range v {
						indexFlag(f)
					}
				case CommandGroup:
					for _, cmd := range v {
						refs.commands[cmd.Name] = struct{}{}
					}
				case *Section:
					walkSections([]Section{*v})
				}
			}
		}
	}
	walkSections(sections)
	return refs
}

func splitNegatableLongFlag(long string) (string, string, bool) {
	prefix, rest, found := strings.Cut(long, "]")
	if !found || !strings.HasPrefix(prefix, "[") || rest == "" {
		return "", "", false
	}
	inversePrefix := strings.TrimPrefix(prefix, "[")
	if inversePrefix == "" {
		return "", "", false
	}
	return rest, inversePrefix + rest, true
}

func (d descRefs) lookupFlag(text string) bool {
	if _, ok := d.flags[text]; ok {
		return true
	}
	name, _, ok := strings.Cut(text, "=")
	if !ok {
		return false
	}
	_, ok = d.flags[name]
	return ok
}

// lookup returns the style to apply to a backticked token, if it resolves
// to a known positional arg or subcommand. Returns false when the token
// matches nothing known, letting the caller fall back to the default
// description-backtick style.
func (d descRefs) lookup(text string, th *theme.Theme) (lipgloss.Style, bool) {
	// Strip <> wrappers so `<name>` and `name` both resolve.
	name := strings.TrimSuffix(strings.TrimPrefix(text, "<"), ">")
	if a, ok := d.args[name]; ok {
		if style := argStyleForRef(a, d.args, th); style != nil {
			return *style, true
		}
	}
	// A known enum value (e.g. `github`) styles like the arg that declares it,
	// so cross-references such as "...for `github` only" match the <provider>
	// arg's color rather than the generic backtick fallback. Checked after the
	// arg-name lookup so a name collision prefers the arg itself.
	if a, ok := d.argEnums[text]; ok {
		if style := argStyleForRef(a, d.args, th); style != nil {
			return *style, true
		}
	}
	if _, ok := d.commands[text]; ok {
		if th.HelpSubcommand != nil {
			return *th.HelpSubcommand, true
		}
		if th.HelpCommand != nil {
			return *th.HelpCommand, true
		}
	}
	// A flag's enum value (e.g. `debug` for --log-level) styles with the flag
	// color, mirroring how the arg case styles with the arg color. Checked
	// after args/commands so a value those own keeps its more specific style.
	if _, ok := d.flagEnums[text]; ok {
		if th.HelpFlag != nil {
			return *th.HelpFlag, true
		}
	}
	// Multi-segment command path: token is the binary itself or starts with
	// "<binary> ", e.g. "mycli", "mycli sub", "mycli sub cmd". Style the
	// whole token with HelpCommand so cross-command references render the
	// same way the Usage line does.
	if d.binary != "" && (text == d.binary || strings.HasPrefix(text, d.binary+" ")) {
		if th.HelpCommand != nil {
			return *th.HelpCommand, true
		}
	}
	return lipgloss.Style{}, false
}

// argStyleForRef mirrors Renderer.argStyle but operates on the args map
// from descRefs, so backtick styling in descriptions picks the same
// optional/required treatment as the Arguments section.
func argStyleForRef(a Arg, all map[string]Arg, th *theme.Theme) *lipgloss.Style {
	if a.Required {
		return th.HelpArg
	}
	for _, o := range all {
		if o.Required {
			return th.HelpArgOptional
		}
	}
	return th.HelpArg
}

func (r *Renderer) flagBacktickBaseStyle() (lipgloss.Style, bool) {
	switch {
	case r.Theme.HelpFlagBacktick != nil:
		return *r.Theme.HelpFlagBacktick, true
	case r.Theme.HelpFlag != nil:
		return *r.Theme.HelpFlag, true
	default:
		return lipgloss.Style{}, false
	}
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
	idx, ok := trailingBalancedSuffixStart(desc, openTok, closeTok)
	if !ok {
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
	return r.renderBackticks(
		strings.TrimRight(prefix, " "),
		nil,
	) + " " + r.renderBackticks(
		note,
		&style,
	), true
}

func trailingBalancedSuffixStart(s, openTok, closeTok string) (int, bool) {
	if openTok == "" || closeTok == "" || !strings.HasSuffix(s, closeTok) {
		return -1, false
	}

	depth := 0
	for i := len(s); i > 0; {
		switch {
		case strings.HasSuffix(s[:i], closeTok):
			depth++
			i -= len(closeTok)
		case strings.HasSuffix(s[:i], openTok):
			depth--
			i -= len(openTok)
			if depth == 0 {
				return i, true
			}
		default:
			i--
		}
	}
	return -1, false
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

	wrapped := gansi.WrapSoft(desc, avail)
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
				rewrapped := gansi.WrapSoft(contText, contAvail)
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
		wrapped := gansi.WrapSoft(prefix, avail)
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
	bracketWrapped := gansi.WrapSoft(bracket, avail)
	bracketLines := strings.Split(bracketWrapped, "\n")
	if len(bracketLines) > 1 {
		contAvail := avail - 1
		if contAvail > 0 {
			contText := strings.Join(bracketLines[1:], " ")
			rewrapped := gansi.WrapSoft(contText, contAvail)
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

// writerIsPlain reports whether w will discard colour styling (no TTY, or
// NO_COLOR / --color=never style downgrades). Used to keep markup delimiters
// like backticks intact when there's no way to render them visually.
func writerIsPlain(w io.Writer) bool {
	cw, ok := w.(*colorprofile.Writer)
	if !ok {
		return false
	}
	return cw.Profile == colorprofile.NoTTY || cw.Profile == colorprofile.Ascii
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
