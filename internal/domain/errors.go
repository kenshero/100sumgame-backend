package domain

import "errors"

// Domain errors
var (
	ErrGameNotFound        = errors.New("game not found")
	ErrPuzzleNotFound      = errors.New("puzzle not found")
	ErrInvalidCell         = errors.New("invalid cell position")
	ErrCellIsPreFilled     = errors.New("cannot modify pre-filled cell")
	ErrInvalidValue        = errors.New("value must be between 1 and 99")
	ErrGameAlreadyComplete = errors.New("game is already completed")
	ErrTokensExhausted     = errors.New("tokens exhausted for this game")
	ErrNoPuzzlesAvailable  = errors.New("no puzzles available")
)
