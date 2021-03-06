package tournament

import (
	"errors"
	"sync"

	"github.com/takama/backer"
	"github.com/takama/backer/datastore"
	"github.com/takama/backer/helper"
	"github.com/takama/backer/model"
	"github.com/takama/backer/player"
)

var (
	// ErrAllreadyFinished appears if the tournament was already finished
	ErrAllreadyFinished = errors.New("Tournament already finished")
	// ErrPlayersAlreadyJoined appears if the tournament was already announced and players already joined
	ErrPlayersAlreadyJoined = errors.New("Could not re-announce the Tournament, players already joined")
	// ErrCouldNotJoinTwice appears if the same player try to join to the tournament twice
	ErrCouldNotJoinTwice = errors.New("Could not join twice to the same tournament")
	// ErrWinnerIsNotMember appears if among winners exists a player who not a tournament member as a player
	ErrWinnerIsNotMember = errors.New("Not a tournament player can not be a winner")
)

// Entry implements Tournament interface
type Entry struct {
	datastore.Controller `json:"-"`
	mutex                sync.RWMutex
	model.Tournament
}

// New returns new Entry which implement Tournament interface
func New(id uint64, ctrl datastore.Controller) (*Entry, error) {
	tx, err := ctrl.Transaction()
	if err != nil {
		tx.Rollback()
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
		tournament = &model.Tournament{ID: id, Bidders: make([]model.Bidder, 0)}
	}
	entry.Tournament = *tournament

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return entry, nil
}

// Find returns Entry with existing Tournament
func Find(id uint64, ctrl datastore.Controller) (*Entry, error) {
	tx, err := ctrl.Transaction()
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	entry := &Entry{Controller: ctrl}

	tournament, err := ctrl.FindTournament(id, tx)
	if err != nil {
		tx.Rollback()
		return nil, err
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
		tx.Rollback()
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

	if len(tournament.Bidders) > 0 {
		return ErrPlayersAlreadyJoined
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
	entry.Tournament.Bidders = tournament.Bidders

	return nil
}

// Join player and backers into a tournament
func (entry *Entry) Join(players ...backer.Player) error {
	tx, err := entry.Controller.Transaction()
	if err != nil {
		tx.Rollback()
		return err
	}

	tournament, err := entry.Controller.FindTournament(entry.Tournament.ID, tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	if tournament.IsFinished {
		tx.Rollback()
		return ErrAllreadyFinished
	}

	var bidder model.Bidder
	bidder.Backers = make([]string, 0)
	contribute := float32(tournament.Deposit / backer.Points(len(players)))
	for idx, participant := range players {
		if _, err := player.ManagePoints(entry.Controller, tx,
			participant.ID(), backer.Points(-contribute)); err != nil {
			tx.Rollback()
			return err
		}
		if idx == 0 {
			for _, member := range tournament.Bidders {
				if member.ID == participant.ID() {
					tx.Rollback()
					return ErrCouldNotJoinTwice
				}
			}
			bidder.ID = participant.ID()
		} else {
			bidder.Backers = append(bidder.Backers, participant.ID())
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
		tx.Rollback()
		return err
	}

	tournament, err := entry.Controller.FindTournament(entry.Tournament.ID, tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	if tournament.IsFinished {
		tx.Rollback()
		return ErrAllreadyFinished
	}

	entry.mutex.Lock()
	defer entry.mutex.Unlock()

	for winner, points := range winners {
		for idx, bidder := range tournament.Bidders {
			if bidder.ID == winner.ID() {
				tournament.Bidders[idx].Winner = true
				tournament.Bidders[idx].Prize = points
				prize := float32(points / backer.Points(len(bidder.Backers)+1))
				if _, err := player.ManagePoints(entry.Controller, tx,
					winner.ID(), backer.Points(prize)); err != nil {
					tx.Rollback()
					return err
				}
				for _, id := range bidder.Backers {
					if _, err := player.ManagePoints(entry.Controller, tx,
						id, backer.Points(prize)); err != nil {
						tx.Rollback()
						return err
					}
				}
				delete(winners, winner)
			}
		}
	}

	if len(winners) != 0 {
		tx.Rollback()
		return ErrWinnerIsNotMember
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
	entry.Tournament.Bidders = tournament.Bidders

	return nil
}
