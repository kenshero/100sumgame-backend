-- Seed: Insert sample puzzles for development
-- Run: docker exec -i sum100-db psql -U postgres -d sum100game < backend/scripts/seed_puzzles.sql

-- 1. Reset Database (Clear all data and reset relations)
TRUNCATE TABLE guest_puzzle_progress CASCADE;
TRUNCATE TABLE guest_set_progress CASCADE;
TRUNCATE TABLE puzzle_pool CASCADE;
TRUNCATE TABLE puzzle_sets CASCADE;

-- 2. Create Set 1 (Using a fixed UUID for easy reference)
INSERT INTO puzzle_sets (id, set_order, difficulty) 
VALUES ('11111111-1111-1111-1111-111111111111', 1, 'medium');

-- 3. Insert 10 Puzzles linked to Set 1
-- Puzzle 1
INSERT INTO puzzle_pool (set_id, grid_solution, prefilled_positions, difficulty) VALUES (
    '11111111-1111-1111-1111-111111111111',
    '[[20, 15, 25, 30, 10], [18, 22, 20, 25, 15], [25, 18, 17, 20, 20], [22, 25, 23, 15, 15], [15, 20, 15, 10, 40]]',
    '[{"row": 0, "col": 0}, {"row": 0, "col": 2}, {"row": 1, "col": 1}, {"row": 1, "col": 4}, {"row": 2, "col": 2}, {"row": 3, "col": 0}, {"row": 3, "col": 3}, {"row": 4, "col": 1}, {"row": 4, "col": 4}]',
    'medium'
);

-- Puzzle 2
INSERT INTO puzzle_pool (set_id, grid_solution, prefilled_positions, difficulty) VALUES (
    '11111111-1111-1111-1111-111111111111',
    '[[15, 25, 20, 25, 15], [20, 20, 25, 20, 15], [25, 15, 20, 20, 20], [20, 25, 15, 20, 20], [20, 15, 20, 15, 30]]',
    '[{"row": 0, "col": 1}, {"row": 0, "col": 3}, {"row": 1, "col": 0}, {"row": 1, "col": 2}, {"row": 2, "col": 1}, {"row": 2, "col": 4}, {"row": 3, "col": 2}, {"row": 4, "col": 0}, {"row": 4, "col": 4}]',
    'medium'
);

-- Puzzle 3
INSERT INTO puzzle_pool (set_id, grid_solution, prefilled_positions, difficulty) VALUES (
    '11111111-1111-1111-1111-111111111111',
    '[[10, 30, 20, 25, 15], [25, 15, 20, 20, 20], [20, 20, 25, 15, 20], [30, 15, 15, 25, 15], [15, 20, 20, 15, 30]]',
    '[{"row": 0, "col": 0}, {"row": 0, "col": 1}, {"row": 1, "col": 3}, {"row": 2, "col": 2}, {"row": 2, "col": 4}, {"row": 3, "col": 1}, {"row": 3, "col": 4}, {"row": 4, "col": 2}, {"row": 4, "col": 3}]',
    'medium'
);

-- Puzzle 4
INSERT INTO puzzle_pool (set_id, grid_solution, prefilled_positions, difficulty) VALUES (
    '11111111-1111-1111-1111-111111111111',
    '[[22, 18, 20, 22, 18], [18, 22, 20, 18, 22], [20, 20, 20, 20, 20], [22, 18, 20, 22, 18], [18, 22, 20, 18, 22]]',
    '[{"row": 0, "col": 0}, {"row": 0, "col": 4}, {"row": 1, "col": 1}, {"row": 1, "col": 3}, {"row": 2, "col": 2}, {"row": 3, "col": 0}, {"row": 3, "col": 4}, {"row": 4, "col": 1}, {"row": 4, "col": 3}]',
    'medium'
);

-- Puzzle 5
INSERT INTO puzzle_pool (set_id, grid_solution, prefilled_positions, difficulty) VALUES (
    '11111111-1111-1111-1111-111111111111',
    '[[12, 28, 18, 24, 18], [24, 16, 22, 20, 18], [20, 20, 20, 20, 20], [26, 18, 22, 16, 18], [18, 18, 18, 20, 26]]',
    '[{"row": 0, "col": 1}, {"row": 0, "col": 3}, {"row": 1, "col": 0}, {"row": 1, "col": 4}, {"row": 2, "col": 2}, {"row": 3, "col": 1}, {"row": 3, "col": 3}, {"row": 4, "col": 0}, {"row": 4, "col": 4}]',
    'medium'
);

