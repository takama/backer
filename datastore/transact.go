package datastore

// Transact contains transaction control methods
type Transact interface {
	Commit() error
	Rollback() error
}
