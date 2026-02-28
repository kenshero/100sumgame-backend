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
	repo                 repository.GameRepository
	puzzleService        *PuzzleService
	puzzleProgressRepo   repository.PuzzleProgressRepository
	guestSetProgressRepo repository.GuestSetProgressRepository
	configService        *ConfigService
	createGameLimiter    *OperationRateLimiter
	fillCellsLimiter     *OperationRateLimiter
	makeMoveLimiter      *OperationRateLimiter
	verifyGameLimiter    *OperationRateLimiter
	submitAnswerLimiter  *OperationRateLimiter
}

// NewGameService creates a new game service
func NewGameService(repo repository.GameRepository, puzzleService *PuzzleService, puzzleProgressRepo repository.PuzzleProgressRepository, guestSetProgressRepo repository.GuestSetProgressRepository, configService *ConfigService) *GameService {
	return &GameService{
		repo:                 repo,
		puzzleService:        puzzleService,
		puzzleProgressRepo:   puzzleProgressRepo,
		guestSetProgressRepo: guestSetProgressRepo,
		configService:        configService,
		createGameLimiter:    NewOperationRateLimiter(10, 1*time.Minute), // 10 games per minute per guest/IP
		fillCellsLimiter:     NewOperationRateLimiter(30, 1*time.Minute), // 30 fill operations per minute per guest/IP
		makeMoveLimiter:      NewOperationRateLimiter(30, 1*time.Minute), // 30 moves per minute per guest/IP
		verifyGameLimiter:    NewOperationRateLimiter(10, 1*time.Minute), // 10 verifications per minute per guest/IP
		submitAnswerLimiter:  NewOperationRateLimiter(30, 1*time.Minute), // 30 submissions per minute per guest/IP
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

// GetByGuestAndPuzzle retrieves the latest game session for a guest and puzzle
func (s *GameService) GetByGuestAndPuzzle(ctx context.Context, guestID, puzzleID uuid.UUID) (*domain.Game, error) {
	game, err := s.repo.GetByGuestAndPuzzle(ctx, guestID, puzzleID)
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

// MakeMove makes a single move - fills a cell and immediately verifies it
func (s *GameService) MakeMove(ctx context.Context, gameID uuid.UUID, row, col, value int) (*domain.MoveResult, error) {
	// Check rate limit BEFORE database operation
	if !s.makeMoveLimiter.AllowGuest(gameID) {
		return nil, domain.ErrRateLimitExceeded
	}

	game, err := s.GetByID(ctx, gameID)
	if err != nil {
		return nil, err
	}

	if game.Status == domain.StatusCompleted {
		return nil, domain.ErrGameAlreadyComplete
	}

	// Validate and fill the cell
	cellInput := domain.CellInput{
		Row:   row,
		Col:   col,
		Value: value,
	}

	if err := s.validateAndFillCell(game, cellInput); err != nil {
		return nil, err
	}

	// Verify the cell immediately
	correctValue := game.GridSolution[row][col]
	var feedback domain.CellFeedback
	var isCorrect bool

	if value == correctValue {
		feedback = domain.FeedbackCorrect
		isCorrect = true
	} else if value < correctValue {
		feedback = domain.FeedbackTooLow
		isCorrect = false
		game.TotalMistakes++
	} else {
		feedback = domain.FeedbackTooHigh
		isCorrect = false
		game.TotalMistakes++
	}

	// Update cell feedback
	game.GridCurrent[row][col].Feedback = feedback

	// Check if game is complete (all cells filled correctly)
	allCorrect := s.checkGameComplete(game)
	var isGameOver bool

	if allCorrect {
		game.Status = domain.StatusCompleted
		isGameOver = true

		// Extract solved positions and mark puzzle as completed for this guest
		solvedPositions := extractSolvedPositions(game)
		if err := s.puzzleProgressRepo.MarkCompleted(ctx, game.GuestID, game.PuzzleID, solvedPositions); err != nil {
			// Log error but don't fail the move
			// The game is still saved, but progress tracking might have failed
		}
	} else {
		isGameOver = false
	}

	game.UpdatedAt = time.Now()

	// Save updated game
	if err := s.repo.Update(ctx, game); err != nil {
		return nil, err
	}

	return &domain.MoveResult{
		Game:          game,
		IsCorrect:     isCorrect,
		Feedback:      feedback,
		TotalMistakes: game.TotalMistakes,
		IsGameOver:    isGameOver,
	}, nil
}

// checkGameComplete checks if all cells are filled correctly
func (s *GameService) checkGameComplete(game *domain.Game) bool {
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			cell := &game.GridCurrent[i][j]

			// Skip pre-filled cells (they're already correct)
			if cell.IsPreFilled {
				continue
			}

			// Check if cell is filled
			if cell.Value == nil {
				return false
			}

			// Check if value is correct
			if *cell.Value != game.GridSolution[i][j] {
				return false
			}
		}
	}
	return true
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

		// Extract solved positions and mark puzzle as completed for this guest
		solvedPositions := extractSolvedPositions(game)
		if err := s.puzzleProgressRepo.MarkCompleted(ctx, game.GuestID, game.PuzzleID, solvedPositions); err != nil {
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

// GetPuzzleStats retrieves statistics for a specific puzzle
func (s *GameService) GetPuzzleStats(ctx context.Context, puzzleID uuid.UUID) (*domain.PuzzleStats, error) {
	stats, err := s.repo.GetPuzzleStats(ctx, puzzleID)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

// GetPlayerStats retrieves statistics for a specific player
func (s *GameService) GetPlayerStats(ctx context.Context, guestID uuid.UUID) (*domain.PlayerStats, error) {
	// Get number of puzzles completed by this guest
	completedCount, err := s.puzzleProgressRepo.GetCompletedCount(ctx, guestID)
	if err != nil {
		return nil, err
	}

	// Get total number of puzzles in the system
	totalPuzzles, err := s.puzzleService.GetTotalCount(ctx)
	if err != nil {
		return nil, err
	}

	return &domain.PlayerStats{
		GuestID:          guestID,
		PuzzlesCompleted: completedCount,
		TotalPuzzles:     totalPuzzles,
	}, nil
}

// SubmitAnswer submits multiple answers for a specific puzzle
func (s *GameService) SubmitAnswer(ctx context.Context, guestID, puzzleID uuid.UUID, answers []domain.CellInput) (*domain.SubmitAnswerResult, error) {
	// Check rate limit BEFORE database operation
	if !s.submitAnswerLimiter.AllowGuest(guestID) {
		return nil, domain.ErrRateLimitExceeded
	}

	// Get game settings for stamina and score configuration
	settings := s.configService.GetSettings()

	// Get puzzle to determine which set it belongs to
	puzzle, err := s.puzzleService.GetByID(ctx, puzzleID)
	if err != nil {
		return nil, err
	}

	// Check if puzzle has a set
	if puzzle.SetID == nil {
		return nil, domain.ErrNoMoreSetsAvailable
	}

	setID := *puzzle.SetID

	// Get guest's set progress
	setProgress, err := s.guestSetProgressRepo.GetByGuestAndSet(ctx, guestID, setID)
	if err != nil {
		return nil, err
	}

	// Regenerate stamina based on time elapsed
	newStamina, newStaminaTime, err := s.guestSetProgressRepo.RegenerateStamina(
		ctx,
		guestID,
		setID,
		setProgress.CurrentStamina,
		settings.StaminaMax,
		settings.StaminaRegenIntervalMinutes,
		settings.StaminaRegenAmount,
		setProgress.LastStaminaUpdate,
	)
	if err != nil {
		return nil, err
	}

	// Update set progress with regenerated stamina
	setProgress.CurrentStamina = newStamina
	setProgress.LastStaminaUpdate = newStaminaTime

	// Check if player has enough stamina (at least 1)
	if setProgress.CurrentStamina < 1 {
		return nil, domain.ErrInsufficientStamina
	}

	// Try to find existing game session for this guest + puzzle
	game, err := s.repo.GetByGuestAndPuzzle(ctx, guestID, puzzleID)

	// If no game session exists, create a new one
	if err != nil {
		// Create new game session
		game = domain.NewGame(puzzle, guestID)
		if err := s.repo.Create(ctx, game); err != nil {
			return nil, err
		}
	}

	// Check if game is already completed
	if game.Status == domain.StatusCompleted {
		return nil, domain.ErrGameAlreadyComplete
	}

	// Mark puzzle as playing for this guest (best-effort)
	if err := s.puzzleProgressRepo.MarkPlaying(ctx, guestID, puzzleID); err != nil {
		// Don't fail submission if progress tracking update fails
	}

	// Process each answer
	result := &domain.SubmitAnswerResult{
		Results: make([]domain.CellVerifyResult, 0),
	}
	mistakeCount := 0

	for _, answer := range answers {
		// Validate and fill cell
		if err := s.validateAndFillCell(game, answer); err != nil {
			return nil, err
		}

		// Verify the cell against solution
		correctValue := game.GridSolution[answer.Row][answer.Col]
		currentValue := answer.Value

		var feedback domain.CellFeedback
		if currentValue == correctValue {
			feedback = domain.FeedbackCorrect
		} else {
			if currentValue < correctValue {
				feedback = domain.FeedbackTooLow
				game.TotalMistakes++
				mistakeCount++
			} else {
				feedback = domain.FeedbackTooHigh
				game.TotalMistakes++
				mistakeCount++
			}
		}

		// Update cell feedback
		game.GridCurrent[answer.Row][answer.Col].Feedback = feedback

		// Add to results
		result.Results = append(result.Results, domain.CellVerifyResult{
			Row:      answer.Row,
			Col:      answer.Col,
			Feedback: feedback,
		})
	}

	// Deduct stamina (1 per submission)
	if err := s.guestSetProgressRepo.DeductStamina(ctx, guestID, setID); err != nil {
		return nil, err
	}

	// Deduct score based on mistakes (10 points per mistake)
	if mistakeCount > 0 {
		scoreDeduction := mistakeCount * settings.ScoreDeductionPerMistake
		if err := s.guestSetProgressRepo.DeductScore(ctx, guestID, setID, scoreDeduction, settings.ScoreMinimum); err != nil {
			// Log error but don't fail the submission
		}
	}

	game.UpdatedAt = time.Now()

	// Extract current solved positions (all correctly answered cells so far)
	solvedPositions := extractSolvedPositions(game)

	// Check if all cells are now correct and complete
	if s.checkGameComplete(game) {
		game.Status = domain.StatusCompleted

		// Mark puzzle as completed for this guest with final solved positions
		if err := s.puzzleProgressRepo.MarkCompleted(ctx, guestID, puzzleID, solvedPositions); err != nil {
			// Log error but don't fail the submission
		}

		// Update set progress for this guest
		s.updateSetProgress(ctx, guestID, game.PuzzleID)
	} else {
		// Game still in progress — persist solved positions so they survive browser refresh
		if len(solvedPositions) > 0 {
			if err := s.puzzleProgressRepo.UpdateSolvedPositions(ctx, guestID, puzzleID, solvedPositions); err != nil {
				// Log error but don't fail the submission
			}
		}
	}

	// Save updated game
	if err := s.repo.Update(ctx, game); err != nil {
		return nil, err
	}

	// Create game result (without solution)
	gameResult := &domain.GameResult{
		ID:            game.ID,
		GuestID:       game.GuestID,
		PuzzleID:      game.PuzzleID,
		GridCurrent:   game.GridCurrent,
		TotalMistakes: game.TotalMistakes,
		Status:        game.Status,
		CreatedAt:     game.CreatedAt,
		UpdatedAt:     game.UpdatedAt,
	}

	result.Game = gameResult
	return result, nil
}

// updateSetProgress updates the guest's set progress when a puzzle is completed
func (s *GameService) updateSetProgress(ctx context.Context, guestID, puzzleID uuid.UUID) {
	// Get puzzle to find which set it belongs to
	puzzle, err := s.puzzleService.GetByID(ctx, puzzleID)
	if err != nil {
		// Log error but don't fail
		return
	}

	// SetID is a pointer, check if it's nil
	if puzzle.SetID == nil {
		return
	}

	setID := *puzzle.SetID

	// Get all puzzles in this set
	setPuzzles, err := s.puzzleService.GetPuzzlesBySet(ctx, setID)
	if err != nil {
		// Log error but don't fail
		return
	}

	// Count how many puzzles in this set are completed by this guest
	completedCount := 0
	for _, p := range setPuzzles {
		hasCompleted, err := s.puzzleProgressRepo.HasCompleted(ctx, guestID, p.ID)
		if err != nil {
			continue
		}
		if hasCompleted {
			completedCount++
		}
	}

	// Update set progress
	if err := s.guestSetProgressRepo.UpdateProgress(ctx, guestID, setID, completedCount); err != nil {
		// Log error but don't fail
		return
	}

	// If all puzzles in set are completed (10 puzzles), mark the set as completed
	if completedCount >= 10 {
		if err := s.guestSetProgressRepo.MarkCompleted(ctx, guestID, setID); err != nil {
			// Log error but don't fail
		}
	}
}
