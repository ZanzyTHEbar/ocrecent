package store

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ZanzyTHEbar/faults-go"

	"github.com/ZanzyTHEbar/ocrecent/internal/model"
	"github.com/ZanzyTHEbar/ocrecent/internal/xdg"
)

// Discover resolves the candidate database paths: $OPENCODE_DATA/opencode.db,
// else `opencode db path`, else the XDG default; plus sibling opencode-*.db
// files and any extra paths. Only existing regular files are returned.
func Discover(extra []string) ([]string, error) {
	var primary string
	var dir string
	if v, ok := xdg.OpenCodeData(); ok {
		dir = v
		primary = filepath.Join(v, "opencode.db")
	} else if p := opencodeDBPath(); p != "" {
		dir = filepath.Dir(p)
		primary = p
	} else {
		primary = xdg.DefaultDB()
		dir = filepath.Dir(primary)
	}

	paths := []string{primary}
	if dir != "" {
		if matches, err := filepath.Glob(filepath.Join(dir, "opencode-*.db")); err == nil {
			paths = append(paths, matches...)
		}
	}
	paths = append(paths, extra...)

	seen := map[string]bool{}
	var out []string
	for _, p := range paths {
		p = filepath.Clean(p)
		if p == "" || seen[p] {
			continue
		}
		seen[p] = true
		if info, err := os.Stat(p); err == nil && info.Mode().IsRegular() {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return nil, faults.New(model.CodeNoDB, "no opencode databases found",
			"paths", strings.Join(paths, ", "))
	}
	return out, nil
}

func opencodeDBPath() string {
	path, err := exec.LookPath("opencode")
	if err != nil {
		return ""
	}
	out, err := exec.Command(path, "db", "path").Output()
	if err != nil {
		return ""
	}
	line, _, _ := strings.Cut(strings.TrimSpace(string(out)), "\n")
	return line
}
