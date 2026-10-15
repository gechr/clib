package cobra

import (
	cobralib "github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CompletionContext describes a native Cobra completion request. Args contains
// parsed positionals preceding ToComplete. Flag is nil for positional completion.
// For comma lists, ToComplete contains the entire word, but the handler returns
// next-item candidates without the already typed comma prefix.
type CompletionContext struct {
	Command    *cobralib.Command
	Args       []string
	Flag       *pflag.Flag
	Kind       string
	ToComplete string
	AfterDash  bool
}

// CompletionHandler returns candidates (optionally value\tdescription) and a
// Cobra directive. Return ShellCompDirectiveNoFileComp for text, or
// ShellCompDirectiveDefault to allow files. complete.Handler and the --@complete
// protocol remain available for previously installed scripts.
type CompletionHandler func(CompletionContext) ([]string, cobralib.ShellCompDirective)
