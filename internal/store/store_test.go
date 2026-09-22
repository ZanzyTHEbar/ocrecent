package store_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/ZanzyTHEbar/faults-go"

	"github.com/ZanzyTHEbar/ocrecent/internal/model"
	"github.com/ZanzyTHEbar/ocrecent/internal/store"
)

const fullSchema = `
CREATE TABLE session (
  id TEXT PRIMARY KEY,
  parent_id TEXT,
  title TEXT,
  time_created INTEGER,
  time_updated INTEGER,
  time_archived INTEGER,
  directory TEXT,
  path TEXT,
  project_id TEXT
);
CREATE TABLE project (
  id TEXT PRIMARY KEY,
  name TEXT,
  worktree TEXT
);`

var fullRows = []string{
	`INSERT INTO session (id, parent_id, title, time_created, time_updated, time_archived, directory, path, project_id)
	 VALUES ('ses_root1', NULL, 'api root', 100, 300, 0, '/home/u/src/api', '', 'p1')`,
	`INSERT INTO session (id, parent_id, title, time_created, time_updated, time_archived, directory, path, project_id)
	 VALUES ('ses_child1', NULL, 'api child', 100, 200, 0, '/home/u/src/api/sub', 'sub', 'p1')`,
	`INSERT INTO session (id, parent_id, title, time_created, time_updated, time_archived, directory, path)
	 VALUES ('ses_agent1', 'ses_root1', 'subagent', 100, 100, 0, '/home/u/src/api', '')`,
	`INSERT INTO session (id, parent_id, title, time_created, time_updated, time_archived, directory, path)
	 VALUES ('ses_archived1', NULL, 'old', 100, 150, 50, '/home/u/src/old', '')`,
	`INSERT INTO session (id, parent_id, title, time_created, time_updated, time_archived, directory, path)
	 VALUES ('ses_root2', NULL, 'web', 100, 250, 0, '/home/u/src/web', '')`,
	`INSERT INTO project (id, name, worktree) VALUES ('p1', 'api', '/home/u/src/api')`,
}

