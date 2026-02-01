package repository

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kenshero/100sumgame/internal/domain"
)

type puzzleRepository struct {
	db *pgxpool.Pool
}

// NewPuzzleRepository creates a new puzzle repository
func NewPuzzleRepository(db *pgxpool.Pool) PuzzleRepository {
	return &puzzleRepository{db: db}
}

func (r *puzzleRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Puzzle, error) {
	query := `
		SELECT id, grid_solution, prefilled_positions, difficulty, created_at
		FROM puzzle_pool
		WHERE id = $1
	`

	var puzzle domain.Puzzle
	var gridJSON, positionsJSON []byte

	err := r.db.QueryRow(ctx, query, id).Scan(
		&puzzle.ID,
		&gridJSON,
		&positionsJSON,
		&puzzle.Difficulty,
		&puzzle.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(gridJSON, &puzzle.GridSolution); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(positionsJSON, &puzzle.PrefilledPositions); err != nil {
		return nil, err
	}

	return &puzzle, nil
}

func (r *puzzleRepository) GetRandom(ctx context.Context) (*domain.Puzzle, error) {
	query := `
		SELECT id, grid_solution, prefilled_positions, difficulty, created_at
		FROM puzzle_pool
		ORDER BY RANDOM()
		LIMIT 1
	`

	var puzzle domain.Puzzle
	var gridJSON, positionsJSON []byte

	err := r.db.QueryRow(ctx, query).Scan(
		&puzzle.ID,
		&gridJSON,
		&positionsJSON,
		&puzzle.Difficulty,
		&puzzle.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(gridJSON, &puzzle.GridSolution); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(positionsJSON, &puzzle.PrefilledPositions); err != nil {
		return nil, err
	}

	return &puzzle, nil
}

func (r *puzzleRepository) GetAvailablePuzzlesForGuest(ctx context.Context, guestID uuid.UUID) ([]*domain.Puzzle, error) {
	// เลือก puzzles ที่ยังไม่เคยเล่นจบของ guest นี้
	query := `
		SELECT p.id, p.grid_solution, p.prefilled_positions, p.difficulty, p.created_at
		FROM puzzle_pool p
		WHERE NOT EXISTS (
			SELECT 1 FROM guest_puzzle_progress gpp
			WHERE gpp.guest_id = $1 AND gpp.puzzle_id = p.id
		)
		ORDER BY RANDOM()
		LIMIT 10
	`

	rows, err := r.db.Query(ctx, query, guestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var puzzles []*domain.Puzzle
	for rows.Next() {
		var puzzle domain.Puzzle
		var gridJSON, positionsJSON []byte

		if err := rows.Scan(
			&puzzle.ID,
			&gridJSON,
			&positionsJSON,
			&puzzle.Difficulty,
			&puzzle.CreatedAt,
		); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(gridJSON, &puzzle.GridSolution); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(positionsJSON, &puzzle.PrefilledPositions); err != nil {
			return nil, err
		}

		puzzles = append(puzzles, &puzzle)
	}

	return puzzles, nil
}

func (r *puzzleRepository) GetAll(ctx context.Context) ([]*domain.Puzzle, error) {
	query := `
		SELECT id, grid_solution, prefilled_positions, difficulty, created_at
		FROM puzzle_pool
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var puzzles []*domain.Puzzle
	for rows.Next() {
		var puzzle domain.Puzzle
		var gridJSON, positionsJSON []byte

		if err := rows.Scan(
			&puzzle.ID,
			&gridJSON,
			&positionsJSON,
			&puzzle.Difficulty,
			&puzzle.CreatedAt,
		); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(gridJSON, &puzzle.GridSolution); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(positionsJSON, &puzzle.PrefilledPositions); err != nil {
			return nil, err
		}

		puzzles = append(puzzles, &puzzle)
	}

	return puzzles, nil
}

func (r *puzzleRepository) Create(ctx context.Context, puzzle *domain.Puzzle) error {
	query := `
		INSERT INTO puzzle_pool (id, grid_solution, prefilled_positions, difficulty, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	gridJSON, err := json.Marshal(puzzle.GridSolution)
	if err != nil {
		return err
	}

	positionsJSON, err := json.Marshal(puzzle.PrefilledPositions)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, query,
		puzzle.ID,
		gridJSON,
		positionsJSON,
		puzzle.Difficulty,
		puzzle.CreatedAt,
	)

	return err
}
