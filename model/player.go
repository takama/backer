package model

import (
	"github.com/takama/backer"
)

// Player data model
type Player struct {
	ID      string
	Balance backer.Points
}
