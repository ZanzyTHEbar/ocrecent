package model

type Session struct {
	ID          string `json:"id"`
	UpdatedMS   int64  `json:"updated_ms"`
	CreatedMS   int64  `json:"created_ms"`
	Title       string `json:"title"`
	Directory   string `json:"directory"`
	Path        string `json:"path,omitempty"`
	ProjectID   string `json:"project_id,omitempty"`
	Worktree    string `json:"worktree,omitempty"`
	ProjectName string `json:"project_name,omitempty"`
	ParentID    string `json:"parent_id,omitempty"`
	ArchivedMS  int64  `json:"archived_ms,omitempty"`
}

type Project struct {
	Key       string `json:"key"`
	Name      string `json:"name"`
	Directory string `json:"directory"`
	ID        string `json:"id"`
	UpdatedMS int64  `json:"updated_ms"`
	Count     int    `json:"count"`
}

type Filter struct {
	Children bool
	Archived bool
	Dir      string
	Extra    []string
	All      bool
}

type Cols map[string]bool

func (c Cols) Has(name string) bool { return c[name] }

type DoctorInfo struct {
	Version        string
	OpenCodePath   string
	OpenCodeDBPath string
	DBs            []string
	Schema         map[string][]string
	Warnings       []string
	Picker         string
	Terminal       string
	RootCount      int
	AllCount       int
}
