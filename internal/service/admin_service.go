package service

import (
	"github.com/kenshero/100sumgame/internal/domain"
)

// AdminService handles admin operations
type AdminService struct {
	configService *ConfigService
}

// NewAdminService creates a new admin service
func NewAdminService(configService *ConfigService) *AdminService {
	return &AdminService{
		configService: configService,
	}
}

// RefreshConfig refreshes game configuration from database
// This should be called after updating database settings
func (s *AdminService) RefreshConfig() (*domain.GameSettings, error) {
	if err := s.configService.RefreshSettings(); err != nil {
		return nil, err
	}
	return s.configService.GetSettings(), nil
}
