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

// PuzzleStatus represents the status of a puzzle for a guest
type PuzzleStatus string

const (
	PuzzleStatusAvailable PuzzleStatus = "AVAILABLE" // Not started yet
	PuzzleStatusPlaying   PuzzleStatus = "PLAYING"   // Currently playing
	PuzzleStatusCompleted PuzzleStatus = "COMPLETED" // Completed
	PuzzleStatusArchived  PuzzleStatus = "ARCHIVED"  // Archived after ad unlock
	PuzzleStatusAdBlock   PuzzleStatus = "AD_BLOCK"  // Available after watching ad
)

// PuzzleWithStatus represents a puzzle with its status for a guest
type PuzzleWithStatus struct {
	Puzzle *Puzzle
	Status PuzzleStatus
}

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

// GameResult represents game data for public API (excludes sensitive info like GridSolution)
type GameResult struct {
	ID            uuid.UUID  `json:"id"`
	GuestID       uuid.UUID  `json:"guest_id"`
	PuzzleID      uuid.UUID  `json:"puzzle_id"`
	GridCurrent   [][]Cell   `json:"grid_current"`
	TotalMistakes int        `json:"total_mistakes"`
	Status        GameStatus `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// SubmitAnswerResult represents result of submitting multiple answers
type SubmitAnswerResult struct {
	Game    *GameResult        `json:"game"`
	Results []CellVerifyResult `json:"results"`
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

// MoveResult represents the result of making a single move
type MoveResult struct {
	Game          *Game        `json:"game"`
	IsCorrect     bool         `json:"is_correct"`
	Feedback      CellFeedback `json:"feedback"`
	TotalMistakes int          `json:"total_mistakes"`
	IsGameOver    bool         `json:"is_game_over"`
}

// PuzzleStats represents statistics for a specific puzzle
type PuzzleStats struct {
	PuzzleID        uuid.UUID `json:"puzzle_id"`
	TotalPlayers    int       `json:"total_players"`
	TotalCompleted  int       `json:"total_completed"`
	AverageMistakes float64   `json:"average_mistakes"`
}

// PlayerStats represents statistics for a specific player
type PlayerStats struct {
	GuestID          uuid.UUID `json:"guest_id"`
	PuzzlesCompleted int       `json:"puzzles_completed"`
	TotalPuzzles     int       `json:"total_puzzles"`
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
