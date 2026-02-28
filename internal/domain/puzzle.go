package domain

import (
	"time"

	"github.com/google/uuid"
)

// GameSettings holds configurable game parameters
type GameSettings struct {
	StaminaMax                  int
	StaminaRegenIntervalMinutes int
	StaminaRegenAmount          int
	InitialScore                int
	ScoreDeductionPerMistake    int
	ScoreMinimum                int
}

// PuzzleSet represents a set/group of puzzles
type PuzzleSet struct {
	ID         uuid.UUID `json:"id"`
	SetOrder   int       `json:"set_order"`
	Difficulty string    `json:"difficulty"`
	CreatedAt  time.Time `json:"created_at"`
}

// GuestSetProgress tracks a guest's progress through a puzzle set
type GuestSetProgress struct {
	GuestID           uuid.UUID  `json:"guest_id"`
	SetID             uuid.UUID  `json:"set_id"`
	PuzzlesCompleted  int        `json:"puzzles_completed"`
	IsUnlocked        bool       `json:"is_unlocked"`
	IsCompleted       bool       `json:"is_completed"`
	UnlockedAt        *time.Time `json:"unlocked_at,omitempty"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`
	CurrentStamina    int        `json:"current_stamina"`
	LastStaminaUpdate time.Time  `json:"last_stamina_update"`
	CurrentScore      int        `json:"current_score"`
}

// Puzzle represents a puzzle template from the pool
type Puzzle struct {
	ID                 uuid.UUID  `json:"id"`
	SetID              *uuid.UUID `json:"set_id,omitempty"` // Foreign key to puzzle_sets
	GridSolution       [][]int    `json:"grid_solution"`
	PrefilledPositions []Position `json:"prefilled_positions"`
	Difficulty         string     `json:"difficulty"`
	CreatedAt          time.Time  `json:"created_at"`
}

// ValidatePuzzle checks if a puzzle solution is valid
// Each row and column must sum to 100
func ValidatePuzzle(grid [][]int) bool {
	size := len(grid)
	if size == 0 {
		return false
	}

	// Check each row
	for i := 0; i < size; i++ {
		if len(grid[i]) != size {
			return false
		}
		sum := 0
		for j := 0; j < size; j++ {
			if grid[i][j] < 1 || grid[i][j] > 99 {
				return false
			}
			sum += grid[i][j]
		}
		if sum != 100 {
			return false
		}
	}

	// Check each column
	for j := 0; j < size; j++ {
		sum := 0
		for i := 0; i < size; i++ {
			sum += grid[i][j]
		}
		if sum != 100 {
			return false
		}
	}

	return true
}
