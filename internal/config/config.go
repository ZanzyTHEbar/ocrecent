package config

import (
	"fmt"
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
	cfg, _ := loadFile(explicit)

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

func loadFile(explicit string) (Config, error) {
	cfg := Default()
	cfg.Path = configPath(explicit)
	data, err := os.ReadFile(cfg.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}
	apply(&cfg, string(data))
	return cfg, nil
}

func configPath(explicit string) string {
	if explicit != "" {
		return explicit
	}
	return filepath.Join(xdg.ConfigHome(), "ocrecent", "config.toml")
}

// Set validates and writes one supported config value. Environment overrides
// are intentionally not included, since they are not persisted configuration.
func Set(path, key, value string) error {
	cfg, err := loadFile(path)
	if err != nil {
		return fmt.Errorf("read config file %q: %w", cfg.Path, err)
	}
	if err := setValue(&cfg, key, value); err != nil {
		return err
	}
	return Write(cfg.Path, cfg)
}

// Write writes the supported config values in the flat format consumed by
// Load. Path identifies the file and Config.Path is not serialized.
func Write(path string, cfg Config) error {
	if path == "" {
		return fmt.Errorf("config path is empty")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	if err := os.WriteFile(path, []byte(format(cfg)), 0o644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}
	return nil
}

func setValue(cfg *Config, key, value string) error {
	switch key {
	case "n":
		n, err := strconv.Atoi(value)
		if err != nil || n < 0 {
			return fmt.Errorf("invalid value for n %q: expected a non-negative integer", value)
		}
		cfg.N = n
	case "picker":
		if err := validateString(key, value); err != nil {
			return err
		}
		cfg.Picker = value
	case "terminal":
		if err := validateString(key, value); err != nil {
			return err
		}
		cfg.Terminal = value
	case "children":
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid value for children %q: expected a boolean", value)
		}
		cfg.Children = b
	case "archived":
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid value for archived %q: expected a boolean", value)
		}
		cfg.Archived = b
	case "notify_urgency":
		if err := validateString(key, value); err != nil {
			return err
		}
		if value != "low" && value != "normal" && value != "critical" {
			return fmt.Errorf("invalid value for notify_urgency %q: expected low, normal, or critical", value)
		}
		cfg.NotifyUrgency = value
	default:
		return fmt.Errorf("unsupported config key %q (supported: n, picker, terminal, children, archived, notify_urgency)", key)
	}
	return nil
}

func validateString(key, value string) error {
	if value == "" || strings.TrimSpace(value) == "" {
		return fmt.Errorf("invalid value for %s: value must not be empty", key)
	}
	if strings.ContainsAny(value, "\r\n") {
		return fmt.Errorf("invalid value for %s: value must be a single line", key)
	}
	return nil
}

func format(cfg Config) string {
	return fmt.Sprintf("n = %d\npicker = %s\nterminal = %s\nchildren = %t\narchived = %t\nnotify_urgency = %s\n",
		cfg.N,
		strconv.Quote(cfg.Picker),
		strconv.Quote(cfg.Terminal),
		cfg.Children,
		cfg.Archived,
		strconv.Quote(cfg.NotifyUrgency),
	)
}

func apply(cfg *Config, text string) {
	for _, line := range strings.Split(text, "\n") {
		line = stripComment(line)
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = parseValue(val)
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

func stripComment(line string) string {
	var quote byte
	escaped := false
	for i := 0; i < len(line); i++ {
		c := line[i]
		if escaped {
			escaped = false
			continue
		}
		if quote != 0 && c == '\\' {
			escaped = true
			continue
		}
		if c == '"' || c == '\'' {
			if quote == 0 {
				quote = c
			} else if quote == c {
				quote = 0
			}
			continue
		}
		if quote == 0 && c == '#' {
			return line[:i]
		}
	}
	return line
}

func parseValue(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		if unquoted, err := strconv.Unquote(value); err == nil {
			return unquoted
		}
	}
	return strings.Trim(value, `"'`)
}
