package tournament

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
	ErrNewTournament    = errors.New("Test new tournament with error")
	ErrFindTournament   = errors.New("Test find tournament with error")
	ErrSaveTournament   = errors.New("Test save tournament with error")
	ErrAlreadyExist     = errors.New("Record already exists")
	ErrNotExist         = errors.New("Record does not exist")
)

func test(t *testing.T, expected bool, messages ...interface{}) {
	if !expected {
		t.Error(messages)
	}
}

type tournamentTxSuccess struct{}

func (ptx tournamentTxSuccess) Commit() error {
	return nil
}

func (ptx tournamentTxSuccess) Rollback() error {
	return nil
}

type tournamentTxFalse struct{}

func (ptx tournamentTxFalse) Commit() error {
	return ErrFalseCommit
}

func (ptx tournamentTxFalse) Rollback() error {
	return ErrFalseRollback
}

type tournamentBundle struct {
	mutex   sync.RWMutex
	tx      db.Transact
	errTx   error
	errNew  error
	errFind error
	errSave error
	records map[uint64]model.Tournament
}

func (trn *tournamentBundle) Transaction() (db.Transact, error) {
	return trn.tx, trn.errTx
}

func (trn *tournamentBundle) NewTournament(ID uint64, tx db.Transact) error {
	trn.mutex.RLock()
	defer trn.mutex.RUnlock()
	_, ok := trn.records[ID]
	if ok {
		return ErrAlreadyExist
	}
	trn.records[ID] = model.Tournament{ID: ID}
	return trn.errNew
}

func (trn *tournamentBundle) FindTournament(ID uint64, tx db.Transact) (*model.Tournament, error) {
	trn.mutex.RLock()
	defer trn.mutex.RUnlock()
	tournament, ok := trn.records[ID]
	if !ok {
		return nil, ErrNotExist
	}
	return &tournament, trn.errFind
}

func (trn *tournamentBundle) SaveTournament(tournament *model.Tournament, tx db.Transact) error {
	trn.mutex.Lock()
	defer trn.mutex.Unlock()

	trn.records[tournament.ID] = *tournament
	return trn.errSave
}

func TestNewTournament(t *testing.T) {

	store := &tournamentBundle{
		tx:      new(tournamentTxSuccess),
		records: make(map[uint64]model.Tournament),
	}
	tournament, err := New(1, store)
	test(t, err == nil, "Expected creating a new tournament, got", err)
	if tournament == nil {
		t.Fatal("Expected tournament entry, got nil")
	}
	test(t, tournament.ID == 1, "Expected tournament id: p1, got", tournament.ID)
	tournamentExists, err := New(1, store)
	test(t, err == nil, "Expected find existing tournament, got", err)
	if tournamentExists == nil {
		t.Fatal("Expected tournament entry, got nil")
	}
	test(t, tournament.ID == tournamentExists.ID, "Expected the tournaments id's are equal, got", tournamentExists.ID)
	store.errTx = ErrFalseTransaction
	_, err = New(2, store)
	test(t, err == ErrFalseTransaction, "Expected", ErrFalseTransaction, "got", err)
	store.tx = new(tournamentTxFalse)
	store.errTx = nil
	_, err = New(3, store)
	test(t, err == ErrFalseCommit, "Expected", ErrFalseCommit, "got", err)
	store.errNew = ErrNewTournament
	_, err = New(4, store)
	test(t, err == ErrNewTournament, "Expected", ErrNewTournament, "got", err)
}

func TestTournamentAnnounce(t *testing.T) {

	store := &tournamentBundle{
		tx:      new(tournamentTxSuccess),
		records: make(map[uint64]model.Tournament),
	}
	tournament, err := New(1, store)
	test(t, err == nil, "Expected creating a new tournament, got", err)
	err = tournament.Announce(1000)
	test(t, err == nil, "Expected announce of the tournament, got", err)
	store.errTx = ErrFalseTransaction
	err = tournament.Announce(2000)
	test(t, err == ErrFalseTransaction, "Expected", ErrFalseTransaction, "got", err)
	store.tx = new(tournamentTxFalse)
	store.errTx = nil
	err = tournament.Announce(3000)
	test(t, err == ErrFalseCommit, "Expected", ErrFalseCommit, "got", err)
	store.tx = new(tournamentTxSuccess)
	store.errFind = ErrFindTournament
	err = tournament.Announce(300)
	test(t, err == ErrFindTournament, "Expected", ErrFindTournament, "got", err)
	store.errFind = nil
	store.errSave = ErrSaveTournament
	err = tournament.Announce(500)
	test(t, err == ErrSaveTournament, "Expected", ErrSaveTournament, "got", err)
}
