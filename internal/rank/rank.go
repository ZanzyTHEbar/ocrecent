package rank

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ZanzyTHEbar/ocrecent/internal/model"
)

// Sessions sorts by UpdatedMS descending, then ID ascending, and cuts to n
// unless all is set.
func Sessions(sessions []model.Session, n int, all bool) []model.Session {
	sort.SliceStable(sessions, func(i, j int) bool {
		if sessions[i].UpdatedMS != sessions[j].UpdatedMS {
			return sessions[i].UpdatedMS > sessions[j].UpdatedMS
		}
		return sessions[i].ID < sessions[j].ID
	})
	if !all && n > 0 && len(sessions) > n {
		sessions = sessions[:n]
	}
	return sessions
}

// Projects collapses sessions into project grain, keyed by worktree when
// present, else directory. The representative row is the one with the max
// UpdatedMS; its ID and Directory are what resume uses.
func Projects(sessions []model.Session, n int, all bool) []model.Project {
	byKey := map[string][]model.Session{}
	var keys []string
	for _, s := range sessions {
		key := projectKey(s)
		if _, ok := byKey[key]; !ok {
			keys = append(keys, key)
		}
		byKey[key] = append(byKey[key], s)
	}

	projects := make([]model.Project, 0, len(keys))
	for _, key := range keys {
		list := byKey[key]
		best := list[0]
		for _, s := range list[1:] {
			if s.UpdatedMS > best.UpdatedMS ||
				(s.UpdatedMS == best.UpdatedMS && s.ID < best.ID) {
				best = s
			}
		}
		name := best.ProjectName
		if name == "" {
			name = filepath.Base(key)
		}
		projects = append(projects, model.Project{
			Key:       key,
			Name:      name,
			Directory: best.Directory,
			ID:        best.ID,
			UpdatedMS: best.UpdatedMS,
			Count:     len(list),
		})
	}
	sort.SliceStable(projects, func(i, j int) bool {
		if projects[i].UpdatedMS != projects[j].UpdatedMS {
			return projects[i].UpdatedMS > projects[j].UpdatedMS
		}
		return projects[i].Key < projects[j].Key
	})
	if !all && n > 0 && len(projects) > n {
		projects = projects[:n]
	}
	return projects
}

func projectKey(s model.Session) string {
	if s.Worktree != "" && (s.Directory == "" || within(s.Worktree, s.Directory)) {
		return s.Worktree
	}
	return s.Directory
}

func within(root, path string) bool {
	root = filepath.Clean(root)
	path = filepath.Clean(path)
	return root == path || strings.HasPrefix(path, root+string(os.PathSeparator))
}
