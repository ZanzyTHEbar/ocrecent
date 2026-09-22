package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ZanzyTHEbar/ocrecent/internal/xdg"
)

type Config struct {
	Path          string
	N             int
	Picker        string
	Terminal      string
	Children      bool
	Archived      bool
	NotifyUrgency string
}

func Default() Config {
	return Config{
		N:             8,
		Picker:        "auto",
		Terminal:      "foot",
		NotifyUrgency: "normal",
	}
}

// Load reads a flat key = value config file, applies env overrides, and never
// fails on a missing or malformed file.
func Load(explicit string) Config {
	cfg := Default()
	if explicit != "" {
		cfg.Path = explicit
	} else {
		cfg.Path = filepath.Join(xdg.ConfigHome(), "ocrecent", "config.toml")
	}
	if data, err := os.ReadFile(cfg.Path); err == nil {
		apply(&cfg, string(data))
	}

	if v := os.Getenv("OCRECENT_N"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.N = n
		}
	}
	if v := os.Getenv("OCRECENT_PICKER"); v != "" {
		cfg.Picker = v
	}
	if v := os.Getenv("OCRECENT_TERMINAL"); v != "" {
		cfg.Terminal = v
	}
	return cfg
}

func apply(cfg *Config, text string) {
	for _, line := range strings.Split(text, "\n") {
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i]
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.Trim(strings.TrimSpace(val), `"'`)
		switch key {
		case "n":
			if n, err := strconv.Atoi(val); err == nil {
				cfg.N = n
			}
		case "picker":
			cfg.Picker = val
		case "terminal":
			cfg.Terminal = val
		case "children":
			if b, err := strconv.ParseBool(val); err == nil {
				cfg.Children = b
			}
		case "archived":
			if b, err := strconv.ParseBool(val); err == nil {
				cfg.Archived = b
			}
		case "notify_urgency":
			cfg.NotifyUrgency = val
		}
	}
}
