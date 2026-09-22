package picker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ZanzyTHEbar/faults-go"

	"github.com/ZanzyTHEbar/ocrecent/internal/model"
)

var CodePicker = faults.Code("ocrecent.picker.unavailable")

var waylandCands = []string{"fuzzel", "wofi", "rofi", "tofi", "bemenu", "fzf"}
var x11Cands = []string{"rofi", "dmenu", "fzf"}

var idRe = regexp.MustCompile(`\[(ses_[^\]]+)\]\s*$`)

var args = map[string][]string{
	"fuzzel": {"-d"},
	"wofi":   {"-d"},
	"rofi":   {"-dmenu"},
	"tofi":   {},
	"bemenu": {},
	"dmenu":  {},
	"fzf":    {},
}

// Picker adapts the package to core.PickerPort.
type Picker struct{}

// Resolve maps an explicit picker name or "auto" to a binary path.
func (Picker) Resolve(explicit string) (string, error) {
	if explicit != "" && explicit != "auto" {
		return explicit, nil
	}
	var cands []string
	switch {
	case os.Getenv("WAYLAND_DISPLAY") != "":
		cands = waylandCands
	case os.Getenv("DISPLAY") != "":
		cands = x11Cands
	default:
		cands = []string{"fzf"}
	}
	for _, name := range cands {
		if p, err := exec.LookPath(name); err == nil {
			return p, nil
		}
	}
	return "", faults.New(CodePicker, "no picker found", "candidates", strings.Join(cands, ", "))
}

// Pick feeds lines to the picker binary and returns the selected session id.
// A cancelled or unrecognized selection is CodePickerCancel.
func (Picker) Pick(bin string, lines []string) (string, error) {
	argv := append([]string{}, args[filepath.Base(bin)]...)
	cmd := exec.Command(bin, argv...)
	cmd.Stdin = strings.NewReader(strings.Join(lines, "\n"))
	out, err := cmd.Output()
	if err != nil {
		if len(out) == 0 {
			return "", faults.New(model.CodePickerCancel, "picker cancelled")
		}
	}
	return parseID(string(out))
}

func parseID(line string) (string, error) {
	m := idRe.FindStringSubmatch(line)
	if m == nil {
		return "", faults.New(model.CodePickerCancel, "picker returned unrecognized line",
			"line", strings.TrimSpace(line))
	}
	return m[1], nil
}

// Line renders one picker entry; the session id is the trailing [ses_…] token.
func Line(name, reltime, path, id string) string {
	return fmt.Sprintf("%s  %s  %s  [%s]", name, reltime, path, id)
}
