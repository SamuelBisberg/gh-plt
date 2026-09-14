package cmd

import (
	"github.com/spf13/cobra"

	"github.com/SamuelBisberg/gh-plt/pkg"
)

func newSearchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search",
		Short: "Browse and fuzzy-search the template catalog",
		Args:  cobra.NoArgs,
		RunE:  runSearch,
	}
}

func runSearch(cmd *cobra.Command, args []string) error {
	theme := pkg.NewTheme()

	all, err := pkg.AllTemplates()
	if err != nil {
		return err
	}

	ctx := pkg.Detect(".")
	t, ok, err := pkg.Search(all, ctx)
	if err != nil {
		return err
	}
	if !ok {
		theme.Infof("No template selected")
		return nil
	}

	return generate(theme, t)
}
