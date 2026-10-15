package complete

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Context describes a native Cobra completion request. Args contains parsed
// positionals preceding ToComplete. Flag is nil for positional completion.
// For comma lists, ToComplete contains the entire word, but the handler returns
// next-item candidates without the already typed comma prefix.
type Context struct {
	Command    *cobra.Command
	Args       []string
	Flag       *pflag.Flag
	Kind       string
	ToComplete string
	AfterDash  bool
}

// ContextHandler returns candidates (optionally value\tdescription) and a Cobra
// directive. Return ShellCompDirectiveNoFileComp for text, or
// ShellCompDirectiveDefault to allow files. Handler and the --@complete protocol
// remain available for previously installed scripts.
type ContextHandler func(Context) ([]string, cobra.ShellCompDirective)

// IsRequest reports whether args (excluding argv[0]) is a native completion
// invocation. Words following the request command are data, not action flags.
func IsRequest(args []string) bool {
	return len(args) > 0 && (args[0] == cobra.ShellCompRequestCmd ||
		args[0] == cobra.ShellCompNoDescRequestCmd)
}
