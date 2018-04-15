package datastore

import (
	"github.com/takama/backer/model"
)

// Controller defines DB interface for Player and Tournament Entry
type Controller interface {
	Transaction() (Transact, error)
	NewPlayer(ID string, tx Transact) error
	FindPlayer(ID string, tx Transact) (*model.Player, error)
	SavePlayer(player *model.Player, tx Transact) error
	NewTournament(ID uint64, tx Transact) error
	FindTournament(ID uint64, tx Transact) (*model.Tournament, error)
	SaveTournament(tournament *model.Tournament, tx Transact) error
}
