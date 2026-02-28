-- Reset Puzzle Sets System
-- This script will clear all puzzle sets, guest set progress, and related data

-- Disable foreign key constraints temporarily
SET session_replication_role = replica;

-- Clear guest set progress
TRUNCATE TABLE guest_set_progress CASCADE;

-- Clear puzzle sets
TRUNCATE TABLE puzzle_sets CASCADE;

-- Clear puzzles (optional - uncomment if you want to clear all puzzles too)
-- TRUNCATE TABLE puzzle_pool CASCADE;

-- Re-enable foreign key constraints
SET session_replication_role = DEFAULT;

-- Verify cleanup
SELECT '✅ Reset complete!' AS status;
SELECT COUNT(*) AS remaining_puzzle_sets FROM puzzle_sets;
SELECT COUNT(*) AS remaining_guest_progress FROM guest_set_progress;