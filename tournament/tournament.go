package tournament

import (
	"errors"
	"sync"

	"github.com/takama/backer"
	"github.com/takama/backer/helper"
	"github.com/takama/backer/model"
)

var (
	// ErrAllreadyFinished appears if tournament was already finished
	ErrAllreadyFinished = errors.New("Tournament already finished")
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

	if tournament.IsFinished {
		return ErrAllreadyFinished
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

// Join player and backers into a tournament
func (entry *Entry) Join(players ...backer.Player) error {
	tx, err := entry.Controller.Transaction()
	if err != nil {
		return err
	}

	tournament, err := entry.Controller.FindTournament(entry.Tournament.ID, tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	if tournament.IsFinished {
		return ErrAllreadyFinished
	}

	var bidder model.Bidder
	contribute := float32(tournament.Deposit / backer.Points(len(players)))
	for idx, player := range players {
		err := player.Take(backer.Points(contribute))
		if err != nil {
			tx.Rollback()
			return err
		}
		if idx == 0 {
			bidder.ID = player.ID()
		} else {
			bidder.Backers = append(bidder.Backers, player)
		}
	}
	tournament.Bidders = append(tournament.Bidders, bidder)

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
	entry.Tournament.Bidders = tournament.Bidders

	return nil
}

// Result tournament prizes and winners
func (entry *Entry) Result(winners map[backer.Player]backer.Points) error {
	tx, err := entry.Controller.Transaction()
	if err != nil {
		return err
	}

	tournament, err := entry.Controller.FindTournament(entry.Tournament.ID, tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	if tournament.IsFinished {
		return ErrAllreadyFinished
	}

	entry.mutex.Lock()
	defer entry.mutex.Unlock()

	for player, points := range winners {
		for idx, bidder := range tournament.Bidders {
			if bidder.ID == player.ID() {
				tournament.Bidders[idx].Winner = true
				prize := float32(points / backer.Points(len(bidder.Backers)+1))
				err := player.Fund(backer.Points(prize))
				if err != nil {
					tx.Rollback()
					return err
				}
				for _, p := range bidder.Backers {
					err := p.Fund(backer.Points(prize))
					if err != nil {
						tx.Rollback()
						return err
					}
				}
			}
		}
	}

	tournament.IsFinished = true

	err = entry.Controller.SaveTournament(tournament, tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	entry.Tournament.IsFinished = tournament.IsFinished

	return nil
}
