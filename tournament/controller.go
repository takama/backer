package tournament

import (
	"github.com/takama/backer/db"
	"github.com/takama/backer/model"
)

// Controller defines DB interface for Entry
type Controller interface {
	Transaction() (db.Transact, error)
	NewTournament(ID uint64, tx db.Transact) error
	FindTournament(ID uint64, tx db.Transact) (*model.Tournament, error)
	SaveTournament(tournament *model.Tournament, tx db.Transact) error
}
