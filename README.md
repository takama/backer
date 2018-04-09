# Backer

![Build Status](https://travis-ci.org/takama/backer.svg?branch=master)](https://travis-ci.org/takama/backer)
[![Contributions Welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat)](https://github.com/takama/backer/issues)
[![Go Report Card](https://goreportcard.com/badge/github.com/takama/backer)](https://goreportcard.com/report/github.com/takama/backer)

Backer service which is allowed players to back each other and get a part the prize in case of a win.

## Logic description

Each player holds certain amount of bonus points. Website funds its players with bonus points based on all kind of activity. Bonus points can traded to goods and represent value like real money.
One of the social products class is a social tournament. This is a competition between players in a multi-player game like poker, bingo, etc)
Entering a tournament requires a player to deposit certain amount of entry fee in bonus points. If a player has not enough point he can ask other players to back him and get a part the prize in case of a win.
In case of multiple backers, they submit equal part of the deposit and share the winning money in the same ration.

## Implementation in Go

Points

```go
// Points can traded to goods and represent value like real money
type Points float32
```

Players

```go
// Player declares players methods
type Player interface {
    ID() string
    Fund(amount Points) error
    Take(amount Points) error
    Balance() (Points, error)
}
```

Tournament

```go
// Tournament declares tournament methods
type Tournament interface {
    Announce(deposit Points) error
    Join(players Player...) error
    Result(winners map[Player]Points) error
}
```

Service

```go
// Service defines methods for service control
type Service interface {
    Reset() error
}
```
