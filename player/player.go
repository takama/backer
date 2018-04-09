package player

import (
	"sync"

	"github.com/takama/backer/model"
)

// Entry implements Player interface
type Entry struct {
	Controller
	mutex sync.RWMutex
	model.Player
}

// New returns new Entry which implement Player interface
func New(id string, ctrl Controller) (*Entry, error) {
	tx, err := ctrl.Transaction()
	if err != nil {
		return nil, err
	}

	entry := &Entry{Controller: ctrl}

	player, err := ctrl.FindPlayer(id, tx)
	if err != nil {
		err = ctrl.NewPlayer(id, tx)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		player = &model.Player{ID: id}
	}
	entry.Player = *player

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return entry, nil
}
