package repository

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kenshero/100sumgame/internal/domain"
)

type gameRepository struct {
	db *pgxpool.Pool
}

// NewGameRepository creates a new game repository
func NewGameRepository(db *pgxpool.Pool) GameRepository {
	return &gameRepository{db: db}
}

func (r *gameRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Game, error) {
	query := `
		SELECT id, guest_id, puzzle_id, grid_current, grid_solution, prefilled_positions,
		       total_mistakes, tokens_used, tokens_limit, status, created_at, updated_at
		FROM game_sessions
		WHERE id = $1
	`

	var game domain.Game
	var gridCurrentJSON, gridSolutionJSON, positionsJSON []byte
	var status string

	err := r.db.QueryRow(ctx, query, id).Scan(
		&game.ID,
		&game.GuestID,
		&game.PuzzleID,
		&gridCurrentJSON,
		&gridSolutionJSON,
		&positionsJSON,
		&game.TotalMistakes,
		&game.TokensUsed,
		&game.TokensLimit,
		&status,
		&game.CreatedAt,
		&game.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	game.Status = domain.GameStatus(status)

	if err := json.Unmarshal(gridCurrentJSON, &game.GridCurrent); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(gridSolutionJSON, &game.GridSolution); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(positionsJSON, &game.PrefilledPositions); err != nil {
		return nil, err
	}

	return &game, nil
}

func (r *gameRepository) Create(ctx context.Context, game *domain.Game) error {
	query := `
		INSERT INTO game_sessions (id, guest_id, puzzle_id, grid_current, grid_solution, prefilled_positions,
		                           total_mistakes, tokens_used, tokens_limit, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	gridCurrentJSON, err := json.Marshal(game.GridCurrent)
	if err != nil {
		return err
	}

	gridSolutionJSON, err := json.Marshal(game.GridSolution)
	if err != nil {
		return err
	}

	positionsJSON, err := json.Marshal(game.PrefilledPositions)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, query,
		game.ID,
		game.GuestID,
		game.PuzzleID,
		gridCurrentJSON,
		gridSolutionJSON,
		positionsJSON,
		game.TotalMistakes,
		game.TokensUsed,
		game.TokensLimit,
		string(game.Status),
		game.CreatedAt,
		game.UpdatedAt,
	)

	return err
}

func (r *gameRepository) Update(ctx context.Context, game *domain.Game) error {
	query := `
		UPDATE game_sessions
		SET grid_current = $2, total_mistakes = $3, tokens_used = $4, status = $5, updated_at = $6
		WHERE id = $1
	`

	gridCurrentJSON, err := json.Marshal(game.GridCurrent)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, query,
		game.ID,
		gridCurrentJSON,
		game.TotalMistakes,
		game.TokensUsed,
		string(game.Status),
		game.UpdatedAt,
	)

	return err
}

func (r *gameRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM game_sessions WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
