package main

import (
	"os"

	clibcobra "github.com/gechr/clib/cli/cobra"
	"github.com/gechr/clib/complete"
	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{Use: "nativeapp", SilenceErrors: true, SilenceUsage: true}
	root.PersistentFlags().String("global", "", "Global text")
	root.Flags().Bool("ancestor-only", false, "Ancestor flag")
	cmd := &cobra.Command{
		Use: "login", Aliases: []string{"l"}, Run: func(*cobra.Command, []string) {},
		Annotations: map[string]string{"clib": "dynamic-args='item'"},
	}
	cmd.Flags().String("format", "", "Format")
	clibcobra.Extend(
		cmd.Flags().Lookup("format"),
		clibcobra.FlagExtra{Enum: []string{"json", "yaml"}},
	)
	cmd.Flags().String("text", "", "Text")
	cmd.Flags().String("target", "", "Target")
	clibcobra.Extend(cmd.Flags().Lookup("target"), clibcobra.FlagExtra{Complete: "predictor=item"})
	cmd.Flags().String("config", "", "Config")
	clibcobra.Extend(cmd.Flags().Lookup("config"), clibcobra.FlagExtra{Extension: "yaml,yml"})
	cmd.Flags().StringSlice("items", nil, "Items")
	clibcobra.Extend(
		cmd.Flags().Lookup("items"),
		clibcobra.FlagExtra{Enum: []string{"a", "b", "c"}},
	)
	for name, hint := range map[string]string{"dir": "dir", "file": "file"} {
		cmd.Flags().String(name, "", name)
		clibcobra.Extend(cmd.Flags().Lookup(name), clibcobra.FlagExtra{Hint: hint})
	}
	other := &cobra.Command{Use: "other", Run: func(*cobra.Command, []string) {}}
	other.Flags().Bool("sibling-only", false, "Sibling flag")
	root.AddCommand(cmd, other)
	gen, err := clibcobra.RegisterCompletions(root, clibcobra.WithCompletionHandler(
		func(c complete.Context) ([]string, cobra.ShellCompDirective) {
			if c.AfterDash {
				return nil, cobra.ShellCompDirectiveDefault
			}
			if c.ToComplete == "hold" {
				return []string{
					"holding",
				}, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
			}
			return []string{"ns:value", "two words", "profile"}, cobra.ShellCompDirectiveNoFileComp
		}))
	if err != nil {
		os.Exit(1)
	}
	root.AddCommand(clibcobra.CompletionCommand(root, func() *complete.Generator { return gen }))
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
