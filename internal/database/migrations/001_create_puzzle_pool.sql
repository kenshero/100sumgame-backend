-- Create puzzle_pool table
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS puzzle_pool (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    grid_solution JSONB NOT NULL,
    prefilled_positions JSONB NOT NULL,
    difficulty VARCHAR(20) DEFAULT 'medium',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Add index for random selection
CREATE INDEX IF NOT EXISTS idx_puzzle_pool_difficulty ON puzzle_pool(difficulty);

-- Add comment
COMMENT ON TABLE puzzle_pool IS 'Pool of pre-generated puzzles with verified solutions';
