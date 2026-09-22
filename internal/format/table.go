package format

import (
	"encoding/json"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/ZanzyTHEbar/ocrecent/internal/model"
)

// JSONRows renders the stable array shape: logical session fields plus
// "kind": "session"|"project". Exactly one of sessions/projects is non-empty.
func JSONRows(sessions []model.Session, projects []model.Project) (string, error) {
	if len(projects) > 0 {
		type row struct {
			Kind string `json:"kind"`
			model.Project
		}
		rows := make([]row, 0, len(projects))
		for _, p := range projects {
			rows = append(rows, row{Kind: "project", Project: p})
		}
		b, err := json.Marshal(rows)
		return string(b), err
	}
	type row struct {
		Kind string `json:"kind"`
		model.Session
	}
	rows := make([]row, 0, len(sessions))
	for _, s := range sessions {
		rows = append(rows, row{Kind: "session", Session: s})
	}
	b, err := json.Marshal(rows)
	return string(b), err
}

func SessionTable(sessions []model.Session, home string, now int64) string {
	var b strings.Builder
	w := tabwriter.NewWriter(&b, 2, 4, 2, ' ', 0)
	fmt.Fprintln(w, "#\tREL\tTITLE\tDIRECTORY\tSESSION")
	for i, s := range sessions {
		dir := HomePath(home, s.Directory)
		if s.Worktree != "" && s.Worktree != s.Directory {
			dir += "  (worktree: " + HomePath(home, s.Worktree) + ")"
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", i+1, RelTime(now, s.UpdatedMS), s.Title, dir, s.ID)
	}
	w.Flush()
	return strings.TrimRight(b.String(), "\n")
}

func ProjectTable(projects []model.Project, home string, now int64) string {
	var b strings.Builder
	w := tabwriter.NewWriter(&b, 2, 4, 2, ' ', 0)
	fmt.Fprintln(w, "#\tREL\tNAME\tPATH\tSESSIONS")
	for i, p := range projects {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%d\n", i+1, RelTime(now, p.UpdatedMS), p.Name, HomePath(home, p.Key), p.Count)
	}
	w.Flush()
	return strings.TrimRight(b.String(), "\n")
}
