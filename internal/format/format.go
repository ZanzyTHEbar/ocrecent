package format

import (
	"fmt"
	"strings"

	"github.com/ZanzyTHEbar/ocrecent/internal/model"
)

func RelTime(now, ms int64) string {
	d := now - ms
	if d < 0 {
		d = 0
	}
	switch {
	case d < 60_000:
		return "now"
	case d < 3_600_000:
		return fmt.Sprintf("%dm", d/60_000)
	case d < 86_400_000:
		return fmt.Sprintf("%dh", d/3_600_000)
	case d < 7*86_400_000:
		return fmt.Sprintf("%dd", d/86_400_000)
	case d < 30*86_400_000:
		return fmt.Sprintf("%dw", d/(7*86_400_000))
	case d < 365*86_400_000:
		return fmt.Sprintf("%dmo", d/(30*86_400_000))
	default:
		return fmt.Sprintf("%dy", d/(365*86_400_000))
	}
}

func HomePath(home, path string) string {
	if home != "" && (path == home || strings.HasPrefix(path, home+"/")) {
		return "~" + strings.TrimPrefix(path, home)
	}
	return path
}

// Print renders the resume command for a session, shell-quoted.
func Print(s model.Session) string {
	return "opencode " + ShellQuote(s.Directory) + " --session " + ShellQuote(s.ID)
}

func ShellQuote(s string) string {
	if s == "" {
		return "''"
	}
	unsafe := func(r rune) bool {
		return !(r == '/' || r == '-' || r == '_' || r == '.' || r == '@' || r == '+' ||
			r == ':' || r == '=' || r == ',' ||
			(r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9'))
	}
	if strings.IndexFunc(s, unsafe) < 0 {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
