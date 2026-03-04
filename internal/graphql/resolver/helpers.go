package resolver

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kenshero/100sumgame/internal/config"
	"github.com/kenshero/100sumgame/internal/domain"
	"github.com/kenshero/100sumgame/internal/graphql/model"
	"github.com/kenshero/100sumgame/internal/middleware"
)

// contextKey is a custom type for context keys
type contextKey string

// RequestContextKey is the key used to store HTTP request in context
const RequestContextKey contextKey = "http_request"

// checkAdminToken validates admin secret token from context
func checkAdminToken(ctx context.Context) error {
	// Get HTTP request from context
	req, ok := ctx.Value(RequestContextKey).(*http.Request)
	if !ok || req == nil {
		return fmt.Errorf("unable to get request from context")
	}

	authHeader := req.Header.Get("Authorization")
	if authHeader == "" {
		return fmt.Errorf("missing authorization header")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == authHeader {
		return fmt.Errorf("invalid authorization header format. Expected: 'Bearer <token>'")
	}

	if token != config.AdminSecretToken {
		return fmt.Errorf("invalid admin token")
	}

	return nil
}

func resolveGuestUUID(ctx context.Context, guestID string) (uuid.UUID, error) {
	if guestID != "" {
		if parsedGuestID, err := uuid.Parse(guestID); err == nil {
			return parsedGuestID, nil
		}
	}

	session := middleware.GetSessionFromContext(ctx)
	if session == nil {
		return uuid.Nil, fmt.Errorf("session not found")
	}

	guestUUID, err := uuid.Parse(session.GuestID)
	if err != nil {
		return uuid.Nil, err
	}

	return guestUUID, nil
}

// Helper function to convert domain Game to GraphQL model
func domainGameToModel(game *domain.Game) *model.Game {
	grid := make([][]*model.Cell, len(game.GridCurrent))
	for i, row := range game.GridCurrent {
		grid[i] = make([]*model.Cell, len(row))
		for j, cell := range row {
			var feedback *model.CellFeedback
			if cell.Feedback != "" {
				f := model.CellFeedback(cell.Feedback)
				feedback = &f
			}

			grid[i][j] = &model.Cell{
				Row:         cell.Row,
				Col:         cell.Col,
				Value:       cell.Value,
				IsPreFilled: cell.IsPreFilled,
				Feedback:    feedback,
			}
		}
	}

	return &model.Game{
		ID:            game.ID.String(),
		GuestID:       game.GuestID.String(),
		PuzzleID:      game.PuzzleID.String(),
		Grid:          grid,
		TotalMistakes: game.TotalMistakes,
		Status:        model.GameStatus(game.Status),
		CreatedAt:     game.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// Helper function to convert domain Puzzle to GraphQL model
func domainPuzzleToModel(puzzle *domain.Puzzle) *model.Puzzle {
	return domainPuzzleToModelWithSolved(puzzle, nil)
}

func domainPuzzleToModelWithSolved(puzzle *domain.Puzzle, solvedPositions []domain.Position) *model.Puzzle {
	// Create Grid with only prefilled positions showing values (rest are 0)
	size := len(puzzle.GridSolution)
	grid := make([][]int, size)
	for i := 0; i < size; i++ {
		grid[i] = make([]int, size)
		// Initialize all to 0
	}

	// Create a map of prefilled positions for quick lookup
	prefilledMap := make(map[string]bool)
	for _, pos := range puzzle.PrefilledPositions {
		prefilledMap[fmt.Sprintf("%d,%d", pos.Row, pos.Col)] = true
	}

	// Create map of solved positions from guest progress
	solvedMap := make(map[string]bool)
	for _, pos := range solvedPositions {
		solvedMap[fmt.Sprintf("%d,%d", pos.Row, pos.Col)] = true
	}

	// Fill in prefilled and solved positions with their solution values
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			key := fmt.Sprintf("%d,%d", i, j)
			if prefilledMap[key] || solvedMap[key] {
				grid[i][j] = puzzle.GridSolution[i][j]
			}
		}
	}

	// Build prefilledPositions: include BOTH original prefilled AND solved positions
	// so the frontend knows which cells are filled and read-only
	allPositions := make([]*model.Position, 0, len(puzzle.PrefilledPositions)+len(solvedPositions))
	for _, pos := range puzzle.PrefilledPositions {
		allPositions = append(allPositions, &model.Position{Row: pos.Row, Col: pos.Col})
	}
	for _, pos := range solvedPositions {
		allPositions = append(allPositions, &model.Position{Row: pos.Row, Col: pos.Col})
	}

	return &model.Puzzle{
		ID:                 puzzle.ID.String(),
		Grid:               grid,
		PrefilledPositions: allPositions,
		CreatedAt:          puzzle.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// Helper function to convert domain grid to GraphQL model grid
func domainGridToModel(grid [][]domain.Cell) [][]*model.Cell {
	modelGrid := make([][]*model.Cell, len(grid))
	for i, row := range grid {
		modelGrid[i] = make([]*model.Cell, len(row))
		for j, cell := range row {
			var feedback *model.CellFeedback
			if cell.Feedback != "" {
				f := model.CellFeedback(cell.Feedback)
				feedback = &f
			}

			modelGrid[i][j] = &model.Cell{
				Row:         cell.Row,
				Col:         cell.Col,
				Value:       cell.Value,
				IsPreFilled: cell.IsPreFilled,
				Feedback:    feedback,
			}
		}
	}
	return modelGrid
}

// Helper function to convert domain PuzzleWithStatus to GraphQL model PuzzleWithStatus
func domainPuzzleWithStatusToModel(puzzleWithStatus *domain.PuzzleWithStatus) *model.PuzzleWithStatus {
	puzzle := puzzleWithStatus.Puzzle
	puzzleModel := domainPuzzleToModelWithSolved(puzzle, puzzleWithStatus.SolvedPositions)
	status := model.PuzzleStatus(puzzleWithStatus.Status)

	// Convert solved positions to model
	solvedPositions := make([]*model.Position, len(puzzleWithStatus.SolvedPositions))
	for i, pos := range puzzleWithStatus.SolvedPositions {
		solvedPositions[i] = &model.Position{
			Row: pos.Row,
			Col: pos.Col,
		}
	}

	// Get set order from puzzle set - we'll need to fetch this
	// For now, return 0 as default
	setOrder := 0

	return &model.PuzzleWithStatus{
		Puzzle:          puzzleModel,
		Status:          status,
		SolvedPositions: solvedPositions,
		SetOrder:        setOrder,
	}
}

// Helper function to convert domain GuestSetProgress to GraphQL model
func domainGuestSetProgressToModel(progress *domain.GuestSetProgress) *model.GuestSetProgress {
	unlockedAt := formatTimePtr(progress.UnlockedAt)
	completedAt := formatTimePtr(progress.CompletedAt)

	return &model.GuestSetProgress{
		GuestID:          progress.GuestID.String(),
		SetID:            progress.SetID.String(),
		PuzzlesCompleted: progress.PuzzlesCompleted,
		IsUnlocked:       progress.IsUnlocked,
		IsCompleted:      progress.IsCompleted,
		IsLastSet:        progress.IsLastSet,
		UnlockedAt:       &unlockedAt,
		CompletedAt:      &completedAt,
		CurrentStamina:   progress.CurrentStamina,
		CurrentScore:     progress.CurrentScore,
	}
}

// Helper function to format time.Time pointer to ISO string
func formatTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02T15:04:05Z07:00")
}
