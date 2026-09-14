package pkg

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/charmbracelet/huh"
)

// WriteStrategy selects how a TemplateFile's rendered content is combined
// with whatever already exists at its Destination.
type WriteStrategy string

const (
	// StrategyOverride replaces the destination's content entirely. This is
	// the default when a TemplateFile leaves Strategy unset.
	StrategyOverride WriteStrategy = "override"
	// StrategyAppend adds the rendered content to the end of the
	// destination's existing content, separated by a newline. If the
	// destination doesn't exist yet, it's created with just the rendered
	// content.
	StrategyAppend WriteStrategy = "append"
	// StrategyMerge deep-merges the rendered content into the destination
	// as JSON objects (both the existing file and the rendered content must
	// parse as JSON objects). If the destination doesn't exist yet, it's
	// created with just the rendered content.
	StrategyMerge WriteStrategy = "merge"
)

// Resolved returns s, or StrategyOverride if s is the zero value - the
// default for a TemplateFile that doesn't declare a strategy.
func (s WriteStrategy) Resolved() WriteStrategy {
	if s == "" {
		return StrategyOverride
	}
	return s
}

// ErrCancelled is returned by WriteFile when the user declines to write.
var ErrCancelled = errors.New("generator: cancelled")

// Render executes body as a text/template using answers as the data
// context.
func Render(body string, answers map[string]any) (string, error) {
	tmpl, err := template.New("workflow").Parse(body)
	if err != nil {
		return "", fmt.Errorf("generator: parsing template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, answers); err != nil {
		return "", fmt.Errorf("generator: rendering template: %w", err)
	}
	return buf.String(), nil
}

// mergedContent computes what should end up on disk at a destination whose
// current content is existing (existed reports whether the destination
// exists at all), combining it with the freshly rendered content per
// strategy. It has no side effects, so it's the piece of WriteFile that can
// be unit tested without a terminal.
func mergedContent(existing []byte, existed bool, content string, strategy WriteStrategy) (string, error) {
	if !existed {
		// Nothing to append to or merge into yet - every strategy reduces
		// to just writing the rendered content.
		return content, nil
	}

	switch strategy.Resolved() {
	case StrategyOverride:
		return content, nil

	case StrategyAppend:
		s := string(existing)
		if len(s) > 0 && !strings.HasSuffix(s, "\n") {
			s += "\n"
		}
		return s + content, nil

	case StrategyMerge:
		return mergeJSON(existing, []byte(content))

	default:
		return "", fmt.Errorf("generator: unknown write strategy %q", strategy)
	}
}

// mergeJSON deep-merges patch into base, both of which must be JSON
// objects: nested objects are merged key by key, and any other value in
// patch (including arrays) replaces base's value at that key.
func mergeJSON(base, patch []byte) (string, error) {
	var baseObj map[string]any
	if err := json.Unmarshal(base, &baseObj); err != nil {
		return "", fmt.Errorf("existing file is not a JSON object: %w", err)
	}

	var patchObj map[string]any
	if err := json.Unmarshal(patch, &patchObj); err != nil {
		return "", fmt.Errorf("rendered content is not a JSON object: %w", err)
	}

	merged, err := json.MarshalIndent(deepMergeJSON(baseObj, patchObj), "", "  ")
	if err != nil {
		return "", err
	}
	return string(merged) + "\n", nil
}

func deepMergeJSON(base, patch map[string]any) map[string]any {
	result := make(map[string]any, len(base)+len(patch))
	for k, v := range base {
		result[k] = v
	}
	for k, pv := range patch {
		if bv, ok := result[k]; ok {
			if bvObj, ok := bv.(map[string]any); ok {
				if pvObj, ok := pv.(map[string]any); ok {
					result[k] = deepMergeJSON(bvObj, pvObj)
					continue
				}
			}
		}
		result[k] = pv
	}
	return result
}

// WriteFile combines content with whatever's already at destination (a path
// relative to the current directory, e.g. ".github/workflows/test.yml",
// "CHANGELOG.md") according to strategy, always asking for confirmation
// before writing - creating a new file is confirmed just as an overwrite,
// append, or merge is. It creates any parent directories needed and returns
// ErrCancelled if the user declines.
func WriteFile(destination, content string, strategy WriteStrategy) (string, error) {
	path := filepath.Clean(destination)

	existing, err := os.ReadFile(path)
	existed := err == nil
	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("generator: checking %s: %w", path, err)
	}

	final, err := mergedContent(existing, existed, content, strategy)
	if err != nil {
		return "", fmt.Errorf("generator: %s: %w", path, err)
	}

	title := fmt.Sprintf("Create %s?", path)
	if existed {
		switch strategy.Resolved() {
		case StrategyAppend:
			title = fmt.Sprintf("Append to %s?", path)
		case StrategyMerge:
			title = fmt.Sprintf("Merge into %s?", path)
		default:
			title = fmt.Sprintf("%s already exists. Overwrite it?", path)
		}
	}

	// Default to "yes" only for a brand-new file: creating one can't lose
	// data, but overwriting, appending to, or merging into one that already
	// exists can, so those default to "no" and require explicit opt-in.
	proceed := !existed
	confirm := huh.NewConfirm().Title(title).Value(&proceed)
	if err := huh.NewForm(huh.NewGroup(confirm)).Run(); err != nil {
		return "", fmt.Errorf("generator: %w", err)
	}
	if !proceed {
		return "", ErrCancelled
	}

	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", fmt.Errorf("generator: creating %s: %w", dir, err)
		}
	}
	if err := os.WriteFile(path, []byte(final), 0o644); err != nil {
		return "", fmt.Errorf("generator: writing %s: %w", path, err)
	}
	return path, nil
}
