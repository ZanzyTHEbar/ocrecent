package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ZanzyTHEbar/ocrecent/internal/config"
	"github.com/ZanzyTHEbar/ocrecent/internal/core"
	"github.com/ZanzyTHEbar/ocrecent/internal/model"
)

type launchStore struct {
	sessions []model.Session
}

func (s launchStore) DBs([]string) ([]string, error) { return []string{"test.db"}, nil }
func (s launchStore) Load(string, model.Filter) ([]model.Session, error) {
	return s.sessions, nil
}
func (s launchStore) Schema(string) (model.Cols, error) { return nil, nil }

type launchPicker struct {
	id           string
	resolveCalls int
	pickCalls    int
}

func (p *launchPicker) Resolve(string) (string, error) {
	p.resolveCalls++
	return "/bin/fzf", nil
}
func (p *launchPicker) Pick(string, []string) (string, error) {
	p.pickCalls++
	return p.id, nil
}

type launchResumer struct {
	dirs []string
	ids  []string
}

func (r *launchResumer) Resume(dir, id string, _ bool) error {
	r.dirs = append(r.dirs, dir)
	r.ids = append(r.ids, id)
	return nil
}
func (r *launchResumer) Spawn([]string) error { return nil }

type launchNotifier struct{}

func (launchNotifier) Send(string, string, string) (string, error) { return "", nil }

func newLaunchTest(t *testing.T) (*bytes.Buffer, *launchPicker, *launchResumer, *core.App) {
	t.Helper()
	sessions := []model.Session{
		{ID: "ses_old", Title: "old", UpdatedMS: 100, Directory: "/work/old"},
		{ID: "ses_new", Title: "new", UpdatedMS: 200, Directory: "/work/new"},
	}
	picker := &launchPicker{id: "ses_new"}
	resumer := &launchResumer{}
	app := core.New(launchStore{sessions: sessions}, picker, launchNotifier{}, resumer)
	app.Now = func() int64 { return 300 }
	out := &bytes.Buffer{}
	return out, picker, resumer, app
}

func executeLaunchTest(t *testing.T, args ...string) (string, *launchPicker, *launchResumer) {
	t.Helper()
	out, picker, resumer, app := newLaunchTest(t)
	root := NewRoot(&CmdParams{
		Cfg:    config.Config{N: 8, Picker: "auto"},
		App:    app,
		Stdout: out,
		Stderr: &bytes.Buffer{},
		InTTY:  true,
	})
	root.SetArgs(args)
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	return out.String(), picker, resumer
}

func TestRootListsByDefaultInTTY(t *testing.T) {
	out, picker, resumer := executeLaunchTest(t)
	if !strings.Contains(out, "ses_new") {
		t.Fatalf("expected session list, got %q", out)
	}
	if picker.pickCalls != 0 || len(resumer.ids) != 0 {
		t.Fatalf("default root command launched a session: picker=%d resumes=%v", picker.pickCalls, resumer.ids)
	}
}

func TestLaunchFlagEnablesRootPicker(t *testing.T) {
	out, picker, resumer := executeLaunchTest(t, "--launch")
	if out != "" {
		t.Fatalf("launch mode should not print a list, got %q", out)
	}
	if picker.pickCalls != 1 || len(resumer.ids) != 1 || resumer.ids[0] != "ses_new" {
		t.Fatalf("launch mode did not resume picked session: picker=%d resumes=%v", picker.pickCalls, resumer.ids)
	}
}

func TestLastListsUnlessLaunchRequested(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantResume bool
	}{
		{name: "list", args: []string{"last"}},
		{name: "launch", args: []string{"last", "--launch"}, wantResume: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, _, resumer := executeLaunchTest(t, tt.args...)
			if tt.wantResume {
				if out != "" || len(resumer.ids) != 1 || resumer.ids[0] != "ses_new" {
					t.Fatalf("last should launch newest session: output=%q resumes=%v", out, resumer.ids)
				}
				return
			}
			if !strings.Contains(out, "ses_new") || len(resumer.ids) != 0 {
				t.Fatalf("last should list without launching: output=%q resumes=%v", out, resumer.ids)
			}
		})
	}
}

func TestPickListsUnlessLaunchRequested(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantResume bool
	}{
		{name: "list", args: []string{"pick"}},
		{name: "launch", args: []string{"pick", "--launch"}, wantResume: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, picker, resumer := executeLaunchTest(t, tt.args...)
			if tt.wantResume {
				if out != "" || picker.pickCalls != 1 || len(resumer.ids) != 1 {
					t.Fatalf("pick should launch selected session: output=%q picks=%d resumes=%v", out, picker.pickCalls, resumer.ids)
				}
				return
			}
			if !strings.Contains(out, "ses_new") || picker.pickCalls != 0 || len(resumer.ids) != 0 {
				t.Fatalf("pick should list without launching: output=%q picks=%d resumes=%v", out, picker.pickCalls, resumer.ids)
			}
		})
	}
}
