# Puzzle Sets System

## Overview

The puzzle sets system organizes puzzles into groups (sets) of 10 puzzles each. Players must complete all 10 puzzles in a set to unlock the next set by watching an ad.

## Database Structure

### Tables

1. **puzzle_sets**
   - Stores puzzle set information
   - Fields: id, set_order, name, puzzles_count, created_at

2. **guest_set_progress**
   - Tracks each guest's progress through puzzle sets
   - Fields: guest_id, set_id, puzzles_completed, is_unlocked, is_completed, unlocked_at, completed_at

3. **puzzle_pool** (updated)
   - Now includes set_id field to link puzzles to sets
   - Each puzzle belongs to exactly one set

## Setup Instructions

### 1. Run Migration

Apply database migration to create new tables:

```bash
# Make sure you're in the project root directory (where docker-compose.yml is)
cd /home/kenshero/projects/100sumgame

# Apply migration 009
docker-compose exec -T db psql -U postgres -d sum100game < backend/internal/database/migrations/009_add_puzzle_sets.sql
```

### 2. Generate Puzzle Sets

Use the seeding script to create puzzle sets with puzzles:

```bash
# Make sure you're in the backend directory
cd /home/kenshero/projects/100sumgame/backend

# Run's seeding script
go run scripts/seed_puzzle_sets.go
```

This will create:
- 3 puzzle sets (Set 1, Set 2, Set 3)
- 10 puzzles per set (30 total puzzles)
- Each puzzle with varying difficulty (EASY, MEDIUM, HARD)

### 3. Reset Puzzle Sets (if needed)

To clear all puzzle sets and start fresh:

```bash
# Make sure you're in project root directory (where docker-compose.yml is)
cd /home/kenshero/projects/100sumgame

# Run's reset script
docker-compose exec -T db psql -U postgres -d sum100game < backend/scripts/reset_puzzle_sets.sql
```

## API Usage

### Get Current Set for Guest

Query to get the current unlocked set for a guest:

```graphql
query GetCurrentSet {
  getCurrentSet {
    set {
      id
      name
      setOrder
      puzzlesCount
    }
    puzzlesCompleted
    isUnlocked
    isCompleted
  }
}
```

### Get Puzzles in Current Set

Query to get available puzzles from the current unlocked set:

```graphql
query GetAvailablePuzzles {
  getAvailablePuzzlesWithSets(limit: 10) {
    puzzle {
      id
      set {
        id
        name
        setOrder
      }
      difficulty
    }
    status
    solvedPositions {
      row
      col
    }
  }
}
```

### Unlock Next Set

Mutation to unlock the next set after watching an ad:

```graphql
mutation UnlockNextSet {
  unlockNextSet {
    set {
      id
      name
      setOrder
    }
    isUnlocked
    unlockedAt
  }
}
```

## Game Flow

1. **New Player**:
   - System automatically unlocks Set 1
   - Player can access all 10 puzzles in Set 1

2. **Playing Puzzles**:
   - Player completes puzzles one by one
   - System tracks progress (puzzles completed / 10)
   - Progress is updated in real-time as puzzles are completed

3. **Set Completion**:
   - When player completes all 10 puzzles in a set:
     - Set is marked as completed
     - Player can see progress: "10/10"
     - Next set remains locked

4. **Unlocking Next Set**:
   - Player clicks "Unlock Next Set" button
   - Player watches an ad (implemented in frontend)
   - Frontend calls `unlockNextSet` mutation
   - System unlocks the next set (Set 2, Set 3, etc.)
   - Player can now access puzzles in the new set

## Progress Tracking

The system tracks:
- Individual puzzle completion status
- Set progress (number of puzzles completed in current set)
- Set unlock status (locked/unlocked)
- Set completion status (completed/not completed)

## Testing

### Verify Puzzle Sets Exist

```sql
SELECT * FROM puzzle_sets ORDER BY set_order;
```

### Verify Puzzles Are Assigned to Sets

```sql
SELECT set_order, COUNT(*) as puzzle_count
FROM puzzle_pool pp
JOIN puzzle_sets ps ON pp.set_id = ps.id
GROUP BY set_order
ORDER BY set_order;
```

### Check Guest Progress

```sql
SELECT * FROM guest_set_progress WHERE guest_id = 'your-guest-id';
```

## Notes

- Each set must have exactly 10 puzzles
- Puzzles are generated automatically with varying difficulty
- Difficulty is based on the number of prefilled cells:
  - 5+ prefilled cells: EASY
  - 4 prefilled cells: MEDIUM
  - 3 or fewer prefilled cells: HARD
- Sets are ordered (Set 1, Set 2, Set 3, etc.)
- Players must unlock sets in order (no skipping)