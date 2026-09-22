package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAppliesConfigFile(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	t.Setenv("OCRECENT_N", "")
	t.Setenv("OCRECENT_PICKER", "")
	t.Setenv("OCRECENT_TERMINAL", "")

	path := filepath.Join(configHome, "ocrecent", "config.toml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(`
n = 3
picker = "rofi"
terminal = "kitty"
children = true
archived = true
notify_urgency = "critical"
`), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := Load("")
	if cfg.Path != path {
		t.Fatalf("config path = %q, want %q", cfg.Path, path)
	}
	if cfg.N != 3 || cfg.Picker != "rofi" || cfg.Terminal != "kitty" ||
		!cfg.Children || !cfg.Archived || cfg.NotifyUrgency != "critical" {
		t.Fatalf("config file values were not applied: %+v", cfg)
	}
}
