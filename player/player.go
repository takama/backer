package player

import (
	"errors"
	"sync"

	"github.com/takama/backer"
	"github.com/takama/backer/datastore"
	"github.com/takama/backer/helper"
	"github.com/takama/backer/model"
)

var (
	// ErrInsufficientPoints appears if player has not enough points
	ErrInsufficientPoints = errors.New("Insufficient points")
)

// Entry implements Player interface
type Entry struct {
	datastore.Controller `json:"-"`
	mutex                sync.RWMutex
	model.Player
}

// New returns new Entry which implement Player interface
func New(id string, ctrl datastore.Controller) (*Entry, error) {
	tx, err := ctrl.Transaction()
	if err != nil {
		tx.Rollback()
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

// Find returns Entry with existing Player
func Find(id string, ctrl datastore.Controller) (*Entry, error) {
	tx, err := ctrl.Transaction()
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	entry := &Entry{Controller: ctrl}

	player, err := ctrl.FindPlayer(id, tx)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	entry.Player = *player

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return entry, nil
}

// Fund funds (add to balance) player with amount
func (entry *Entry) Fund(amount backer.Points) error {
	tx, err := entry.Controller.Transaction()
	if err != nil {
		tx.Rollback()
		return err
	}

	player, err := entry.Controller.FindPlayer(entry.Player.ID, tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	player.Balance = backer.Points(
		helper.RoundPrice(float32(player.Balance) + helper.TruncatePrice(float32(amount))))
	err = entry.Controller.SavePlayer(player, tx)
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
	entry.Player.Balance = player.Balance

	return nil
}

// Take takes points from player account
func (entry *Entry) Take(amount backer.Points) error {
	tx, err := entry.Controller.Transaction()
	if err != nil {
		tx.Rollback()
		return err
	}

	player, err := entry.Controller.FindPlayer(entry.Player.ID, tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	if player.Balance < amount {
		tx.Rollback()
		return ErrInsufficientPoints
	}

	player.Balance = backer.Points(
		helper.RoundPrice(float32(player.Balance) - helper.TruncatePrice(float32(amount))))
	err = entry.Controller.SavePlayer(player, tx)
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
	entry.Player.Balance = player.Balance

	return nil
}

// Balance gets current points
func (entry *Entry) Balance() (backer.Points, error) {
	player, err := entry.Controller.FindPlayer(entry.Player.ID, nil)
	if err != nil {
		return 0, err
	}

	entry.mutex.Lock()
	defer entry.mutex.Unlock()
	entry.Player.Balance = player.Balance

	return player.Balance, nil
}

// ID returns player ID
func (entry *Entry) ID() string {
	entry.mutex.RLock()
	defer entry.mutex.RUnlock()
	return entry.Player.ID
}
