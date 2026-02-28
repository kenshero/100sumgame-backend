-- Add solved_positions column to guest_puzzle_progress table
-- This stores the positions of correctly answered cells for completed puzzles
ALTER TABLE guest_puzzle_progress ADD COLUMN solved_positions JSONB;

-- Add comment
COMMENT ON COLUMN guest_puzzle_progress.solved_positions IS 'Array of {row, col} positions that were correctly answered when puzzle was completed';