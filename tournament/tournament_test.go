package tournament

import (
	"errors"
	"sync"
	"testing"

	"github.com/takama/backer"
	"github.com/takama/backer/db"
	"github.com/takama/backer/model"
	"github.com/takama/backer/player"
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

type playerTx struct{}

func (ptx playerTx) Commit() error {
	return nil
}

func (ptx playerTx) Rollback() error {
	return nil
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
	pb.tx = new(playerTx)
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

func TestFindTournament(t *testing.T) {

	store := &tournamentBundle{
		tx:      new(tournamentTxSuccess),
		records: make(map[uint64]model.Tournament),
	}
	_, err := Find(1, store)
	test(t, err != nil, "Expected getting error, got nil")

	tournament, err := New(1, store)
	test(t, err == nil, "Expected creating a new tournament, got", err)
	if tournament == nil {
		t.Fatal("Expected tournament entry, got nil")
	}
	test(t, tournament.ID == 1, "Expected tournament id: p1, got", tournament.ID)
	tournamentExists, err := Find(1, store)
	test(t, err == nil, "Expected find existing tournament, got", err)
	if tournamentExists == nil {
		t.Fatal("Expected tournament entry, got nil")
	}
	test(t, tournament.ID == tournamentExists.ID, "Expected the tournaments id's are equal, got", tournamentExists.ID)
	store.errTx = ErrFalseTransaction
	_, err = Find(1, store)
	test(t, err == ErrFalseTransaction, "Expected", ErrFalseTransaction, "got", err)
	store.tx = new(tournamentTxFalse)
	store.errTx = nil
	_, err = Find(1, store)
	test(t, err == ErrFalseCommit, "Expected", ErrFalseCommit, "got", err)
	store.errNew = ErrNewTournament
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
	store.errSave = nil
	err = tournament.Result(nil)
	err = tournament.Announce(700)
	test(t, err == ErrAllreadyFinished, "Expected", ErrAllreadyFinished, "got", err)
}

func TestTournamentJoin(t *testing.T) {

	playerStore := &playerBundle{
		records: make(map[string]model.Player),
	}
	playerP1, err := player.New("p1", playerStore)
	test(t, err == nil, "Expected creating a new player, got", err)

	store := &tournamentBundle{
		tx:      new(tournamentTxSuccess),
		records: make(map[uint64]model.Tournament),
	}
	tournament, err := New(1, store)
	test(t, err == nil, "Expected creating a new tournament, got", err)
	err = tournament.Announce(1000)
	test(t, err == nil, "Expected announce of the tournament, got", err)

	err = tournament.Join(playerP1)
	test(t, err == player.ErrInsufficientPoints, "Expected", player.ErrInsufficientPoints, "got", err)
	err = playerP1.Fund(1000)
	test(t, err == nil, "Expected fund 1000 to the player, got", err)

	store.errTx = ErrFalseTransaction
	err = tournament.Join(playerP1)
	test(t, err == ErrFalseTransaction, "Expected", ErrFalseTransaction, "got", err)
	store.errTx = nil
	store.errFind = ErrFindTournament
	err = tournament.Join(playerP1)
	test(t, err == ErrFindTournament, "Expected", ErrFindTournament, "got", err)
	store.errFind = nil
	store.errSave = ErrSaveTournament
	err = tournament.Join(playerP1)
	test(t, err == ErrSaveTournament, "Expected", ErrSaveTournament, "got", err)
	err = playerP1.Fund(1000)
	test(t, err == nil, "Expected fund 1000 to the player, got", err)
	store.errSave = nil
	store.tx = new(tournamentTxFalse)
	err = tournament.Join(playerP1)
	test(t, err == ErrFalseCommit, "Expected", ErrFalseCommit, "got", err)
	err = playerP1.Fund(1000)
	test(t, err == nil, "Expected fund 1000 to the player, got", err)
	store.tx = new(tournamentTxSuccess)

	err = tournament.Join(playerP1)
	test(t, err == nil, "Expected join a player, got", err)

	playerP2, err := player.New("p2", playerStore)
	test(t, err == nil, "Expected creating a new player, got", err)
	playerB1, err := player.New("b1", playerStore)
	test(t, err == nil, "Expected creating a new player, got", err)
	playerB2, err := player.New("b2", playerStore)
	test(t, err == nil, "Expected creating a new player, got", err)
	playerB3, err := player.New("b3", playerStore)
	test(t, err == nil, "Expected creating a new player, got", err)
	err = tournament.Join(playerP2, playerB1, playerB2, playerB3)
	test(t, err == player.ErrInsufficientPoints, "Expected", player.ErrInsufficientPoints, "got", err)
	err = playerP2.Fund(500)
	test(t, err == nil, "Expected fund 500 to the player, got", err)
	err = playerB1.Fund(300)
	test(t, err == nil, "Expected fund 300 to the player, got", err)
	err = playerB2.Fund(300)
	test(t, err == nil, "Expected fund 300 to the player, got", err)
	err = playerB3.Fund(300)
	test(t, err == nil, "Expected fund 300 to the player, got", err)
	err = tournament.Join(playerP2, playerB1, playerB2, playerB3)
	test(t, err == nil, "Expected join a player, got", err)

	err = tournament.Result(nil)
	playerP3, err := player.New("p3", playerStore)
	test(t, err == nil, "Expected creating a new player, got", err)
	err = tournament.Join(playerP3)
	test(t, err == ErrAllreadyFinished, "Expected", ErrAllreadyFinished, "got", err)
}

func TestTournamentResult(t *testing.T) {

	playerStore := &playerBundle{
		records: make(map[string]model.Player),
	}
	playerP1, err := player.New("p1", playerStore)
	test(t, err == nil, "Expected creating a new player, got", err)

	store := &tournamentBundle{
		tx:      new(tournamentTxSuccess),
		records: make(map[uint64]model.Tournament),
	}
	tournament, err := New(1, store)
	test(t, err == nil, "Expected creating a new tournament, got", err)
	err = tournament.Announce(1000)
	test(t, err == nil, "Expected announce of the tournament, got", err)

	err = playerP1.Fund(1000)
	test(t, err == nil, "Expected fund 1000 to the player, got", err)
	err = tournament.Join(playerP1)
	test(t, err == nil, "Expected join a player, got", err)

	playerP2, err := player.New("p2", playerStore)
	test(t, err == nil, "Expected creating a new player, got", err)
	playerB1, err := player.New("b1", playerStore)
	test(t, err == nil, "Expected creating a new player, got", err)
	playerB2, err := player.New("b2", playerStore)
	test(t, err == nil, "Expected creating a new player, got", err)
	playerB3, err := player.New("b3", playerStore)
	test(t, err == nil, "Expected creating a new player, got", err)
	err = playerP2.Fund(500)
	test(t, err == nil, "Expected fund 500 to the player, got", err)
	err = playerB1.Fund(300)
	test(t, err == nil, "Expected fund 300 to the player, got", err)
	err = playerB2.Fund(300)
	test(t, err == nil, "Expected fund 300 to the player, got", err)
	err = playerB3.Fund(300)
	test(t, err == nil, "Expected fund 300 to the player, got", err)
	err = tournament.Join(playerP2, playerB1, playerB2, playerB3)
	test(t, err == nil, "Expected join a player, got", err)

	winners := make(map[backer.Player]backer.Points)
	winners[playerP2] = 2000

	store.errTx = ErrFalseTransaction
	err = tournament.Result(winners)
	test(t, err == ErrFalseTransaction, "Expected", ErrFalseTransaction, "got", err)
	store.errTx = nil
	store.errFind = ErrFindTournament
	err = tournament.Result(winners)
	test(t, err == ErrFindTournament, "Expected", ErrFindTournament, "got", err)
	store.errFind = nil

	err = tournament.Result(winners)
	test(t, err == nil, "Expected result of the tournament, got", err)

	balance, err := playerP1.Balance()
	test(t, err == nil, "Expected check balance of the player, got", err)
	test(t, balance == 0, "Expected 0 points for the player, got", balance)
	balance, err = playerP2.Balance()
	test(t, err == nil, "Expected check balance of the player, got", err)
	test(t, balance == 750, "Expected 750 points for the player, got", balance)
	balance, err = playerB1.Balance()
	test(t, err == nil, "Expected check balance of the player, got", err)
	test(t, balance == 550, "Expected 550 points for the player, got", balance)
	balance, err = playerB2.Balance()
	test(t, err == nil, "Expected check balance of the player, got", err)
	test(t, balance == 550, "Expected 550 points for the player, got", balance)
	balance, err = playerB3.Balance()
	test(t, err == nil, "Expected check balance of the player, got", err)
	test(t, balance == 550, "Expected 550 points for the player, got", balance)

	err = tournament.Result(winners)
	test(t, err == ErrAllreadyFinished, "Expected", ErrAllreadyFinished, "got", err)

	tournament, err = New(2, store)
	test(t, err == nil, "Expected creating a new tournament, got", err)
	store.errSave = ErrSaveTournament
	err = tournament.Result(winners)
	test(t, err == ErrSaveTournament, "Expected", ErrSaveTournament, "got", err)
	store.errSave = nil

	tournament, err = New(3, store)
	test(t, err == nil, "Expected creating a new tournament, got", err)
	store.tx = new(tournamentTxFalse)
	err = tournament.Result(winners)
	test(t, err == ErrFalseCommit, "Expected", ErrFalseCommit, "got", err)
	store.tx = new(tournamentTxSuccess)

	tournament, err = New(4, store)
	test(t, err == nil, "Expected creating a new tournament, got", err)
	err = tournament.Announce(1000)
	test(t, err == nil, "Expected announce of the tournament, got", err)
	err = playerP2.Fund(250)
	test(t, err == nil, "Expected fund 1000 to the player, got", err)
	err = tournament.Join(playerP2)
	test(t, err == nil, "Expected join a player, got", err)
	playerStore.errTx = ErrFalseTransaction
	err = tournament.Result(winners)
	test(t, err == ErrFalseTransaction, "Expected", ErrFalseTransaction, "got", err)
	playerStore.errTx = nil

	tournament, err = New(5, store)
	test(t, err == nil, "Expected creating a new tournament, got", err)
	err = tournament.Announce(1000)
	test(t, err == nil, "Expected announce of the tournament, got", err)
	err = playerP1.Fund(250)
	test(t, err == nil, "Expected fund 250 to the player, got", err)
	err = tournament.Join(playerP1, playerB1, playerB2, playerB3)
	test(t, err == nil, "Expected join players, got", err)
	winners = make(map[backer.Player]backer.Points)
	delete(playerStore.records, "b3")
	winners[playerP1] = 1000
	err = tournament.Result(winners)
	test(t, err == ErrNotExist, "Expected", ErrNotExist, "got", err)

	tournament, err = New(6, store)
	test(t, err == nil, "Expected creating a new tournament, got", err)
	err = tournament.Announce(1000)
	test(t, err == nil, "Expected announce of the tournament, got", err)
	err = playerP1.Fund(400)
	test(t, err == nil, "Expected fund 250 to the player, got", err)
	err = tournament.Join(playerP1, playerB1, playerB2)
	test(t, err == nil, "Expected join players, got", err)
	err = playerP2.Fund(1000)
	test(t, err == nil, "Expected fund 1000 to the player, got", err)
	err = tournament.Join(playerP2)
	test(t, err == nil, "Expected join players, got", err)

	winners = make(map[backer.Player]backer.Points)
	winners[playerP1] = 2000
	err = tournament.Result(winners)
	test(t, err == nil, "Expected result of the tournament, got", err)

	balance, err = playerP1.Balance()
	test(t, err == nil, "Expected check balance of the player, got", err)
	test(t, balance == 983.33, "Expected 983.33 points for the player, got", balance)
	balance, err = playerB1.Balance()
	test(t, err == nil, "Expected check balance of the player, got", err)
	test(t, balance == 883.33, "Expected 883.33 points for the player, got", balance)
	balance, err = playerB2.Balance()
	test(t, err == nil, "Expected check balance of the player, got", err)
	test(t, balance == 883.33, "Expected 883.33 points for the player, got", balance)
}
