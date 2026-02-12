package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/kenshero/100sumgame/internal/domain"
)

// PuzzleRepository defines the interface for puzzle data access
type PuzzleRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Puzzle, error)
	GetRandom(ctx context.Context) (*domain.Puzzle, error)
	GetAvailablePuzzlesForGuest(ctx context.Context, guestID uuid.UUID) ([]*domain.Puzzle, error)
	GetAll(ctx context.Context) ([]*domain.Puzzle, error)
	GetTotalCount(ctx context.Context) (int, error)
	Create(ctx context.Context, puzzle *domain.Puzzle) error
}

// GameRepository defines the interface for game data access
type GameRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Game, error)
	GetByGuestAndPuzzle(ctx context.Context, guestID, puzzleID uuid.UUID) (*domain.Game, error)
	Create(ctx context.Context, game *domain.Game) error
	Update(ctx context.Context, game *domain.Game) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetPuzzleStats(ctx context.Context, puzzleID uuid.UUID) (*domain.PuzzleStats, error)
}

// LeaderboardRepository defines the interface for leaderboard data access
type LeaderboardRepository interface {
	GetTop(ctx context.Context, limit int) ([]*domain.LeaderboardEntry, error)
	Create(ctx context.Context, entry *domain.LeaderboardEntry) error
	GetRank(ctx context.Context, mistakes int) (int, error)
}

// PuzzleProgressRepository defines the interface for guest puzzle progress tracking
type PuzzleProgressRepository interface {
	MarkCompleted(ctx context.Context, guestID, puzzleID uuid.UUID) error
	GetCompletedPuzzles(ctx context.Context, guestID uuid.UUID) ([]uuid.UUID, error)
	GetCompletedCount(ctx context.Context, guestID uuid.UUID) (int, error)
	HasCompleted(ctx context.Context, guestID, puzzleID uuid.UUID) (bool, error)

	// New methods for ad reward system
	GetAvailablePuzzlesForGuest(ctx context.Context, guestID uuid.UUID, limit int) ([]*domain.PuzzleWithStatus, error)
	MarkArchived(ctx context.Context, guestID uuid.UUID) error
	MarkAvailable(ctx context.Context, guestID uuid.UUID) error
}
