package tournament

import (
	"errors"
	"testing"

	"github.com/takama/backer"
	"github.com/takama/backer/datastore"
	"github.com/takama/backer/player"
)

var (
	ErrFalseTransaction = errors.New("Test false transaction")
	ErrFalseCommit      = errors.New("Test false commit")
	ErrFalseRollback    = errors.New("Test false rollback")
	ErrNewTournament    = errors.New("Test new tournament with error")
	ErrFindTournament   = errors.New("Test find tournament with error")
	ErrSaveTournament   = errors.New("Test save tournament with error")
)

func test(t *testing.T, expected bool, messages ...interface{}) {
	if !expected {
		t.Error(messages)
	}
}

func TestNewTournament(t *testing.T) {

	store := new(datastore.Stub)
	store.Reset()
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
	store.ErrTx = append(store.ErrTx, ErrFalseTransaction)
	_, err = New(2, store)
	test(t, err == ErrFalseTransaction, "Expected", ErrFalseTransaction, "got", err)
	store.ErrTxCmt = append(store.ErrTxCmt, ErrFalseCommit)
	_, err = New(3, store)
	test(t, err == ErrFalseCommit, "Expected", ErrFalseCommit, "got", err)
	store.ErrNew = append(store.ErrNew, ErrNewTournament)
	_, err = New(4, store)
	test(t, err == ErrNewTournament, "Expected", ErrNewTournament, "got", err)
}

func TestFindTournament(t *testing.T) {

	store := new(datastore.Stub)
	store.Reset()
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
	store.ErrTx = append(store.ErrTx, ErrFalseTransaction)
	_, err = Find(1, store)
	test(t, err == ErrFalseTransaction, "Expected", ErrFalseTransaction, "got", err)
	store.ErrTxCmt = append(store.ErrTxCmt, ErrFalseCommit)
	_, err = Find(1, store)
	test(t, err == ErrFalseCommit, "Expected", ErrFalseCommit, "got", err)
	store.ErrFind = append(store.ErrFind, ErrFindTournament)
	_, err = Find(1, store)
}

func TestTournamentAnnounce(t *testing.T) {

	store := new(datastore.Stub)
	store.Reset()
	tournament, err := New(1, store)
	test(t, err == nil, "Expected creating a new tournament, got", err)
	err = tournament.Announce(1000)
	test(t, err == nil, "Expected announce of the tournament, got", err)
	playerP1, err := player.New("p1", store)
	test(t, err == nil, "Expected creating a new player, got", err)
	err = playerP1.Fund(1000)
	test(t, err == nil, "Expected fund 1000 to the player, got", err)
	err = tournament.Join(playerP1)
	test(t, err == nil, "Expected join a player, got", err)
	err = tournament.Announce(1000)
	test(t, err == ErrPlayersAlreadyJoined, "Expected disable to re-announce of the tournament, got", err)
	tournament, err = New(2, store)
	test(t, err == nil, "Expected creating a new tournament, got", err)
	store.ErrTx = append(store.ErrTx, ErrFalseTransaction)
	err = tournament.Announce(2000)
	test(t, err == ErrFalseTransaction, "Expected", ErrFalseTransaction, "got", err)
	store.ErrTxCmt = append(store.ErrTxCmt, ErrFalseCommit)
	err = tournament.Announce(3000)
	test(t, err == ErrFalseCommit, "Expected", ErrFalseCommit, "got", err)
	store.ErrFind = append(store.ErrFind, ErrFindTournament)
	err = tournament.Announce(300)
	test(t, err == ErrFindTournament, "Expected", ErrFindTournament, "got", err)
	store.ErrSave = append(store.ErrSave, ErrSaveTournament)
	err = tournament.Announce(500)
	test(t, err == ErrSaveTournament, "Expected", ErrSaveTournament, "got", err)
	err = tournament.Result(nil)
	err = tournament.Announce(700)
	test(t, err == ErrAllreadyFinished, "Expected", ErrAllreadyFinished, "got", err)
}

