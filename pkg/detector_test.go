package pkg

import "testing"

func TestDetect(t *testing.T) {
	tests := []struct {
		name  string
		files []string
		want  ProjectContext
	}{
		{
			name:  "laravel",
			files: []string{"composer.json", "artisan"},
			want:  ProjectContext{Language: "php", Framework: "laravel", PackageManager: "composer"},
		},
		{
			name:  "generic php",
			files: []string{"composer.json"},
			want:  ProjectContext{Language: "php", PackageManager: "composer"},
		},
		{
			name:  "js with pnpm",
			files: []string{"package.json", "pnpm-lock.yaml"},
			want:  ProjectContext{Language: "js", PackageManager: "pnpm"},
		},
		{
			name:  "js with yarn",
			files: []string{"package.json", "yarn.lock"},
			want:  ProjectContext{Language: "js", PackageManager: "yarn"},
		},
		{
			name:  "js with npm",
			files: []string{"package.json", "package-lock.json"},
			want:  ProjectContext{Language: "js", PackageManager: "npm"},
		},
		{
			name:  "js with no lockfile",
			files: []string{"package.json"},
			want:  ProjectContext{Language: "js", PackageManager: "npm"},
		},
		{
			name:  "python with uv",
			files: []string{"pyproject.toml", "uv.lock"},
			want:  ProjectContext{Language: "python", PackageManager: "uv"},
		},
		{
			name:  "python with no lockfile",
			files: []string{"pyproject.toml"},
			want:  ProjectContext{Language: "python", PackageManager: "pip"},
		},
		{
			name:  "unrecognized",
			files: nil,
			want:  ProjectContext{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for _, f := range tt.files {
				touchFile(t, dir, f)
			}
			got := Detect(dir)
			if got != tt.want {
				t.Errorf("Detect() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
