package service

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kenshero/100sumgame/internal/domain"
	"github.com/kenshero/100sumgame/internal/repository"
)

// PuzzleService handles puzzle-related business logic
type PuzzleService struct {
	repo                 repository.PuzzleRepository
	PuzzleProgressRepo   repository.PuzzleProgressRepository
	gameRepo             repository.GameRepository
	PuzzleSetRepo        repository.PuzzleSetRepository
	GuestSetProgressRepo repository.GuestSetProgressRepository
}

// NewPuzzleService creates a new puzzle service
func NewPuzzleService(
	repo repository.PuzzleRepository,
	puzzleProgressRepo repository.PuzzleProgressRepository,
	gameRepo repository.GameRepository,
	puzzleSetRepo repository.PuzzleSetRepository,
	guestSetProgressRepo repository.GuestSetProgressRepository,
) *PuzzleService {
	return &PuzzleService{
		repo:                 repo,
		PuzzleProgressRepo:   puzzleProgressRepo,
		gameRepo:             gameRepo,
		PuzzleSetRepo:        puzzleSetRepo,
		GuestSetProgressRepo: guestSetProgressRepo,
	}
}

// GetByID retrieves a puzzle by ID
func (s *PuzzleService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Puzzle, error) {
	return s.repo.GetByID(ctx, id)
}

// GetRandom retrieves a random puzzle from pool
func (s *PuzzleService) GetRandom(ctx context.Context) (*domain.Puzzle, error) {
	puzzle, err := s.repo.GetRandom(ctx)
	if err != nil {
		return nil, domain.ErrNoPuzzlesAvailable
	}
	return puzzle, nil
}

// GetRandomForGuest retrieves a random puzzle that guest hasn't played much
// Returns puzzles sorted by least played by this guest
func (s *PuzzleService) GetRandomForGuest(ctx context.Context, guestID uuid.UUID) (*domain.Puzzle, error) {
	puzzles, err := s.repo.GetAvailablePuzzlesForGuest(ctx, guestID)
	if err != nil {
		return nil, domain.ErrNoPuzzlesAvailable
	}
	if len(puzzles) == 0 {
		return nil, domain.ErrNoPuzzlesAvailable
	}
	// Random เลือกจาก top 10 ที่เล่นน้อยสุด
	// ถ้ามีน้อยกว่า 10 ก็เลือกจากทั้งหมด
	maxPuzzles := 10
	if len(puzzles) < maxPuzzles {
		maxPuzzles = len(puzzles)
	}

	// Random เลือก index
	randomIndex := rand.Intn(maxPuzzles)
	return puzzles[randomIndex], nil
}

// GetAll retrieves all puzzles
func (s *PuzzleService) GetAll(ctx context.Context) ([]*domain.Puzzle, error) {
	return s.repo.GetAll(ctx)
}

// GetTotalCount retrieves total number of puzzles
func (s *PuzzleService) GetTotalCount(ctx context.Context) (int, error) {
	return s.repo.GetTotalCount(ctx)
}

// GetPuzzlesBySet retrieves all puzzles belonging to a specific set
func (s *PuzzleService) GetPuzzlesBySet(ctx context.Context, setID uuid.UUID) ([]*domain.Puzzle, error) {
	return s.repo.GetPuzzlesBySet(ctx, setID)
}

