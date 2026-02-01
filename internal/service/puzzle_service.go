package service

import (
	"context"
	"math/rand"

	"github.com/google/uuid"
	"github.com/kenshero/100sumgame/internal/domain"
	"github.com/kenshero/100sumgame/internal/repository"
)

// PuzzleService handles puzzle-related business logic
type PuzzleService struct {
	repo repository.PuzzleRepository
}

// NewPuzzleService creates a new puzzle service
func NewPuzzleService(repo repository.PuzzleRepository) *PuzzleService {
	return &PuzzleService{repo: repo}
}

// GetByID retrieves a puzzle by ID
func (s *PuzzleService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Puzzle, error) {
	return s.repo.GetByID(ctx, id)
}

// GetRandom retrieves a random puzzle from the pool
func (s *PuzzleService) GetRandom(ctx context.Context) (*domain.Puzzle, error) {
	puzzle, err := s.repo.GetRandom(ctx)
	if err != nil {
		return nil, domain.ErrNoPuzzlesAvailable
	}
	return puzzle, nil
}

// GetRandomForGuest retrieves a random puzzle that the guest hasn't played much
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

// Create creates a new puzzle
func (s *PuzzleService) Create(ctx context.Context, puzzle *domain.Puzzle) error {
	// Validate the puzzle solution
	if !domain.ValidatePuzzle(puzzle.GridSolution) {
		return domain.ErrInvalidValue
	}
	return s.repo.Create(ctx, puzzle)
}
