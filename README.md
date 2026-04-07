# clib

A reusable Go library that plugs into existing CLI frameworks to add helpers, shell completions, and more polished help output.

## Packages

| Package      | Description                                                                    |
| ------------ | ------------------------------------------------------------------------------ |
| `ansi`       | Terminal-aware ANSI output                                                     |
| `cli/cobra`  | [Cobra](https://github.com/spf13/cobra) framework adapters                     |
| `cli/kong`   | [Kong](https://github.com/alecthomas/kong) framework adapters                  |
| `cli/urfave` | [urfave/cli](https://github.com/urfave/cli) framework adapters                 |
| `complete`   | Shell completion generation (bash, zsh, fish)                                  |
| `help`       | Structured help rendering with themed output                                   |
| `human`      | Human-friendly formatting                                                      |
| `shell`      | Shell detection                                                                |
| `terminal`   | Terminal detection                                                             |
| `theme`      | Configurable theme (via [lipgloss](https://github.com/charmbracelet/lipgloss)) |

## Installation

```text
go get github.com/gechr/clib
```

## Usage

### ANSI Output

Auto-detect terminal support, or force/disable ANSI output:

```go
w := ansi.Auto()                          // detect from os.Stdout
w := ansi.Auto(os.Stdout, os.Stderr)      // all must be terminals
w := ansi.Force()                         // always emit ANSI
w := ansi.Never()                         // plain text only
w := ansi.New(ansi.WithTerminal(true))    // manual configuration

w.Hyperlink("https://example.com", "click here")  // OSC 8 hyperlink

// Control how hyperlinks render in non-terminal output:
w = ansi.New(ansi.WithHyperlinkFallback(ansi.HyperlinkFallbackMarkdown))
// HyperlinkFallbackExpanded (default) → "text (url)"
// HyperlinkFallbackMarkdown           → "[text](url)"
// HyperlinkFallbackText               → "text"
// HyperlinkFallbackURL                → "url"
```

### Terminal Detection

```go
if terminal.Is(os.Stdout) {
    // stdout is a terminal
}

width := terminal.Width(os.Stdout)  // column count, 0 if not a terminal
```

### Theme

Create a theme with defaults or customize individual styles:

```go
// Use defaults.
th := theme.Default()

// Or customize with options.
th := theme.Default().With(
    theme.WithRed(lipgloss.NewStyle().Foreground(lipgloss.Color("9"))),
    theme.WithEntityColors([]lipgloss.Color{"208", "51", "226"}),
)
```

### Help Rendering

Build structured help output. Content blocks implement the `help.Content`
interface: `FlagGroup`, `Args`, `Usage`, `Text`, `Examples`, `CommandGroup`,
and nested `*Section`.

```go
hr := help.NewRenderer(th)

sections := []help.Section{
    {Title: "Flags", Content: []help.Content{
        help.FlagGroup{
            {Short: "q", Long: "query", Placeholder: "text", Desc: "Filter results by query"},
            {
                Short:        "f",
                Long:         "format",
                Placeholder:  "format",
                Desc:         "Output format",
                Enum:         []string{"table", "json", "yaml"},
                EnumDefault:  "table",
            },
        },
    }},
    {Title: "Examples", Content: []help.Content{
        help.Examples{
            {Comment: "List items in JSON format", Command: "catalog list --format json"},
        },
    }},
}

if err := hr.Render(os.Stdout, sections); err != nil {
    return err
}
```

Flag fields use bare names - the renderer adds dashes (`-`/`--`) and angle
brackets (`<`/`>`) automatically. Sections can be nested by including a
`*help.Section` as content.

### Framework Adapters

Each adapter provides the same core capabilities for its framework:
`Extend` (or struct tags for Kong), `FlagMeta`, `Sections`/`HelpFunc`, `NewCompletion`,
and `CSVFlag`.

#### [Kong](https://github.com/alecthomas/kong)

Annotate your CLI struct with Kong-style tags and a `clib:"..."` tag for
`clib`-specific metadata (grouping, descriptions, completions, placeholders):

```go
type CLI struct {
    clib.CompletionFlags

    Query  string        `name:"query"  short:"q" help:"Filter results by query" clib:"group='Filters',terse='Query',complete='predictor=query'"`
    Fields clib.CSVFlag  `name:"fields" help:"Fields to show"                    clib:"complete='predictor=field,comma'"`
    Format string        `name:"format" short:"f" help:"Output format" default:"table" enum:"table,json,yaml"`
}

flags := clib.Reflect(&CLI{})
gen := complete.NewGenerator("catalog").FromFlags(flags)
gen.Install("fish", false)
```

Integrate with Kong's help system using `HelpPrinter` or `HelpPrinterFunc`:

```go
r := help.NewRenderer(th)
k := konglib.Must(&cli,
  konglib.Help(clib.HelpPrinterFunc(r, clib.NodeSectionsFunc())),
)
```

#### [Cobra](https://github.com/spf13/cobra)

Use `Extend` to attach `clib` metadata to pflag flags:

```go
f := root.Flags()
f.StringP("query", "q", "", "Filter results by query")
f.StringP("format", "f", "table", "Output format")
cobraFields := &cobracli.CSVFlag{}
f.Var(cobraFields, "fields", "Fields to show")

cobracli.Extend(f.Lookup("query"), cobracli.FlagExtra{
  Group: "Filters", Placeholder: "text", Complete: "predictor=query",
})
cobracli.Extend(f.Lookup("format"), cobracli.FlagExtra{
  Group: "Output", Placeholder: "format", Enum: []string{"table", "json", "yaml"}, EnumDefault: "table",
})
cobracli.Extend(f.Lookup("fields"), cobracli.FlagExtra{
  Complete: "predictor=field,comma",
})

// Auto-grouped help using extras.
root.SetHelpFunc(cobracli.HelpFunc(r, cobracli.Sections))

// Completion flags (hidden).
comp := cobracli.NewCompletion(root)
gen := complete.NewGenerator("catalog").FromFlags(cobracli.FlagMeta(root))
handled, err := comp.Handle(gen, nil)
```

#### [urfave/cli](https://github.com/urfave/cli)

Use `Extend` to attach `clib` metadata to [urfave/cli](https://github.com/urfave/cli) flags:

```go
queryFlag := &clilib.StringFlag{Name: "query", Aliases: []string{"q"}, Usage: "Filter results by query"}
formatFlag := &clilib.StringFlag{Name: "format", Aliases: []string{"f"}, Usage: "Output format", Value: "table"}
fieldsFlag := &clilib.GenericFlag{Name: "fields", Usage: "Fields to show", Value: &cliurfave.CSVFlag{}}

cliurfave.Extend(queryFlag, cliurfave.FlagExtra{
  Group: "Filters", Placeholder: "text", Complete: "predictor=query",
})
cliurfave.Extend(formatFlag, cliurfave.FlagExtra{
  Group: "Output", Placeholder: "format", Enum: []string{"table", "json", "yaml"}, EnumDefault: "table",
})
cliurfave.Extend(fieldsFlag, cliurfave.FlagExtra{
  Complete: "predictor=field,comma",
})

// Custom help using clib themed renderer.
clilib.HelpPrinter = cliurfave.HelpPrinter(r, cliurfave.Sections)

// Completion flags (hidden).
comp := cliurfave.NewCompletion(root)
gen := complete.NewGenerator("catalog").FromFlags(cliurfave.FlagMeta(cmd))
handled, err := comp.Handle(gen, nil)
```

### Completions

The `complete` package generates shell completion scripts from flag
metadata. The `complete` tag (Kong) or `FlagExtra.Complete` field (Cobra/urfave)
controls completion behavior:

- `predictor=<name>` - dynamic completion via `<app> --@complete=<name>`
- `comma` - comma-separated multi-value mode
- `values=<space-separated>` - static completion values

The `terse` key provides a very short description for completions (falls back to `help`).

### Time-Ago

```go
styled  := th.RenderTimeAgo(someTime, true)     // colored based on thresholds
plain   := human.FormatTimeAgo(someTime)        // "3 hours ago"
compact := human.FormatTimeAgoCompact(someTime) // "3h ago"
```

### Path Formatting

```go
human.ContractHome("/Users/alice/Documents")  // "~/Documents"
```

### Enum Formatting

```go
// Styled enum with shortcut letters: [text, json, yaml]
th.FmtEnum([]theme.EnumValue{
  {Name: "text", Bold: "t"},
  {Name: "json", Bold: "j"},
  {Name: "yaml", Bold: "y"},
})

// With default annotation: [text, json, yaml] (default: text)
th.FmtEnumDefault("text", []theme.EnumValue{
  {Name: "text", Bold: "t"},
  {Name: "json", Bold: "j"},
  {Name: "yaml", Bold: "y"},
})

// Dim default/note annotations for flag descriptions.
th.DimDefault("30")    // (default: 30)
th.DimNote("required") // (required)
```

### Markdown Rendering

Render short markdown strings for inline display with themed code spans:

```go
th.RenderMarkdown("Use `--verbose` for debug output")
```

### Shell Detection

```go
shell := shell.Detect() // COMPLETE_SHELL env -> parent process -> SHELL env
```

## Examples

Working examples for each framework adapter are in the `examples/` directory:

- [`examples/cobra`](examples/cobra/main.go)
- [`examples/kong`](examples/kong/main.go)
- [`examples/urfave`](examples/urfave/main.go)
