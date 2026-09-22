package term

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ZanzyTHEbar/faults-go"
)

var CodeTerminal = faults.Code("ocrecent.terminal")

var known = []string{"foot", "kitty", "alacritty", "ghostty", "wezterm"}

// Detect resolves the configured terminal (or "auto") to an absolute binary
// path, since callers exec the result directly.
func Detect(configured string) (string, error) {
	if configured != "" && configured != "auto" {
		p, err := exec.LookPath(configured)
		if err != nil {
			return "", faults.New(CodeTerminal, "terminal not found", "terminal", configured)
		}
		return p, nil
	}
	for _, name := range known {
		if p, err := exec.LookPath(name); err == nil {
			return p, nil
		}
	}
	return "", faults.New(CodeTerminal, "no terminal found", "candidates", known)
}

// Command builds the argv for spawning argv inside a terminal binary.
func Command(termBin string, argv []string) []string {
	switch filepath.Base(termBin) {
	case "alacritty", "ghostty":
		return append([]string{termBin, "-e"}, argv...)
	case "wezterm":
		return append([]string{termBin, "start", "--"}, argv...)
	default: // foot, kitty, and most others take the command directly
		return append([]string{termBin}, argv...)
	}
}

func IsTTY(f *os.File) bool {
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