// GetAvailableForGuest retrieves all available puzzles for a guest with their status
// Returns puzzles ordered by ID (not random) with status for each puzzle
// Limit parameter controls maximum number of puzzles returned
func (s *PuzzleService) GetAvailableForGuest(ctx context.Context, guestID uuid.UUID, limit int) ([]*domain.PuzzleWithStatus, error) {
	if limit <= 0 {
		limit = 20 // Default limit
	}

	puzzles, err := s.PuzzleProgressRepo.GetAvailablePuzzlesForGuest(ctx, guestID, limit)
	if err != nil {
		return nil, domain.ErrNoPuzzlesAvailable
	}

	if len(puzzles) == 0 {
		return []*domain.PuzzleWithStatus{}, nil
	}

	for _, puzzleWithStatus := range puzzles {
		if puzzleWithStatus.Puzzle == nil {
			continue
		}

		game, err := s.gameRepo.GetByGuestAndPuzzle(ctx, guestID, puzzleWithStatus.Puzzle.ID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}
			return nil, err
		}

		if puzzleWithStatus.Status == domain.PuzzleStatusAvailable && game.Status == domain.StatusPlaying {
			puzzleWithStatus.Status = domain.PuzzleStatusPlaying
		}

		// Extract solved positions from game session; merge with DB data as fallback
		gameSolved := extractSolvedPositions(game)
		if len(gameSolved) > 0 {
			puzzleWithStatus.SolvedPositions = gameSolved
		}
		// If gameSolved is empty, keep SolvedPositions from DB (guest_puzzle_progress.solved_positions)
	}

	return puzzles, nil
}

func extractSolvedPositions(game *domain.Game) []domain.Position {
	if game == nil || len(game.GridCurrent) == 0 || len(game.GridSolution) == 0 {
		return nil
	}

	solved := make([]domain.Position, 0)
	for rowIdx, row := range game.GridCurrent {
		if rowIdx >= len(game.GridSolution) {
			continue
		}

		for colIdx, cell := range row {
			if cell.IsPreFilled || cell.Value == nil {
				continue
			}

			if colIdx >= len(game.GridSolution[rowIdx]) {
				continue
			}

			if *cell.Value == game.GridSolution[rowIdx][colIdx] {
				solved = append(solved, domain.Position{Row: rowIdx, Col: colIdx})
			}
		}
	}

	return solved
}

// Create creates a new puzzle
func (s *PuzzleService) Create(ctx context.Context, puzzle *domain.Puzzle) error {
	// Validate puzzle solution
	if !domain.ValidatePuzzle(puzzle.GridSolution) {
		return domain.ErrInvalidValue
	}
	return s.repo.Create(ctx, puzzle)
}

// GetCurrentSet gets current unlocked set for a guest
// If no unlocked set exists, it creates and unlocks first set automatically
func (s *PuzzleService) GetCurrentSet(ctx context.Context, guestID uuid.UUID) (*domain.GuestSetProgress, error) {
	// Try to get unlocked set
	setProgress, err := s.GuestSetProgressRepo.GetUnlockedSet(ctx, guestID)
	if err == nil && setProgress != nil {
		return setProgress, nil
	}

	// If no unlocked set, create and unlock first set
	firstSet, err := s.PuzzleSetRepo.GetByOrder(ctx, 1)
	if err != nil {
		return nil, domain.ErrNoPuzzlesAvailable
	}

	now := time.Now()
	newProgress := &domain.GuestSetProgress{
		GuestID:          guestID,
		SetID:            firstSet.ID,
		PuzzlesCompleted: 0,
		IsUnlocked:       true,
		IsCompleted:      false,
		UnlockedAt:       &now,
	}

	if err := s.GuestSetProgressRepo.Create(ctx, newProgress); err != nil {
		return nil, err
	}

	return newProgress, nil
}

