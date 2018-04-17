package datastore

import (
	"errors"
	"sync"

	"github.com/takama/backer/model"
)

var (
	// ErrAlreadyExist appears for existing records when try to create new one
	ErrAlreadyExist = errors.New("Record already exists")
	// ErrRecordNotFound appears if record does not exist
	ErrRecordNotFound = errors.New("Record not found")
)

// Stub in-memory controller
type Stub struct {
	mutex       sync.RWMutex
	tx          Transact
	transact    transactData
	ErrReset    []error
	ErrMigUp    []error
	ErrMigDn    []error
	ErrTx       []error
	ErrTxCmt    []error
	ErrTxRbk    []error
	ErrNew      []error
	ErrFind     []error
	ErrSave     []error
	ErrDelete   []error
	players     map[string]model.Player
	tournaments map[uint64]model.Tournament
}

type transactData struct {
	mutex       sync.RWMutex
	players     map[string]model.Player
	tournaments map[uint64]model.Tournament
}

// Ready returns connection state
func (stub *Stub) Ready() bool {
	return true
}

// Reset makes the DB initialization
func (stub *Stub) Reset() error {
	var err error
	stub.players = make(map[string]model.Player)
	stub.tournaments = make(map[uint64]model.Tournament)
	stub.tx = stub
	if len(stub.ErrReset) == 0 {
		return nil
	}
	err, stub.ErrReset = stub.ErrReset[len(stub.ErrReset)-1], stub.ErrReset[:len(stub.ErrReset)-1]
	return err
}

// MigrateUp migrates DB schema
func (stub *Stub) MigrateUp() error {
	return stub.Reset()
}

// MigrateDown remove DB schema and data
func (stub *Stub) MigrateDown() error {
	return stub.Reset()
}

// Transaction returns DB transaction control
func (stub *Stub) Transaction() (Transact, error) {
	var err error
	stub.mutex.RLock()
	defer stub.mutex.RUnlock()
	stub.transact.players = make(map[string]model.Player)
	stub.transact.tournaments = make(map[uint64]model.Tournament)
	for idx, val := range stub.players {
		stub.transact.players[idx] = val
	}
	for idx, val := range stub.tournaments {
		stub.transact.tournaments[idx] = val
	}
	if len(stub.ErrTx) == 0 {
		return stub.tx, nil
	}
	err, stub.ErrTx = stub.ErrTx[len(stub.ErrTx)-1], stub.ErrTx[:len(stub.ErrTx)-1]
	return stub.tx, err
}

// Commit confirms all changes during a transaction
func (stub *Stub) Commit() error {
	var err error
	stub.transact.players = make(map[string]model.Player)
	stub.transact.tournaments = make(map[uint64]model.Tournament)
	if len(stub.ErrTxCmt) == 0 {
		return nil
	}
	err, stub.ErrTxCmt = stub.ErrTxCmt[len(stub.ErrTxCmt)-1], stub.ErrTxCmt[:len(stub.ErrTxCmt)-1]
	return err
}

// Rollback undo all changes during a transaction
func (stub *Stub) Rollback() error {
	var err error
	stub.mutex.RLock()
	defer stub.mutex.RUnlock()
	stub.Reset()
	for idx, val := range stub.transact.players {
		stub.players[idx] = val
	}
	for idx, val := range stub.transact.tournaments {
		stub.tournaments[idx] = val
	}
	stub.transact.players = make(map[string]model.Player)
	stub.transact.tournaments = make(map[uint64]model.Tournament)
	if len(stub.ErrTxRbk) == 0 {
		return nil
	}
	err, stub.ErrTxRbk = stub.ErrTxRbk[len(stub.ErrTxRbk)-1], stub.ErrTxRbk[:len(stub.ErrTxRbk)-1]
	return err
}

// NewPlayer creates a new player with specified ID
func (stub *Stub) NewPlayer(ID string, tx Transact) error {
	var err error
	stub.mutex.RLock()
	defer stub.mutex.RUnlock()
	_, ok := stub.players[ID]
	if ok {
		return ErrAlreadyExist
	}
	stub.players[ID] = model.Player{ID: ID}
	if len(stub.ErrNew) == 0 {
		return nil
	}
	err, stub.ErrNew = stub.ErrNew[len(stub.ErrNew)-1], stub.ErrNew[:len(stub.ErrNew)-1]
	return err
}

