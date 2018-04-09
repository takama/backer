package model

import (
	"github.com/takama/backer"
)

// Tournament data model
type Tournament struct {
	ID         uint64
	Deposit    backer.Points
	IsFinished bool
	Bidders    []Bidder
}

// Bidder data model
type Bidder struct {
	ID      string
	Winner  bool
	Prize   backer.Points
	Backers []backer.Player
}
