package format

import (
	"strings"
	"testing"

	"github.com/ZanzyTHEbar/ocrecent/internal/model"
)

func TestRelTime(t *testing.T) {
	now := int64(1_000_000_000_000)
	cases := []struct {
		ms   int64
		want string
	}{
		{now - 30_000, "now"},
		{now - 120_000, "2m"},
		{now - 3_600_000, "1h"},
		{now - 86_400_000, "1d"},
		{now - 7*86_400_000, "1w"},
		{now - 30*86_400_000, "1mo"},
		{now - 365*86_400_000, "1y"},
	}
	for _, c := range cases {
		if got := RelTime(now, c.ms); got != c.want {
			t.Errorf("RelTime(%d): want %q got %q", c.ms, c.want, got)
		}
	}
}

func TestHomePath(t *testing.T) {
	if got := HomePath("/home/u", "/home/u/src/x"); got != "~/src/x" {
		t.Errorf("got %q", got)
	}
	if got := HomePath("/home/u", "/opt/x"); got != "/opt/x" {
		t.Errorf("got %q", got)
	}
}

func TestShellQuote(t *testing.T) {
	cases := map[string]string{
		"/home/u/src/x": "/home/u/src/x",
		"/a b/c":        "'/a b/c'",
		"quote's":       `'quote'\''s'`,
		"-flag":         "-flag",
		"a=b:c,d@e+f/g": "a=b:c,d@e+f/g",
	}
	for in, want := range cases {
		if got := ShellQuote(in); got != want {
			t.Errorf("ShellQuote(%q): want %q got %q", in, want, got)
		}
	}
}

func TestPrint(t *testing.T) {
	s := model.Session{ID: "ses_1", Directory: "/home/u/src x"}
	if got, want := Print(s), "opencode '/home/u/src x' --session ses_1"; got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestJSONRows(t *testing.T) {
	out, err := JSONRows([]model.Session{{ID: "ses_1", UpdatedMS: 5}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `"kind":"session"`) || !strings.Contains(out, `"ses_1"`) {
		t.Fatalf("session rows malformed: %s", out)
	}
	out, err = JSONRows(nil, []model.Project{{Key: "/w", Name: "w", ID: "ses_9"}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `"kind":"project"`) || !strings.Contains(out, `"ses_9"`) {
		t.Fatalf("project rows malformed: %s", out)
	}
}
