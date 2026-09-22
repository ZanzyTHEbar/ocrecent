package core

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ZanzyTHEbar/faults-go"

	"github.com/ZanzyTHEbar/ocrecent/internal/format"
	"github.com/ZanzyTHEbar/ocrecent/internal/model"
	"github.com/ZanzyTHEbar/ocrecent/internal/picker"
	"github.com/ZanzyTHEbar/ocrecent/internal/rank"
)

type App struct {
	Store    StorePort
	Picker   PickerPort
	Notifier NotifierPort
	Resumer  ResumerPort
	Now      func() int64
	Warnings []string
}

func New(store StorePort, p PickerPort, n NotifierPort, r ResumerPort) *App {
	return &App{
		Store:    store,
		Picker:   p,
		Notifier: n,
		Resumer:  r,
		Now:      func() int64 { return time.Now().UnixMilli() },
	}
}

// Load merges every discovered database; on session-ID collision the row
// with the greater UpdatedMS wins.
func (a *App) Load(f model.Filter) ([]model.Session, error) {
	paths, err := a.Store.DBs(f.Extra)
	if err != nil {
		return nil, err
	}
	var all []model.Session
	a.Warnings = nil
	for _, p := range paths {
		list, err := a.Store.Load(p, f)
		if err != nil {
			a.Warnings = append(a.Warnings, fmt.Sprintf("%s: %v", p, err))
			continue
		}
		all = mergeSessions(all, list)
	}
	return all, nil
}

func mergeSessions(dst, src []model.Session) []model.Session {
	idx := map[string]int{}
	for i, s := range dst {
		idx[s.ID] = i
	}
	for _, s := range src {
		if i, ok := idx[s.ID]; ok {
			if s.UpdatedMS > dst[i].UpdatedMS {
				dst[i] = s
			}
			continue
		}
		idx[s.ID] = len(dst)
		dst = append(dst, s)
	}
	return dst
}

func (a *App) Sessions(f model.Filter, n int) ([]model.Session, error) {
	all, err := a.Load(f)
	if err != nil {
		return nil, err
	}
	return rank.Sessions(all, n, f.All), nil
}

func (a *App) Projects(f model.Filter, n int) ([]model.Project, error) {
	all, err := a.Load(f)
	if err != nil {
		return nil, err
	}
	return rank.Projects(all, n, f.All), nil
}

// Resolve maps an index (1-based, into the ranked listing) or an exact
// session id to a session.
func (a *App) Resolve(idOrIndex string, f model.Filter, n int) (model.Session, error) {
	sessions, err := a.Load(f)
	if err != nil {
		return model.Session{}, err
	}
	if idx, err := strconv.Atoi(idOrIndex); err == nil {
		sessions = rank.Sessions(sessions, n, f.All)
		if idx >= 1 && idx <= len(sessions) {
			return sessions[idx-1], nil
		}
		return model.Session{}, faults.New(model.CodeNotFound, "index out of range",
			"index", idx, "total", len(sessions))
	}
	for _, s := range rank.Sessions(sessions, 0, true) {
		if s.ID == idOrIndex {
			return s, nil
		}
	}
	return model.Session{}, faults.New(model.CodeNotFound, "session not found", "id", idOrIndex)
}

func (a *App) ResumeSession(idOrIndex string, f model.Filter, n int, inTTY bool) error {
	s, err := a.Resolve(idOrIndex, f, n)
	if err != nil {
		return err
	}
	return a.Resumer.Resume(s.Directory, s.ID, inTTY)
}

func (a *App) ResumeLast(f model.Filter, inTTY bool) error {
	sessions, err := a.Sessions(f, 1)
	if err != nil {
		return err
	}
	if len(sessions) == 0 {
		return faults.New(model.CodeNotFound, "no sessions found")
	}
	return a.Resumer.Resume(sessions[0].Directory, sessions[0].ID, inTTY)
}

// Pick resolves the picker, presents sessions or projects, and resumes the
// selection.
func (a *App) Pick(f model.Filter, n int, pickerExplicit string, projectsGrain, inTTY bool) error {
	s, err := a.PickSession(f, n, pickerExplicit, projectsGrain)
	if err != nil {
		return err
	}
	return a.Resumer.Resume(s.Directory, s.ID, inTTY)
}

func (a *App) PickSession(f model.Filter, n int, pickerExplicit string, projectsGrain bool) (model.Session, error) {
	lines := []string{}
	sessions := map[string]model.Session{}
	home := os.Getenv("HOME")
	now := a.Now()
	if projectsGrain {
		ps, err := a.Projects(f, n)
		if err != nil {
			return model.Session{}, err
		}
		if len(ps) == 0 {
			return model.Session{}, faults.New(model.CodeNotFound, "no projects found")
		}
		for _, p := range ps {
			lines = append(lines, picker.Line(p.Name, format.RelTime(now, p.UpdatedMS),
				format.HomePath(home, p.Key), p.ID))
			sessions[p.ID] = model.Session{ID: p.ID, Title: p.Name, Directory: p.Directory}
		}
	} else {
		ss, err := a.Sessions(f, n)
		if err != nil {
			return model.Session{}, err
		}
		if len(ss) == 0 {
			return model.Session{}, faults.New(model.CodeNotFound, "no sessions found")
		}
		for _, s := range ss {
			lines = append(lines, picker.Line(s.Title, format.RelTime(now, s.UpdatedMS),
				format.HomePath(home, s.Directory), s.ID))
			sessions[s.ID] = s
		}
	}

	bin, err := a.Picker.Resolve(pickerExplicit)
	if err != nil {
		return model.Session{}, err
	}
	id, err := a.Picker.Pick(bin, lines)
	if err != nil {
		return model.Session{}, err
	}
	s, ok := sessions[id]
	if !ok {
		return model.Session{}, faults.New(model.CodeNotFound, "picked session not found", "id", id)
	}
	return s, nil
}

// Notify posts the notification and returns the chosen action key.
func (a *App) Notify(f model.Filter, n int, urgency string) (string, error) {
	ps, err := a.Projects(f, n)
	if err != nil {
		return "", err
	}
	title := fmt.Sprintf("OpenCode · %d recent project(s)", len(ps))
	now := a.Now()
	var body strings.Builder
	for _, p := range ps {
		suffix := "sessions"
		if p.Count == 1 {
			suffix = "session"
		}
		fmt.Fprintf(&body, "%s  %s  (%d %s)\n", format.RelTime(now, p.UpdatedMS), p.Name, p.Count, suffix)
	}
	if len(ps) == 0 {
		body.WriteString("no recent sessions")
	}
	return a.Notifier.Send(title, strings.TrimRight(body.String(), "\n"), urgency)
}

// DispatchNotify runs the notify action: pick, projects (terminal list), or
// last (resume newest).
func (a *App) DispatchNotify(action string, f model.Filter, n int, pickerExplicit string) error {
	switch action {
	case "pick":
		return a.Pick(f, n, pickerExplicit, true, false)
	case "projects":
		exe, err := os.Executable()
		if err != nil {
			return faults.Wrap(model.CodeNotFound, "locate own binary", err)
		}
		return a.Resumer.Spawn([]string{exe, "projects", "--all"})
	case "last":
		return a.ResumeLast(f, false)
	}
	return nil
}
