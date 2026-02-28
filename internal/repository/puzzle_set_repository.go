package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kenshero/100sumgame/internal/domain"
)

type puzzleSetRepository struct {
	db *pgxpool.Pool
}

// NewPuzzleSetRepository creates a new puzzle set repository
func NewPuzzleSetRepository(db *pgxpool.Pool) PuzzleSetRepository {
	return &puzzleSetRepository{db: db}
}

func (r *puzzleSetRepository) Create(ctx context.Context, set *domain.PuzzleSet) error {
	query := `
		INSERT INTO puzzle_sets (id, set_order, difficulty, created_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.Exec(ctx, query, set.ID, set.SetOrder, set.Difficulty, set.CreatedAt)
	return err
}

func (r *puzzleSetRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.PuzzleSet, error) {
	query := `
		SELECT id, set_order, difficulty, created_at
		FROM puzzle_sets
		WHERE id = $1
	`
	row := r.db.QueryRow(ctx, query, id)

	var set domain.PuzzleSet
	err := row.Scan(&set.ID, &set.SetOrder, &set.Difficulty, &set.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &set, nil
}

func (r *puzzleSetRepository) GetByOrder(ctx context.Context, order int) (*domain.PuzzleSet, error) {
	query := `
		SELECT id, set_order, difficulty, created_at
		FROM puzzle_sets
		WHERE set_order = $1
	`
	row := r.db.QueryRow(ctx, query, order)

	var set domain.PuzzleSet
	err := row.Scan(&set.ID, &set.SetOrder, &set.Difficulty, &set.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &set, nil
}

func (r *puzzleSetRepository) GetAll(ctx context.Context) ([]*domain.PuzzleSet, error) {
	query := `
		SELECT id, set_order, difficulty, created_at
		FROM puzzle_sets
		ORDER BY set_order
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sets []*domain.PuzzleSet
	for rows.Next() {
		var set domain.PuzzleSet
		if err := rows.Scan(&set.ID, &set.SetOrder, &set.Difficulty, &set.CreatedAt); err != nil {
			return nil, err
		}
		sets = append(sets, &set)
	}

	return sets, nil
}

func (r *puzzleSetRepository) GetNextSet(ctx context.Context, currentOrder int) (*domain.PuzzleSet, error) {
	query := `
		SELECT id, set_order, difficulty, created_at
		FROM puzzle_sets
		WHERE set_order > $1
		ORDER BY set_order
		LIMIT 1
	`
	row := r.db.QueryRow(ctx, query, currentOrder)

	var set domain.PuzzleSet
	err := row.Scan(&set.ID, &set.SetOrder, &set.Difficulty, &set.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &set, nil
}
