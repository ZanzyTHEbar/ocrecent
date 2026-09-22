package rank

import (
	"testing"

	"github.com/ZanzyTHEbar/ocrecent/internal/model"
)

func TestSessionsSortsDescStable(t *testing.T) {
	in := []model.Session{
		{ID: "a", UpdatedMS: 100},
		{ID: "d", UpdatedMS: 300},
		{ID: "c", UpdatedMS: 200},
		{ID: "b", UpdatedMS: 300},
	}
	got := Sessions(in, 8, false)
	want := []string{"b", "d", "c", "a"}
	for i, id := range want {
		if got[i].ID != id {
			t.Fatalf("pos %d: want %s got %s", i, id, got[i].ID)
		}
	}
}

func TestProjectsUseDeterministicTieBreakers(t *testing.T) {
	in := []model.Session{
		{ID: "ses_z", UpdatedMS: 100, Directory: "/work/z", Worktree: "/work/z"},
		{ID: "ses_a", UpdatedMS: 100, Directory: "/work/a", Worktree: "/work/a"},
	}
	got := Projects(in, 8, false)
	if len(got) != 2 || got[0].Key != "/work/a" || got[1].Key != "/work/z" {
		t.Fatalf("project tie ordering is not deterministic: %+v", got)
	}

	in = []model.Session{
		{ID: "ses_z", UpdatedMS: 100, Directory: "/work/same", Worktree: "/work/same"},
		{ID: "ses_a", UpdatedMS: 100, Directory: "/work/same/sub", Worktree: "/work/same"},
	}
	got = Projects(in, 8, false)
	if len(got) != 1 || got[0].ID != "ses_a" {
		t.Fatalf("representative tie ordering is not deterministic: %+v", got)
	}
}

func TestSessionsLimit(t *testing.T) {
	in := []model.Session{
		{ID: "a", UpdatedMS: 1}, {ID: "b", UpdatedMS: 2}, {ID: "c", UpdatedMS: 3},
	}
	if got := Sessions(in, 2, false); len(got) != 2 || got[0].ID != "c" || got[1].ID != "b" {
		t.Fatalf("bad limit: %v", got)
	}
	if got := Sessions(in, 2, true); len(got) != 3 {
		t.Fatalf("all should ignore limit: %d", len(got))
	}
}

func TestProjectsCollapse(t *testing.T) {
	in := []model.Session{
		{ID: "s1", UpdatedMS: 300, Directory: "/w/proj", Worktree: "/w/proj", ProjectName: "proj"},
		{ID: "s2", UpdatedMS: 200, Directory: "/w/proj/sub", Worktree: "/w/proj"},
		{ID: "s3", UpdatedMS: 100, Directory: "/w/other", Worktree: "", ProjectName: "other"},
	}
	got := Projects(in, 8, false)
	if len(got) != 2 {
		t.Fatalf("want 2 groups, got %d", len(got))
	}
	if got[0].Key != "/w/proj" || got[0].ID != "s1" || got[0].Count != 2 || got[0].Name != "proj" {
		t.Fatalf("bad group 0: %+v", got[0])
	}
	if got[1].Key != "/w/other" || got[1].Count != 1 || got[1].Name != "other" {
		t.Fatalf("bad group 1: %+v", got[1])
	}
}

func TestProjectsNameFallsBackToBase(t *testing.T) {
	in := []model.Session{{ID: "s1", UpdatedMS: 1, Directory: "/w/named", Worktree: "/w/named"}}
	got := Projects(in, 8, false)
	if got[0].Name != "named" {
		t.Fatalf("want base name fallback, got %q", got[0].Name)
	}
}

func TestProjectsDoNotTrustUnrelatedWorktree(t *testing.T) {
	in := []model.Session{
		{ID: "global", UpdatedMS: 200, Directory: "/other/project", Worktree: "/current/project"},
		{ID: "child", UpdatedMS: 100, Directory: "/current/project/sub", Worktree: "/current/project"},
	}
	got := Projects(in, 8, false)
	if len(got) != 2 || got[0].Key != "/other/project" || got[1].Key != "/current/project" {
		t.Fatalf("unrelated worktree collapsed sessions: %+v", got)
	}
}