// FindPlayer finds existing player by specified ID
func (stub *Stub) FindPlayer(ID string, tx Transact) (*model.Player, error) {
	var err error
	stub.mutex.RLock()
	defer stub.mutex.RUnlock()
	player, ok := stub.players[ID]
	if !ok {
		return nil, ErrRecordNotFound
	}
	if len(stub.ErrFind) == 0 {
		return &player, nil
	}
	err, stub.ErrFind = stub.ErrFind[len(stub.ErrFind)-1], stub.ErrFind[:len(stub.ErrFind)-1]
	return &player, err
}

// SavePlayer saves a Player model
func (stub *Stub) SavePlayer(player *model.Player, tx Transact) error {
	var err error
	stub.mutex.Lock()
	defer stub.mutex.Unlock()

	stub.players[player.ID] = *player
	if len(stub.ErrSave) == 0 {
		return nil
	}
	err, stub.ErrSave = stub.ErrSave[len(stub.ErrSave)-1], stub.ErrSave[:len(stub.ErrSave)-1]
	return err
}

// DeletePlayer delete player by specified ID
func (stub *Stub) DeletePlayer(ID string, tx Transact) error {
	var err error
	stub.mutex.Lock()
	defer stub.mutex.Unlock()

	delete(stub.players, ID)
	if len(stub.ErrDelete) == 0 {
		return nil
	}
	err, stub.ErrDelete = stub.ErrDelete[len(stub.ErrDelete)-1], stub.ErrDelete[:len(stub.ErrDelete)-1]
	return err
}

// NewTournament creates a new tournament with specified ID
func (stub *Stub) NewTournament(ID uint64, tx Transact) error {
	var err error
	stub.mutex.RLock()
	defer stub.mutex.RUnlock()
	_, ok := stub.tournaments[ID]
	if ok {
		return ErrAlreadyExist
	}
	stub.tournaments[ID] = model.Tournament{ID: ID, Bidders: make([]model.Bidder, 0)}
	if len(stub.ErrNew) == 0 {
		return nil
	}
	err, stub.ErrNew = stub.ErrNew[len(stub.ErrNew)-1], stub.ErrNew[:len(stub.ErrNew)-1]
	return err
}

// FindTournament finds existing tournament by specified ID
func (stub *Stub) FindTournament(ID uint64, tx Transact) (*model.Tournament, error) {
	var err error
	stub.mutex.RLock()
	defer stub.mutex.RUnlock()
	tournament, ok := stub.tournaments[ID]
	if !ok {
		return nil, ErrRecordNotFound
	}
	if len(stub.ErrFind) == 0 {
		return &tournament, nil
	}
	err, stub.ErrFind = stub.ErrFind[len(stub.ErrFind)-1], stub.ErrFind[:len(stub.ErrFind)-1]
	return &tournament, err
}

// SaveTournament saves a Tournament model
func (stub *Stub) SaveTournament(tournament *model.Tournament, tx Transact) error {
	var err error
	stub.mutex.Lock()
	defer stub.mutex.Unlock()

	stub.tournaments[tournament.ID] = *tournament
	if len(stub.ErrSave) == 0 {
		return nil
	}
	err, stub.ErrSave = stub.ErrSave[len(stub.ErrSave)-1], stub.ErrSave[:len(stub.ErrSave)-1]
	return err
}

// DeleteTournament delete tournament by specified ID
func (stub *Stub) DeleteTournament(ID uint64, tx Transact) error {
	var err error
	stub.mutex.Lock()
	defer stub.mutex.Unlock()

	delete(stub.tournaments, ID)
	if len(stub.ErrDelete) == 0 {
		return nil
	}
	err, stub.ErrDelete = stub.ErrDelete[len(stub.ErrDelete)-1], stub.ErrDelete[:len(stub.ErrDelete)-1]
	return err
}
