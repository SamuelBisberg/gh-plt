package cmd

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/SamuelBisberg/gh-plt/pkg"
)

func newConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "View or change gh-plt's default settings",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfig()
		},
	}
}

func runConfig() error {
	theme := pkg.NewTheme()

	cfg, err := pkg.LoadConfig()
	if err != nil {
		return err
	}
	if cfg.Installers == nil {
		cfg.Installers = map[string]string{}
	}

	langs, err := pkg.AllLanguages()
	if err != nil {
		return err
	}

	// One group per question, so the form steps through them one at a time
	// instead of showing every setting stacked on a single page.
	groups := make([]*huh.Group, 0, len(langs)+1)
	installerVals := make(map[string]*string, len(langs))

	for _, lang := range langs {
		if len(lang.PackageManagers) < 2 {
			continue // nothing to choose between
		}

		names := make([]string, len(lang.PackageManagers))
		for i, pm := range lang.PackageManagers {
			names[i] = pm.Name
		}

		val := cfg.Installers[lang.Name]
		installerVals[lang.Name] = &val

		groups = append(groups, huh.NewGroup(
			huh.NewSelect[string]().
				Title(fmt.Sprintf("%s installer", lang.Name)).
				Options(huh.NewOptions(names...)...).
				Value(&val),
		))
	}

	groups = append(groups, huh.NewGroup(
		huh.NewSelect[pkg.Versioning]().
			Title("Action versioning").
			Options(huh.NewOptions(pkg.VersioningTag, pkg.VersioningHash)...).
			Value(&cfg.Actions.Versioning),
	))

	if err := huh.NewForm(groups...).Run(); err != nil {
		return err
	}

	for name, val := range installerVals {
		cfg.Installers[name] = *val
	}

	if err := cfg.Save(); err != nil {
		return err
	}

	path, _ := pkg.ConfigPath()
	theme.Successf("Saved config to %s", path)
	return nil
}
