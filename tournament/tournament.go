package tournament

import (
	"sync"

	"github.com/takama/backer"
	"github.com/takama/backer/helper"
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

// Announce tournament with specified deposit
func (entry *Entry) Announce(deposit backer.Points) error {
	tx, err := entry.Controller.Transaction()
	if err != nil {
		return err
	}

	tournament, err := entry.Controller.FindTournament(entry.Tournament.ID, tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	tournament.Deposit = backer.Points(helper.TruncatePrice(float32(deposit)))
	err = entry.Controller.SaveTournament(tournament, tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	entry.mutex.Lock()
	defer entry.mutex.Unlock()
	entry.Tournament.Deposit = tournament.Deposit

	return nil
}
