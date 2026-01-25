package domain

import (
	"time"

	"github.com/google/uuid"
)

// LeaderboardEntry represents an entry in the leaderboard
type LeaderboardEntry struct {
	ID            uuid.UUID `json:"id"`
	GameSessionID uuid.UUID `json:"game_session_id"`
	GuestID       uuid.UUID `json:"guest_id"`
	Username      string    `json:"username"`
	Mistakes      int       `json:"mistakes"`
	CreatedAt     time.Time `json:"created_at"`
}

// NewLeaderboardEntry creates a new leaderboard entry
func NewLeaderboardEntry(gameID uuid.UUID, guestID uuid.UUID, username string, mistakes int) *LeaderboardEntry {
	return &LeaderboardEntry{
		ID:            uuid.New(),
		GameSessionID: gameID,
		GuestID:       guestID,
		Username:      username,
		Mistakes:      mistakes,
		CreatedAt:     time.Now(),
	}
}
