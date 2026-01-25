package service

import (
	"context"
	"fmt"

	"github.com/kenshero/100sumgame/internal/domain"
)

// AIService handles AI-related functionality using Gemini
type AIService struct {
	apiKey string
}

// NewAIService creates a new AI service
func NewAIService(apiKey string) *AIService {
	return &AIService{apiKey: apiKey}
}

// ChatRequest represents a chat request
type ChatRequest struct {
	GameID  string `json:"game_id"`
	Message string `json:"message"`
}

// ChatResponse represents a chat response
type ChatResponse struct {
	Message         string `json:"message"`
	TokensUsed      int    `json:"tokens_used"`
	TokensRemaining int    `json:"tokens_remaining"`
}

// Chat processes a chat message and returns AI response
func (s *AIService) Chat(ctx context.Context, game *domain.Game, message string) (*ChatResponse, error) {
	// TODO: Implement Gemini integration
	// For now, return a placeholder response

	tokensUsed := len(message) / 4 // Rough estimate
	tokensRemaining := game.TokensLimit - game.TokensUsed - tokensUsed

	if tokensRemaining < 0 {
		return nil, domain.ErrTokensExhausted
	}

	// Placeholder response
	response := &ChatResponse{
		Message:         fmt.Sprintf("I'm your Game Master! You asked: '%s'. Let me analyze the grid and give you a hint...", message),
		TokensUsed:      tokensUsed,
		TokensRemaining: tokensRemaining,
	}

	return response, nil
}

// GetSystemPrompt returns the system prompt for the AI
func (s *AIService) GetSystemPrompt() string {
	return `You are a friendly Game Master for the Sum-100 puzzle game.

RULES:
- Respond in the SAME LANGUAGE the user writes in
- Give hints about logic and possible number ranges only
- NEVER reveal exact answers
- Be encouraging and supportive
- Keep responses concise

GAME CONTEXT:
- Players fill a 5x5 grid where each row and column must sum to 100
- Numbers must be between 1 and 99
- Some cells are pre-filled and cannot be changed
- Players can verify their answers to see which cells are correct, too high, or too low

AVAILABLE FUNCTIONS:
- fill_cells: Fill numbers into grid cells
- verify_grid: Check all answers
- get_current_state: Get current grid state`
}
