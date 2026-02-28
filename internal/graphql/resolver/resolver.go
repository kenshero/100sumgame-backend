package resolver

import (
	"github.com/kenshero/100sumgame/internal/service"
)

// Resolver is the root resolver for GraphQL
type Resolver struct {
	GameService        *service.GameService
	PuzzleService      *service.PuzzleService
	LeaderboardService *service.LeaderboardService
	AdminService       *service.AdminService
	// AIService          *service.AIService  // TODO: Add back when implementing AI chat
}
