package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kenshero/100sumgame/internal/domain"
)

var (
	ErrInsufficientStamina = errors.New("insufficient stamina")
)

type guestSetProgressRepository struct {
	db *pgxpool.Pool
}

// NewGuestSetProgressRepository creates a new guest set progress repository
func NewGuestSetProgressRepository(db *pgxpool.Pool) GuestSetProgressRepository {
	return &guestSetProgressRepository{db: db}
}

func (r *guestSetProgressRepository) Create(ctx context.Context, progress *domain.GuestSetProgress) error {
	query := `
		INSERT INTO guest_set_progress (guest_id, set_id, puzzles_completed, is_unlocked, is_completed, unlocked_at, completed_at, current_stamina, last_stamina_update, current_score)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.Exec(ctx, query,
		progress.GuestID,
		progress.SetID,
		progress.PuzzlesCompleted,
		progress.IsUnlocked,
		progress.IsCompleted,
		progress.UnlockedAt,
		progress.CompletedAt,
		progress.CurrentStamina,
		progress.LastStaminaUpdate,
		progress.CurrentScore,
	)
	return err
}

// CreateIfNotExists inserts progress only if it doesn't exist, otherwise does nothing
// Uses ON CONFLICT DO NOTHING to prevent duplicate key errors
func (r *guestSetProgressRepository) CreateIfNotExists(ctx context.Context, progress *domain.GuestSetProgress) error {
	query := `
		INSERT INTO guest_set_progress (guest_id, set_id, puzzles_completed, is_unlocked, is_completed, unlocked_at, completed_at, current_stamina, last_stamina_update, current_score)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (guest_id, set_id) DO NOTHING
	`
	_, err := r.db.Exec(ctx, query,
		progress.GuestID,
		progress.SetID,
		progress.PuzzlesCompleted,
		progress.IsUnlocked,
		progress.IsCompleted,
		progress.UnlockedAt,
		progress.CompletedAt,
		progress.CurrentStamina,
		progress.LastStaminaUpdate,
		progress.CurrentScore,
	)
	return err
}

func (r *guestSetProgressRepository) GetByGuestAndSet(ctx context.Context, guestID, setID uuid.UUID) (*domain.GuestSetProgress, error) {
	query := `
		SELECT guest_id, set_id, puzzles_completed, is_unlocked, is_completed, unlocked_at, completed_at, current_stamina, last_stamina_update, current_score
		FROM guest_set_progress
		WHERE guest_id = $1 AND set_id = $2
	`
	row := r.db.QueryRow(ctx, query, guestID, setID)

	var progress domain.GuestSetProgress
	err := row.Scan(
		&progress.GuestID,
		&progress.SetID,
		&progress.PuzzlesCompleted,
		&progress.IsUnlocked,
		&progress.IsCompleted,
		&progress.UnlockedAt,
		&progress.CompletedAt,
		&progress.CurrentStamina,
		&progress.LastStaminaUpdate,
		&progress.CurrentScore,
	)
	if err != nil {
		return nil, err
	}
	return &progress, nil
}

func (r *guestSetProgressRepository) GetByGuest(ctx context.Context, guestID uuid.UUID) ([]*domain.GuestSetProgress, error) {
	query := `
		SELECT gsp.guest_id, gsp.set_id, gsp.puzzles_completed, gsp.is_unlocked, gsp.is_completed, gsp.unlocked_at, gsp.completed_at, gsp.current_stamina, gsp.last_stamina_update, gsp.current_score
		FROM guest_set_progress gsp
		INNER JOIN puzzle_sets ps ON gsp.set_id = ps.id
		WHERE gsp.guest_id = $1
		ORDER BY ps.set_order
	`
	rows, err := r.db.Query(ctx, query, guestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var progresses []*domain.GuestSetProgress
	for rows.Next() {
		var progress domain.GuestSetProgress
		if err := rows.Scan(
			&progress.GuestID,
			&progress.SetID,
			&progress.PuzzlesCompleted,
			&progress.IsUnlocked,
			&progress.IsCompleted,
			&progress.UnlockedAt,
			&progress.CompletedAt,
			&progress.CurrentStamina,
			&progress.LastStaminaUpdate,
			&progress.CurrentScore,
		); err != nil {
			return nil, err
		}
		progresses = append(progresses, &progress)
	}

	return progresses, nil
}

func (r *guestSetProgressRepository) GetUnlockedSet(ctx context.Context, guestID uuid.UUID) (*domain.GuestSetProgress, error) {
	query := `
		SELECT gsp.guest_id, gsp.set_id, gsp.puzzles_completed, gsp.is_unlocked, gsp.is_completed, gsp.unlocked_at, gsp.completed_at, gsp.current_stamina, gsp.last_stamina_update, gsp.current_score
		FROM guest_set_progress gsp
		INNER JOIN puzzle_sets ps ON gsp.set_id = ps.id
		WHERE gsp.guest_id = $1 AND gsp.is_unlocked = true AND gsp.is_completed = false
		ORDER BY ps.set_order ASC
		LIMIT 1
	`
	row := r.db.QueryRow(ctx, query, guestID)

	var progress domain.GuestSetProgress
	err := row.Scan(
		&progress.GuestID,
		&progress.SetID,
		&progress.PuzzlesCompleted,
		&progress.IsUnlocked,
		&progress.IsCompleted,
		&progress.UnlockedAt,
		&progress.CompletedAt,
		&progress.CurrentStamina,
		&progress.LastStaminaUpdate,
		&progress.CurrentScore,
	)
	if err != nil {
		return nil, err
	}
	return &progress, nil
}

func (r *guestSetProgressRepository) UpdateProgress(ctx context.Context, guestID, setID uuid.UUID, puzzlesCompleted int) error {
	query := `
		UPDATE guest_set_progress
		SET puzzles_completed = $3
		WHERE guest_id = $1 AND set_id = $2
	`
	_, err := r.db.Exec(ctx, query, guestID, setID, puzzlesCompleted)
	return err
}

func (r *guestSetProgressRepository) MarkUnlocked(ctx context.Context, guestID, setID uuid.UUID) error {
	now := time.Now()
	query := `
		UPDATE guest_set_progress
		SET is_unlocked = true, unlocked_at = $3, current_stamina = 35, current_score = 500, last_stamina_update = $3
		WHERE guest_id = $1 AND set_id = $2
	`
	_, err := r.db.Exec(ctx, query, guestID, setID, now)
	return err
}

func (r *guestSetProgressRepository) MarkCompleted(ctx context.Context, guestID, setID uuid.UUID) error {
	now := time.Now()
	query := `
		UPDATE guest_set_progress
		SET is_completed = true, completed_at = $3
		WHERE guest_id = $1 AND set_id = $2
	`
	_, err := r.db.Exec(ctx, query, guestID, setID, now)
	return err
}

// RegenerateStamina calculates and updates stamina based on time elapsed
func (r *guestSetProgressRepository) RegenerateStamina(ctx context.Context, guestID, setID uuid.UUID, currentStamina, maxStamina, regenIntervalMinutes, regenAmount int, lastUpdate time.Time) (int, time.Time, error) {
	// Calculate time elapsed since last update
	elapsedMinutes := int(time.Since(lastUpdate).Minutes())

	// Calculate how many stamina points to regenerate
	intervalsPassed := elapsedMinutes / regenIntervalMinutes
	staminaToRegen := intervalsPassed * regenAmount

	// Don't exceed max stamina
	newStamina := currentStamina + staminaToRegen
	if newStamina > maxStamina {
		newStamina = maxStamina
	}

	// If stamina didn't change, return current values
	if newStamina == currentStamina {
		return currentStamina, lastUpdate, nil
	}

	// Update database
	now := time.Now()
	query := `
		UPDATE guest_set_progress
		SET current_stamina = $3, last_stamina_update = $4
		WHERE guest_id = $1 AND set_id = $2
	`
	_, err := r.db.Exec(ctx, query, guestID, setID, newStamina, now)

	return newStamina, now, err
}

// DeductStamina reduces stamina by 1
func (r *guestSetProgressRepository) DeductStamina(ctx context.Context, guestID, setID uuid.UUID) error {
	query := `
		UPDATE guest_set_progress
		SET current_stamina = current_stamina - 1, last_stamina_update = NOW()
		WHERE guest_id = $1 AND set_id = $2 AND current_stamina > 0
	`
	result, err := r.db.Exec(ctx, query, guestID, setID)
	if err != nil {
		return err
	}

	// Check if any row was updated (stamina > 0)
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrInsufficientStamina
	}

	return nil
}

// DeductScore reduces score by given amount
func (r *guestSetProgressRepository) DeductScore(ctx context.Context, guestID, setID uuid.UUID, amount int, minimumScore int) error {
	query := `
		UPDATE guest_set_progress
		SET current_score = GREATEST($4, current_score - $3)
		WHERE guest_id = $1 AND set_id = $2
	`
	_, err := r.db.Exec(ctx, query, guestID, setID, amount, minimumScore)
	return err
}
