-- Add status column to guest_puzzle_progress table
ALTER TABLE guest_puzzle_progress ADD COLUMN status VARCHAR(20) DEFAULT 'COMPLETED';

-- Update existing records to have COMPLETED status
UPDATE guest_puzzle_progress SET status = 'COMPLETED' WHERE status IS NULL;

-- Make the column NOT NULL
ALTER TABLE guest_puzzle_progress ALTER COLUMN status SET NOT NULL;

-- Create index on guest_id for faster queries
CREATE INDEX idx_guest_puzzle_progress_guest ON guest_puzzle_progress(guest_id);

-- Create index on status for filtering
CREATE INDEX idx_guest_puzzle_progress_status ON guest_puzzle_progress(status);