// GetAvailableForGuestWithSets retrieves available puzzles for a guest based on their current unlocked set
// Returns puzzles from current unlocked set, ordered by set_order and puzzle id
func (s *PuzzleService) GetAvailableForGuestWithSets(ctx context.Context, guestID uuid.UUID, limit int) ([]*domain.PuzzleWithStatus, error) {
	if limit <= 0 {
		limit = 20
	}

	// Get current unlocked set for guest
	setProgress, err := s.GetCurrentSet(ctx, guestID)
	if err != nil {
		return nil, err
	}

	// Get puzzles from this set
	// We need to modify puzzle progress repo to filter by set_id
	// For now, let's use a different approach
	puzzles, err := s.PuzzleProgressRepo.GetAvailablePuzzlesForGuest(ctx, guestID, limit)
	if err != nil {
		return nil, domain.ErrNoPuzzlesAvailable
	}

	if len(puzzles) == 0 {
		return []*domain.PuzzleWithStatus{}, nil
	}

	// Filter puzzles to only include those from current set
	filteredPuzzles := make([]*domain.PuzzleWithStatus, 0)
	for _, puzzleWithStatus := range puzzles {
		if puzzleWithStatus.Puzzle == nil {
			continue
		}
		// Check if puzzle belongs to current set
		if puzzleWithStatus.Puzzle.SetID != nil && *puzzleWithStatus.Puzzle.SetID == setProgress.SetID {
			filteredPuzzles = append(filteredPuzzles, puzzleWithStatus)
		}
	}

	// Update playing status and solved positions
	for _, puzzleWithStatus := range filteredPuzzles {
		if puzzleWithStatus.Puzzle == nil {
			continue
		}

		game, err := s.gameRepo.GetByGuestAndPuzzle(ctx, guestID, puzzleWithStatus.Puzzle.ID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}
			return nil, err
		}

		if puzzleWithStatus.Status == domain.PuzzleStatusAvailable && game.Status == domain.StatusPlaying {
			puzzleWithStatus.Status = domain.PuzzleStatusPlaying
		}

		// Extract solved positions from game session
		gameSolved := extractSolvedPositions(game)
		if len(gameSolved) > 0 {
			puzzleWithStatus.SolvedPositions = gameSolved
		}
	}

	return filteredPuzzles, nil
}

// UnlockNextSet unlocks next set for a guest after watching an ad
func (s *PuzzleService) UnlockNextSet(ctx context.Context, guestID uuid.UUID) (*domain.GuestSetProgress, error) {
	// Get current unlocked set to find its order
	currentSetProgress, err := s.GuestSetProgressRepo.GetUnlockedSet(ctx, guestID)
	if err != nil {
		return nil, err
	}

	// Get current set to find its order
	currentSet, err := s.PuzzleSetRepo.GetByID(ctx, currentSetProgress.SetID)
	if err != nil {
		return nil, err
	}

	// Get next set
	nextSet, err := s.PuzzleSetRepo.GetNextSet(ctx, currentSet.SetOrder)
	if err != nil {
		return nil, domain.ErrNoMoreSetsAvailable
	}

	// Create progress for next set
	now := time.Now()
	newProgress := &domain.GuestSetProgress{
		GuestID:          guestID,
		SetID:            nextSet.ID,
		PuzzlesCompleted: 0,
		IsUnlocked:       true,
		IsCompleted:      false,
		UnlockedAt:       &now,
	}

	if err := s.GuestSetProgressRepo.Create(ctx, newProgress); err != nil {
		return nil, err
	}

	return newProgress, nil
}

// IncrementSetProgress increments puzzle completion count for a guest's current set
func (s *PuzzleService) IncrementSetProgress(ctx context.Context, guestID, puzzleID uuid.UUID) error {
	// Get puzzle to find which set it belongs to
	puzzle, err := s.repo.GetByID(ctx, puzzleID)
	if err != nil {
		return err
	}

	if puzzle.SetID == nil {
		return nil // Puzzle doesn't belong to any set
	}

	// Get current progress for this set
	progress, err := s.GuestSetProgressRepo.GetByGuestAndSet(ctx, guestID, *puzzle.SetID)
	if err != nil {
		return err
	}

	// Increment count
	newCount := progress.PuzzlesCompleted + 1
	if err := s.GuestSetProgressRepo.UpdateProgress(ctx, guestID, *puzzle.SetID, newCount); err != nil {
		return err
	}

	// Check if set is complete (10 puzzles)
	if newCount >= 10 {
		return s.GuestSetProgressRepo.MarkCompleted(ctx, guestID, *puzzle.SetID)
	}

	return nil
}
