# clib

A reusable Go library that plugs into existing CLI frameworks to add helpers, shell completions, and more polished help output.

## Packages

| Package      | Description                                                                    |
| ------------ | ------------------------------------------------------------------------------ |
| `cli/cobra`  | [Cobra](https://github.com/spf13/cobra) framework adapters                     |
| `cli/kong`   | [Kong](https://github.com/alecthomas/kong) framework adapters                  |
| `cli/urfave` | [urfave/cli](https://github.com/urfave/cli) framework adapters                 |
| `complete`   | Shell completion generation (bash, zsh, fish, pwsh, elvish, nu)                |
| `help`       | Structured help rendering with themed output                                   |
| `theme`      | Configurable theme (via [lipgloss](https://github.com/charmbracelet/lipgloss)) |

## Installation

```text
go get github.com/gechr/clib
```

## Usage

### Theme

Use the built-in light/dark pair, or build an explicit pair for your terminal
backgrounds:

```go
// Auto-detect the terminal background and select from clib's defaults.
th := theme.Auto()

// Or require an explicit light/dark pair.
themes := theme.MustPair(
    theme.Light().With(
        theme.WithRed(lipgloss.NewStyle().Foreground(lipgloss.Color("#a11d33"))),
        theme.WithEntityColors([]color.Color{lipgloss.Color("#5d4037"), lipgloss.Color("#00695c")}),
    ),
    theme.Dark().With(
        theme.WithRed(lipgloss.NewStyle().Foreground(lipgloss.Color("9"))),
        theme.WithEntityColors([]color.Color{lipgloss.Color("208"), lipgloss.Color("51")}),
    ),
)
th = themes.Auto()
```

To let users pick themes via the environment, build the pair from `<PREFIX>_THEME_LIGHT` and `<PREFIX>_THEME_DARK` (both required):

```go
// Reads CLIB_THEME_LIGHT and CLIB_THEME_DARK (override the prefix with SetEnvPrefix).
themes, err := theme.PairFromEnv()
if err != nil {
    return err
}
th := themes.Auto()
```

Setting the single `<PREFIX>_THEME` (e.g. `CLIB_THEME=dracula`) forces that theme regardless of the terminal background, taking precedence over the light/dark pair. Any `Auto()` call honors it, including the default `theme.Auto()` and `help.NewRenderer(nil)`. An unknown name is ignored and detection proceeds as normal.

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

Commands that forward to another command can be marked as aliases. They are removed from the `Commands` section and shown immediately below it in an `Aliases` section, aligned to the same description column:

```text
Commands

  serve    Serve requests

Aliases

  publish  Alias for tool release
```

Declare an alias using the adapter's command metadata:

- Kong: `` Publish CmdPublish `cmd:"" clib:"alias='tool release'"` ``
- Cobra: `cobracli.ExtendCommand(publishCmd, cobracli.CommandExtra{Alias: "tool release"})`
- urfave/cli: `cliurfave.ExtendCommand(publishCmd, cliurfave.CommandExtra{Alias: "tool release"})`

The forwarding behavior remains the application's responsibility; this metadata only controls how the command appears in generated help.

Pass `WithInlineCommandAliases()` to the adapter's section builder to keep alias commands in `Commands` instead of creating a separate `Aliases` section, or `WithHideCommandAliases()` to omit them from help.

#### [Kong](https://github.com/alecthomas/kong)

Annotate your CLI struct with Kong-style tags and a `clib:"..."` tag for
`clib`-specific metadata (grouping, descriptions, completions, placeholders):

```go
type CLI struct {
    clib.CompletionFlags

    Query  string        `name:"query"  short:"q" help:"Filter results by query" clib:"group='Filters',terse='Query',complete='predictor=query',order=keep"`
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
  clib.CSVFlagPlaceholders(),
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
  Group: "Filters", Placeholder: "text", Complete: "predictor=query", Order: complete.OrderKeep,
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
  Group: "Filters", Placeholder: "text", Complete: "predictor=query", Order: complete.OrderKeep,
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

Bash completions handle values containing word-break characters such as `@`, `:`,
and `=`, including `--flag=value`, without changing the shell's word-break settings.
Regenerate installed completion scripts to pick up generator improvements.

The `terse` key provides a very short description for completions (falls back to `help`).
Use `order=keep` (Kong) or `Order: complete.OrderKeep` (Cobra/urfave) to preserve fish completion order for a flag by emitting `complete -k`.
Use `order=shell` or `Order: complete.OrderShell` to force the shell's normal ordering for a flag.
Use `complete.WithOrder(complete.OrderKeep)` to make keep-order the generator default.

Flags hidden from `--help` are omitted from completions by default. Use `complete-hidden` (Kong) or `FlagExtra.CompleteHidden` (Cobra/urfave) to still offer a specific hidden flag as a completion, or `complete.WithIncludeHidden()` to offer every hidden flag across the command tree.

#### Native Cobra completions

Call `RegisterCompletions` after building the command tree and registering any native Cobra callbacks. It returns a generator whose Bash, Fish, and Zsh scripts ask the application through `__complete`. Existing native callbacks take precedence. Text and number values suppress file completion by default; `Hint: "file"`, `Hint: "dir"`, and `Extension` opt into filesystem completion.

```go
import (
    clibcobra "github.com/gechr/clib/cli/cobra"
    "github.com/gechr/clib/complete"
    "github.com/spf13/cobra"
)

root := &cobra.Command{
    Use: "app [profiles...]",
    Args: cobra.ArbitraryArgs,
    Run: func(*cobra.Command, []string) {},
    Annotations: map[string]string{"clib": "dynamic-args='profile'"},
}
root.Flags().String("profile", "", "Profile to use")
clibcobra.Extend(root.Flags().Lookup("profile"), clibcobra.FlagExtra{
    Complete: "predictor=profile",
})
gen, err := clibcobra.RegisterCompletions(root,
    clibcobra.WithCompletionHandler(func(c complete.Context) ([]string, cobra.ShellCompDirective) {
        if c.AfterDash {
            return nil, cobra.ShellCompDirectiveDefault
        }
        return []string{"development\tLocal development", "production\tProduction"},
            cobra.ShellCompDirectiveNoFileComp
    }),
)
if err != nil {
    return err
}
root.AddCommand(clibcobra.CompletionCommand(root, func() *complete.Generator { return gen }))
return root.Execute()
```

Use `app completion bash`, `app completion fish`, or `app completion zsh` to print a script, or pass the returned generator to the existing install/preflight helpers. Bash supports 3.2 and newer without the bash-completion package. Regenerate installed scripts to switch to native completion. Keep your legacy `complete.Handler` and `--@complete` handling for users with older scripts.

`complete.Context` carries `Command`, parsed positional `Args`, `Flag` (nil for positionals), predictor `Kind`, `ToComplete`, and `AfterDash`. Predictors return values with optional tab-separated descriptions and a Cobra directive. Comma lists retain the typed prefix and exclude already selected items. Their predictors receive the entire word in `ToComplete` and return next-item candidates without the comma prefix; the bridge keeps the cursor in the list by suppressing a trailing space. The final `dynamic-args` predictor repeats, subject to the command's positional limit; after `--`, the predictor decides what to complete.

`clibcobra.CommandMeta(cmd)` exposes `DynamicArgs` and `PathArgs` for any command, including the root. Native callbacks can use `clibcobra.AfterDash(cmd)` directly. The bridge preserves existing root pre-run hooks; install it after setting them, and use `complete.IsRequest(os.Args[1:])` to bypass application initialization during completion. `Preflight()` ignores both `__complete` and `__completeNoDesc` requests, including any install, uninstall, or print flags being completed.

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

## Examples

Working examples for each framework adapter are in the `examples/` directory:

- [`examples/cobra`](examples/cobra/main.go)
- [`examples/kong`](examples/kong/main.go)
- [`examples/urfave`](examples/urfave/main.go)
