package config

import (
	"os"
	"path/filepath"
	"strings"
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

func TestWriteRoundTripsQuotedValues(t *testing.T) {
	for _, name := range []string{"OCRECENT_N", "OCRECENT_PICKER", "OCRECENT_TERMINAL"} {
		t.Setenv(name, "")
	}
	path := filepath.Join(t.TempDir(), "ocrecent", "config.toml")
	want := Config{
		N:             4,
		Picker:        "picker#one",
		Terminal:      `terminal "one"`,
		Children:      true,
		Archived:      true,
		NotifyUrgency: "low",
	}
	if err := Write(path, want); err != nil {
		t.Fatal(err)
	}

	got := Load(path)
	if got.N != want.N || got.Picker != want.Picker || got.Terminal != want.Terminal ||
		got.Children != want.Children || got.Archived != want.Archived || got.NotifyUrgency != want.NotifyUrgency {
		t.Fatalf("written config did not round trip: got %+v, want %+v", got, want)
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(contents), `picker = "picker#one"`) {
		t.Fatalf("written config did not use quoted values: %q", contents)
	}
}

func TestSetValidatesAndWritesValue(t *testing.T) {
	for _, test := range []struct {
		key   string
		value string
		check func(Config) bool
	}{
		{"n", "0", func(cfg Config) bool { return cfg.N == 0 }},
		{"children", "true", func(cfg Config) bool { return cfg.Children }},
		{"archived", "true", func(cfg Config) bool { return cfg.Archived }},
		{"picker", "fzf", func(cfg Config) bool { return cfg.Picker == "fzf" }},
		{"terminal", "kitty", func(cfg Config) bool { return cfg.Terminal == "kitty" }},
		{"notify_urgency", "critical", func(cfg Config) bool { return cfg.NotifyUrgency == "critical" }},
	} {
		t.Run(test.key, func(t *testing.T) {
			for _, name := range []string{"OCRECENT_N", "OCRECENT_PICKER", "OCRECENT_TERMINAL"} {
				t.Setenv(name, "")
			}
			path := filepath.Join(t.TempDir(), "config.toml")
			if err := Set(path, test.key, test.value); err != nil {
				t.Fatal(err)
			}
			if cfg := Load(path); !test.check(cfg) {
				t.Fatalf("config value was not written: %+v", cfg)
			}
		})
	}
}

func TestSetRejectsUnsupportedValue(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	for _, test := range []struct {
		key   string
		value string
	}{
		{"unknown", "value"},
		{"n", "-1"},
		{"children", "maybe"},
		{"notify_urgency", "urgent"},
	} {
		t.Run(test.key+"/"+test.value, func(t *testing.T) {
			if err := Set(path, test.key, test.value); err == nil {
				t.Fatalf("Set(%q, %q) succeeded", test.key, test.value)
			}
		})
	}
}
