package player

import (
	"errors"
	"testing"

	"github.com/takama/backer/db"
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

func TestNewPlayer(t *testing.T) {

	store := new(db.Stub)
	store.Reset()
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
	store.ErrTx = append(store.ErrTx, ErrFalseTransaction)
	_, err = New("p2", store)
	test(t, err == ErrFalseTransaction, "Expected", ErrFalseTransaction, "got", err)
	store.ErrTxCmt = append(store.ErrTxCmt, ErrFalseCommit)
	_, err = New("p3", store)
	test(t, err == ErrFalseCommit, "Expected", ErrFalseCommit, "got", err)
	store.ErrNew = append(store.ErrNew, ErrNewPlayer)
	_, err = New("p4", store)
	test(t, err == ErrNewPlayer, "Expected", ErrNewPlayer, "got", err)
}

func TestFindPlayer(t *testing.T) {

	store := new(db.Stub)
	store.Reset()
	_, err := Find("p1", store)
	test(t, err != nil, "Expected getting error, got nil")

	entry, err := New("p1", store)
	test(t, err == nil, "Expected creating a new player, got", err)
	if entry == nil {
		t.Fatal("Expected player entry, got nil")
	}
	test(t, entry.Player.ID == "p1", "Expected player id: p1, got", entry.Player.ID)
	entryExists, err := Find("p1", store)
	test(t, err == nil, "Expected find existing player, got", err)
	if entryExists == nil {
		t.Fatal("Expected player entry, got nil")
	}
	test(t, entry.Player.ID == entryExists.Player.ID,
		"Expected the players id's are equal, got", entryExists.Player.ID)
	test(t, entry.Player.Balance == entryExists.Player.Balance,
		"Expected the players balances are equal, got", entryExists.Player.Balance)
	store.ErrTx = append(store.ErrTx, ErrFalseTransaction)
	_, err = Find("p1", store)
	test(t, err == ErrFalseTransaction, "Expected", ErrFalseTransaction, "got", err)
	store.ErrTxCmt = append(store.ErrTxCmt, ErrFalseCommit)
	_, err = Find("p1", store)
	test(t, err == ErrFalseCommit, "Expected", ErrFalseCommit, "got", err)
}

func TestPlayerFund(t *testing.T) {

	store := new(db.Stub)
	store.Reset()
	entry, err := New("p1", store)
	test(t, err == nil, "Expected creating a new player, got", err)
	err = entry.Fund(300)
	test(t, err == nil, "Expected fund 300 to the player, got", err)
	points, err := entry.Balance()
	test(t, err == nil, "Expected check balance of the player, got", err)
	test(t, points == 300, "Expected 300 points for the player, got", points)
	store.ErrTx = append(store.ErrTx, ErrFalseTransaction)
	err = entry.Fund(10)
	test(t, err == ErrFalseTransaction, "Expected", ErrFalseTransaction, "got", err)
	store.ErrTxCmt = append(store.ErrTxCmt, ErrFalseCommit)
	err = entry.Fund(20)
	test(t, err == ErrFalseCommit, "Expected", ErrFalseCommit, "got", err)
	store.ErrFind = append(store.ErrFind, ErrFindPlayer)
	err = entry.Fund(30)
	test(t, err == ErrFindPlayer, "Expected", ErrFindPlayer, "got", err)
	store.ErrSave = append(store.ErrSave, ErrSavePlayer)
	err = entry.Fund(40)
	test(t, err == ErrSavePlayer, "Expected", ErrSavePlayer, "got", err)
}

func TestPlayerTake(t *testing.T) {

	store := new(db.Stub)
	store.Reset()
	entry, err := New("p3", store)
	test(t, err == nil, "Expected creating a new player, got", err)
	err = entry.Fund(300)
	test(t, err == nil, "Expected fund 300 to the player, got", err)
	err = entry.Take(400)
	test(t, err != nil, "Expected take more than player balance, got success")
	err = entry.Take(200)
	test(t, err == nil, "Expected take amount from player balance, got", err)
	balance, err := entry.Balance()
	test(t, err == nil, "Expected check balance of the player, got", err)
	test(t, balance == 100, "Expected 100 points for the player, got", balance)
	store.ErrTx = append(store.ErrTx, ErrFalseTransaction)
	err = entry.Take(10)
	test(t, err == ErrFalseTransaction, "Expected", ErrFalseTransaction, "got", err)
	store.ErrTxCmt = append(store.ErrTxCmt, ErrFalseCommit)
	err = entry.Take(20)
	test(t, err == ErrFalseCommit, "Expected", ErrFalseCommit, "got", err)
	store.ErrFind = append(store.ErrFind, ErrFindPlayer)
	err = entry.Take(30)
	test(t, err == ErrFindPlayer, "Expected", ErrFindPlayer, "got", err)
	store.ErrSave = append(store.ErrSave, ErrSavePlayer)
	err = entry.Take(40)
	test(t, err == ErrSavePlayer, "Expected", ErrSavePlayer, "got", err)
}

func TestPlayerBalance(t *testing.T) {

	store := new(db.Stub)
	store.Reset()
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
	store.ErrFind = append(store.ErrFind, ErrFindPlayer)
	_, err = entry.Balance()
	test(t, err == ErrFindPlayer, "Expected", ErrFindPlayer, "got", err)
}

func TestPlayerID(t *testing.T) {

	store := new(db.Stub)
	store.Reset()
	entry, err := New("p1", store)
	test(t, err == nil, "Expected creating a new player, got", err)
	id := entry.ID()
	test(t, id == entry.Player.ID, "Expected the player id,", entry.Player.ID, " got", id)
}
