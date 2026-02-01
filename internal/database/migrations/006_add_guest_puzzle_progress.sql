-- Migration: Add guest_puzzle_progress table
-- Tracks which puzzles each guest has completed

CREATE TABLE IF NOT EXISTS guest_puzzle_progress (
    guest_id UUID NOT NULL,
    puzzle_id UUID NOT NULL REFERENCES puzzle_pool(id) ON DELETE CASCADE,
    completed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (guest_id, puzzle_id)
);

-- Index for efficient lookup of uncompleted puzzles for a guest
CREATE INDEX IF NOT EXISTS idx_guest_puzzle_progress_guest_id ON guest_puzzle_progress(guest_id);

COMMENT ON TABLE guest_puzzle_progress IS 'Tracks puzzles completed by each guest';
COMMENT ON COLUMN guest_puzzle_progress.completed_at IS 'When the guest completed this puzzle';