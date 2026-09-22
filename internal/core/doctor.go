package core

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ZanzyTHEbar/faults-go"

	"github.com/ZanzyTHEbar/ocrecent/internal/model"
	"github.com/ZanzyTHEbar/ocrecent/internal/term"
	"github.com/ZanzyTHEbar/ocrecent/internal/xdg"
)

// Install writes the user systemd units and XDG autostart entry, returning
// the paths written. It never enables timers.
func (a *App) Install() ([]string, error) {
	exe, err := os.Executable()
	if err != nil {
		return nil, faults.Wrap(model.CodeNotFound, "locate own binary", err)
	}
	systemdDir := filepath.Join(xdg.ConfigHome(), "systemd", "user")
	autostartDir := filepath.Join(xdg.ConfigHome(), "autostart")

	service, desktop := launcherContents(exe)
	timer := `[Unit]
Description=Notify recent OpenCode sessions after login

[Timer]
OnStartupSec=8
Unit=ocrecent-notify.service

[Install]
WantedBy=timers.target
`

	paths := map[string]string{
		filepath.Join(systemdDir, "ocrecent-notify.service"): service,
		filepath.Join(systemdDir, "ocrecent-notify.timer"):   timer,
		filepath.Join(autostartDir, "ocrecent.desktop"):      desktop,
	}
	written := make([]string, 0, len(paths))
	for path, content := range paths {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return written, faults.Wrap(model.CodeUsage, "create config dir", err, "path", path)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return written, faults.Wrap(model.CodeUsage, "write unit", err, "path", path)
		}
		written = append(written, path)
	}
	sort.Strings(written)
	return written, nil
}

func launcherContents(exe string) (service, desktop string) {
	service = `[Unit]
Description=OpenCode recent session notification

[Service]
Type=oneshot
ExecStart=` + systemdQuote(exe) + ` notify
`
	desktop = `[Desktop Entry]
Type=Application
Name=ocrecent
Comment=Show recent OpenCode sessions
Exec=` + desktopQuote(exe) + ` notify
X-GNOME-Autostart-enabled=true
`
	return service, desktop
}

func systemdQuote(value string) string {
	return `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(value) + `"`
}

func desktopQuote(value string) string {
	return `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`, "`", "\x5c`", "$", "\x5c$").Replace(value) + `"`
}

// Uninstall removes the files written by Install.
func (a *App) Uninstall() ([]string, error) {
	paths := []string{
		filepath.Join(xdg.ConfigHome(), "systemd", "user", "ocrecent-notify.service"),
		filepath.Join(xdg.ConfigHome(), "systemd", "user", "ocrecent-notify.timer"),
		filepath.Join(xdg.ConfigHome(), "autostart", "ocrecent.desktop"),
	}
	var removed []string
	for _, path := range paths {
		if err := os.Remove(path); err == nil {
			removed = append(removed, path)
		} else if !os.IsNotExist(err) {
			return removed, faults.Wrap(model.CodeUsage, "remove unit", err, "path", path)
		}
	}
	return removed, nil
}

// Doctor gathers the support surface: version, config, opencode presence,
// every discovered DB with its column set, picker, terminal, and counts.
func (a *App) Doctor(f model.Filter, pickerExplicit, version string) (model.DoctorInfo, error) {
	info := model.DoctorInfo{
		Version: version,
		Schema:  map[string][]string{},
	}
	if p, err := exec.LookPath("opencode"); err == nil {
		info.OpenCodePath = p
		if out, err := exec.Command(p, "db", "path").Output(); err == nil {
			info.OpenCodeDBPath = trimLine(string(out))
		}
	}

	paths, err := a.Store.DBs(f.Extra)
	if err == nil {
		info.DBs = paths
		for _, path := range paths {
			cols, err := a.Store.Schema(path)
			if err != nil {
				info.Warnings = append(info.Warnings, path+": "+err.Error())
				continue
			}
			names := make([]string, 0, len(cols))
			for name := range cols {
				names = append(names, name)
			}
			sort.Strings(names)
			info.Schema[path] = names
		}
	} else {
		info.Warnings = append(info.Warnings, err.Error())
	}

	root, rerr := a.Load(f)
	if rerr == nil {
		info.RootCount = len(root)
	} else {
		info.Warnings = append(info.Warnings, rerr.Error())
	}
	info.Warnings = append(info.Warnings, a.Warnings...)

	everything, aerr := a.Load(model.Filter{Children: true, Archived: true, Extra: f.Extra})
	if aerr == nil {
		info.AllCount = len(everything)
	} else {
		info.Warnings = append(info.Warnings, aerr.Error())
	}
	info.Warnings = append(info.Warnings, a.Warnings...)

	if bin, err := a.Picker.Resolve(pickerExplicit); err == nil {
		info.Picker = bin
	} else {
		info.Picker = "unavailable: " + err.Error()
	}
	if bin, err := term.Detect("auto"); err == nil {
		info.Terminal = bin
	} else {
		info.Terminal = "unavailable"
	}
	return info, nil
}

func trimLine(s string) string {
	for i, r := range s {
		if r == '\n' {
			return s[:i]
		}
	}
	return s
}
