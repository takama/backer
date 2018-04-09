package player

import (
	"github.com/takama/backer/db"
	"github.com/takama/backer/model"
)

// Controller defines DB interface for Player Entry
type Controller interface {
	Transaction() (db.Transact, error)
	NewPlayer(ID string, tx db.Transact) error
	FindPlayer(ID string, tx db.Transact) (*model.Player, error)
	SavePlayer(player *model.Player, tx db.Transact) error
}
