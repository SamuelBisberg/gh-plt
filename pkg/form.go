package pkg

import (
	"errors"
	"fmt"

	"github.com/charmbracelet/huh"
)

// PromptVariables builds and runs an interactive huh.Form from a template's
// declared variables, returning the collected answers keyed by variable
// name. It returns an empty map without prompting if vars is empty.
func PromptVariables(vars []Variable) (map[string]any, error) {
	answers := make(map[string]any, len(vars))
	if len(vars) == 0 {
		return answers, nil
	}

	boolVals := make(map[string]*bool, len(vars))
	strVals := make(map[string]*string, len(vars))
	// One group per variable, so the form steps through them one at a time
	// instead of showing every prompt stacked on a single page.
	groups := make([]*huh.Group, 0, len(vars))

	for _, v := range vars {
		switch v.Type {
		case "bool":
			def, _ := v.Default.(bool)
			b := def
			boolVals[v.Name] = &b
			groups = append(groups, huh.NewGroup(huh.NewConfirm().Title(v.Prompt).Value(&b)))
		case "string":
			def, _ := v.Default.(string)
			s := def
			strVals[v.Name] = &s
			groups = append(groups, huh.NewGroup(huh.NewInput().Title(v.Prompt).Value(&s)))
		default:
			return nil, fmt.Errorf("template variable %q: unsupported type %q", v.Name, v.Type)
		}
	}

	if err := huh.NewForm(groups...).Run(); err != nil {
		return nil, fmt.Errorf("form: %w", err)
	}

	for name, b := range boolVals {
		answers[name] = *b
	}
	for name, s := range strVals {
		answers[name] = *s
	}
	return answers, nil
}

// PromptDestination asks the user to confirm or rename a template file's
// output path before it's generated, pre-filled with its declared default
// (def). Pressing enter without editing keeps the default; clearing the
// field entirely also falls back to it, rather than writing to an empty
// path.
func PromptDestination(source, def string) (string, error) {
	dest := def
	input := huh.NewInput().
		Title(fmt.Sprintf("Output path for %s?", source)).
		Value(&dest)

	if err := huh.NewForm(huh.NewGroup(input)).Run(); err != nil {
		return "", fmt.Errorf("form: %w", err)
	}

	if dest == "" {
		return def, nil
	}
	return dest, nil
}

// PromptLanguage asks the user to pick a language from the ecosystem
// catalog, for when it couldn't be detected from the project directory. It
// returns ok=false if the user aborted (e.g. Ctrl+C) without choosing one.
func PromptLanguage(langs []Language) (name string, ok bool, err error) {
	options := make([]huh.Option[string], 0, len(langs))
	for _, l := range langs {
		options = append(options, huh.NewOption(l.Name, l.Name))
	}

	var choice string
	prompt := huh.NewSelect[string]().
		Title("Select a language").
		Options(options...).
		Value(&choice)

	if err := huh.NewForm(huh.NewGroup(prompt)).Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("form: %w", err)
	}
	return choice, true, nil
}

// PromptFramework asks the user to pick a framework from those a language
// declares, or "(none)" to proceed without one. It returns ok=true with
// name="" immediately, without prompting, if frameworks is empty - there's
// nothing to choose between. It returns ok=false if the user aborted (e.g.
// Ctrl+C) without choosing.
func PromptFramework(frameworks []Framework) (name string, ok bool, err error) {
	if len(frameworks) == 0 {
		return "", true, nil
	}

	options := []huh.Option[string]{huh.NewOption("(none)", "")}
	for _, f := range frameworks {
		options = append(options, huh.NewOption(f.Name, f.Name))
	}

	var choice string
	prompt := huh.NewSelect[string]().
		Title("Select a framework").
		Options(options...).
		Value(&choice)

	if err := huh.NewForm(huh.NewGroup(prompt)).Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("form: %w", err)
	}
	return choice, true, nil
}
