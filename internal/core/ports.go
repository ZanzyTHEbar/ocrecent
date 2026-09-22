package core

import "github.com/ZanzyTHEbar/ocrecent/internal/model"

type StorePort interface {
	DBs(extra []string) ([]string, error)
	Load(path string, f model.Filter) ([]model.Session, error)
	Schema(path string) (model.Cols, error)
}

type PickerPort interface {
	Resolve(explicit string) (string, error)
	Pick(bin string, lines []string) (string, error)
}

type NotifierPort interface {
	Send(title, body, urgency string) (string, error)
}

type ResumerPort interface {
	Resume(dir, id string, inTTY bool) error
	Spawn(argv []string) error
}
