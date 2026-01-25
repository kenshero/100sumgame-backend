-- Seed: Insert sample puzzles for development
-- Run: docker exec -i sum100-db psql -U postgres -d sum100game < backend/scripts/seed_puzzles.sql

-- Puzzle 1: Medium difficulty
-- Each row and column sums to 100
INSERT INTO puzzle_pool (grid_solution, prefilled_positions, difficulty) VALUES (
    '[[20, 15, 25, 30, 10], [18, 22, 20, 25, 15], [25, 18, 17, 20, 20], [22, 25, 23, 15, 15], [15, 20, 15, 10, 40]]',
    '[{"row": 0, "col": 0}, {"row": 0, "col": 2}, {"row": 1, "col": 1}, {"row": 1, "col": 4}, {"row": 2, "col": 2}, {"row": 3, "col": 0}, {"row": 3, "col": 3}, {"row": 4, "col": 1}, {"row": 4, "col": 4}]',
    'medium'
);

-- Puzzle 2: Medium difficulty
INSERT INTO puzzle_pool (grid_solution, prefilled_positions, difficulty) VALUES (
    '[[15, 25, 20, 25, 15], [20, 20, 25, 20, 15], [25, 15, 20, 20, 20], [20, 25, 15, 20, 20], [20, 15, 20, 15, 30]]',
    '[{"row": 0, "col": 1}, {"row": 0, "col": 3}, {"row": 1, "col": 0}, {"row": 1, "col": 2}, {"row": 2, "col": 1}, {"row": 2, "col": 4}, {"row": 3, "col": 2}, {"row": 4, "col": 0}, {"row": 4, "col": 4}]',
    'medium'
);

-- Puzzle 3: Medium difficulty
INSERT INTO puzzle_pool (grid_solution, prefilled_positions, difficulty) VALUES (
    '[[10, 30, 20, 25, 15], [25, 15, 20, 20, 20], [20, 20, 25, 15, 20], [30, 15, 15, 25, 15], [15, 20, 20, 15, 30]]',
    '[{"row": 0, "col": 0}, {"row": 0, "col": 1}, {"row": 1, "col": 3}, {"row": 2, "col": 2}, {"row": 2, "col": 4}, {"row": 3, "col": 1}, {"row": 3, "col": 4}, {"row": 4, "col": 2}, {"row": 4, "col": 3}]',
    'medium'
);

-- Puzzle 4: Medium difficulty
INSERT INTO puzzle_pool (grid_solution, prefilled_positions, difficulty) VALUES (
    '[[22, 18, 20, 22, 18], [18, 22, 20, 18, 22], [20, 20, 20, 20, 20], [22, 18, 20, 22, 18], [18, 22, 20, 18, 22]]',
    '[{"row": 0, "col": 0}, {"row": 0, "col": 4}, {"row": 1, "col": 1}, {"row": 1, "col": 3}, {"row": 2, "col": 2}, {"row": 3, "col": 0}, {"row": 3, "col": 4}, {"row": 4, "col": 1}, {"row": 4, "col": 3}]',
    'medium'
);

-- Puzzle 5: Medium difficulty
INSERT INTO puzzle_pool (grid_solution, prefilled_positions, difficulty) VALUES (
    '[[12, 28, 18, 24, 18], [24, 16, 22, 20, 18], [20, 20, 20, 20, 20], [26, 18, 22, 16, 18], [18, 18, 18, 20, 26]]',
    '[{"row": 0, "col": 1}, {"row": 0, "col": 3}, {"row": 1, "col": 0}, {"row": 1, "col": 4}, {"row": 2, "col": 2}, {"row": 3, "col": 1}, {"row": 3, "col": 3}, {"row": 4, "col": 0}, {"row": 4, "col": 4}]',
    'medium'
);

SELECT 'Seeded ' || COUNT(*) || ' puzzles' as result FROM puzzle_pool;
