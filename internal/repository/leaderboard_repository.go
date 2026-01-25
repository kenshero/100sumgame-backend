package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kenshero/100sumgame/internal/domain"
)

type leaderboardRepository struct {
	db *pgxpool.Pool
}

// NewLeaderboardRepository creates a new leaderboard repository
func NewLeaderboardRepository(db *pgxpool.Pool) LeaderboardRepository {
	return &leaderboardRepository{db: db}
}

func (r *leaderboardRepository) GetTop(ctx context.Context, limit int) ([]*domain.LeaderboardEntry, error) {
	query := `
		SELECT id, game_session_id, guest_id, username, mistakes, created_at
		FROM leaderboard
		ORDER BY mistakes ASC, created_at ASC
		LIMIT $1
	`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*domain.LeaderboardEntry
	for rows.Next() {
		var entry domain.LeaderboardEntry
		if err := rows.Scan(
			&entry.ID,
			&entry.GameSessionID,
			&entry.GuestID,
			&entry.Username,
			&entry.Mistakes,
			&entry.CreatedAt,
		); err != nil {
			return nil, err
		}
		entries = append(entries, &entry)
	}

	return entries, nil
}

func (r *leaderboardRepository) Create(ctx context.Context, entry *domain.LeaderboardEntry) error {
	query := `
		INSERT INTO leaderboard (id, game_session_id, guest_id, username, mistakes, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Exec(ctx, query,
		entry.ID,
		entry.GameSessionID,
		entry.GuestID,
		entry.Username,
		entry.Mistakes,
		entry.CreatedAt,
	)

	return err
}

func (r *leaderboardRepository) GetRank(ctx context.Context, mistakes int) (int, error) {
	query := `
		SELECT COUNT(*) + 1
		FROM leaderboard
		WHERE mistakes < $1
	`

	var rank int
	err := r.db.QueryRow(ctx, query, mistakes).Scan(&rank)
	if err != nil {
		return 0, err
	}

	return rank, nil
}