-- Puzzle 6 (Duplicate of 1 for testing 10 puzzles)
INSERT INTO puzzle_pool (set_id, grid_solution, prefilled_positions, difficulty) VALUES (
    '11111111-1111-1111-1111-111111111111',
    '[[20, 15, 25, 30, 10], [18, 22, 20, 25, 15], [25, 18, 17, 20, 20], [22, 25, 23, 15, 15], [15, 20, 15, 10, 40]]',
    '[{"row": 0, "col": 0}, {"row": 0, "col": 2}, {"row": 1, "col": 1}, {"row": 1, "col": 4}, {"row": 2, "col": 2}, {"row": 3, "col": 0}, {"row": 3, "col": 3}, {"row": 4, "col": 1}, {"row": 4, "col": 4}]',
    'medium'
);

-- Puzzle 7 (Duplicate of 2 for testing 10 puzzles)
INSERT INTO puzzle_pool (set_id, grid_solution, prefilled_positions, difficulty) VALUES (
    '11111111-1111-1111-1111-111111111111',
    '[[15, 25, 20, 25, 15], [20, 20, 25, 20, 15], [25, 15, 20, 20, 20], [20, 25, 15, 20, 20], [20, 15, 20, 15, 30]]',
    '[{"row": 0, "col": 1}, {"row": 0, "col": 3}, {"row": 1, "col": 0}, {"row": 1, "col": 2}, {"row": 2, "col": 1}, {"row": 2, "col": 4}, {"row": 3, "col": 2}, {"row": 4, "col": 0}, {"row": 4, "col": 4}]',
    'medium'
);

-- Puzzle 8 (Duplicate of 3 for testing 10 puzzles)
INSERT INTO puzzle_pool (set_id, grid_solution, prefilled_positions, difficulty) VALUES (
    '11111111-1111-1111-1111-111111111111',
    '[[10, 30, 20, 25, 15], [25, 15, 20, 20, 20], [20, 20, 25, 15, 20], [30, 15, 15, 25, 15], [15, 20, 20, 15, 30]]',
    '[{"row": 0, "col": 0}, {"row": 0, "col": 1}, {"row": 1, "col": 3}, {"row": 2, "col": 2}, {"row": 2, "col": 4}, {"row": 3, "col": 1}, {"row": 3, "col": 4}, {"row": 4, "col": 2}, {"row": 4, "col": 3}]',
    'medium'
);

-- Puzzle 9 (Duplicate of 4 for testing 10 puzzles)
INSERT INTO puzzle_pool (set_id, grid_solution, prefilled_positions, difficulty) VALUES (
    '11111111-1111-1111-1111-111111111111',
    '[[22, 18, 20, 22, 18], [18, 22, 20, 18, 22], [20, 20, 20, 20, 20], [22, 18, 20, 22, 18], [18, 22, 20, 18, 22]]',
    '[{"row": 0, "col": 0}, {"row": 0, "col": 4}, {"row": 1, "col": 1}, {"row": 1, "col": 3}, {"row": 2, "col": 2}, {"row": 3, "col": 0}, {"row": 3, "col": 4}, {"row": 4, "col": 1}, {"row": 4, "col": 3}]',
    'medium'
);

-- Puzzle 10 (Duplicate of 5 for testing 10 puzzles)
INSERT INTO puzzle_pool (set_id, grid_solution, prefilled_positions, difficulty) VALUES (
    '11111111-1111-1111-1111-111111111111',
    '[[12, 28, 18, 24, 18], [24, 16, 22, 20, 18], [20, 20, 20, 20, 20], [26, 18, 22, 16, 18], [18, 18, 18, 20, 26]]',
    '[{"row": 0, "col": 1}, {"row": 0, "col": 3}, {"row": 1, "col": 0}, {"row": 1, "col": 4}, {"row": 2, "col": 2}, {"row": 3, "col": 1}, {"row": 3, "col": 3}, {"row": 4, "col": 0}, {"row": 4, "col": 4}]',
    'medium'
);

SELECT 'Seeded ' || COUNT(*) || ' puzzles into Set 1' as result FROM puzzle_pool;
