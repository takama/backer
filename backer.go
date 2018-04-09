package backer

// Points can traded to goods and represent value like real money
type Points float32

// Player declares players methods
type Player interface {
	ID() string
	Fund(amount Points) error
	Take(amount Points) error
	Balance() (Points, error)
}

// Tournament declares tournament methods
type Tournament interface {
	Announce(deposit Points) error
	Join(players ...Player) error
	Result(winners map[Player]Points) error
}

// Service defines methods for service control
type Service interface {
	Reset() error
}
