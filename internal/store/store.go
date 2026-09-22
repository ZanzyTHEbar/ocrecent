package store

import (
	"github.com/ZanzyTHEbar/faults-go"

	"github.com/ZanzyTHEbar/ocrecent/internal/model"
)

// Store adapts the package functions to the core.StorePort shape.
type Store struct{}

func (Store) DBs(extra []string) ([]string, error) { return Discover(extra) }

func (Store) Load(path string, f model.Filter) ([]model.Session, error) {
	db, err := Open(path)
	if err != nil {
		return nil, faults.Wrap(model.CodeStoreUnread, "open database", err, "path", path)
	}
	defer db.Close()
	return List(db, f)
}

func (Store) Schema(path string) (model.Cols, error) {
	db, err := Open(path)
	if err != nil {
		return nil, faults.Wrap(model.CodeStoreUnread, "open database", err, "path", path)
	}
	defer db.Close()
	return Inspect(db, "session")
}
