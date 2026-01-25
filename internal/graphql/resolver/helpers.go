package resolver

import (
	"fmt"

	"github.com/kenshero/100sumgame/internal/domain"
	"github.com/kenshero/100sumgame/internal/graphql/model"
)

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
		ID:              game.ID.String(),
		Grid:            grid,
		TotalMistakes:   game.TotalMistakes,
		Status:          model.GameStatus(game.Status),
		TokensUsed:      game.TokensUsed,
		TokensRemaining: game.TokensLimit - game.TokensUsed,
		CreatedAt:       game.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// Helper function to convert domain Puzzle to GraphQL model
func domainPuzzleToModel(puzzle *domain.Puzzle) *model.Puzzle {
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

	// Fill in only the prefilled positions with their solution values
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			if prefilledMap[fmt.Sprintf("%d,%d", i, j)] {
				grid[i][j] = puzzle.GridSolution[i][j]
			}
		}
	}

	return &model.Puzzle{
		ID:        puzzle.ID.String(),
		Grid:      grid,
		Solution:  puzzle.GridSolution,
		CreatedAt: puzzle.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
