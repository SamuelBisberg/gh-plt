package pkg

import (
	"os"
	"path/filepath"
	"testing"
)

func touchFile(t *testing.T, dir, name string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), nil, 0o644); err != nil {
		t.Fatalf("touch %s: %v", name, err)
	}
}

func TestAllLanguagesParsesEmbeddedCatalog(t *testing.T) {
	langs, err := AllLanguages()
	if err != nil {
		t.Fatalf("AllLanguages() error = %v", err)
	}

	names := map[string]bool{}
	for _, l := range langs {
		names[l.Name] = true
	}
	for _, want := range []string{"php", "js", "python"} {
		if !names[want] {
			t.Errorf("AllLanguages() missing language %q, got %+v", want, langs)
		}
	}
}

func TestDetectIn(t *testing.T) {
	tests := []struct {
		name          string
		files         []string
		wantLang      string
		wantFramework string
		wantPM        string
	}{
		{"laravel", []string{"composer.json", "artisan"}, "php", "laravel", "composer"},
		{"generic php", []string{"composer.json"}, "php", "", "composer"},
		{"js pnpm", []string{"package.json", "pnpm-lock.yaml"}, "js", "", "pnpm"},
		{"js default", []string{"package.json"}, "js", "", "npm"},
		{"python uv", []string{"pyproject.toml", "uv.lock"}, "python", "", "uv"},
		{"python default", []string{"pyproject.toml"}, "python", "", "pip"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for _, f := range tt.files {
				touchFile(t, dir, f)
			}

			lang, framework, pm, ok, err := DetectIn(dir)
			if err != nil {
				t.Fatalf("DetectIn() error = %v", err)
			}
			if !ok {
				t.Fatal("DetectIn() ok = false, want true")
			}
			if lang.Name != tt.wantLang || framework.Name != tt.wantFramework || pm.Name != tt.wantPM {
				t.Errorf("DetectIn() = (%q, %q, %q), want (%q, %q, %q)",
					lang.Name, framework.Name, pm.Name, tt.wantLang, tt.wantFramework, tt.wantPM)
			}
		})
	}
}

func TestDetectInUnrecognized(t *testing.T) {
	_, _, _, ok, err := DetectIn(t.TempDir())
	if err != nil {
		t.Fatalf("DetectIn() error = %v", err)
	}
	if ok {
		t.Fatal("DetectIn() ok = true for an empty directory, want false")
	}
}
