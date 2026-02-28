-- Migration: Add puzzle sets system
-- Organizes puzzles into sets/groups of 10 puzzles each
-- Tracks guest progress through sets and unlocks after ads

-- Create puzzle_sets table
CREATE TABLE IF NOT EXISTS puzzle_sets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    set_order INT NOT NULL UNIQUE,
    difficulty VARCHAR(20) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create index for efficient set ordering queries
CREATE INDEX idx_puzzle_sets_order ON puzzle_sets(set_order);

-- Add set_id to puzzle_pool table
ALTER TABLE puzzle_pool ADD COLUMN set_id UUID REFERENCES puzzle_sets(id) ON DELETE CASCADE;

-- Create index for efficient puzzle selection by set
CREATE INDEX idx_puzzle_pool_set_id ON puzzle_pool(set_id);

-- Create guest_set_progress table to track guest progress through sets
CREATE TABLE IF NOT EXISTS guest_set_progress (
    guest_id UUID NOT NULL,
    set_id UUID NOT NULL REFERENCES puzzle_sets(id) ON DELETE CASCADE,
    puzzles_completed INT DEFAULT 0,
    is_unlocked BOOLEAN DEFAULT FALSE,
    is_completed BOOLEAN DEFAULT FALSE,
    unlocked_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    PRIMARY KEY (guest_id, set_id)
);

-- Create indexes for efficient guest set progress queries
CREATE INDEX idx_guest_set_progress_guest ON guest_set_progress(guest_id);
CREATE INDEX idx_guest_set_progress_unlocked ON guest_set_progress(guest_id, is_unlocked);

-- Add comments
COMMENT ON TABLE puzzle_sets IS 'Sets/groups of puzzles, each set contains 10 puzzles';
COMMENT ON COLUMN puzzle_sets.set_order IS 'Order of the set (1, 2, 3, ...) - determines unlock sequence';
COMMENT ON COLUMN puzzle_sets.difficulty IS 'Difficulty level of puzzles in this set (easy, medium, hard, expert)';
COMMENT ON COLUMN puzzle_pool.set_id IS 'Foreign key to puzzle_sets - which set this puzzle belongs to';
COMMENT ON TABLE guest_set_progress IS 'Tracks guest progress through puzzle sets';
COMMENT ON COLUMN guest_set_progress.puzzles_completed IS 'Number of puzzles completed in this set (0-10)';
COMMENT ON COLUMN guest_set_progress.is_unlocked IS 'Whether this set is unlocked for the guest';
COMMENT ON COLUMN guest_set_progress.is_completed IS 'Whether all 10 puzzles in this set are completed';