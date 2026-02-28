package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kenshero/100sumgame/internal/domain"
)

type settingsRepository struct {
	db *pgxpool.Pool
}

// NewSettingsRepository creates a new settings repository
func NewSettingsRepository(db *pgxpool.Pool) SettingsRepository {
	return &settingsRepository{db: db}
}

// SettingsRepository defines the interface for settings operations
type SettingsRepository interface {
	GetAllSettings(ctx context.Context) (map[string]string, error)
	UpdateSetting(ctx context.Context, key, value string) error
}

// GetAllSettings retrieves all settings from database
func (r *settingsRepository) GetAllSettings(ctx context.Context) (map[string]string, error) {
	query := `
		SELECT key, value
		FROM game_settings
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		settings[key] = value
	}

	return settings, nil
}

// UpdateSetting updates a single setting in database
func (r *settingsRepository) UpdateSetting(ctx context.Context, key, value string) error {
	query := `
		UPDATE game_settings
		SET value = $2, updated_at = NOW()
		WHERE key = $1
	`
	_, err := r.db.Exec(ctx, query, key, value)
	return err
}

// ParseSettings converts settings map to GameSettings struct
func ParseSettings(settings map[string]string) *domain.GameSettings {
	result := &domain.GameSettings{}

	// Parse stamina_max
	if val, ok := settings["stamina_max"]; ok {
		result.StaminaMax = parseInt(val, 35)
	}

	// Parse stamina_regen_interval_minutes
	if val, ok := settings["stamina_regen_interval_minutes"]; ok {
		result.StaminaRegenIntervalMinutes = parseInt(val, 5)
	}

	// Parse stamina_regen_amount
	if val, ok := settings["stamina_regen_amount"]; ok {
		result.StaminaRegenAmount = parseInt(val, 1)
	}

	// Parse initial_score
	if val, ok := settings["initial_score"]; ok {
		result.InitialScore = parseInt(val, 500)
	}

	// Parse score_deduction_per_mistake
	if val, ok := settings["score_deduction_per_mistake"]; ok {
		result.ScoreDeductionPerMistake = parseInt(val, 10)
	}

	// Parse score_minimum
	if val, ok := settings["score_minimum"]; ok {
		result.ScoreMinimum = parseInt(val, 0)
	}

	return result
}

// parseInt is a helper to parse string to int with default value
func parseInt(s string, defaultValue int) int {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	if err != nil {
		return defaultValue
	}
	return result
}
