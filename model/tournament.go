package model

import (
	"github.com/takama/backer"
)

// Tournament data model
type Tournament struct {
	ID         uint64        `json:"id"`
	Deposit    backer.Points `json:"deposit"`
	IsFinished bool          `json:"is_finished"`
	Bidders    []Bidder      `json:"bidders"`
}

// Bidder data model
type Bidder struct {
	ID      string          `json:"id"`
	Winner  bool            `json:"winner"`
	Prize   backer.Points   `json:"prize"`
	Backers []backer.Player `json:"backers"`
}
