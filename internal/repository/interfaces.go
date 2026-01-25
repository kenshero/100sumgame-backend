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
	GetAll(ctx context.Context) ([]*domain.Puzzle, error)
	Create(ctx context.Context, puzzle *domain.Puzzle) error
}

// GameRepository defines the interface for game data access
type GameRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Game, error)
	Create(ctx context.Context, game *domain.Game) error
	Update(ctx context.Context, game *domain.Game) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// LeaderboardRepository defines the interface for leaderboard data access
type LeaderboardRepository interface {
	GetTop(ctx context.Context, limit int) ([]*domain.LeaderboardEntry, error)
	Create(ctx context.Context, entry *domain.LeaderboardEntry) error
	GetRank(ctx context.Context, mistakes int) (int, error)
}
