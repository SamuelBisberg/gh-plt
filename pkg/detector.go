package pkg

// ProjectContext describes what kind of project was found in a directory.
type ProjectContext struct {
	Language       string // e.g. "php", "" if undetected
	Framework      string // e.g. "laravel", "" if generic/undetected
	PackageManager string // e.g. "uv", "" if undetected
}

// Detect scans dir for well-known project marker files, using the
// ecosystem catalog, and returns the best guess at its ProjectContext. A
// zero-value ProjectContext (empty Language) means nothing recognizable was
// found.
func Detect(dir string) ProjectContext {
	lang, framework, pm, ok, err := DetectIn(dir)
	if err != nil || !ok {
		return ProjectContext{}
	}
	return ProjectContext{
		Language:       lang.Name,
		Framework:      framework.Name,
		PackageManager: pm.Name,
	}
}
