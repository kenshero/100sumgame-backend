# Fix for Duplicate Key Error in GetCurrentSet

## Problem

When calling `GetCurrentSet` API, if a player has completed all puzzle sets, the system would throw:
```
ERROR: duplicate key value violates unique constraint "guest_set_progress_pkey" (SQLSTATE 23505)
```

## Root Cause

The `GetCurrentSet()` service had logic flaw:

1. **Original Logic:**
   ```go
   func GetCurrentSet(guestID) {
       // Try to get unlocked AND not completed set
       setProgress := GetUnlockedSet(guestID)
       
       if setProgress != nil {
           return setProgress  // ✅ Normal case
       }
       
       // ❌ BUG: If setProgress is nil (e.g., completed last set),
       // assumes it's a new player and tries to create set 1 again!
       CreateNewSet(guestID, setID: 1)  // ← Duplicate key!
   }
   ```

2. **Why it failed:**
   - When player completed last set → `is_completed = true`
   - `GetUnlockedSet()` only returns sets with `is_completed = false`
   - Returns `nil` → Service thinks it's a new player
   - Tries to insert set 1 → **Duplicate key error!**

## Solution

Added check for completed sets before creating new ones:

### **New Logic:**

```go
func GetCurrentSet(guestID) (*GuestSetProgress, error) {
    // Step 1: Try to get unlocked AND not completed set
    setProgress := GetUnlockedSet(guestID)
    if setProgress != nil {
        return setProgress  // ✅ Normal case
    }
    
    // Step 2: Check if player has completed sets
    allProgress := GetByGuest(guestID)
    if len(allProgress) > 0 {
        lastSet := allProgress[len(allProgress)-1]  // Get last set (highest order)
        if lastSet.IsCompleted {
            // ✅ Player completed all sets - return last completed set
            lastSet.IsLastSet = true
            return lastSet, nil
        }
    }
    
    // Step 3: Only create new set if no progress at all
    CreateNewSet(guestID, setID: 1)
}
```

## What Changed

### Before:
```go
func GetCurrentSet(guestID) {
    setProgress := GetUnlockedSet(guestID)
    if setProgress != nil {
        return setProgress
    }
    
    // ❌ Directly create new set - no check for completed sets
    CreateNewSet(guestID, setID: 1)
}
```

### After:
```go
func GetCurrentSet(guestID) {
    setProgress := GetUnlockedSet(guestID)
    if setProgress != nil {
        return setProgress
    }
    
    // ✅ Check for completed sets first
    allProgress := GetByGuest(guestID)
    if len(allProgress) > 0 {
        lastSet := allProgress[len(allProgress)-1]
        if lastSet.IsCompleted {
            lastSet.IsLastSet = true
            return lastSet, nil
        }
    }
    
    // ✅ Only create new set if no progress at all
    CreateNewSet(guestID, setID: 1)
}
```

## Flow Diagram

### **Player Progress Through Sets:**

```
New Player:
  GetCurrentSet() → No progress → Create Set 1 → Return Set 1
  
Playing Set 1:
  GetCurrentSet() → GetUnlockedSet() → Returns Set 1 ✅
  
Completed Set 1 (but not unlocked Set 2):
  GetCurrentSet() → GetUnlockedSet() → Returns nil (no unlocked sets)
                → GetByGuest() → Returns Set 1 (isCompleted: true)
                → Returns Set 1 with isLastSet=false ✅
                
Completed Set 50 (last set):
  GetCurrentSet() → GetUnlockedSet() → Returns nil (no unlocked sets)
                → GetByGuest() → Returns Set 50 (isCompleted: true)
                → Returns Set 50 with isLastSet=true ✅
                → Frontend hides unlock button, shows completion message
```

## Testing

### **Test Case 1: New Player**
```graphql
query GetCurrentSet {
  getCurrentSet(guestId: "new-guest-uuid") {
    puzzlesCompleted
    isUnlocked
    isCompleted
    isLastSet
  }
}
```

**Expected:**
```json
{
  "getCurrentSet": {
    "puzzlesCompleted": 0,
    "isUnlocked": true,
    "isCompleted": false,
    "isLastSet": false
  }
}
```

### **Test Case 2: Player Completed All Sets**
```graphql
query GetCurrentSet {
  getCurrentSet(guestId: "completed-guest-uuid") {
    puzzlesCompleted
    isUnlocked
    isCompleted
    isLastSet
  }
}
```

**Expected:**
```json
{
  "getCurrentSet": {
    "puzzlesCompleted": 10,
    "isUnlocked": true,
    "isCompleted": true,
    "isLastSet": true
  }
}
```

**Frontend Should:**
- Hide "Unlock Next Set" button
- Show "Congratulations! You've completed all sets!" message

### **Test Case 3: Player Completed Set 1, Not Unlocked Set 2**
```graphql
query GetCurrentSet {
  getCurrentSet(guestId: "completed-set1-guest-uuid") {
    puzzlesCompleted
    isUnlocked
    isCompleted
    isLastSet
  }
}
```

**Expected:**
```json
{
  "getCurrentSet": {
    "puzzlesCompleted": 10,
    "isUnlocked": true,
    "isCompleted": true,
    "isLastSet": false
  }
}
```

**Frontend Should:**
- Show "Unlock Next Set" button
- Unlock via `unlockPuzzlesAfterAd` mutation

## Related Changes

1. **Domain Model** (`internal/domain/puzzle.go`):
   - Added `IsLastSet` field to `GuestSetProgress`

2. **Service Layer** (`internal/service/puzzle_service.go`):
   - Fixed `GetCurrentSet()` to handle completed sets
   - Calculate `IsLastSet` based on set order

3. **GraphQL** (`internal/graphql/schema/schema.graphqls`):
   - Added `isLastSet` field to `GuestSetProgress` type

## Benefits

✅ **No more duplicate key errors**  
✅ **Proper handling of completed sets**  
✅ **Frontend can detect end game**  
✅ **Backward compatible** - doesn't break existing logic  
✅ **Clean separation** - new player vs completed player logic  

## Files Modified

- `internal/service/puzzle_service.go` - Fixed GetCurrentSet() logic
- `internal/domain/puzzle.go` - Added IsLastSet field
- `internal/graphql/schema/schema.graphqls` - Added isLastSet to schema
- `internal/graphql/resolver/helpers.go` - Added IsLastSet mapping