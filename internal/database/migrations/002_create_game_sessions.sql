-- Migration: Create game_sessions table
-- Run: docker exec -i sum100-db psql -U postgres -d sum100game < backend/internal/database/migrations/002_create_game_sessions.sql

CREATE TABLE IF NOT EXISTS game_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    puzzle_id UUID REFERENCES puzzle_pool(id),
    grid_current JSONB NOT NULL,            -- Current state of grid with player values
    grid_solution JSONB NOT NULL,           -- Copy of solution for verification
    prefilled_positions JSONB NOT NULL,     -- Positions that cannot be changed
    total_mistakes INT DEFAULT 0,           -- Running total of mistakes
    tokens_used INT DEFAULT 0,              -- AI tokens consumed
    tokens_limit INT DEFAULT 1000,          -- Max tokens per game
    status VARCHAR(20) DEFAULT 'playing',   -- playing, completed
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for active games
CREATE INDEX IF NOT EXISTS idx_game_sessions_status ON game_sessions(status);
CREATE INDEX IF NOT EXISTS idx_game_sessions_created_at ON game_sessions(created_at DESC);

COMMENT ON TABLE game_sessions IS 'Active and completed game sessions';
COMMENT ON COLUMN game_sessions.grid_current IS 'Current grid state with player-entered values';
COMMENT ON COLUMN game_sessions.total_mistakes IS 'Cumulative count of incorrect cells across all verifications';
