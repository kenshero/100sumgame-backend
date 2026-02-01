package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type puzzleProgressRepository struct {
	db *pgxpool.Pool
}

// NewPuzzleProgressRepository creates a new puzzle progress repository
func NewPuzzleProgressRepository(db *pgxpool.Pool) PuzzleProgressRepository {
	return &puzzleProgressRepository{db: db}
}

func (r *puzzleProgressRepository) MarkCompleted(ctx context.Context, guestID, puzzleID uuid.UUID) error {
	query := `
		INSERT INTO guest_puzzle_progress (guest_id, puzzle_id)
		VALUES ($1, $2)
		ON CONFLICT (guest_id, puzzle_id) DO NOTHING
	`

	_, err := r.db.Exec(ctx, query, guestID, puzzleID)
	return err
}

func (r *puzzleProgressRepository) GetCompletedPuzzles(ctx context.Context, guestID uuid.UUID) ([]uuid.UUID, error) {
	query := `
		SELECT puzzle_id
		FROM guest_puzzle_progress
		WHERE guest_id = $1
		ORDER BY completed_at DESC
	`

	rows, err := r.db.Query(ctx, query, guestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var puzzleIDs []uuid.UUID
	for rows.Next() {
		var puzzleID uuid.UUID
		if err := rows.Scan(&puzzleID); err != nil {
			return nil, err
		}
		puzzleIDs = append(puzzleIDs, puzzleID)
	}

	return puzzleIDs, nil
}

func (r *puzzleProgressRepository) HasCompleted(ctx context.Context, guestID, puzzleID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM guest_puzzle_progress
			WHERE guest_id = $1 AND puzzle_id = $2
		)
	`

	var completed bool
	err := r.db.QueryRow(ctx, query, guestID, puzzleID).Scan(&completed)
	return completed, err
}
