package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kenshero/100sumgame/internal/domain"
	"github.com/kenshero/100sumgame/internal/repository"
)

// GameService handles game-related business logic
type GameService struct {
	repo               repository.GameRepository
	puzzleService      *PuzzleService
	puzzleProgressRepo repository.PuzzleProgressRepository
	createGameLimiter  *OperationRateLimiter
	fillCellsLimiter   *OperationRateLimiter
	verifyGameLimiter  *OperationRateLimiter
}

// NewGameService creates a new game service
func NewGameService(repo repository.GameRepository, puzzleService *PuzzleService, puzzleProgressRepo repository.PuzzleProgressRepository) *GameService {
	return &GameService{
		repo:               repo,
		puzzleService:      puzzleService,
		puzzleProgressRepo: puzzleProgressRepo,
		createGameLimiter:  NewOperationRateLimiter(10, 1*time.Minute), // 10 games per minute per guest/IP
		fillCellsLimiter:   NewOperationRateLimiter(30, 1*time.Minute), // 30 fill operations per minute per guest/IP
		verifyGameLimiter:  NewOperationRateLimiter(10, 1*time.Minute), // 10 verifications per minute per guest/IP
	}
}

// CreateGame creates a new game with a random puzzle for guest
func (s *GameService) CreateGame(ctx context.Context, guestID uuid.UUID) (*domain.Game, error) {
	// Check rate limit BEFORE database operation
	if !s.createGameLimiter.AllowGuest(guestID) {
		return nil, domain.ErrRateLimitExceeded
	}

	// Get a random puzzle that guest hasn't completed yet
	puzzle, err := s.puzzleService.GetRandomForGuest(ctx, guestID)
	if err != nil {
		return nil, err
	}

	// Create new game from puzzle
	game := domain.NewGame(puzzle, guestID)

	// Save to database
	if err := s.repo.Create(ctx, game); err != nil {
		return nil, err
	}

	return game, nil
}

// GetByID retrieves a game by ID
func (s *GameService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Game, error) {
	game, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, domain.ErrGameNotFound
	}
	return game, nil
}

// FillCells fills cells with values
func (s *GameService) FillCells(ctx context.Context, gameID uuid.UUID, cells []domain.CellInput) (*domain.Game, error) {
	// Check rate limit BEFORE database operation
	if !s.fillCellsLimiter.AllowGuest(gameID) {
		return nil, domain.ErrRateLimitExceeded
	}

	game, err := s.GetByID(ctx, gameID)
	if err != nil {
		return nil, err
	}

	if game.Status == domain.StatusCompleted {
		return nil, domain.ErrGameAlreadyComplete
	}

	// Validate and fill cells
	for _, cell := range cells {
		if err := s.validateAndFillCell(game, cell); err != nil {
			return nil, err
		}
	}

	game.UpdatedAt = time.Now()

	// Save updated game
	if err := s.repo.Update(ctx, game); err != nil {
		return nil, err
	}

	return game, nil
}

func (s *GameService) validateAndFillCell(game *domain.Game, cell domain.CellInput) error {
	// Validate position
	if cell.Row < 0 || cell.Row >= 5 || cell.Col < 0 || cell.Col >= 5 {
		return domain.ErrInvalidCell
	}

	// Check if cell is pre-filled
	if game.GridCurrent[cell.Row][cell.Col].IsPreFilled {
		return domain.ErrCellIsPreFilled
	}

	// Validate value range
	if cell.Value < 1 || cell.Value > 99 {
		return domain.ErrInvalidValue
	}

	// Fill the cell
	game.GridCurrent[cell.Row][cell.Col].Value = &cell.Value
	game.GridCurrent[cell.Row][cell.Col].Feedback = domain.FeedbackNone

	return nil
}

// VerifyGame verifies all cells in the game
func (s *GameService) VerifyGame(ctx context.Context, gameID uuid.UUID) (*domain.Game, *domain.VerifyResult, error) {
	// Check rate limit BEFORE database operation
	if !s.verifyGameLimiter.AllowGuest(gameID) {
		return nil, nil, domain.ErrRateLimitExceeded
	}

	game, err := s.GetByID(ctx, gameID)
	if err != nil {
		return nil, nil, err
	}

	if game.Status == domain.StatusCompleted {
		return nil, nil, domain.ErrGameAlreadyComplete
	}

	result := &domain.VerifyResult{
		Results:    make([]domain.CellVerifyResult, 0),
		AllCorrect: true,
		Mistakes:   0,
	}

	// Check each non-prefilled cell
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			cell := &game.GridCurrent[i][j]

			// Skip pre-filled cells
			if cell.IsPreFilled {
				continue
			}

			// Skip empty cells
			if cell.Value == nil {
				result.AllCorrect = false
				continue
			}

			correctValue := game.GridSolution[i][j]
			currentValue := *cell.Value

			var feedback domain.CellFeedback
			if currentValue == correctValue {
				feedback = domain.FeedbackCorrect
			} else {
				if currentValue < correctValue {
					feedback = domain.FeedbackTooLow
					result.AllCorrect = false
					result.Mistakes++
				} else {
					feedback = domain.FeedbackTooHigh
					result.AllCorrect = false
					result.Mistakes++
				}
			}

			cell.Feedback = feedback
			result.Results = append(result.Results, domain.CellVerifyResult{
				Row:      i,
				Col:      j,
				Feedback: feedback,
			})
		}
	}

	// Update total mistakes
	game.TotalMistakes += result.Mistakes
	game.UpdatedAt = time.Now()

	// Check if game is complete
	if result.AllCorrect {
		game.Status = domain.StatusCompleted

		// Mark puzzle as completed for this guest
		if err := s.puzzleProgressRepo.MarkCompleted(ctx, game.GuestID, game.PuzzleID); err != nil {
			// Log error but don't fail the verification
			// The game is still saved, but progress tracking might have failed
			// This can be retried later if needed
		}
	}

	// Save updated game
	if err := s.repo.Update(ctx, game); err != nil {
		return nil, nil, err
	}

	return game, result, nil
}
