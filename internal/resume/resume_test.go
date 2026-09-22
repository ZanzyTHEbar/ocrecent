package resume

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestResumeArgv runs itself as a child process; the child unix.Exec's the
// fake opencode script, which records its argv. The parent then asserts the
// exact command line: opencode <dir> --session <id>.
func TestResumeArgv(t *testing.T) {
	if os.Getenv("OCRECENT_RESUME_HELPER") == "1" {
		r := Resumer{Terminal: "foot"}
		if err := r.Resume("/tmp/work dir", "ses_abc", true); err != nil {
			os.WriteFile(os.Getenv("OCRECENT_RECORD"), []byte("ERR: "+err.Error()), 0o644)
		}
		return
	}

	dir := t.TempDir()
	record := filepath.Join(dir, "argv.txt")
	script := filepath.Join(dir, "opencode")
	os.WriteFile(script, []byte("#!/bin/sh\necho \"$@\" > \"$OCRECENT_RECORD\"\n"), 0o755)

	cmd := exec.Command(os.Args[0], "-test.run=TestResumeArgv")
	cmd.Env = append(os.Environ(),
		"OCRECENT_RESUME_HELPER=1",
		"OCRECENT_RECORD="+record,
		"PATH="+dir+":"+os.Getenv("PATH"),
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("child failed: %v\n%s", err, out)
	}

	data, err := os.ReadFile(record)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := strings.TrimSpace(string(data)), "/tmp/work dir --session ses_abc"; got != want {
		t.Fatalf("argv = %q, want %q", got, want)
	}
}

// TestSpawnTerminalArgv runs itself as a child that must replace itself with
// a fake terminal binary; the fake records its argv.
func TestSpawnTerminalArgv(t *testing.T) {
	if os.Getenv("OCRECENT_SPAWN_HELPER") == "1" {
		r := Resumer{Terminal: "foot"}
		if err := r.Spawn([]string{"opencode", "projects", "--all"}); err != nil {
			os.WriteFile(os.Getenv("OCRECENT_RECORD"), []byte("ERR: "+err.Error()), 0o644)
		}
		return
	}

	dir := t.TempDir()
	record := filepath.Join(dir, "argv.txt")
	for _, name := range []string{"foot", "opencode"} {
		os.WriteFile(filepath.Join(dir, name),
			[]byte("#!/bin/sh\necho \"$@\" > \"$OCRECENT_RECORD\"\n"), 0o755)
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestSpawnTerminalArgv")
	cmd.Env = append(os.Environ(),
		"OCRECENT_SPAWN_HELPER=1",
		"OCRECENT_RECORD="+record,
		"PATH="+dir+":"+os.Getenv("PATH"),
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("child failed: %v\n%s", err, out)
	}

	data, err := os.ReadFile(record)
	if err != nil {
		t.Fatal(err)
	}
	// foot passes the command directly.
	fields := strings.Fields(strings.TrimSpace(string(data)))
	if len(fields) != 3 || filepath.Base(fields[0]) != "opencode" ||
		fields[1] != "projects" || fields[2] != "--all" {
		t.Fatalf("argv = %q, want opencode projects --all", data)
	}
}
