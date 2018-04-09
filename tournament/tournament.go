package tournament

import (
	"sync"

	"github.com/takama/backer/model"
)

// Entry implements Tournament interface
type Entry struct {
	Controller
	mutex sync.RWMutex
	model.Tournament
}

// New returns new Entry which implement Tournament interface
func New(id uint64, ctrl Controller) (*Entry, error) {
	tx, err := ctrl.Transaction()
	if err != nil {
		return nil, err
	}

	entry := &Entry{Controller: ctrl}

	tournament, err := ctrl.FindTournament(id, tx)
	if err != nil {
		err = ctrl.NewTournament(id, tx)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		tournament = &model.Tournament{ID: id}
	}
	entry.Tournament = *tournament

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return entry, nil
}