func seed(t *testing.T, schema string, inserts []string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "opencode.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if schema != "" {
		if _, err := db.Exec(schema); err != nil {
			t.Fatal(err)
		}
	}
	for _, ins := range inserts {
		if _, err := db.Exec(ins); err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func openRO(t *testing.T, path string) *sql.DB {
	t.Helper()
	db, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func ids(t *testing.T, sessions []model.Session) []string {
	t.Helper()
	out := make([]string, len(sessions))
	for i, s := range sessions {
		out[i] = s.ID
	}
	sort.Strings(out)
	return out
}

func expectIDs(t *testing.T, got []model.Session, want ...string) {
	t.Helper()
	g, w := ids(t, got), want
	sort.Strings(w)
	if strings.Join(g, ",") != strings.Join(w, ",") {
		t.Fatalf("got %v, want %v", g, w)
	}
}

func TestListDefaultFilters(t *testing.T) {
	db := openRO(t, seed(t, fullSchema, fullRows))
	sessions, err := store.List(db, model.Filter{})
	if err != nil {
		t.Fatal(err)
	}
	expectIDs(t, sessions, "ses_root1", "ses_child1", "ses_root2")
}

func TestListChildren(t *testing.T) {
	db := openRO(t, seed(t, fullSchema, fullRows))
	sessions, err := store.List(db, model.Filter{Children: true})
	if err != nil {
		t.Fatal(err)
	}
	expectIDs(t, sessions, "ses_root1", "ses_child1", "ses_agent1", "ses_root2")
}

func TestListArchived(t *testing.T) {
	db := openRO(t, seed(t, fullSchema, fullRows))
	sessions, err := store.List(db, model.Filter{Archived: true})
	if err != nil {
		t.Fatal(err)
	}
	expectIDs(t, sessions, "ses_root1", "ses_child1", "ses_root2", "ses_archived1")
}

func TestListDirFilter(t *testing.T) {
	db := openRO(t, seed(t, fullSchema, fullRows))
	sessions, err := store.List(db, model.Filter{Dir: "/home/u/src/api"})
	if err != nil {
		t.Fatal(err)
	}
	expectIDs(t, sessions, "ses_root1", "ses_child1")

	sessions, err = store.List(db, model.Filter{Dir: "/home/u/src/api/sub"})
	if err != nil {
		t.Fatal(err)
	}
	expectIDs(t, sessions, "ses_child1")

	sessions, err = store.List(db, model.Filter{Dir: "/home/u/src/api2"})
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 0 {
		t.Fatalf("got %v, want none", ids(t, sessions))
	}
}

func TestListProjectJoin(t *testing.T) {
	db := openRO(t, seed(t, fullSchema, fullRows))
	sessions, err := store.List(db, model.Filter{})
	if err != nil {
		t.Fatal(err)
	}
	byID := map[string]model.Session{}
	for _, s := range sessions {
		byID[s.ID] = s
	}
	root := byID["ses_root1"]
	if root.Worktree != "/home/u/src/api" || root.ProjectName != "api" {
		t.Fatalf("bad project join: %+v", root)
	}
	child := byID["ses_child1"]
	if child.Worktree != "/home/u/src/api" || child.Path != "sub" || child.Directory != "/home/u/src/api/sub" {
		t.Fatalf("bad path-scoped row: %+v", child)
	}
	web := byID["ses_root2"]
	if web.Worktree != "" || web.ProjectName != "" {
		t.Fatalf("unexpected project data without join: %+v", web)
	}
}

func TestListClearsUnrelatedProjectWorktree(t *testing.T) {
	db := openRO(t, seed(t, fullSchema, []string{
		`INSERT INTO session (id, parent_id, title, time_created, time_updated, time_archived, directory, path, project_id)
			VALUES ('ses_global', NULL, 'global', 100, 200, 0, '/home/u/other', '', 'global')`,
		`INSERT INTO project (id, name, worktree) VALUES ('global', '/mnt/current', '/mnt/current')`,
	}))
	sessions, err := store.List(db, model.Filter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 1 || sessions[0].Worktree != "" || sessions[0].ProjectName != "" {
		t.Fatalf("unrelated project metadata leaked into session: %+v", sessions)
	}
}

func TestListMissingProjectTable(t *testing.T) {
	schema := `CREATE TABLE session (
  id TEXT PRIMARY KEY, parent_id TEXT, title TEXT,
  time_created INTEGER, time_updated INTEGER, time_archived INTEGER,
  directory TEXT, path TEXT, project_id TEXT);`
	db := openRO(t, seed(t, schema, fullRows[:5]))
	sessions, err := store.List(db, model.Filter{})
	if err != nil {
		t.Fatal(err)
	}
	expectIDs(t, sessions, "ses_root1", "ses_child1", "ses_root2")
	for _, s := range sessions {
		if s.Worktree != "" || s.ProjectName != "" {
			t.Fatalf("expected empty project fields, got %+v", s)
		}
	}
}

func TestListProjectTableWithoutSessionProjectID(t *testing.T) {
	schema := `CREATE TABLE session (
  id TEXT PRIMARY KEY, title TEXT,
  time_created INTEGER, time_updated INTEGER, time_archived INTEGER,
  directory TEXT, path TEXT);
CREATE TABLE project (id TEXT PRIMARY KEY, name TEXT, worktree TEXT);`
	db := openRO(t, seed(t, schema, []string{
		`INSERT INTO session (id, title, time_created, time_updated, time_archived, directory, path)
		 VALUES ('ses_no_project_id', 'legacy', 100, 200, 0, '/home/u/src/legacy', '')`,
		`INSERT INTO project (id, name, worktree) VALUES ('p1', 'unused', '/home/u/src/legacy')`,
	}))
	sessions, err := store.List(db, model.Filter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 1 || sessions[0].ID != "ses_no_project_id" {
		t.Fatalf("unexpected sessions: %+v", sessions)
	}
	if sessions[0].ProjectID != "" || sessions[0].Worktree != "" || sessions[0].ProjectName != "" {
		t.Fatalf("project fields should be empty without session.project_id: %+v", sessions[0])
	}
}

func TestListMissingOptionalColumns(t *testing.T) {
	schema := `CREATE TABLE session (
  id TEXT PRIMARY KEY, title TEXT,
  time_created INTEGER, time_updated INTEGER, directory TEXT);`
	inserts := []string{
		`INSERT INTO session (id, title, time_created, time_updated, directory)
		 VALUES ('ses_old1', 'legacy', 100, 300, '/home/u/src/legacy')`,
		`INSERT INTO session (id, title, time_created, time_updated, directory)
		 VALUES ('ses_old2', 'legacy two', 100, 200, '/home/u/src/legacy2')`,
	}
	db := openRO(t, seed(t, schema, inserts))
	sessions, err := store.List(db, model.Filter{})
	if err != nil {
		t.Fatal(err)
	}
	expectIDs(t, sessions, "ses_old1", "ses_old2")
	for _, s := range sessions {
		if s.ArchivedMS != 0 || s.ParentID != "" || s.Path != "" {
			t.Fatalf("expected zero optional fields, got %+v", s)
		}
	}
}

func TestListMissingUpdatedFallsBackToCreated(t *testing.T) {
	schema := `CREATE TABLE session (
  id TEXT PRIMARY KEY, title TEXT,
  time_created INTEGER, directory TEXT);`
	db := openRO(t, seed(t, schema, []string{
		`INSERT INTO session (id, title, time_created, directory)
		 VALUES ('ses_fb1', 'no updated', 500, '/home/u/src/fb')`,
	}))
	sessions, err := store.List(db, model.Filter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 1 || sessions[0].UpdatedMS != 500 {
		t.Fatalf("got %+v", sessions)
	}
}

func TestListMissingSessionTable(t *testing.T) {
	db := openRO(t, seed(t, `CREATE TABLE other (x TEXT);`, nil))
	if _, err := store.List(db, model.Filter{}); err == nil {
		t.Fatal("expected error for missing session table")
	} else if !faults.IsCode(err, model.CodeStoreUnread) {
		t.Fatalf("expected store.unreadable, got %v", err)
	}
}

func TestListSessionWithoutID(t *testing.T) {
	db := openRO(t, seed(t, `CREATE TABLE session (title TEXT);`, nil))
	if _, err := store.List(db, model.Filter{}); err == nil {
		t.Fatal("expected error for session without id")
	}
}

func TestOpenReadOnlyCreatesNoSideFiles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "opencode.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE session (id TEXT PRIMARY KEY, time_updated INTEGER);
		INSERT INTO session VALUES ('ses_x', 1);`); err != nil {
		t.Fatal(err)
	}
	db.Close()

	before, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	ro := openRO(t, path)
	rows, err := ro.Query("SELECT COUNT(*) FROM session")
	if err != nil {
		t.Fatal(err)
	}
	rows.Close()

	after, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(before) != len(after) {
		t.Fatalf("read-only open created files: before=%d after=%d", len(before), len(after))
	}
}

func TestDiscover(t *testing.T) {
	dir := t.TempDir()
	primary := filepath.Join(dir, "opencode.db")
	sibling := filepath.Join(dir, "opencode-extra.db")
	for _, p := range []string{primary, sibling} {
		if err := os.WriteFile(p, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("OPENCODE_DATA wins", func(t *testing.T) {
		t.Setenv("OPENCODE_DATA", dir)
		paths, err := store.Discover(nil)
		if err != nil {
			t.Fatal(err)
		}
		if len(paths) != 2 || paths[0] != primary {
			t.Fatalf("got %v", paths)
		}
	})

	t.Run("XDG default", func(t *testing.T) {
		data := t.TempDir()
		dbpath := filepath.Join(data, "opencode", "opencode.db")
		if err := os.MkdirAll(filepath.Dir(dbpath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(dbpath, nil, 0o644); err != nil {
			t.Fatal(err)
		}
		t.Setenv("OPENCODE_DATA", "")
		t.Setenv("XDG_DATA_HOME", data)
		paths, err := store.Discover(nil)
		if err != nil {
			t.Fatal(err)
		}
		if len(paths) != 1 || paths[0] != dbpath {
			t.Fatalf("got %v", paths)
		}
	})

	t.Run("opencode db path", func(t *testing.T) {
		bin := t.TempDir()
		script := filepath.Join(bin, "opencode")
		dbpath := filepath.Join(t.TempDir(), "via-cli.db")
		if err := os.WriteFile(dbpath, nil, 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(script, []byte("#!/bin/sh\necho "+dbpath+"\n"), 0o755); err != nil {
			t.Fatal(err)
		}
		t.Setenv("OPENCODE_DATA", "")
		t.Setenv("XDG_DATA_HOME", filepath.Join(t.TempDir(), "unused"))
		t.Setenv("PATH", bin+":"+os.Getenv("PATH"))
		paths, err := store.Discover(nil)
		if err != nil {
			t.Fatal(err)
		}
		if len(paths) != 1 || paths[0] != dbpath {
			t.Fatalf("got %v", paths)
		}
	})

	t.Run("no databases", func(t *testing.T) {
		t.Setenv("OPENCODE_DATA", filepath.Join(t.TempDir(), "empty"))
		if _, err := store.Discover(nil); err == nil {
			t.Fatal("expected error when nothing found")
		}
	})
}
