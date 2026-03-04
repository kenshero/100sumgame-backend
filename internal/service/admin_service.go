package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/kenshero/100sumgame/internal/domain"
	"github.com/kenshero/100sumgame/internal/repository"
)

// AdminService handles admin operations
type AdminService struct {
	configService        *ConfigService
	puzzleSetRepo        repository.PuzzleSetRepository
	puzzleRepo           repository.PuzzleRepository
	puzzleProgressRepo   repository.PuzzleProgressRepository
	guestSetProgressRepo repository.GuestSetProgressRepository
}

// NewAdminService creates a new admin service
func NewAdminService(
	configService *ConfigService,
	puzzleSetRepo repository.PuzzleSetRepository,
	puzzleRepo repository.PuzzleRepository,
	puzzleProgressRepo repository.PuzzleProgressRepository,
	guestSetProgressRepo repository.GuestSetProgressRepository,
) *AdminService {
	return &AdminService{
		configService:        configService,
		puzzleSetRepo:        puzzleSetRepo,
		puzzleRepo:           puzzleRepo,
		puzzleProgressRepo:   puzzleProgressRepo,
		guestSetProgressRepo: guestSetProgressRepo,
	}
}

// RefreshConfig refreshes game configuration from database
// This should be called after updating database settings
func (s *AdminService) RefreshConfig() (*domain.GameSettings, error) {
	if err := s.configService.RefreshSettings(); err != nil {
		return nil, err
	}
	return s.configService.GetSettings(), nil
}

// ForceCompletePuzzles forces completion of all puzzles in the current unlocked set for a guest
// This is an admin-only operation for testing purposes
func (s *AdminService) ForceCompletePuzzles(ctx context.Context, guestID uuid.UUID) (*domain.GuestSetProgress, error) {
	// Get the current unlocked set for this guest
	setProgress, err := s.guestSetProgressRepo.GetUnlockedSet(ctx, guestID)
	if err != nil {
		return nil, errors.New("no unlocked set found for this guest")
	}

	// Check if set is already completed
	if setProgress.IsCompleted {
		return nil, errors.New("set is already completed")
	}

	// Get all puzzles in this set
	puzzles, err := s.puzzleRepo.GetPuzzlesBySet(ctx, setProgress.SetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get puzzles for set: %w", err)
	}

	// Mark all puzzles as completed
	for _, puzzle := range puzzles {
		// Mark puzzle as completed with empty solved positions (admin forced completion)
		if err := s.puzzleProgressRepo.MarkCompleted(ctx, guestID, puzzle.ID, []domain.Position{}); err != nil {
			// Log error but continue with other puzzles
			continue
		}
	}

	// Update set progress to completed
	if err := s.guestSetProgressRepo.MarkCompleted(ctx, guestID, setProgress.SetID); err != nil {
		return nil, fmt.Errorf("failed to mark set as completed: %w", err)
	}

	// Get updated progress
	updatedProgress, err := s.guestSetProgressRepo.GetByGuestAndSet(ctx, guestID, setProgress.SetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated progress: %w", err)
	}

	return updatedProgress, nil
}
