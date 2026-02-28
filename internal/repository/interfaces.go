package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kenshero/100sumgame/internal/domain"
)

// PuzzleRepository defines interface for puzzle data access
type PuzzleRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Puzzle, error)
	GetRandom(ctx context.Context) (*domain.Puzzle, error)
	GetAvailablePuzzlesForGuest(ctx context.Context, guestID uuid.UUID) ([]*domain.Puzzle, error)
	GetAll(ctx context.Context) ([]*domain.Puzzle, error)
	GetTotalCount(ctx context.Context) (int, error)
	Create(ctx context.Context, puzzle *domain.Puzzle) error
	GetPuzzlesBySet(ctx context.Context, setID uuid.UUID) ([]*domain.Puzzle, error)
}

// GameRepository defines interface for game data access
type GameRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Game, error)
	GetByGuestAndPuzzle(ctx context.Context, guestID, puzzleID uuid.UUID) (*domain.Game, error)
	Create(ctx context.Context, game *domain.Game) error
	Update(ctx context.Context, game *domain.Game) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetPuzzleStats(ctx context.Context, puzzleID uuid.UUID) (*domain.PuzzleStats, error)
}

// LeaderboardRepository defines interface for leaderboard data access
type LeaderboardRepository interface {
	GetTop(ctx context.Context, limit int) ([]*domain.LeaderboardEntry, error)
	Create(ctx context.Context, entry *domain.LeaderboardEntry) error
	GetRank(ctx context.Context, mistakes int) (int, error)
}

// PuzzleProgressRepository defines interface for guest puzzle progress tracking
type PuzzleProgressRepository interface {
	MarkCompleted(ctx context.Context, guestID, puzzleID uuid.UUID, solvedPositions []domain.Position) error
	MarkPlaying(ctx context.Context, guestID, puzzleID uuid.UUID) error
	GetCompletedPuzzles(ctx context.Context, guestID uuid.UUID) ([]uuid.UUID, error)
	GetCompletedCount(ctx context.Context, guestID uuid.UUID) (int, error)
	HasCompleted(ctx context.Context, guestID, puzzleID uuid.UUID) (bool, error)

	// UpdateSolvedPositions saves correctly answered cell positions for a playing puzzle
	UpdateSolvedPositions(ctx context.Context, guestID, puzzleID uuid.UUID, solvedPositions []domain.Position) error

	// New methods for ad reward system
	GetAvailablePuzzlesForGuest(ctx context.Context, guestID uuid.UUID, limit int) ([]*domain.PuzzleWithStatus, error)
	MarkArchived(ctx context.Context, guestID uuid.UUID) error
	MarkAvailable(ctx context.Context, guestID uuid.UUID) error
}

// PuzzleSetRepository defines interface for puzzle set management
type PuzzleSetRepository interface {
	Create(ctx context.Context, set *domain.PuzzleSet) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.PuzzleSet, error)
	GetByOrder(ctx context.Context, order int) (*domain.PuzzleSet, error)
	GetAll(ctx context.Context) ([]*domain.PuzzleSet, error)
	GetNextSet(ctx context.Context, currentOrder int) (*domain.PuzzleSet, error)
}

// GuestSetProgressRepository defines interface for guest set progress tracking
type GuestSetProgressRepository interface {
	Create(ctx context.Context, progress *domain.GuestSetProgress) error
	GetByGuestAndSet(ctx context.Context, guestID, setID uuid.UUID) (*domain.GuestSetProgress, error)
	GetByGuest(ctx context.Context, guestID uuid.UUID) ([]*domain.GuestSetProgress, error)
	GetUnlockedSet(ctx context.Context, guestID uuid.UUID) (*domain.GuestSetProgress, error)
	UpdateProgress(ctx context.Context, guestID, setID uuid.UUID, puzzlesCompleted int) error
	MarkUnlocked(ctx context.Context, guestID, setID uuid.UUID) error
	MarkCompleted(ctx context.Context, guestID, setID uuid.UUID) error

	// Stamina and score methods
	RegenerateStamina(ctx context.Context, guestID, setID uuid.UUID, currentStamina, maxStamina, regenIntervalMinutes, regenAmount int, lastUpdate time.Time) (int, time.Time, error)
	DeductStamina(ctx context.Context, guestID, setID uuid.UUID) error
	DeductScore(ctx context.Context, guestID, setID uuid.UUID, amount int, minimumScore int) error
}