func TestTournamentJoin(t *testing.T) {

	store := new(datastore.Stub)
	store.Reset()
	playerP1, err := player.New("p1", store)
	test(t, err == nil, "Expected creating a new player, got", err)

	tournament, err := New(1, store)
	test(t, err == nil, "Expected creating a new tournament, got", err)
	err = tournament.Announce(1000)
	test(t, err == nil, "Expected announce of the tournament, got", err)

	err = tournament.Join(playerP1)
	test(t, err == player.ErrInsufficientPoints, "Expected", player.ErrInsufficientPoints, "got", err)
	err = playerP1.Fund(1000)
	test(t, err == nil, "Expected fund 1000 to the player, got", err)
	balance, err := playerP1.Balance()
	test(t, err == nil, "Expected check balance of the player, got", err)
	test(t, balance == 1000, "Expected 1000 points for the player, got", balance)

	store.ErrTx = append(store.ErrTx, ErrFalseTransaction)
	err = tournament.Join(playerP1)
	test(t, err == ErrFalseTransaction, "Expected", ErrFalseTransaction, "got", err)
	balance, err = playerP1.Balance()
	test(t, err == nil, "Expected check balance of the player, got", err)
	test(t, balance == 1000, "Expected 1000 points for the player, got", balance)
	store.ErrFind = append(store.ErrFind, ErrFindTournament)
	err = tournament.Join(playerP1)
	test(t, err == ErrFindTournament, "Expected", ErrFindTournament, "got", err)
	balance, err = playerP1.Balance()
	test(t, err == nil, "Expected check balance of the player, got", err)
	test(t, balance == 1000, "Expected 1000 points for the player, got", balance)
	store.ErrSave = append(store.ErrSave, ErrSaveTournament, nil)
	err = tournament.Join(playerP1)
	test(t, err == ErrSaveTournament, "Expected", ErrSaveTournament, "got", err)
	balance, err = playerP1.Balance()
	test(t, err == nil, "Expected check balance of the player, got", err)
	test(t, balance == 1000, "Expected 1000 points for the player, got", balance)
	store.ErrTxCmt = append(store.ErrTxCmt, ErrFalseCommit, nil)
	err = tournament.Join(playerP1)
	test(t, err == ErrFalseCommit, "Expected", ErrFalseCommit, "got", err)
	balance, err = playerP1.Balance()
	test(t, err == nil, "Expected check balance of the player, got", err)
	test(t, balance == 0, "Expected 0 points for the player, got", balance)
	err = playerP1.Fund(1000)
	test(t, err == nil, "Expected fund 1000 to the player, got", err)
	balance, err = playerP1.Balance()
	test(t, err == nil, "Expected check balance of the player, got", err)
	test(t, balance == 1000, "Expected 1000 points for the player, got", balance)

	err = tournament.Join(playerP1)
	test(t, err == ErrCouldNotJoinTwice, "Expected", ErrCouldNotJoinTwice, "got", err)
	balance, err = playerP1.Balance()
	test(t, err == nil, "Expected check balance of the player, got", err)
	test(t, balance == 1000, "Expected 1000 points for the player, got", balance)

	tournament, err = New(2, store)
	test(t, err == nil, "Expected creating a new tournament, got", err)
	err = tournament.Announce(1000)
	test(t, err == nil, "Expected announce of the tournament, got", err)

	err = tournament.Join(playerP1)
	test(t, err == nil, "Expected join a player, got", err)
	balance, err = playerP1.Balance()
	test(t, err == nil, "Expected check balance of the player, got", err)
	test(t, balance == 0, "Expected 0 points for the player, got", balance)

	playerP2, err := player.New("p2", store)
	test(t, err == nil, "Expected creating a new player, got", err)
	playerB1, err := player.New("b1", store)
	test(t, err == nil, "Expected creating a new player, got", err)
	playerB2, err := player.New("b2", store)
	test(t, err == nil, "Expected creating a new player, got", err)
	playerB3, err := player.New("b3", store)
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
	test(t, err == nil, "Expected join players, got", err)

	err = tournament.Result(nil)
	playerP3, err := player.New("p3", store)
	test(t, err == nil, "Expected creating a new player, got", err)
	err = tournament.Join(playerP3)
	test(t, err == ErrAllreadyFinished, "Expected", ErrAllreadyFinished, "got", err)
}

