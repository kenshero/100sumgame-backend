package service

import (
	"context"
	"sync"

	"github.com/kenshero/100sumgame/internal/domain"
	"github.com/kenshero/100sumgame/internal/repository"
)

// ConfigService manages game configuration with caching
type ConfigService struct {
	repo     repository.SettingsRepository
	settings *domain.GameSettings
	mu       sync.RWMutex
}

// NewConfigService creates a new config service
func NewConfigService(repo repository.SettingsRepository) *ConfigService {
	return &ConfigService{
		repo: repo,
	}
}

// LoadSettings loads settings from database (should be called on server startup)
func (s *ConfigService) LoadSettings() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	settingsMap, err := s.repo.GetAllSettings(context.Background())
	if err != nil {
		return err
	}

	s.settings = repository.ParseSettings(settingsMap)
	return nil
}

// GetSettings returns current settings from cache
func (s *ConfigService) GetSettings() *domain.GameSettings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.settings
}

// RefreshSettings reloads settings from database
func (s *ConfigService) RefreshSettings() error {
	return s.LoadSettings()
}
