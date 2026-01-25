package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/kenshero/100sumgame/internal/domain"
	"github.com/kenshero/100sumgame/internal/repository"
)

// LeaderboardService handles leaderboard-related business logic
type LeaderboardService struct {
	repo repository.LeaderboardRepository
}

// NewLeaderboardService creates a new leaderboard service
func NewLeaderboardService(repo repository.LeaderboardRepository) *LeaderboardService {
	return &LeaderboardService{repo: repo}
}

// GetTop retrieves top leaderboard entries
func (s *LeaderboardService) GetTop(ctx context.Context, limit int) ([]*domain.LeaderboardEntry, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	return s.repo.GetTop(ctx, limit)
}

// Submit submits a game result to the leaderboard
func (s *LeaderboardService) Submit(ctx context.Context, gameID uuid.UUID, guestID uuid.UUID, username string, mistakes int) (*domain.LeaderboardEntry, int, error) {
	// Validate username
	if username == "" {
		username = fmt.Sprintf("Guest_%s", guestID.String()[:8])
	}

	// Create leaderboard entry
	entry := domain.NewLeaderboardEntry(gameID, guestID, username, mistakes)

	if err := s.repo.Create(ctx, entry); err != nil {
		return nil, 0, err
	}

	// Get rank
	rank, err := s.repo.GetRank(ctx, mistakes)
	if err != nil {
		return entry, 0, nil // Return entry without rank on error
	}

	return entry, rank, nil
}

// GetRank gets the rank for a given number of mistakes
func (s *LeaderboardService) GetRank(ctx context.Context, mistakes int) (int, error) {
	return s.repo.GetRank(ctx, mistakes)
}