func TestTournamentResult(t *testing.T) {

	store := new(datastore.Stub)
	store.Reset()
	playerP1, err := player.New("p1", store)
	test(t, err == nil, "Expected creating a new player, got", err)

	tournament, err := New(1, store)
	test(t, err == nil, "Expected creating a new tournament, got", err)
	err = tournament.Announce(1000)
	test(t, err == nil, "Expected announce of the tournament, got", err)

	err = playerP1.Fund(1000)
	test(t, err == nil, "Expected fund 1000 to the player, got", err)
	err = tournament.Join(playerP1)
	test(t, err == nil, "Expected join a player, got", err)

	playerP2, err := player.New("p2", store)
	test(t, err == nil, "Expected creating a new player, got", err)
	playerB1, err := player.New("b1", store)
	test(t, err == nil, "Expected creating a new player, got", err)
	playerB2, err := player.New("b2", store)
	test(t, err == nil, "Expected creating a new player, got", err)
	playerB3, err := player.New("b3", store)
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

	store.ErrTx = append(store.ErrTx, ErrFalseTransaction)
	err = tournament.Result(winners)
	test(t, err == ErrFalseTransaction, "Expected", ErrFalseTransaction, "got", err)
	store.ErrFind = append(store.ErrFind, ErrFindTournament)
	err = tournament.Result(winners)
	test(t, err == ErrFindTournament, "Expected", ErrFindTournament, "got", err)

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
	store.ErrSave = append(store.ErrSave, ErrSaveTournament)
	err = tournament.Result(winners)
	test(t, err == ErrSaveTournament, "Expected", ErrSaveTournament, "got", err)

	tournament, err = New(3, store)
	test(t, err == nil, "Expected creating a new tournament, got", err)
	store.ErrTxCmt = append(store.ErrTxCmt, ErrFalseCommit)
	err = tournament.Result(winners)
	test(t, err == ErrFalseCommit, "Expected", ErrFalseCommit, "got", err)

	tournament, err = New(4, store)
	test(t, err == nil, "Expected creating a new tournament, got", err)
	err = tournament.Announce(1000)
	test(t, err == nil, "Expected announce of the tournament, got", err)
	err = playerP2.Fund(250)
	test(t, err == nil, "Expected fund 250 to the player, got", err)
	balance, err = playerP2.Balance()
	test(t, err == nil, "Expected check balance of the player, got", err)
	test(t, balance == 1000, "Expected 1000 points for the player, got", balance)
	err = tournament.Join(playerP2)
	test(t, err == nil, "Expected join a player, got", err)
	balance, err = playerP2.Balance()
	test(t, err == nil, "Expected check balance of the player, got", err)
	test(t, balance == 0, "Expected 0 points for the player, got", balance)
	store.ErrTx = append(store.ErrTx, ErrFalseTransaction)
	err = tournament.Result(winners)
	test(t, err == ErrFalseTransaction, "Expected", ErrFalseTransaction, "got", err)

	tournament, err = New(5, store)
	test(t, err == nil, "Expected creating a new tournament, got", err)
	err = tournament.Announce(1000)
	test(t, err == nil, "Expected announce of the tournament, got", err)
	err = playerP1.Fund(250)
	test(t, err == nil, "Expected fund 250 to the player, got", err)
	balance, err = playerP1.Balance()
	test(t, err == nil, "Expected check balance of the player, got", err)
	test(t, balance == 250, "Expected 250 points for the player, got", balance)
	err = tournament.Join(playerP1, playerB1, playerB2, playerB3)
	test(t, err == nil, "Expected join players, got", err)
	winners = make(map[backer.Player]backer.Points)
	store.ErrFind = append(store.ErrFind, datastore.ErrRecordNotFound, nil)
	winners[playerP1] = 1000
	err = tournament.Result(winners)
	test(t, err == datastore.ErrRecordNotFound, "Expected", datastore.ErrRecordNotFound, "got", err)
	balance, err = playerP1.Balance()
	test(t, err == nil, "Expected check balance of the player, got", err)
	test(t, balance == 0, "Expected 0 points for the player, got", balance)
	balance, err = playerB1.Balance()
	test(t, err == nil, "Expected check balance of the player, got", err)
	test(t, balance == 300, "Expected 300 points for the player, got", balance)
	balance, err = playerB2.Balance()
	test(t, err == nil, "Expected check balance of the player, got", err)
	test(t, balance == 300, "Expected 300 points for the player, got", balance)

	store.ErrFind = append(store.ErrFind, datastore.ErrRecordNotFound, nil, nil)
	winners[playerP1] = 1000
	err = tournament.Result(winners)
	test(t, err == datastore.ErrRecordNotFound, "Expected", datastore.ErrRecordNotFound, "got", err)
	balance, err = playerP1.Balance()
	test(t, err == nil, "Expected check balance of the player, got", err)
	test(t, balance == 0, "Expected 0 points for the player, got", balance)
	balance, err = playerB1.Balance()
	test(t, err == nil, "Expected check balance of the player, got", err)
	test(t, balance == 300, "Expected 300 points for the player, got", balance)
	balance, err = playerB2.Balance()
	test(t, err == nil, "Expected check balance of the player, got", err)
	test(t, balance == 300, "Expected 300 points for the player, got", balance)

	store.ErrFind = append(store.ErrFind, datastore.ErrRecordNotFound, nil, nil, nil)
	winners[playerP1] = 1000
	err = tournament.Result(winners)
	test(t, err == datastore.ErrRecordNotFound, "Expected", datastore.ErrRecordNotFound, "got", err)

	tournament, err = New(6, store)
	test(t, err == nil, "Expected creating a new tournament, got", err)
	err = tournament.Announce(1000)
	test(t, err == nil, "Expected announce of the tournament, got", err)
	err = playerP1.Fund(600)
	test(t, err == nil, "Expected fund 600 to the player, got", err)
	balance, err = playerP1.Balance()
	test(t, err == nil, "Expected check balance of the player, got", err)
	test(t, balance == 600, "Expected 600 points for the player, got", balance)
	err = playerB1.Fund(200)
	test(t, err == nil, "Expected fund 200 to the player, got", err)
	balance, err = playerB1.Balance()
	test(t, err == nil, "Expected check balance of the player, got", err)
	test(t, balance == 500, "Expected 500 points for the player, got", balance)
	err = playerB2.Fund(200)
	test(t, err == nil, "Expected fund 200 to the player, got", err)
	balance, err = playerB2.Balance()
	test(t, err == nil, "Expected check balance of the player, got", err)
	test(t, balance == 500, "Expected 500 points for the player, got", balance)
	err = tournament.Join(playerP1, playerB1, playerB2)
	test(t, err == nil, "Expected join players, got", err)
	err = playerP2.Fund(1000)
	test(t, err == nil, "Expected fund 1000 to the player, got", err)
	balance, err = playerP2.Balance()
	test(t, err == nil, "Expected check balance of the player, got", err)
	test(t, balance == 1000, "Expected 1000 points for the player, got", balance)
	err = tournament.Join(playerP2)
	test(t, err == nil, "Expected join players, got", err)

	winners = make(map[backer.Player]backer.Points)
	winners[playerP1] = 2000
	err = tournament.Result(winners)
	test(t, err == nil, "Expected result of the tournament, got", err)

	balance, err = playerP1.Balance()
	test(t, err == nil, "Expected check balance of the player, got", err)
	test(t, balance == 933.33, "Expected 933.33 points for the player, got", balance)
	balance, err = playerB1.Balance()
	test(t, err == nil, "Expected check balance of the player, got", err)
	test(t, balance == 833.33, "Expected 833.33 points for the player, got", balance)
	balance, err = playerB2.Balance()
	test(t, err == nil, "Expected check balance of the player, got", err)
	test(t, balance == 833.33, "Expected 833.33 points for the player, got", balance)
}
