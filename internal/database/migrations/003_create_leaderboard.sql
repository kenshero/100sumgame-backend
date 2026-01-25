-- Migration: Create leaderboard table
-- Run: docker exec -i sum100-db psql -U postgres -d sum100game < backend/internal/database/migrations/003_create_leaderboard.sql

CREATE TABLE IF NOT EXISTS leaderboard (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_session_id UUID REFERENCES game_sessions(id),
    username VARCHAR(50) NOT NULL,
    mistakes INT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for ranking (sorted by mistakes ascending, then by date)
CREATE INDEX IF NOT EXISTS idx_leaderboard_ranking ON leaderboard(mistakes ASC, created_at ASC);

-- Index for username search
CREATE INDEX IF NOT EXISTS idx_leaderboard_username ON leaderboard(username);

COMMENT ON TABLE leaderboard IS 'Global leaderboard for completed games';
COMMENT ON COLUMN leaderboard.mistakes IS 'Total mistakes - lower is better';
