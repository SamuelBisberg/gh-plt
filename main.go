package main

import (
	"os"

	"github.com/SamuelBisberg/gh-plt/cmd"
	"github.com/SamuelBisberg/gh-plt/pkg"
)

func main() {
	if err := cmd.NewRootCmd().Execute(); err != nil {
		pkg.NewTheme().Errorf("%s", err)
		os.Exit(1)
	}
}
