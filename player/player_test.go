package player

import (
	"errors"
	"sync"
	"testing"

	"github.com/takama/backer/db"
	"github.com/takama/backer/model"
)

var (
	ErrFalseTransaction = errors.New("Test false transaction")
	ErrFalseCommit      = errors.New("Test false commit")
	ErrFalseRollback    = errors.New("Test false rollback")
	ErrNewPlayer        = errors.New("Test new player with error")
	ErrFindPlayer       = errors.New("Test find player with error")
	ErrSavePlayer       = errors.New("Test save player with error")
	ErrAlreadyExist     = errors.New("Record already exists")
	ErrNotExist         = errors.New("Record does not exist")
)

func test(t *testing.T, expected bool, messages ...interface{}) {
	if !expected {
		t.Error(messages)
	}
}

type playerTxSuccess struct{}

func (ptx playerTxSuccess) Commit() error {
	return nil
}

func (ptx playerTxSuccess) Rollback() error {
	return nil
}

type playerTxFalse struct{}

func (ptx playerTxFalse) Commit() error {
	return ErrFalseCommit
}

func (ptx playerTxFalse) Rollback() error {
	return ErrFalseRollback
}

type playerBundle struct {
	mutex   sync.RWMutex
	tx      db.Transact
	errTx   error
	errNew  error
	errFind error
	errSave error
	records map[string]model.Player
}

func (pb *playerBundle) Transaction() (db.Transact, error) {
	return pb.tx, pb.errTx
}

func (pb *playerBundle) NewPlayer(ID string, tx db.Transact) error {
	pb.mutex.RLock()
	defer pb.mutex.RUnlock()
	_, ok := pb.records[ID]
	if ok {
		return ErrAlreadyExist
	}
	pb.records[ID] = model.Player{ID: ID}
	return pb.errNew
}

func (pb *playerBundle) FindPlayer(ID string, tx db.Transact) (*model.Player, error) {
	pb.mutex.RLock()
	defer pb.mutex.RUnlock()
	player, ok := pb.records[ID]
	if !ok {
		return nil, ErrNotExist
	}
	return &player, pb.errFind
}

func (pb *playerBundle) SavePlayer(player *model.Player, tx db.Transact) error {
	pb.mutex.Lock()
	defer pb.mutex.Unlock()

	pb.records[player.ID] = *player
	return pb.errSave
}

func TestNewPlayer(t *testing.T) {

	store := &playerBundle{
		tx:      new(playerTxSuccess),
		records: make(map[string]model.Player),
	}
	entry, err := New("p1", store)
	test(t, err == nil, "Expected creating a new player, got", err)
	if entry == nil {
		t.Fatal("Expected player entry, got nil")
	}
	test(t, entry.Player.ID == "p1", "Expected player id: p1, got", entry.Player.ID)
	entryExists, err := New("p1", store)
	test(t, err == nil, "Expected find existing player, got", err)
	if entryExists == nil {
		t.Fatal("Expected player entry, got nil")
	}
	test(t, entry.Player.ID == entryExists.Player.ID,
		"Expected the players id's are equal, got", entryExists.Player.ID)
	test(t, entry.Player.Balance == entryExists.Player.Balance,
		"Expected the players balances are equal, got", entryExists.Player.Balance)
	store.errTx = ErrFalseTransaction
	_, err = New("p2", store)
	test(t, err == ErrFalseTransaction, "Expected", ErrFalseTransaction, "got", err)
	store.tx = new(playerTxFalse)
	store.errTx = nil
	_, err = New("p3", store)
	test(t, err == ErrFalseCommit, "Expected", ErrFalseCommit, "got", err)
	store.errNew = ErrNewPlayer
	_, err = New("p4", store)
	test(t, err == ErrNewPlayer, "Expected", ErrNewPlayer, "got", err)
}

func TestPlayerFund(t *testing.T) {

	store := &playerBundle{
		tx:      new(playerTxSuccess),
		records: make(map[string]model.Player),
	}
	entry, err := New("p1", store)
	test(t, err == nil, "Expected creating a new player, got", err)
	err = entry.Fund(300)
	test(t, err == nil, "Expected fund 300 to the player, got", err)
	points, err := entry.Balance()
	test(t, err == nil, "Expected check balance of the player, got", err)
	test(t, points == 300, "Expected 300 points for the player, got", points)
	store.errTx = ErrFalseTransaction
	err = entry.Fund(10)
	test(t, err == ErrFalseTransaction, "Expected", ErrFalseTransaction, "got", err)
	store.tx = new(playerTxFalse)
	store.errTx = nil
	err = entry.Fund(20)
	test(t, err == ErrFalseCommit, "Expected", ErrFalseCommit, "got", err)
	store.tx = new(playerTxSuccess)
	store.errFind = ErrFindPlayer
	err = entry.Fund(30)
	test(t, err == ErrFindPlayer, "Expected", ErrFindPlayer, "got", err)
	store.errFind = nil
	store.errSave = ErrSavePlayer
	err = entry.Fund(40)
	test(t, err == ErrSavePlayer, "Expected", ErrSavePlayer, "got", err)
}

func TestPlayerBalance(t *testing.T) {

	store := &playerBundle{
		tx:      new(playerTxSuccess),
		records: make(map[string]model.Player),
	}
	entry, err := New("p4", store)
	test(t, err == nil, "Expected creating a new player, got", err)
	balance, err := entry.Balance()
	test(t, err == nil, "Expected check balance of the player, got", err)
	test(t, balance == 0, "Expected 0 points for the player, got", balance)
	err = entry.Fund(50.99)
	test(t, err == nil, "Expected fund 50.99 to the player, got", err)
	balance, err = entry.Balance()
	test(t, err == nil, "Expected check balance of the player, got", err)
	test(t, balance == 50.99, "Expected 300 points for the player, got", balance)
	store.errFind = ErrFindPlayer
	_, err = entry.Balance()
	test(t, err == ErrFindPlayer, "Expected", ErrFindPlayer, "got", err)
}
