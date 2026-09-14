package pkg

import (
	"encoding/json"
	"testing"
)

func TestMergedContentNewFile(t *testing.T) {
	for _, strategy := range []WriteStrategy{StrategyOverride, StrategyAppend, StrategyMerge, ""} {
		got, err := mergedContent(nil, false, "hello", strategy)
		if err != nil {
			t.Fatalf("mergedContent(strategy=%q) error = %v", strategy, err)
		}
		if got != "hello" {
			t.Errorf("mergedContent(strategy=%q) = %q, want %q (a missing file just gets the rendered content)", strategy, got, "hello")
		}
	}
}

func TestMergedContentOverride(t *testing.T) {
	got, err := mergedContent([]byte("old"), true, "new", StrategyOverride)
	if err != nil {
		t.Fatalf("mergedContent() error = %v", err)
	}
	if got != "new" {
		t.Errorf("mergedContent() = %q, want %q", got, "new")
	}
}

func TestMergedContentAppend(t *testing.T) {
	tests := []struct {
		name     string
		existing string
		want     string
	}{
		{"trailing newline", "line one\n", "line one\nline two"},
		{"no trailing newline", "line one", "line one\nline two"},
		{"empty existing", "", "line two"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mergedContent([]byte(tt.existing), true, "line two", StrategyAppend)
			if err != nil {
				t.Fatalf("mergedContent() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("mergedContent() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMergedContentMerge(t *testing.T) {
	existing := `{"name": "app", "scripts": {"build": "old-build"}, "keep": true}`
	incoming := `{"scripts": {"test": "jest"}, "version": "1.0.0"}`

	got, err := mergedContent([]byte(existing), true, incoming, StrategyMerge)
	if err != nil {
		t.Fatalf("mergedContent() error = %v", err)
	}

	var merged map[string]any
	if err := json.Unmarshal([]byte(got), &merged); err != nil {
		t.Fatalf("merged output is not valid JSON: %v\n%s", err, got)
	}

	if merged["name"] != "app" {
		t.Errorf("merged[\"name\"] = %v, want existing value preserved", merged["name"])
	}
	if merged["keep"] != true {
		t.Errorf("merged[\"keep\"] = %v, want existing value preserved", merged["keep"])
	}
	if merged["version"] != "1.0.0" {
		t.Errorf("merged[\"version\"] = %v, want incoming value added", merged["version"])
	}
	scripts, ok := merged["scripts"].(map[string]any)
	if !ok {
		t.Fatalf("merged[\"scripts\"] = %v, want an object", merged["scripts"])
	}
	if scripts["build"] != "old-build" {
		t.Errorf("scripts[\"build\"] = %v, want existing nested value preserved", scripts["build"])
	}
	if scripts["test"] != "jest" {
		t.Errorf("scripts[\"test\"] = %v, want incoming nested value added", scripts["test"])
	}
}

func TestMergedContentMergeNonJSON(t *testing.T) {
	_, err := mergedContent([]byte("not json"), true, "also not json", StrategyMerge)
	if err == nil {
		t.Fatal("mergedContent() error = nil, want an error for non-JSON content")
	}
}

func TestMergedContentUnknownStrategy(t *testing.T) {
	_, err := mergedContent([]byte("old"), true, "new", WriteStrategy("bogus"))
	if err == nil {
		t.Fatal("mergedContent() error = nil, want an error for an unknown strategy")
	}
}

func TestWriteStrategyResolved(t *testing.T) {
	if got := WriteStrategy("").Resolved(); got != StrategyOverride {
		t.Errorf("Resolved() = %q, want %q", got, StrategyOverride)
	}
	if got := StrategyAppend.Resolved(); got != StrategyAppend {
		t.Errorf("Resolved() = %q, want %q", got, StrategyAppend)
	}
}
