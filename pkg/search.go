package pkg

import (
	"errors"
	"fmt"
	"sort"

	"github.com/charmbracelet/huh"
)

// Search prompts the user to pick a template from the catalog via an
// inline, filterable huh.Select. Templates matching ctx's detected language
// are bubbled to the top. It returns the chosen Template, or ok=false if
// the user aborted (e.g. Ctrl+C) without choosing one.
func Search(templates []Template, ctx ProjectContext) (t Template, ok bool, err error) {
	ordered := make([]Template, len(templates))
	copy(ordered, templates)
	sort.SliceStable(ordered, func(i, j int) bool {
		mi := ctx.Language != "" && ordered[i].Lang == ctx.Language
		mj := ctx.Language != "" && ordered[j].Lang == ctx.Language
		return mi && !mj
	})

	// Template itself isn't comparable (it holds a []TemplateFile), so the
	// select is keyed on Path and mapped back to the full Template after.
	byPath := make(map[string]Template, len(ordered))
	options := make([]huh.Option[string], 0, len(ordered))
	for _, tpl := range ordered {
		label := fmt.Sprintf("%s %s", tpl.Icon, tpl.Path)
		if tpl.Description != "" {
			label += " — " + tpl.Description
		}
		options = append(options, huh.NewOption(label, tpl.Path))
		byPath[tpl.Path] = tpl
	}

	title := "Select a template"
	if ctx.Language != "" {
		title = fmt.Sprintf("Select a template (detected: %s)", ctx.Language)
	}

	var choice string
	prompt := huh.NewSelect[string]().
		Title(title).
		Options(options...).
		Filtering(true).
		Height(len(options) + 2).
		Value(&choice)

	if err := huh.NewForm(huh.NewGroup(prompt)).Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return Template{}, false, nil
		}
		return Template{}, false, fmt.Errorf("search: %w", err)
	}

	return byPath[choice], true, nil
}
