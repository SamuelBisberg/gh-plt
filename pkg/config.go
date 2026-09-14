package pkg

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Versioning selects how resolved GitHub Action dependencies are pinned.
type Versioning string

const (
	VersioningTag  Versioning = "tag"
	VersioningHash Versioning = "hash"
)

// Config holds all persisted gh-plt settings.
type Config struct {
	// Installers maps a language name (e.g. "python") to the preferred
	// package manager to install its dependencies with (e.g. "uv"). The set
	// of valid languages/package managers comes from the ecosystem catalog,
	// not from this struct.
	Installers map[string]string `yaml:"installers"`
	Actions    struct {
		Versioning Versioning `yaml:"versioning"`
	} `yaml:"actions"`
}

// defaultConfig builds a Config using each ecosystem language's default
// package manager.
func defaultConfig() *Config {
	cfg := &Config{Installers: map[string]string{}}
	cfg.Actions.Versioning = VersioningTag

	langs, err := AllLanguages()
	if err != nil {
		return cfg
	}
	for _, lang := range langs {
		if pm, ok := lang.DefaultPackageManager(); ok {
			cfg.Installers[lang.Name] = pm.Name
		}
	}
	return cfg
}

// ConfigPath returns the on-disk location of the config file, honoring
// XDG_CONFIG_HOME / os.UserConfigDir so it resolves to ~/.config/gh/plt.yaml
// on Linux and the platform-appropriate equivalent elsewhere.
func ConfigPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "gh", "plt.yaml"), nil
}

// LoadConfig reads the config file, returning defaults if none exists yet.
func LoadConfig() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return defaultConfig(), nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	cfg := defaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config at %s: %w", path, err)
	}
	return cfg, nil
}

// Save writes the config file, creating its parent directory if needed.
func (c *Config) Save() error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("encoding config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("writing config at %s: %w", path, err)
	}
	return nil
}
