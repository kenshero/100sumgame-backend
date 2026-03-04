# Filter getAvailablePuzzles by Current Set

## Problem

The `getAvailablePuzzles` API was returning puzzles from multiple sets instead of just the current unlocked set, causing the frontend to receive 20 puzzles from 2 different sets at once.

## Root Cause

### **Original Behavior:**

```go
func GetAvailableForGuest(ctx, guestID, limit) {
    // ❌ Get ALL puzzles for guest, no set filtering
    puzzles := repo.GetAvailablePuzzlesForGuest(ctx, guestID, limit)
    return puzzles
}
```

**Result:**
- Returns up to 20 puzzles from ANY set
- If guest has puzzles in Set 1 and Set 2 → returns mix from both
- Breaks the intended flow: "unlock set → play puzzles in that set → complete → unlock next"

## Solution

Modified `GetAvailableForGuest` to filter puzzles by the guest's current unlocked set only.

### **New Behavior:**

```go
func GetAvailableForGuest(ctx, guestID, limit) {
    // 1. Get current unlocked set for guest
    setProgress := GetCurrentSet(ctx, guestID)
    
    // 2. Get all puzzles for guest
    puzzles := repo.GetAvailablePuzzlesForGuest(ctx, guestID, limit)
    
    // 3. Filter puzzles to only include those from current set
    filteredPuzzles := make([]*PuzzleWithStatus, 0)
    for _, puzzle := range puzzles {
        if puzzle.Puzzle.SetID != nil && 
           *puzzle.Puzzle.SetID == setProgress.SetID {
            filteredPuzzles = append(filteredPuzzles, puzzle)
        }
    }
    
    return filteredPuzzles
}
```

## Code Changes

### **File Modified:**

`internal/service/puzzle_service.go` - `GetAvailableForGuest` function

### **Before:**
```go
func (s *PuzzleService) GetAvailableForGuest(ctx context.Context, guestID uuid.UUID, limit int) ([]*domain.PuzzleWithStatus, error) {
    if limit <= 0 {
        limit = 20
    }

    puzzles, err := s.PuzzleProgressRepo.GetAvailablePuzzlesForGuest(ctx, guestID, limit)
    if err != nil {
        return nil, domain.ErrNoPuzzlesAvailable
    }

    if len(puzzles) == 0 {
        return []*domain.PuzzleWithStatus{}, nil
    }

    // Process puzzles...
    for _, puzzleWithStatus := range puzzles {
        // Update status and solved positions...
    }

    return puzzles, nil
}
```

### **After:**
```go
func (s *PuzzleService) GetAvailableForGuest(ctx context.Context, guestID uuid.UUID, limit int) ([]*domain.PuzzleWithStatus, error) {
    if limit <= 0 {
        limit = 20
    }

    // ✅ Step 1: Get current unlocked set for guest
    setProgress, err := s.GetCurrentSet(ctx, guestID)
    if err != nil {
        return nil, err
    }

    // Step 2: Get all puzzles for guest
    puzzles, err := s.PuzzleProgressRepo.GetAvailablePuzzlesForGuest(ctx, guestID, limit)
    if err != nil {
        return nil, domain.ErrNoPuzzlesAvailable
    }

    if len(puzzles) == 0 {
        return []*domain.PuzzleWithStatus{}, nil
    }

    // ✅ Step 3: Filter puzzles to only include those from current unlocked set
    filteredPuzzles := make([]*domain.PuzzleWithStatus, 0)
    for _, puzzleWithStatus := range puzzles {
        if puzzleWithStatus.Puzzle == nil {
            continue
        }
        // Check if puzzle belongs to current unlocked set
        if puzzleWithStatus.Puzzle.SetID != nil && *puzzleWithStatus.Puzzle.SetID == setProgress.SetID {
            filteredPuzzles = append(filteredPuzzles, puzzleWithStatus)
        }
    }

    // Step 4: Update playing status and solved positions for filtered puzzles
    for _, puzzleWithStatus := range filteredPuzzles {
        // Update status and solved positions...
    }

    return filteredPuzzles, nil
}
```

## Game Flow

### **Complete Player Journey:**

```
1. New Player:
   getCurrentSet() → Set 1 unlocked
   getAvailablePuzzles() → 10 puzzles from Set 1 ✅

2. Playing Set 1:
   getAvailablePuzzles() → 10 puzzles from Set 1 ✅
   Player completes 5/10 puzzles...

3. Completed Set 1:
   getCurrentSet() → Set 1 completed
   getAvailablePuzzles() → 10 puzzles from Set 1 ✅
   Frontend shows "Unlock Next Set" button

4. Unlock Set 2:
   unlockPuzzlesAfterAd() → Set 2 unlocked
   getCurrentSet() → Set 2 unlocked
   getAvailablePuzzles() → 10 puzzles from Set 2 ✅

5. Playing Set 2:
   getAvailablePuzzles() → 10 puzzles from Set 2 ✅
   Player completes all 10 puzzles...

6. Completed Set 50 (last set):
   getCurrentSet() → Set 50 completed, isLastSet=true
   getAvailablePuzzles() → 10 puzzles from Set 50 ✅
   Frontend shows "Congratulations!" message
```

## API Response Examples

### **Scenario 1: New Player - Set 1**

**Request:**
```graphql
query GetPuzzles {
  getAvailablePuzzles(guestId: "new-guest-uuid", limit: 20) {
    puzzle {
      id
      set {
        setOrder
      }
    }
    status
  }
}
```

