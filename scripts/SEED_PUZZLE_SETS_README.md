# Puzzle Sets Seed Script

## Overview

This script generates and seeds puzzle sets with 10 puzzles each into the database. The number of sets is configurable via command-line argument.

## Features

- **Configurable sets**: Specify how many sets to generate (default: 50)
- **10 puzzles per set**: Each set contains 10 puzzles
- **Random values per cell**: 1-80 (configurable)
- **Random prefilled positions**: 3-5 cells per puzzle
- **Valid grids**: All puzzles sum to 100 exactly
- **Progress indicator**: Shows progress for each set
- **Command-line flags**: Easy to specify number of sets

## Configuration

Edit the constants at the top of `seed_puzzle_sets.go`:

```go
const (
    maxVal        = 80 // Maximum value per cell (1-80)
    minPrefill    = 3  // Minimum prefilled cells
    maxPrefill    = 5  // Maximum prefilled cells
    puzzlesPerSet = 10 // Puzzles per set
)
```

## How to Use

### 1. Make sure PostgreSQL is running

```bash
docker-compose up -d postgres
```

### 2. Run the seed script

**Default (50 sets):**
```bash
cd scripts
go run seed_puzzle_sets.go
```

**Custom number of sets:**
```bash
cd scripts

# Generate 10 sets
go run seed_puzzle_sets.go -sets=10

# Generate 1 set (for testing)
go run seed_puzzle_sets.go -sets=1

# Generate 100 sets
go run seed_puzzle_sets.go -sets=100
```

**Show help:**
```bash
go run seed_puzzle_sets.go -help
```

### 3. Expected Output

```
Connected to database successfully
🎯 Creating 50 puzzle sets with 10 puzzles each...
   Cell value range: 1-80
   Prefilled cells: 3-5

[1/50] Creating Set 1...
✓ Set 1 completed with 10 puzzles
[2/50] Creating Set 2...
✓ Set 2 completed with 10 puzzles
...

✅ All puzzle sets and puzzles created successfully!
📊 Total: 50 sets × 10 puzzles = 500 puzzles
🎲 Cell values: 1-80
📍 Prefilled cells: 3-5
```

## Database Connection

The script connects to PostgreSQL using this connection string:

```go
const dbURL = "postgres://postgres:postgres@localhost:5432/sum100game?sslmode=disable"
```

If you're using a different database configuration, update the `dbURL` constant.

## Cleaning Up

To delete all existing puzzle sets and start fresh, uncomment these lines in `main()`:

```go
// Delete existing puzzle sets and puzzles (uncomment to clean slate)
// _, err = db.Exec(ctx, "DELETE FROM guest_set_progress")
// if err != nil {
//     log.Fatalf("Failed to delete guest set progress: %v", err)
// }
// _, err = db.Exec(ctx, "DELETE FROM puzzle_sets CASCADE")
// if err != nil {
//     log.Fatalf("Failed to delete puzzle sets: %v", err)
// }
```

Then run the script again to seed fresh puzzles.

## Algorithm Details

### Grid Generation

1. **Generate random values** for first 24 cells (1 to maxVal)
2. **Calculate sum** of first 24 cells
3. **Determine last cell value**: `100 - sum`
4. **Validate last cell** is within range (1 to maxVal)
5. **Redistribute if needed**: If last cell is out of range, adjust other cells

This ensures:
- ✅ Every cell has a value between 1 and maxVal
- ✅ All 25 cells sum to exactly 100
- ✅ Values are randomly distributed
- ✅ No invalid or impossible puzzles

### Prefilled Positions

- **Random selection**: 3-5 cells are randomly chosen to be prefilled
- **Uniform distribution**: All 25 positions have equal probability
- **Player perspective**: These cells are shown with values filled in

## Database Schema

### puzzle_sets

| Column      | Type         | Description              |
|-------------|--------------|--------------------------|
| id          | UUID         | Unique set ID            |
| set_order   | INT          | Order (1-50)            |
| difficulty  | VARCHAR(20) | Difficulty level           |
| created_at  | TIMESTAMP    | Creation timestamp         |

### puzzle_pool

| Column               | Type        | Description                          |
|----------------------|-------------|--------------------------------------|
| id                   | UUID        | Unique puzzle ID                      |
| set_id               | UUID        | Foreign key to puzzle_sets            |
| grid_solution         | JSONB       | 5x5 solution grid                   |
| prefilled_positions   | JSONB       | Positions with prefilled values        |
| difficulty           | VARCHAR(20) | Difficulty level                     |
| created_at          | TIMESTAMP   | Creation timestamp                    |

## Troubleshooting

### Connection Failed

```
Failed to connect to database: ...
```

**Solution**: 
- Check PostgreSQL is running: `docker-compose ps`
- Verify database URL in `seed_puzzle_sets.go`
- Check database exists and is accessible

### Insert Failed

```
Failed to insert puzzle X for set Y: ...
```

**Solution**:
- Check migration 009 was applied: `CREATE TABLE puzzle_sets`
- Verify database schema is correct
- Check for duplicate set_order values

## Customization

### Change Number of Sets

Use command-line flag:

```bash
# Generate 10 sets
go run seed_puzzle_sets.go -sets=10

# Generate 1 set
go run seed_puzzle_sets.go -sets=1

# Generate 100 sets
go run seed_puzzle_sets.go -sets=100
```

**Note:** The number of sets must be between 1 and 1000.

### Change Difficulty

All puzzles use "MEDIUM" difficulty. To vary by set:

```go
difficulty := "MEDIUM"
if setOrder <= 10 {
    difficulty = "EASY"
} else if setOrder > 40 {
    difficulty = "HARD"
}
```

### Change Cell Value Range

```go
const maxVal = 80 // Increase to 99 for harder puzzles
```

## Performance

- **Generation time**: ~1-2 seconds per set
- **Total time**: ~1-2 minutes for 50 sets
- **Database size**: ~500 puzzles × ~1KB each = ~500KB

## Notes

- All puzzles are mathematically valid (sum to 100)
- Values are randomly generated each run
- Same seed script will produce different puzzles
- Puzzle IDs are random UUIDs
- Set orders are sequential (1-50)