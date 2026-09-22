package core

import (
	"errors"
	"strings"
	"testing"

	"github.com/ZanzyTHEbar/faults-go"

	"github.com/ZanzyTHEbar/ocrecent/internal/model"
)

type fakeStore struct {
	paths []string
	dbs   map[string][]model.Session
	err   error
}

func (f fakeStore) DBs(extra []string) ([]string, error) { return f.paths, f.err }
func (f fakeStore) Load(path string, fl model.Filter) ([]model.Session, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.dbs[path], nil
}
func (f fakeStore) Schema(path string) (model.Cols, error) { return nil, nil }

type fakePicker struct {
	bin string
	id  string
	err error
}

func (f fakePicker) Resolve(explicit string) (string, error) { return f.bin, f.err }
func (f fakePicker) Pick(bin string, lines []string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return f.id, nil
}

type fakeNotifier struct{ action string }

func (f fakeNotifier) Send(title, body, urgency string) (string, error) {
	return f.action, nil
}

type captureNotifier struct {
	title   string
	body    string
	urgency string
}

func (n *captureNotifier) Send(title, body, urgency string) (string, error) {
	n.title = title
	n.body = body
	n.urgency = urgency
	return "", nil
}

type fakeResumer struct {
	dirs []string
	ids  []string
}

func (f *fakeResumer) Resume(dir, id string, inTTY bool) error {
	f.dirs = append(f.dirs, dir)
	f.ids = append(f.ids, id)
	return nil
}
func (f *fakeResumer) Spawn(argv []string) error { return nil }

func sess(id string, updated int64) model.Session {
	return model.Session{ID: id, UpdatedMS: updated, Directory: "/work/" + id}
}

func TestMergeTwoDBsSameIDNewerWins(t *testing.T) {
	store := fakeStore{
		paths: []string{"/a.db", "/b.db"},
		dbs: map[string][]model.Session{
			"/a.db": {sess("ses_x", 100), sess("ses_y", 50)},
			"/b.db": {sess("ses_x", 200)},
		},
	}
	app := New(store, fakePicker{}, fakeNotifier{}, &fakeResumer{})
	got, err := app.Sessions(model.Filter{}, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 merged sessions, got %d", len(got))
	}
	for _, s := range got {
		if s.ID == "ses_x" && s.UpdatedMS != 200 {
			t.Errorf("newer UpdatedMS should win, got %d", s.UpdatedMS)
		}
	}
}

func TestMergeSkipsUnreadableDB(t *testing.T) {
	store := pathErrStore{
		paths:   []string{"/a.db", "/bad.db"},
		dbs:     map[string][]model.Session{"/a.db": {sess("ses_x", 100)}},
		errPath: "/bad.db",
	}
	app := New(store, fakePicker{}, fakeNotifier{}, &fakeResumer{})
	got, err := app.Sessions(model.Filter{}, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("want 1 session, got %d", len(got))
	}
	if len(app.Warnings) != 1 {
		t.Fatalf("want 1 warning, got %v", app.Warnings)
	}
}

type pathErrStore struct {
	paths   []string
	dbs     map[string][]model.Session
	errPath string
}

func (f pathErrStore) DBs(extra []string) ([]string, error) { return f.paths, nil }
func (f pathErrStore) Load(path string, fl model.Filter) ([]model.Session, error) {
	if path == f.errPath {
		return nil, errors.New("unreadable")
	}
	return f.dbs[path], nil
}
func (f pathErrStore) Schema(path string) (model.Cols, error) { return nil, nil }

func TestResolveIndexAndID(t *testing.T) {
	store := fakeStore{
		paths: []string{"/a.db"},
		dbs: map[string][]model.Session{"/a.db": {
			sess("ses_b", 200),
			sess("ses_a", 100),
			sess("ses_old", 1),
		}},
	}
	app := New(store, fakePicker{}, fakeNotifier{}, &fakeResumer{})

	s, err := app.Resolve("1", model.Filter{}, 2)
	if err != nil || s.ID != "ses_b" {
		t.Fatalf("index 1 should be newest: got %v, %v", s.ID, err)
	}
	s, err = app.Resolve("ses_old", model.Filter{}, 2)
	if err != nil || s.ID != "ses_old" {
		t.Fatalf("old id lookup failed: got %v, %v", s.ID, err)
	}
	s, err = app.Resolve("ses_a", model.Filter{}, 2)
	if err != nil || s.ID != "ses_a" {
		t.Fatalf("id lookup failed: got %v, %v", s.ID, err)
	}
	_, err = app.Resolve("3", model.Filter{}, 2)
	if !faults.IsCode(err, model.CodeNotFound) {
		t.Fatalf("out of range should be CodeNotFound, got %v", err)
	}
	_, err = app.Resolve("ses_zzz", model.Filter{}, 10)
	if !faults.IsCode(err, model.CodeNotFound) {
		t.Fatalf("unknown id should be CodeNotFound, got %v", err)
	}
}

