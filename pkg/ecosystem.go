package pkg

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	embedded "github.com/SamuelBisberg/gh-plt/templates"
)

const ecosystemCatalogFile = "ecosystems.yaml"

// Framework is a marker-detected specialization of a Language, e.g.
// "laravel" within "php".
type Framework struct {
	Name    string   `yaml:"name"`
	Markers []string `yaml:"markers"`
}

// PackageManager describes one way to install a dev dependency for a
// Language: Lockfile (if present in the project directory) signals this is
// the manager in use; Install is the argv template, with "{package}"
// substituted for the dependency name.
type PackageManager struct {
	Name     string   `yaml:"name"`
	Lockfile string   `yaml:"lockfile"`
	Default  bool     `yaml:"default"`
	Install  []string `yaml:"install"`
}

// Language is one entry in the ecosystem catalog.
type Language struct {
	Name            string           `yaml:"name"`
	Markers         []string         `yaml:"markers"`
	Frameworks      []Framework      `yaml:"frameworks"`
	PackageManagers []PackageManager `yaml:"package_managers"`
}

// DefaultPackageManager returns the package manager flagged `default: true`,
// or the first declared one if none is flagged.
func (l Language) DefaultPackageManager() (PackageManager, bool) {
	for _, pm := range l.PackageManagers {
		if pm.Default {
			return pm, true
		}
	}
	if len(l.PackageManagers) > 0 {
		return l.PackageManagers[0], true
	}
	return PackageManager{}, false
}

// PackageManager returns the named package manager, if this language
// declares one by that name.
func (l Language) PackageManager(name string) (PackageManager, bool) {
	for _, pm := range l.PackageManagers {
		if pm.Name == name {
			return pm, true
		}
	}
	return PackageManager{}, false
}

// Command returns the argv that installs pkg via this package manager.
func (pm PackageManager) Command(pkg string) []string {
	argv := make([]string, len(pm.Install))
	for i, part := range pm.Install {
		argv[i] = strings.ReplaceAll(part, "{package}", pkg)
	}
	return argv
}

// AllLanguages parses the embedded ecosystem catalog, sorted by Name for
// stable, deterministic detection order.
func AllLanguages() ([]Language, error) {
	data, err := embedded.FS.ReadFile(ecosystemCatalogFile)
	if err != nil {
		return nil, fmt.Errorf("ecosystem: reading %s: %w", ecosystemCatalogFile, err)
	}

	var catalog struct {
		Languages []Language `yaml:"languages"`
	}
	if err := yaml.Unmarshal(data, &catalog); err != nil {
		return nil, fmt.Errorf("ecosystem: parsing %s: %w", ecosystemCatalogFile, err)
	}

	languages := catalog.Languages
	sort.Slice(languages, func(i, j int) bool { return languages[i].Name < languages[j].Name })
	return languages, nil
}

// FindLanguage returns the named Language from the catalog.
func FindLanguage(name string) (Language, bool, error) {
	all, err := AllLanguages()
	if err != nil {
		return Language{}, false, err
	}
	for _, lang := range all {
		if lang.Name == name {
			return lang, true, nil
		}
	}
	return Language{}, false, nil
}

func markerExists(dir, name string) bool {
	_, err := os.Stat(filepath.Join(dir, name))
	return err == nil
}

func anyMarkerExists(dir string, markers []string) bool {
	for _, m := range markers {
		if markerExists(dir, m) {
			return true
		}
	}
	return false
}

// DetectIn returns the first Language in the catalog whose markers are
// present in dir, along with the matched Framework (if any of its markers
// are also present) and the PackageManager to use (the first one whose
// lockfile is present, falling back to the language's default).
func DetectIn(dir string) (lang Language, framework Framework, pm PackageManager, matched bool, err error) {
	all, err := AllLanguages()
	if err != nil {
		return Language{}, Framework{}, PackageManager{}, false, err
	}

	for _, l := range all {
		if !anyMarkerExists(dir, l.Markers) {
			continue
		}

		for _, fw := range l.Frameworks {
			if anyMarkerExists(dir, fw.Markers) {
				framework = fw
				break
			}
		}

		for _, candidate := range l.PackageManagers {
			if candidate.Lockfile != "" && markerExists(dir, candidate.Lockfile) {
				pm = candidate
				return l, framework, pm, true, nil
			}
		}
		if def, ok := l.DefaultPackageManager(); ok {
			pm = def
		}
		return l, framework, pm, true, nil
	}

	return Language{}, Framework{}, PackageManager{}, false, nil
}
