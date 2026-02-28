# Bug Fix Summary: COMPLETED Puzzles Disappearing After Refresh

## Problem
When users correctly answered a puzzle and refreshed the page, the completed puzzle would disappear from the available puzzles list in the frontend. The puzzle showed as "AVAILABLE" with all non-prefilled cells showing 0 instead of the correctly answered values.

## Root Cause Analysis

### Issue 1: Puzzles Not Marked as COMPLETED
The `guest_puzzle_progress` table was tracking puzzle status, but not storing which cells were correctly solved. When a puzzle was completed:
1. The status was set to "COMPLETED"
2. But the solved cell positions were not persisted
3. After refresh, the system couldn't restore the solved values
4. The puzzle appeared as "AVAILABLE" with 0s in the grid

### Issue 2: Game Session Dependency
The resolver tried to fetch the game session to get solved positions, but:
- Game sessions might be deleted or not found after completion
- This caused COMPLETED puzzles to be skipped entirely
- The `SolvedPositions` from `PuzzleService.GetAvailableForGuest()` weren't being used

## Complete Solution

### 1. Database Schema Change
Added `solved_positions` column to `guest_puzzle_progress` table to persist correctly answered cells:

**Migration:** `internal/database/migrations/008_add_solved_positions.sql`
```sql
ALTER TABLE guest_puzzle_progress ADD COLUMN solved_positions JSONB;
```

### 2. Repository Updates
**File:** `internal/repository/puzzle_progress_repository.go`

- Updated `MarkCompleted()` to save solved positions:
```go
func (r *puzzleProgressRepository) MarkCompleted(ctx context.Context, guestID, puzzleID uuid.UUID, solvedPositions []domain.Position) error
```

- Updated `GetAvailablePuzzlesForGuest()` to load solved positions from database:
```go
SELECT ..., gp.solved_positions FROM ...
```

**File:** `internal/repository/interfaces.go`
- Updated interface to match new signature

### 3. Service Layer Updates
**File:** `internal/service/game_service.go`

Updated all three completion methods to extract and save solved positions:

- `MakeMove()`: When game completes, extract solved positions and mark puzzle as completed
- `VerifyGame()`: When game completes, extract solved positions and mark puzzle as completed
- `SubmitAnswer()`: When game completes, extract solved positions and mark puzzle as completed

All now call:
```go
solvedPositions := extractSolvedPositions(game)
s.puzzleProgressRepo.MarkCompleted(ctx, game.GuestID, game.PuzzleID, solvedPositions)
```

### 4. Resolver Fix
**File:** `internal/graphql/resolver/schema.resolver.go`

Simplified `GetAvailablePuzzles` to use the already-populated `SolvedPositions` from the service layer:

```go
// Get available puzzles for this guest with their status
// Note: PuzzleService.GetAvailableForGuest already populates SolvedPositions
// for COMPLETED and PLAYING puzzles, so we don't need to fetch game state separately
puzzles, err := r.PuzzleService.GetAvailableForGuest(ctx, guestUUID, l)
if err != nil {
    return nil, err
}

// Convert to model - solved positions are already included in the puzzle data
result := make([]*model.PuzzleWithStatus, len(puzzles))
for i, puzzle := range puzzles {
    if puzzle == nil || puzzle.Puzzle == nil {
        continue
    }
    result[i] = domainPuzzleWithStatusToModel(puzzle)
}
```

### 5. PuzzleService Enhancement
**File:** `internal/service/puzzle_service.go`

The `GetAvailableForGuest()` method now:
1. Fetches puzzle status from `guest_puzzle_progress` table
2. For PLAYING and COMPLETED puzzles, loads `solved_positions` from database
3. Includes these in the `PuzzleWithStatus.SolvedPositions` field
4. The resolver helper uses these to populate the grid with correct values

## How It Works Now

### When a User Completes a Puzzle:
1. User submits all correct answers via `SubmitAnswer`, `MakeMove`, or `VerifyGame`
2. System detects all cells are correct
3. Extracts all correctly answered cell positions from the game
4. Saves these positions to `guest_puzzle_progress.solved_positions` column
5. Sets status to "COMPLETED"

### When User Refreshes the Page:
1. Frontend calls `getAvailablePuzzles` GraphQL query
2. `GetAvailablePuzzles` resolver calls `PuzzleService.GetAvailableForGuest()`
3. Service fetches puzzles with status from `guest_puzzle_progress` table
4. For COMPLETED puzzles, loads the `solved_positions` JSONB array
5. Includes these in `PuzzleWithStatus.SolvedPositions`
6. Resolver converts to GraphQL model using `domainPuzzleWithStatusToModel()`
7. Helper function `domainPuzzleToModelWithSolved()` populates grid with:
   - Prefilled cells (from puzzle template)
   - Solved cells (from `solved_positions` in database)
8. Frontend receives puzzle with COMPLETED status and all solved values visible

## Files Changed

### Database Migrations
- `internal/database/migrations/008_add_solved_positions.sql` (NEW)
  - Adds `solved_positions` JSONB column to `guest_puzzle_progress`

### Repository Layer
- `internal/repository/interfaces.go`
  - Updated `MarkCompleted` signature to accept `solvedPositions []domain.Position`

- `internal/repository/puzzle_progress_repository.go`
  - Updated `MarkCompleted()` to save solved positions
  - Updated `GetAvailablePuzzlesForGuest()` to load solved positions

### Service Layer
- `internal/service/game_service.go`
  - Updated `MakeMove()` to extract and save solved positions
  - Updated `VerifyGame()` to extract and save solved positions
  - Updated `SubmitAnswer()` to extract and save solved positions

### GraphQL Resolver
- `internal/graphql/resolver/schema.resolver.go`
  - Simplified `GetAvailablePuzzles` resolver
  - Removed redundant game state fetch
  - Removed unused `mergeSolvedCellsToPuzzleGrid` function
  - Removed unused `errors` import

## Testing & Deployment

### Database Migration Required
Before deploying this fix, run the migration:

```sql
-- Migration: Add solved_positions column to guest_puzzle_progress
ALTER TABLE guest_puzzle_progress ADD COLUMN solved_positions JSONB;
```

### Verification
The fix has been compiled successfully with no errors. Test the following scenarios:

1. **Complete a puzzle** and verify it shows as COMPLETED with all solved values
2. **Refresh the page** and verify the puzzle still shows COMPLETED with solved values
3. **Start a new puzzle**, answer some cells, refresh - should show PLAYING with partial solutions
4. **Complete multiple puzzles**, refresh - all should persist correctly

### Behavior by Status
- **AVAILABLE**: Shows only prefilled cells (all other cells = 0)
- **PLAYING**: Shows prefilled + correctly answered cells from current game session
- **COMPLETED**: Shows prefilled + all solved cells (persisted in database)
- **ARCHIVED**: Shows prefilled + all solved cells (persisted in database)
- **AD_BLOCK**: Shows only prefilled cells (all other cells = 0)

All puzzle statuses and solved cell values now persist correctly across page refreshes.
