package pkg

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/huh"
)

// InstallMissing prompts the user, once per required package, to install it
// via the package manager appropriate for that requirement's language, then
// runs the installer with output streamed to the terminal. A requirement is
// skipped if its language doesn't match ctx's detected project language, or
// if that language/package-manager isn't in the ecosystem catalog.
func InstallMissing(requires []Requirement, ctx ProjectContext) error {
	for _, req := range requires {
		if ctx.Language != "" && req.Language != ctx.Language {
			continue
		}

		lang, ok, err := FindLanguage(req.Language)
		if err != nil {
			return fmt.Errorf("executor: %w", err)
		}
		if !ok {
			continue
		}

		pm, ok := lang.PackageManager(ctx.PackageManager)
		if !ok {
			pm, ok = lang.DefaultPackageManager()
			if !ok {
				continue
			}
		}
		argv := pm.Command(req.Package)

		install := false
		confirm := huh.NewConfirm().
			Title(fmt.Sprintf("This template requires %s. Run `%s` now?", req.Package, strings.Join(argv, " "))).
			Value(&install)
		if err := huh.NewForm(huh.NewGroup(confirm)).Run(); err != nil {
			return fmt.Errorf("executor: %w", err)
		}
		if !install {
			continue
		}

		cmd := exec.Command(argv[0], argv[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("executor: running %s: %w", strings.Join(argv, " "), err)
		}
	}
	return nil
}
