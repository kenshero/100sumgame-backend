package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kenshero/100sumgame/internal/domain"
)

type puzzleProgressRepository struct {
	db *pgxpool.Pool
}

// NewPuzzleProgressRepository creates a new puzzle progress repository
func NewPuzzleProgressRepository(db *pgxpool.Pool) PuzzleProgressRepository {
	return &puzzleProgressRepository{db: db}
}

func (r *puzzleProgressRepository) MarkCompleted(ctx context.Context, guestID, puzzleID uuid.UUID, solvedPositions []domain.Position) error {
	query := `
		INSERT INTO guest_puzzle_progress (guest_id, puzzle_id, status, completed_at, solved_positions)
		VALUES ($1, $2, 'COMPLETED', NOW(), $3)
		ON CONFLICT (guest_id, puzzle_id)
		DO UPDATE SET status = 'COMPLETED', completed_at = NOW(), solved_positions = $3
	`

	_, err := r.db.Exec(ctx, query, guestID, puzzleID, solvedPositions)
	return err
}

func (r *puzzleProgressRepository) MarkPlaying(ctx context.Context, guestID, puzzleID uuid.UUID) error {
	query := `
		INSERT INTO guest_puzzle_progress (guest_id, puzzle_id, status, completed_at)
		VALUES ($1, $2, 'PLAYING', NOW())
		ON CONFLICT (guest_id, puzzle_id)
		DO UPDATE SET status = CASE
			WHEN guest_puzzle_progress.status = 'COMPLETED' THEN 'COMPLETED'
			ELSE 'PLAYING'
		END
	`

	_, err := r.db.Exec(ctx, query, guestID, puzzleID)
	return err
}

func (r *puzzleProgressRepository) UpdateSolvedPositions(ctx context.Context, guestID, puzzleID uuid.UUID, solvedPositions []domain.Position) error {
	query := `
		INSERT INTO guest_puzzle_progress (guest_id, puzzle_id, status, solved_positions)
		VALUES ($1, $2, 'PLAYING', $3)
		ON CONFLICT (guest_id, puzzle_id)
		DO UPDATE SET solved_positions = $3
		WHERE guest_puzzle_progress.status != 'COMPLETED'
	`

	_, err := r.db.Exec(ctx, query, guestID, puzzleID, solvedPositions)
	return err
}

func (r *puzzleProgressRepository) GetCompletedPuzzles(ctx context.Context, guestID uuid.UUID) ([]uuid.UUID, error) {
	query := `
		SELECT puzzle_id
		FROM guest_puzzle_progress
		WHERE guest_id = $1 AND status = 'COMPLETED'
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

func (r *puzzleProgressRepository) GetCompletedCount(ctx context.Context, guestID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(DISTINCT puzzle_id)
		FROM guest_puzzle_progress
		WHERE guest_id = $1 AND status = 'COMPLETED'
	`

	var count int
	err := r.db.QueryRow(ctx, query, guestID).Scan(&count)
	return count, err
}

func (r *puzzleProgressRepository) HasCompleted(ctx context.Context, guestID, puzzleID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM guest_puzzle_progress
			WHERE guest_id = $1 AND puzzle_id = $2 AND status = 'COMPLETED'
		)
	`

	var completed bool
	err := r.db.QueryRow(ctx, query, guestID, puzzleID).Scan(&completed)
	return completed, err
}

// GetAvailablePuzzlesForGuest gets puzzles with their status for a guest
func (r *puzzleProgressRepository) GetAvailablePuzzlesForGuest(ctx context.Context, guestID uuid.UUID, limit int) ([]*domain.PuzzleWithStatus, error) {
	query := `
		WITH all_puzzles AS (
			SELECT id, set_id, grid_solution, prefilled_positions, created_at
			FROM puzzle_pool
			ORDER BY id ASC
			LIMIT $1
		),
		guest_progress AS (
			SELECT puzzle_id, status, solved_positions
			FROM guest_puzzle_progress
			WHERE guest_id = $2
		)
		SELECT 
			ap.id,
			ap.set_id,
			ap.grid_solution,
			ap.prefilled_positions,
			ap.created_at,
			COALESCE(gp.status, 'AVAILABLE') as status,
			gp.solved_positions
		FROM all_puzzles ap
		LEFT JOIN guest_progress gp ON ap.id = gp.puzzle_id
	`

	rows, err := r.db.Query(ctx, query, limit, guestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var puzzles []*domain.PuzzleWithStatus
	for rows.Next() {
		var p domain.PuzzleWithStatus
		var puzzle domain.Puzzle
		var status string
		var solvedPositions []domain.Position

		err := rows.Scan(
			&puzzle.ID,
			&puzzle.SetID,
			&puzzle.GridSolution,
			&puzzle.PrefilledPositions,
			&puzzle.CreatedAt,
			&status,
			&solvedPositions,
		)
		if err != nil {
			return nil, err
		}

		p.Puzzle = &puzzle
		p.Status = domain.PuzzleStatus(status)
		p.SolvedPositions = solvedPositions
		puzzles = append(puzzles, &p)
	}

	return puzzles, nil
}

// MarkArchived marks all COMPLETED puzzles as ARCHIVED for a guest
func (r *puzzleProgressRepository) MarkArchived(ctx context.Context, guestID uuid.UUID) error {
	query := `
		UPDATE guest_puzzle_progress
		SET status = 'ARCHIVED'
		WHERE guest_id = $1 AND status = 'COMPLETED'
	`

	_, err := r.db.Exec(ctx, query, guestID)
	return err
}

// MarkAvailable marks all AD_BLOCK puzzles as AVAILABLE for a guest
func (r *puzzleProgressRepository) MarkAvailable(ctx context.Context, guestID uuid.UUID) error {
	query := `
		UPDATE guest_puzzle_progress
		SET status = 'AVAILABLE'
		WHERE guest_id = $1 AND status = 'AD_BLOCK'
	`

	_, err := r.db.Exec(ctx, query, guestID)
	return err
}
