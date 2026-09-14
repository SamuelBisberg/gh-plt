package pkg

import "testing"

func TestAllTemplatesFindsEmbeddedTemplates(t *testing.T) {
	all, err := AllTemplates()
	if err != nil {
		t.Fatalf("AllTemplates() error = %v", err)
	}

	want := map[string]struct {
		lang, framework, kind, icon string
		numFiles                    int
	}{
		"php/laravel/ci": {"php", "laravel", "ci", "🧪", 1},
		"js/ci":          {"js", "", "ci", "🧪", 1},
		"js/setup":       {"js", "", "setup", "⚙️", 2},
		"python/ci":      {"python", "", "ci", "🧪", 1},
		"release":        {"", "", "release", "🚀", 2},
	}

	if len(all) != len(want) {
		t.Fatalf("AllTemplates() returned %d templates, want %d: %+v", len(all), len(want), all)
	}

	for _, tpl := range all {
		w, ok := want[tpl.Path]
		if !ok {
			t.Errorf("unexpected template path %q", tpl.Path)
			continue
		}
		if tpl.Lang != w.lang || tpl.Framework != w.framework || tpl.Kind != w.kind || tpl.Icon != w.icon {
			t.Errorf("template %q = %+v, want lang=%s framework=%s kind=%s icon=%s", tpl.Path, tpl, w.lang, w.framework, w.kind, w.icon)
		}
		if len(tpl.Files) != w.numFiles {
			t.Errorf("template %q has %d files, want %d", tpl.Path, len(tpl.Files), w.numFiles)
		}
		if tpl.Description == "" {
			t.Errorf("template %q has no Description", tpl.Path)
		}
	}
}

func TestFindTemplateMultiFile(t *testing.T) {
	tpl, ok, err := FindTemplate("release")
	if err != nil {
		t.Fatalf("FindTemplate() error = %v", err)
	}
	if !ok {
		t.Fatal("FindTemplate(\"release\") not found")
	}
	if len(tpl.Files) != 2 {
		t.Fatalf("Files = %+v, want 2 entries", tpl.Files)
	}

	destinations := map[string]bool{}
	for _, f := range tpl.Files {
		destinations[f.Destination] = true

		raw, err := ReadTemplateFile(tpl, f)
		if err != nil {
			t.Fatalf("ReadTemplateFile(%q) error = %v", f.Source, err)
		}
		if len(raw) == 0 {
			t.Errorf("ReadTemplateFile(%q) returned empty content", f.Source)
		}
	}

	for _, want := range []string{".github/workflows/release.yml", "CHANGELOG.md"} {
		if !destinations[want] {
			t.Errorf("Files missing destination %q, got %+v", want, tpl.Files)
		}
	}
}

func TestFindTemplateSingleFile(t *testing.T) {
	tpl, ok, err := FindTemplate("php/laravel/ci")
	if err != nil {
		t.Fatalf("FindTemplate() error = %v", err)
	}
	if !ok {
		t.Fatal("FindTemplate(\"php/laravel/ci\") not found")
	}
	if len(tpl.Files) != 1 {
		t.Fatalf("Files = %+v, want 1 entry", tpl.Files)
	}

	f := tpl.Files[0]
	if len(f.Requires) == 0 {
		t.Error("Files[0].Requires is empty, want at least one dependency")
	}
	if len(f.Variables) == 0 {
		t.Error("Files[0].Variables is empty, want at least one variable")
	}

	raw, err := ReadTemplateFile(tpl, f)
	if err != nil {
		t.Fatalf("ReadTemplateFile() error = %v", err)
	}
	if len(raw) == 0 {
		t.Error("ReadTemplateFile() returned empty content")
	}
}

func TestFindTemplateWriteStrategies(t *testing.T) {
	tpl, ok, err := FindTemplate("js/setup")
	if err != nil {
		t.Fatalf("FindTemplate() error = %v", err)
	}
	if !ok {
		t.Fatal("FindTemplate(\"js/setup\") not found")
	}

	strategies := map[string]WriteStrategy{}
	for _, f := range tpl.Files {
		strategies[f.Destination] = f.Strategy
	}
	if strategies[".gitignore"] != StrategyAppend {
		t.Errorf("Files[.gitignore].Strategy = %q, want %q", strategies[".gitignore"], StrategyAppend)
	}
	if strategies["package.json"] != StrategyMerge {
		t.Errorf("Files[package.json].Strategy = %q, want %q", strategies["package.json"], StrategyMerge)
	}
}

func TestFilterTemplates(t *testing.T) {
	all, err := AllTemplates()
	if err != nil {
		t.Fatalf("AllTemplates() error = %v", err)
	}

	tests := []struct {
		name          string
		lang          string
		framework     string
		wantPaths     []string
		wantAllUnlike bool // want == all, no filtering
	}{
		{name: "no language filters nothing", lang: "", wantAllUnlike: true},
		{
			name:      "php with laravel",
			lang:      "php",
			framework: "laravel",
			wantPaths: []string{"php/laravel/ci", "release"},
		},
		{
			name:      "php without a framework excludes framework-specific templates",
			lang:      "php",
			framework: "",
			wantPaths: []string{"release"},
		},
		{
			name:      "js",
			lang:      "js",
			framework: "",
			wantPaths: []string{"js/ci", "js/setup", "release"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterTemplates(all, tt.lang, tt.framework)
			if tt.wantAllUnlike {
				if len(got) != len(all) {
					t.Errorf("FilterTemplates() returned %d templates, want all %d unfiltered", len(got), len(all))
				}
				return
			}

			gotPaths := map[string]bool{}
			for _, tpl := range got {
				gotPaths[tpl.Path] = true
			}
			if len(gotPaths) != len(tt.wantPaths) {
				t.Errorf("FilterTemplates() = %+v, want paths %v", got, tt.wantPaths)
			}
			for _, want := range tt.wantPaths {
				if !gotPaths[want] {
					t.Errorf("FilterTemplates() missing %q, got %+v", want, got)
				}
			}
		})
	}
}

func TestFindTemplateMissing(t *testing.T) {
	_, ok, err := FindTemplate("does/not/exist")
	if err != nil {
		t.Fatalf("FindTemplate() error = %v", err)
	}
	if ok {
		t.Fatal("FindTemplate(\"does/not/exist\") = found, want not found")
	}
}
