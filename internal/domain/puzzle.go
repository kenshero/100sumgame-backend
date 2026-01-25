package domain

import (
	"time"

	"github.com/google/uuid"
)

// Puzzle represents a puzzle template from the pool
type Puzzle struct {
	ID                 uuid.UUID  `json:"id"`
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
