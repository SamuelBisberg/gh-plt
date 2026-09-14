package pkg

import (
	"path/filepath"
	"testing"
)

func TestLoadConfigDefaultsWhenMissing(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if cfg.Actions.Versioning != VersioningTag {
		t.Errorf("Actions.Versioning = %q, want %q", cfg.Actions.Versioning, VersioningTag)
	}
	want := map[string]string{"php": "composer", "js": "npm", "python": "pip"}
	for lang, installer := range want {
		if got := cfg.Installers[lang]; got != installer {
			t.Errorf("Installers[%q] = %q, want %q", lang, got, installer)
		}
	}
}

func TestConfigSaveThenLoadRoundTrips(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	cfg.Installers["python"] = "uv"
	cfg.Actions.Versioning = VersioningHash
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	path, err := ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath() error = %v", err)
	}
	if filepath.Base(path) != "plt.yaml" {
		t.Errorf("ConfigPath() = %q, want a plt.yaml file", path)
	}

	got, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if got.Installers["python"] != "uv" || got.Actions.Versioning != VersioningHash {
		t.Errorf("LoadConfig() after Save() = %+v, want the saved values", got)
	}
}
