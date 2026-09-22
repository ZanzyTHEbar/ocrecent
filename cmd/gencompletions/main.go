// Command gencompletions writes bash and zsh completions generated from the
// same cobra command table as --help.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ZanzyTHEbar/ocrecent/internal/cli"
	"github.com/ZanzyTHEbar/ocrecent/internal/config"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: gencompletions <outdir>")
		os.Exit(1)
	}
	out := os.Args[1]
	root := cli.NewRoot(&cli.CmdParams{Cfg: config.Default()})

	bashDir := filepath.Join(out, "bash")
	zshDir := filepath.Join(out, "zsh")
	for _, d := range []string{bashDir, zshDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	if err := root.GenBashCompletionFile(filepath.Join(bashDir, "ocrecent")); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := root.GenZshCompletionFile(filepath.Join(zshDir, "_ocrecent")); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
