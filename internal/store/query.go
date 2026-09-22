package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ZanzyTHEbar/faults-go"

	"github.com/ZanzyTHEbar/ocrecent/internal/model"
)

// List scans the logical sessions from db under filter. A database missing
// the session table or its id column is an error; the caller skips it and
// reports the warning.
func List(db *sql.DB, f model.Filter) ([]model.Session, error) {
	sessCols, err := Inspect(db, "session")
	if err != nil {
		return nil, err
	}
	if len(sessCols) == 0 {
		return nil, faults.New(model.CodeStoreUnread, "session table missing")
	}
	if !sessCols.Has("id") {
		return nil, faults.New(model.CodeStoreUnread, "session table has no id column")
	}

	projCols, _ := Inspect(db, "project")
	hasProject := sessCols.Has("project_id") && projCols.Has("id")

	pick := func(name, expr, fallback string) string {
		if sessCols.Has(name) {
			return expr
		}
		return fallback
	}

	worktree, projName := "''", "''"
	join := ""
	if hasProject {
		join = "LEFT JOIN project p ON p.id = s.project_id"
		if projCols.Has("worktree") {
			worktree = "p.worktree"
		}
		if projCols.Has("name") {
			projName = "p.name"
		}
	}

	where := []string{"1=1"}
	if !f.Children && sessCols.Has("parent_id") {
		where = append(where, "(s.parent_id IS NULL OR s.parent_id = '')")
	}
	if !f.Archived && sessCols.Has("time_archived") {
		where = append(where, "(s.time_archived IS NULL OR s.time_archived = 0)")
	}

	q := fmt.Sprintf(`SELECT
  s.id,
  %s,
  %s,
  %s,
  %s,
  %s,
  %s,
  %s,
  %s,
  %s,
  %s
FROM session s
%s
WHERE %s`,
		pick("time_updated", "COALESCE(s.time_updated, s.time_created, 0)", "COALESCE(s.time_created, 0)"),
		pick("time_created", "COALESCE(s.time_created, 0)", "0"),
		pick("title", "s.title", "''"),
		pick("directory", "s.directory", "''"),
		pick("path", "s.path", "''"),
		pick("project_id", "s.project_id", "''"),
		pick("parent_id", "s.parent_id", "''"),
		pick("time_archived", "s.time_archived", "0"),
		worktree,
		projName,
		join,
		strings.Join(where, " AND "),
	)

	rows, err := db.Query(q)
	if err != nil {
		return nil, faults.Wrap(model.CodeStoreUnread, "query sessions", err)
	}
	defer rows.Close()

	var sessions []model.Session
	for rows.Next() {
		var s model.Session
		var updated, created, archived sql.NullInt64
		var title, dir, path, projectID, parentID, wt, name sql.NullString
		if err := rows.Scan(&s.ID, &updated, &created, &title, &dir, &path,
			&projectID, &parentID, &archived, &wt, &name); err != nil {
			return nil, faults.Wrap(model.CodeStoreUnread, "scan session", err)
		}
		s.UpdatedMS = updated.Int64
		s.CreatedMS = created.Int64
		s.ArchivedMS = archived.Int64
		s.Title = title.String
		s.Directory = dir.String
		s.Path = path.String
		s.ProjectID = projectID.String
		s.ParentID = parentID.String
		s.Worktree = wt.String
		s.ProjectName = name.String
		if s.Worktree != "" && s.Directory != "" && !within(s.Worktree, s.Directory) {
			s.Worktree = ""
			s.ProjectName = ""
		}
		sessions = append(sessions, s)
	}
	if err := rows.Err(); err != nil {
		return nil, faults.Wrap(model.CodeStoreUnread, "iterate sessions", err)
	}

	if f.Dir != "" {
		sessions = filterDir(sessions, filepath.Clean(f.Dir))
	}
	return sessions, nil
}

func within(root, path string) bool {
	root = filepath.Clean(root)
	path = filepath.Clean(path)
	return root == path || strings.HasPrefix(path, root+string(os.PathSeparator))
}

func filterDir(sessions []model.Session, dir string) []model.Session {
	out := sessions[:0]
	for _, s := range sessions {
		d := filepath.Clean(s.Directory)
		if d == dir || strings.HasPrefix(d, dir+string(os.PathSeparator)) {
			out = append(out, s)
		}
	}
	return out
}
