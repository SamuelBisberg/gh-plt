package cmd

import (
	"errors"
	"fmt"

	"github.com/SamuelBisberg/gh-plt/pkg"
)

// collectVariables returns the union of every file's declared Variables, in
// first-seen order and deduplicated by name, so a name shared by multiple
// files (e.g. a "coverage" flag used by both a workflow and a changelog
// snippet) is only prompted for once.
func collectVariables(files []pkg.TemplateFile) []pkg.Variable {
	seen := make(map[string]bool)
	var vars []pkg.Variable
	for _, f := range files {
		for _, v := range f.Variables {
			if !seen[v.Name] {
				seen[v.Name] = true
				vars = append(vars, v)
			}
		}
	}
	return vars
}

// collectRequirements returns the union of every file's declared Requires,
// deduplicated so an identical dependency needed by two files is only
// installed once.
func collectRequirements(files []pkg.TemplateFile) []pkg.Requirement {
	seen := make(map[pkg.Requirement]bool)
	var reqs []pkg.Requirement
	for _, f := range files {
		for _, r := range f.Requires {
			if !seen[r] {
				seen[r] = true
				reqs = append(reqs, r)
			}
		}
	}
	return reqs
}

// generate runs the full scaffolding pipeline for a selected template:
// collect variables and dependencies across all of its files -> prompt once
// -> install missing dependencies -> for each file, confirm or rename its
// output path, render, resolve action versions, and write it.
func generate(theme *pkg.Theme, t pkg.Template) error {
	answers, err := pkg.PromptVariables(collectVariables(t.Files))
	if err != nil {
		return err
	}

	ctx := pkg.Detect(".")
	if reqs := collectRequirements(t.Files); len(reqs) > 0 {
		if err := pkg.InstallMissing(reqs, ctx); err != nil {
			return err
		}
	}

	cfg, err := pkg.LoadConfig()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	for _, f := range t.Files {
		dest, err := pkg.PromptDestination(f.Source, f.Destination)
		if err != nil {
			return err
		}

		raw, err := pkg.ReadTemplateFile(t, f)
		if err != nil {
			return err
		}

		rendered, err := pkg.Render(string(raw), answers)
		if err != nil {
			return err
		}

		rendered, err = pkg.ResolveVersions(rendered, cfg.Actions.Versioning)
		if err != nil {
			return err
		}

		path, err := pkg.WriteFile(dest, rendered, f.Strategy)
		if errors.Is(err, pkg.ErrCancelled) {
			theme.Warningf("Skipped %s", dest)
			continue
		}
		if err != nil {
			return err
		}

		theme.Successf("Wrote %s", path)
	}
	return nil
}