**Response:**
```json
{
  "getAvailablePuzzles": [
    { "puzzle": { "set": { "setOrder": 1 } }, "status": "AVAILABLE" },
    { "puzzle": { "set": { "setOrder": 1 } }, "status": "AVAILABLE" },
    { "puzzle": { "set": { "setOrder": 1 } }, "status": "AVAILABLE" },
    ...
    // ✅ Exactly 10 puzzles, ALL from Set 1
  ]
}
```

### **Scenario 2: Completed Set 1, Unlocked Set 2**

**Request:**
```graphql
query GetPuzzles {
  getAvailablePuzzles(guestId: "guest-uuid", limit: 20) {
    puzzle {
      id
      set {
        setOrder
      }
    }
    status
  }
}
```

**Response:**
```json
{
  "getAvailablePuzzles": [
    { "puzzle": { "set": { "setOrder": 2 } }, "status": "AVAILABLE" },
    { "puzzle": { "set": { "setOrder": 2 } }, "status": "AVAILABLE" },
    ...
    // ✅ 10 puzzles from Set 2 only
    // ❌ NOT from Set 1 (even though completed)
  ]
}
```

### **Scenario 3: Completed All Sets**

**Request:**
```graphql
query GetPuzzles {
  getAvailablePuzzles(guestId: "completed-guest-uuid", limit: 20) {
    puzzle {
      id
      set {
        setOrder
      }
    }
    status
  }
}
```

**Response:**
```json
{
  "getAvailablePuzzles": [
    { "puzzle": { "set": { "setOrder": 50 } }, "status": "COMPLETED" },
    { "puzzle": { "set": { "setOrder": 50 } }, "status": "COMPLETED" },
    ...
    // ✅ 10 puzzles from Set 50 (last set)
    // All marked as COMPLETED
  ]
}
```

## Frontend Integration

### **Required Queries:**

```graphql
# 1. Get current set info
query GetCurrentSet {
  getCurrentSet(guestId: "uuid") {
    setId
    puzzlesCompleted
    isUnlocked
    isCompleted
    isLastSet
  }
}

# 2. Get puzzles from current set
query GetPuzzles {
  getAvailablePuzzles(guestId: "uuid", limit: 20) {
    puzzle {
      id
      set {
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

### **Frontend Logic:**

```javascript
function PuzzleScreen() {
  const { data: setInfo } = useQuery(GET_CURRENT_SET);
  const { data: puzzles } = useQuery(GET_PUZZLES);
  
  // Set is completed
  if (setInfo?.getCurrentSet?.isCompleted) {
    // Check if last set
    if (setInfo?.getCurrentSet?.isLastSet) {
      return <CongratulationsScreen />;
    }
    return <UnlockNextSetButton />;
  }
  
  // Display puzzles from current set
  return <PuzzleList puzzles={puzzles.getAvailablePuzzles} />;
}
```

## Benefits

✅ **Correct Flow** - Players must unlock sets sequentially  
✅ **Puzzles per Set** - Exactly 10 puzzles per set  
✅ **Clear Progress** - Easy to track progress through sets  
✅ **Predictable** - Frontend knows which set's puzzles to display  
✅ **Unlock Mechanism** - Works as designed (watch ad → unlock next set)  
✅ **End Game Detection** - `isLastSet` field shows when all sets completed  

## Testing

### **Test Case 1: New Player**

```bash
# Create new guest
# Expected: getAvailablePuzzles returns 10 puzzles from Set 1
```

### **Test Case 2: Partial Completion**

```bash
# Complete 5 puzzles in Set 1
# Expected: getAvailablePuzzles returns 10 puzzles from Set 1 (5 completed, 5 available)
```

### **Test Case 3: Set Completed**

```bash
# Complete all 10 puzzles in Set 1
# Expected: getAvailablePuzzles returns 10 puzzles from Set 1 (all completed)
# Expected: getCurrentSet returns isCompleted=true
```

### **Test Case 4: Next Set Unlocked**

```bash
# Unlock Set 2 via unlockPuzzlesAfterAd
# Expected: getAvailablePuzzles returns 10 puzzles from Set 2 (all available)
# Expected: getCurrentSet returns isCompleted=false
```

### **Test Case 5: All Sets Completed**

```bash
# Complete all 50 sets
# Expected: getAvailablePuzzles returns 10 puzzles from Set 50 (all completed)
# Expected: getCurrentSet returns isLastSet=true, isCompleted=true
```

## Breaking Changes

⚠️ **Note:** This is a breaking change from the previous behavior.

### **Before:**
- `getAvailablePuzzles` returned puzzles from multiple sets
- Limit of 20 could include puzzles from 2 different sets

### **After:**
- `getAvailablePuzzles` returns puzzles ONLY from current unlocked set
- Limit of 20 will return max 10 puzzles (puzzles per set)

### **Impact:**
- Frontend will only receive puzzles from the current unlocked set
- Must rely on `getCurrentSet` to know which set is active
- Must call `unlockPuzzlesAfterAd` to progress to next set

## Related Changes

- `internal/service/puzzle_service.go` - Modified `GetAvailableForGuest` to filter by set
- `DUPLICATE_KEY_FIX.md` - Fixed GetCurrentSet to handle completed sets
- `SEED_PUZZLE_SETS_DUPLICATE_FIX.md` - Fixed seed script to handle duplicates

## Summary

This change ensures that the `getAvailablePuzzles` API only returns puzzles from the guest's current unlocked set, enforcing the intended game flow where players:
1. Start with Set 1 (10 puzzles)
2. Complete all puzzles in Set 1
3. Watch an ad to unlock Set 2
4. Complete all puzzles in Set 2
5. Repeat until all sets are completed

This provides a clear progression system and prevents players from seeing puzzles from sets they haven't unlocked yet.