package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/SamuelBisberg/gh-plt/pkg"
)

func newAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add [template]",
		Short: "Generate a template, suggesting one if no path is given",
		Long: "Generate a specific template by its catalog path, bypassing suggestions,\n" +
			"e.g. `gh plt add php/laravel/ci`. Without a path, detects (or prompts for)\n" +
			"the project's language and framework, then suggests matching templates.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				return runAddPath(args[0])
			}
			return runAddSuggest()
		},
	}
}

func runAddPath(path string) error {
	theme := pkg.NewTheme()

	t, ok, err := pkg.FindTemplate(path)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("no template found at %q (run `gh plt search` to browse the catalog)", path)
	}

	return generate(theme, t)
}

// runAddSuggest resolves the current project's language and framework -
// detecting them where possible, prompting for whichever couldn't be
// detected - then narrows the catalog to templates that apply and lets the
// user pick one.
func runAddSuggest() error {
	theme := pkg.NewTheme()

	ctx := pkg.Detect(".")
	lang, framework := ctx.Language, ctx.Framework

	if lang == "" {
		langs, err := pkg.AllLanguages()
		if err != nil {
			return err
		}
		chosen, ok, err := pkg.PromptLanguage(langs)
		if err != nil {
			return err
		}
		if !ok {
			theme.Infof("No language selected")
			return nil
		}
		lang = chosen
	}

	if framework == "" {
		l, ok, err := pkg.FindLanguage(lang)
		if err != nil {
			return err
		}
		if ok && len(l.Frameworks) > 0 {
			chosen, ok, err := pkg.PromptFramework(l.Frameworks)
			if err != nil {
				return err
			}
			if !ok {
				theme.Infof("No framework selected")
				return nil
			}
			framework = chosen
		}
	}

	all, err := pkg.AllTemplates()
	if err != nil {
		return err
	}

	suggested := pkg.FilterTemplates(all, lang, framework)
	if len(suggested) == 0 {
		return fmt.Errorf("no templates found for %s (run `gh plt search` to browse the full catalog)", lang)
	}

	t, ok, err := pkg.Search(suggested, pkg.ProjectContext{Language: lang, Framework: framework})
	if err != nil {
		return err
	}
	if !ok {
		theme.Infof("No template selected")
		return nil
	}

	return generate(theme, t)
}
