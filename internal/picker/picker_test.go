package picker

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ZanzyTHEbar/faults-go"

	"github.com/ZanzyTHEbar/ocrecent/internal/model"
)

func TestParseID(t *testing.T) {
	id, err := parseID("My session  2h  ~/src  [ses_abc123]")
	if err != nil || id != "ses_abc123" {
		t.Fatalf("got %q %v", id, err)
	}
	if _, err := parseID("no id here"); !faults.IsCode(err, model.CodePickerCancel) {
		t.Fatalf("want CodePickerCancel, got %v", err)
	}
}

func TestLine(t *testing.T) {
	got := Line("title", "2h", "~/src", "ses_1")
	want := "title  2h  ~/src  [ses_1]"
	if got != want {
		t.Fatalf("got %q", got)
	}
}

func TestPickSelectsAndFeedsStdin(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "fzf")
	record := filepath.Join(dir, "stdin.txt")
	os.WriteFile(script, []byte("#!/bin/sh\ncat > "+record+"\necho 'a line  [ses_zzz]'\n"), 0o755)

	lines := []string{"one  [ses_1]", "two  [ses_2]"}
	id, err := Picker{}.Pick(script, lines)
	if err != nil || id != "ses_zzz" {
		t.Fatalf("got %q %v", id, err)
	}
	data, _ := os.ReadFile(record)
	if string(data) != "one  [ses_1]\ntwo  [ses_2]" {
		t.Fatalf("stdin mismatch: %q", data)
	}
}

func TestPickCancel(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "rofi")
	os.WriteFile(script, []byte("#!/bin/sh\nexit 1\n"), 0o755)

	_, err := Picker{}.Pick(script, []string{"x"})
	if !faults.IsCode(err, model.CodePickerCancel) {
		t.Fatalf("want CodePickerCancel, got %v", err)
	}
}

func TestResolveAutoUsesEnv(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"fzf", "rofi"} {
		os.WriteFile(filepath.Join(dir, name), []byte("#!/bin/sh\n"), 0o755)
	}
	t.Setenv("PATH", dir)
	t.Setenv("WAYLAND_DISPLAY", "")
	t.Setenv("DISPLAY", "")
	bin, err := Picker{}.Resolve("auto")
	if err != nil || filepath.Base(bin) != "fzf" {
		t.Fatalf("no display should pick fzf, got %q %v", bin, err)
	}
	t.Setenv("DISPLAY", ":0")
	bin, err = Picker{}.Resolve("auto")
	if err != nil || filepath.Base(bin) != "rofi" {
		t.Fatalf("DISPLAY should pick rofi, got %q %v", bin, err)
	}
	bin, err = Picker{}.Resolve("custom-picker")
	if err != nil || bin != "custom-picker" {
		t.Fatalf("explicit should pass through, got %q %v", bin, err)
	}
}

func TestResolveAutoNoPicker(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	t.Setenv("WAYLAND_DISPLAY", "")
	t.Setenv("DISPLAY", "")
	_, err := Picker{}.Resolve("auto")
	if !faults.IsCode(err, CodePicker) {
		t.Fatalf("want CodePicker, got %v", err)
	}
}

func TestResolveExplicitNotInPath(t *testing.T) {
	bin, err := Picker{}.Resolve("definitely-not-a-picker")
	if err != nil || bin != "definitely-not-a-picker" {
		t.Fatalf("explicit resolves as-is, got %q %v", bin, err)
	}
}
