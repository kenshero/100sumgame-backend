# Fix for unlockPuzzlesAfterAd "no rows in result set" Error

## Problem Summary

The `unlockPuzzlesAfterAd` API was throwing "failed to unlock next set: no rows in result set" error when:
- User completed Set 1 (is_completed=true)
- User tried to unlock Set 2
- Database had only 1 row in `guest_set_progress` table for that user

## Root Cause

The `UnlockNextSet()` function called `GetUnlockedSet()` which queries:
```sql
WHERE is_unlocked = true AND is_completed = false
```

When Set 1 was completed (is_completed=true), this query returned no rows, causing the error.

## Solution: Hybrid Approach

We implemented a Hybrid Approach that combines the benefits of lazy initialization with eager progress tracking:

### Changes Made

#### 1. Fixed `UnlockNextSet()` Function

Added fallback logic to handle completed sets:
- If no unlocked set found, search for the last completed set
- Use that set's order to find the next available set
- Create progress for the next set

```go
func (s *PuzzleService) UnlockNextSet(ctx context.Context, guestID uuid.UUID) (*domain.GuestSetProgress, error) {
    currentSetProgress, err := s.GuestSetProgressRepo.GetUnlockedSet(ctx, guestID)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            // No unlocked set - get last completed set
            allProgress, err := s.GuestSetProgressRepo.GetByGuest(ctx, guestID)
            if err != nil {
                return nil, err
            }
            if len(allProgress) == 0 {
                return nil, domain.ErrNoPuzzlesAvailable
            }
            currentSetProgress = allProgress[len(allProgress)-1]
        } else {
            return nil, err
        }
    }
    // ... rest of the logic
}
```

#### 2. Added `EnsureAllSetsProgress()` Function

New function that ensures every user has progress records for ALL sets in the system:
- Called automatically in `GetCurrentSet()`
- Creates progress for any missing sets
- New users: Set 1 is unlocked, others are locked
- Existing users: New sets start locked

```go
func (s *PuzzleService) EnsureAllSetsProgress(ctx context.Context, guestID uuid.UUID) error {
    allSets, _ := s.PuzzleSetRepo.GetAll(ctx)
    userProgress, _ := s.GuestSetProgressRepo.GetByGuest(ctx, guestID)
    
    userSetIDs := make(map[uuid.UUID]bool)
    for _, progress := range userProgress {
        userSetIDs[progress.SetID] = true
    }
    
    now := time.Now()
    isNewUser := len(userProgress) == 0
    
    for _, set := range allSets {
        if !userSetIDs[set.ID] {
            isFirstSet := (isNewUser && set.SetOrder == 1)
            newProgress := &domain.GuestSetProgress{
                GuestID:    guestID,
                SetID:      set.ID,
                IsUnlocked: isFirstSet, // Only unlock first set for new users
                // ... other fields
            }
            s.GuestSetProgressRepo.Create(ctx, newProgress)
        }
    }
    return nil
}
```

#### 3. Updated `GetCurrentSet()` Function

Modified to call `EnsureAllSetsProgress()` first, ensuring users always see all available sets:
```go
func (s *PuzzleService) GetCurrentSet(ctx context.Context, guestID uuid.UUID) (*domain.GuestSetProgress, error) {
    // Ensure progress exists for all sets (syncs new sets automatically)
    if err := s.EnsureAllSetsProgress(ctx, guestID); err != nil {
        return nil, err
    }
    // ... rest of the logic
}
```

## How It Works

### For New Users
1. User enters website → calls `getCurrentSet`
2. `EnsureAllSetsProgress()` creates progress for ALL sets (Set 1, 2, 3, ...)
   - Set 1: `is_unlocked=true`
   - Set 2+: `is_unlocked=false`
3. User sees puzzles from Set 1

### For Existing Users Completing Sets
1. User completes Set 1 → calls `unlockPuzzlesAfterAd`
2. `UnlockNextSet()` finds no unlocked set (Set 1 is completed)
3. Falls back to last completed set (Set 1)
4. Creates progress for Set 2 with `is_unlocked=true`
5. User can now play Set 2

### When Admin Adds New Sets
1. Admin seeds new sets (e.g., Set 4, 5) into `puzzle_sets`
2. User completes Set 3 → calls `getCurrentSet`
3. `EnsureAllSetsProgress()` detects missing Set 4, 5
4. Creates progress for Set 4, 5 with `is_unlocked=false`
5. User sees `isLastSet=false` (because Set 4, 5 exist)
6. When user unlocks Set 4, progress already exists

## Benefits

✅ **Fixes the bug** - `unlockPuzzlesAfterAd` now works correctly for completed sets  
✅ **Users see all sets** - Automatically syncs when new sets are added  
✅ **Accurate isLastSet** - Always reflects the actual number of sets in system  
✅ **Resource efficient** - Doesn't sync all users at once (Option C)  
✅ **Backward compatible** - Works with existing user data  
✅ **Admin friendly** - New sets are automatically available to users as they progress  

## Database Impact

### Before Fix
```
guest_set_progress:
- Set 1: is_unlocked=true, is_completed=true
```

### After First Call to getCurrentSet
```
guest_set_progress:
- Set 1: is_unlocked=true, is_completed=true
- Set 2: is_unlocked=false, is_completed=false  ← Auto-created
```

### After unlockPuzzlesAfterAd
```
guest_set_progress:
- Set 1: is_unlocked=true, is_completed=true
- Set 2: is_unlocked=true, is_completed=false  ← Unlocked
```

## Testing the Fix

### Test Case 1: New User
```graphql
# First time user calls getCurrentSet
query {
  getCurrentSet(guestId: "NEW_USER_ID") {
    setId
    isUnlocked
    isLastSet
  }
}
# Expected: Returns Set 1 (unlocked), isLastSet=false (if >1 set)
```

### Test Case 2: Completed User Unlocks Next Set
```graphql
# User completed Set 1, calls unlockPuzzlesAfterAd
mutation {
  unlockPuzzlesAfterAd(guestId: "COMPLETED_USER_ID") {
    setId
    isUnlocked
    isLastSet
  }
}
# Expected: Returns Set 2 (unlocked), no error
```

### Test Case 3: Admin Adds New Sets
```sql
-- Add Set 3 to puzzle_sets
INSERT INTO puzzle_sets (id, set_order, difficulty, created_at)
VALUES (uuid_generate_v4(), 3, 'medium', NOW());
```

```graphql
# Existing user completes Set 2, calls getCurrentSet
query {
  getCurrentSet(guestId: "EXISTING_USER_ID") {
    setId
    isLastSet
  }
}
# Expected: Returns Set 2, isLastSet=false (Set 3 exists)
```

## Migration Notes

- **No database migration needed** - Changes are purely application logic
- **Existing users** will be auto-synced on their next `getCurrentSet` call
- **New sets** added by admin will be visible to users as they progress

## Future Enhancements

If needed, could add:
- Background job to sync all users when new sets are added (instead of lazy sync)
- API endpoint for admin to force sync specific users
- Analytics on set completion rates

## Files Modified

- `internal/service/puzzle_service.go`
  - Modified `UnlockNextSet()` - Added fallback for completed sets
  - Added `EnsureAllSetsProgress()` - New function to sync all sets
  - Added `getUnlockedAt()` - Helper function
  - Modified `GetCurrentSet()` - Calls `EnsureAllSetsProgress()` first