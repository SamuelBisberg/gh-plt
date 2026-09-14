// Package cmd wires gh-plt's cobra command tree.
package cmd

import (
	"github.com/spf13/cobra"
)

// Version is injected at build time (see main.go / .github/workflows/release.yml).
var Version = "dev"

// NewRootCmd builds the `gh plt` root command and all its subcommands.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "plt",
		Short: "Smart template scaffolding for GitHub Actions workflows",
		Long: "gh-plt generates best-practice GitHub Actions workflows and composite\n" +
			"actions from a curated, context-aware template catalog.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       Version,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	root.AddCommand(newAddCmd(), newSearchCmd(), newConfigCmd())
	return root
}