func TestProjectsLimitAfterCollapse(t *testing.T) {
	store := fakeStore{
		paths: []string{"/a.db"},
		dbs: map[string][]model.Session{"/a.db": {
			{ID: "ses_a1", UpdatedMS: 300, Directory: "/work/a", Worktree: "/work/a"},
			{ID: "ses_a2", UpdatedMS: 200, Directory: "/work/a/sub", Worktree: "/work/a"},
			{ID: "ses_b1", UpdatedMS: 100, Directory: "/work/b", Worktree: "/work/b"},
		}},
	}
	app := New(store, fakePicker{}, fakeNotifier{}, &fakeResumer{})
	projects, err := app.Projects(model.Filter{}, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 2 || projects[0].Key != "/work/a" || projects[1].Key != "/work/b" {
		t.Fatalf("projects were limited before collapse: %+v", projects)
	}
}

func TestResumeLastPicksNewest(t *testing.T) {
	store := fakeStore{
		paths: []string{"/a.db"},
		dbs:   map[string][]model.Session{"/a.db": {sess("ses_old", 1), sess("ses_new", 2)}},
	}
	r := &fakeResumer{}
	app := New(store, fakePicker{}, fakeNotifier{}, r)
	if err := app.ResumeLast(model.Filter{}, false); err != nil {
		t.Fatal(err)
	}
	if len(r.ids) != 1 || r.ids[0] != "ses_new" || r.dirs[0] != "/work/ses_new" {
		t.Fatalf("want ses_new /work/ses_new, got %v %v", r.ids, r.dirs)
	}
}

func TestPickProjectsGrain(t *testing.T) {
	store := fakeStore{
		paths: []string{"/a.db"},
		dbs: map[string][]model.Session{
			"/a.db": {
				{ID: "ses_1", UpdatedMS: 300, Title: "t1", Directory: "/w/proj", Worktree: "/w/proj", ProjectName: "proj"},
				{ID: "ses_2", UpdatedMS: 200, Title: "t2", Directory: "/w/proj/sub", Worktree: "/w/proj"},
			},
		},
	}
	p := fakePicker{bin: "/bin/fzf", id: "ses_1"}
	r := &fakeResumer{}
	app := New(store, p, fakeNotifier{}, r)
	if err := app.Pick(model.Filter{}, 10, "auto", true, false); err != nil {
		t.Fatal(err)
	}
	if len(r.ids) != 1 || r.ids[0] != "ses_1" {
		t.Fatalf("want ses_1 resumed, got %v", r.ids)
	}
}

func TestPickEmptyStore(t *testing.T) {
	store := fakeStore{paths: []string{"/a.db"}, dbs: map[string][]model.Session{}}
	app := New(store, fakePicker{}, fakeNotifier{}, &fakeResumer{})
	err := app.Pick(model.Filter{}, 10, "auto", false, false)
	if !faults.IsCode(err, model.CodeNotFound) {
		t.Fatalf("want CodeNotFound, got %v", err)
	}
}

func TestNotifyUsesCompactProjectSummary(t *testing.T) {
	store := fakeStore{
		paths: []string{"/a.db"},
		dbs: map[string][]model.Session{"/a.db": {
			{ID: "ses_a1", UpdatedMS: 300, Directory: "/work/a", Worktree: "/work/a", ProjectName: "alpha"},
			{ID: "ses_a2", UpdatedMS: 200, Directory: "/work/a/sub", Worktree: "/work/a", ProjectName: "alpha"},
			{ID: "ses_b1", UpdatedMS: 100, Directory: "/work/b", Worktree: "/work/b", ProjectName: "beta"},
		}},
	}
	n := &captureNotifier{}
	app := New(store, fakePicker{}, n, &fakeResumer{})
	app.Now = func() int64 { return 300 }
	if _, err := app.Notify(model.Filter{}, 10, "normal"); err != nil {
		t.Fatal(err)
	}
	if n.title != "OpenCode · 2 recent project(s)" || n.urgency != "normal" {
		t.Fatalf("unexpected notification metadata: %+v", n)
	}
	if n.body != "now  alpha  (2 sessions)\nnow  beta  (1 session)" {
		t.Fatalf("unexpected notification body: %q", n.body)
	}
	if strings.Contains(n.body, "/work/") {
		t.Fatalf("notification should not dump paths: %q", n.body)
	}
}

func TestDispatchNotifyLast(t *testing.T) {
	store := fakeStore{
		paths: []string{"/a.db"},
		dbs:   map[string][]model.Session{"/a.db": {sess("ses_x", 42)}},
	}
	r := &fakeResumer{}
	app := New(store, fakePicker{}, fakeNotifier{}, r)
	if err := app.DispatchNotify("last", model.Filter{}, 8, "auto"); err != nil {
		t.Fatal(err)
	}
	if len(r.ids) != 1 || r.ids[0] != "ses_x" {
		t.Fatalf("want ses_x resumed, got %v", r.ids)
	}
}
