package datastore

// Store contains DB store control methods
type Store interface {
	Ready() bool
	Reset() error
	MigrateUp() error
	MigrateDown() error
}
