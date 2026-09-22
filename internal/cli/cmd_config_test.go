package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ZanzyTHEbar/ocrecent/internal/config"
)

func TestConfigSetWritesWithoutLaunching(t *testing.T) {
	for _, name := range []string{"OCRECENT_N", "OCRECENT_PICKER", "OCRECENT_TERMINAL"} {
		t.Setenv(name, "")
	}
	path := filepath.Join(t.TempDir(), "ocrecent", "config.toml")
	var out bytes.Buffer
	root := NewRoot(&CmdParams{
		Cfg:    config.Config{Path: path},
		Stdout: &out,
	})
	root.SetArgs([]string{"config", "set", "n", "4"})

	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if got := config.Load(path); got.N != 4 {
		t.Fatalf("config n = %d, want 4", got.N)
	}
	if !strings.Contains(out.String(), "wrote "+path) {
		t.Fatalf("command output = %q, want written path", out.String())
	}
}

func TestConfigSetReportsValidationError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	root := NewRoot(&CmdParams{
		Cfg:    config.Config{Path: path},
		Stdout: &bytes.Buffer{},
	})
	root.SetArgs([]string{"config", "set", "n", "not-a-number"})

	err := root.Execute()
	if err == nil || !strings.Contains(err.Error(), "invalid value for n") {
		t.Fatalf("error = %v, want n validation error", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("invalid command created config file: %v", err)
	}
}
