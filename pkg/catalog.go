package pkg

import (
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	embedded "github.com/SamuelBisberg/gh-plt/templates"
)

const (
	templatesRoot = "."
	manifestName  = "template.yaml"
)

var kindIcons = map[string]string{
	"build":   "🏗️",
	"ci":      "🧪",
	"release": "🚀",
	"setup":   "⚙️",
}

func iconFor(kind string) string {
	if icon, ok := kindIcons[kind]; ok {
		return icon
	}
	return "📄"
}

// TemplateFile is one output file within a Template: Source is its content
// file relative to the template's own directory, Destination is where it's
// written relative to the project root (e.g. ".github/workflows/ci.yml",
// "CHANGELOG.md"). Requires and Variables are scoped to this file alone, so
// a bundle's files can each need their own dependencies and prompts.
//
// Strategy controls how this file's rendered content is combined with
// whatever's already at Destination: "override" (the default) replaces it
// entirely; "append" adds the rendered content to the end, for a file like
// a CHANGELOG that's meant to grow; "merge" deep-merges it in as JSON, for
// a partial file like package.json where only some keys are ours to set.
type TemplateFile struct {
	Source      string        `yaml:"source"`
	Destination string        `yaml:"destination"`
	Strategy    WriteStrategy `yaml:"strategy"`
	Requires    []Requirement `yaml:"requires"`
	Variables   []Variable    `yaml:"variables"`
}

// Requirement is one local dependency a template file needs installed
// before its generated content will actually pass, tagged with the
// ecosystem it belongs to so the executor can pick the right package
// manager for it (e.g. a Python file's "pytest-cov" should never be run
// through npm).
type Requirement struct {
	Language string `yaml:"language"` // "python", "js", or "php"
	Package  string `yaml:"package"`
}

// Variable describes one user-facing prompt a template file needs answered
// before it can be rendered.
type Variable struct {
	Name    string `yaml:"name"`
	Type    string `yaml:"type"` // "bool" or "string"
	Prompt  string `yaml:"prompt"`
	Default any    `yaml:"default"`
}

// Template describes one scaffoldable bundle discovered in the embedded
// catalog: one or more output Files sharing a catalog Path (e.g.
// "php/laravel/release" might write both a workflow and a CHANGELOG).
type Template struct {
	// Path is the catalog identifier used by `gh plt add <path>`, e.g.
	// "php/laravel/ci", or "js/ci" for a framework-less template, or
	// "release" for one that isn't specific to any language.
	Path string
	// Lang is "" for a template that applies regardless of project
	// language (it lives directly under templates/, e.g.
	// templates/release/).
	Lang string
	// Framework is "" for a template that isn't specific to any framework
	// (it lives directly under its language, e.g. templates/js/ci/).
	Framework   string
	Kind        string
	Icon        string
	Description string
	Files       []TemplateFile
	dir         string // this template's directory in the embedded FS
}

// AllTemplates walks the embedded template filesystem for manifest files
// and returns every discovered Template, sorted by Path for stable output.
// A manifest one level deep (templates/<kind>/template.yaml) applies to any
// language; two levels deep (templates/<lang>/<kind>/template.yaml) is a
// framework-less template for that language; three levels deep
// (templates/<lang>/<framework>/<kind>/template.yaml) is framework-specific.
func AllTemplates() ([]Template, error) {
	var list []Template
	err := fs.WalkDir(embedded.FS, templatesRoot, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || path.Base(p) != manifestName {
			return nil
		}

		dir := path.Dir(p)
		parts := strings.Split(dir, "/")

		var lang, framework, kind string
		switch len(parts) {
		case 1:
			kind = parts[0]
		case 2:
			lang, kind = parts[0], parts[1]
		case 3:
			lang, framework, kind = parts[0], parts[1], parts[2]
		default:
			return nil
		}

		data, err := embedded.FS.ReadFile(p)
		if err != nil {
			return fmt.Errorf("catalog: reading %s: %w", p, err)
		}

		var m struct {
			Description string         `yaml:"description"`
			Files       []TemplateFile `yaml:"files"`
		}
		if err := yaml.Unmarshal(data, &m); err != nil {
			return fmt.Errorf("catalog: parsing %s: %w", p, err)
		}

		list = append(list, Template{
			Path:        dir,
			Lang:        lang,
			Framework:   framework,
			Kind:        kind,
			Icon:        iconFor(kind),
			Description: m.Description,
			Files:       m.Files,
			dir:         dir,
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("catalog: %w", err)
	}

	sort.Slice(list, func(i, j int) bool { return list[i].Path < list[j].Path })
	return list, nil
}

// FindTemplate returns the Template whose Path matches the given catalog
// identifier (e.g. "php/laravel/ci").
func FindTemplate(id string) (Template, bool, error) {
	all, err := AllTemplates()
	if err != nil {
		return Template{}, false, err
	}

	id = strings.Trim(path.Clean(id), "/")
	for _, t := range all {
		if t.Path == id {
			return t, true, nil
		}
	}
	return Template{}, false, nil
}

// FilterTemplates returns the subset of all that's usable for a project of
// the given language and framework: templates with no language restriction
// (they apply everywhere), templates matching lang with no framework
// restriction, and templates matching both lang and framework exactly.
// Passing lang == "" returns all of them, unfiltered.
func FilterTemplates(all []Template, lang, framework string) []Template {
	if lang == "" {
		return all
	}

	var filtered []Template
	for _, t := range all {
		switch {
		case t.Lang == "":
			filtered = append(filtered, t)
		case t.Lang == lang && t.Framework == "":
			filtered = append(filtered, t)
		case t.Lang == lang && framework != "" && t.Framework == framework:
			filtered = append(filtered, t)
		}
	}
	return filtered
}

// ReadTemplateFile returns f's raw embedded content, within Template t.
func ReadTemplateFile(t Template, f TemplateFile) ([]byte, error) {
	data, err := embedded.FS.ReadFile(path.Join(t.dir, f.Source))
	if err != nil {
		return nil, fmt.Errorf("catalog: reading %s: %w", path.Join(t.dir, f.Source), err)
	}
	return data, nil
}
