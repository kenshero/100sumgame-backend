package domain

import (
	"time"

	"github.com/google/uuid"
)

// Cell represents a single cell in the grid
type Cell struct {
	Row         int          `json:"row"`
	Col         int          `json:"col"`
	Value       *int         `json:"value"`
	IsPreFilled bool         `json:"is_pre_filled"`
	Feedback    CellFeedback `json:"feedback,omitempty"`
}

// CellFeedback represents the feedback for a cell after verification
type CellFeedback string

const (
	FeedbackNone    CellFeedback = ""
	FeedbackCorrect CellFeedback = "CORRECT"
	FeedbackTooLow  CellFeedback = "TOO_LOW"
	FeedbackTooHigh CellFeedback = "TOO_HIGH"
)

// GameStatus represents the status of a game
type GameStatus string

const (
	StatusPlaying   GameStatus = "PLAYING"
	StatusCompleted GameStatus = "COMPLETED"
)

// Game represents a game session
type Game struct {
	ID                 uuid.UUID  `json:"id"`
	GuestID            uuid.UUID  `json:"guest_id"`
	PuzzleID           uuid.UUID  `json:"puzzle_id"`
	GridCurrent        [][]Cell   `json:"grid_current"`
	GridSolution       [][]int    `json:"grid_solution"`
	PrefilledPositions []Position `json:"prefilled_positions"`
	TotalMistakes      int        `json:"total_mistakes"`
	TokensUsed         int        `json:"tokens_used"`
	TokensLimit        int        `json:"tokens_limit"`
	Status             GameStatus `json:"status"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// Position represents a row/col position
type Position struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

// CellInput represents input for filling a cell
type CellInput struct {
	Row   int `json:"row"`
	Col   int `json:"col"`
	Value int `json:"value"`
}

// VerifyResult represents the result of verifying a game
type VerifyResult struct {
	Results    []CellVerifyResult `json:"results"`
	AllCorrect bool               `json:"all_correct"`
	Mistakes   int                `json:"mistakes"`
}

// CellVerifyResult represents the verification result for a single cell
type CellVerifyResult struct {
	Row      int          `json:"row"`
	Col      int          `json:"col"`
	Feedback CellFeedback `json:"feedback"`
}

// NewGame creates a new game from a puzzle
func NewGame(puzzle *Puzzle, guestID uuid.UUID) *Game {
	grid := make([][]Cell, 5)
	for i := range grid {
		grid[i] = make([]Cell, 5)
		for j := range grid[i] {
			grid[i][j] = Cell{
				Row:         i,
				Col:         j,
				Value:       nil,
				IsPreFilled: false,
			}
		}
	}

	// Set pre-filled cells
	for _, pos := range puzzle.PrefilledPositions {
		value := puzzle.GridSolution[pos.Row][pos.Col]
		grid[pos.Row][pos.Col].Value = &value
		grid[pos.Row][pos.Col].IsPreFilled = true
	}

	return &Game{
		ID:                 uuid.New(),
		GuestID:            guestID,
		PuzzleID:           puzzle.ID,
		GridCurrent:        grid,
		GridSolution:       puzzle.GridSolution,
		PrefilledPositions: puzzle.PrefilledPositions,
		TotalMistakes:      0,
		TokensUsed:         0,
		TokensLimit:        1000,
		Status:             StatusPlaying,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
}
