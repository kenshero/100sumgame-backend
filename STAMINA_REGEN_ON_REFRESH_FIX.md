# Stamina Regeneration on Page Refresh Fix

## Date
March 6, 2026

## Problem

When a user refreshed the web page, the stamina was NOT being recalculated. This caused issues when:
1. User's stamina ran out (0)
2. User waited for stamina to regenerate (e.g., 10 minutes)
3. User refreshed the page
4. Frontend still showed stamina = 0 (old value)
5. User couldn't submit answers because frontend thought they had 0 stamina

The stamina was only being regenerated when `SubmitAnswer` API was called, not on page refresh.

## Root Cause

The `GetCurrentSet()` GraphQL API was simply querying the database and returning the stored values without:
1. Checking how much time had passed since `last_stamina_update`
2. Calculating how much stamina should have regenerated
3. Updating the database with the new stamina value

## Solution

Modified the `GetCurrentSet()` function in `PuzzleService` to automatically regenerate stamina based on elapsed time before returning the progress data.

## Files Modified

### 1. `internal/service/puzzle_service.go`

#### Added `configService` dependency to `PuzzleService` struct:
```go
type PuzzleService struct {
    repo                 repository.PuzzleRepository
    PuzzleProgressRepo   repository.PuzzleProgressRepository
    gameRepo             repository.GameRepository
    PuzzleSetRepo        repository.PuzzleSetRepository
    GuestSetProgressRepo repository.GuestSetProgressRepository
    configService        *ConfigService  // ← NEW
}
```

#### Updated `NewPuzzleService()` constructor:
```go
func NewPuzzleService(
    repo repository.PuzzleRepository,
    puzzleProgressRepo repository.PuzzleProgressRepository,
    gameRepo repository.GameRepository,
    puzzleSetRepo repository.PuzzleSetRepository,
    guestSetProgressRepo repository.GuestSetProgressRepository,
    configService *ConfigService,  // ← NEW PARAMETER
) *PuzzleService {
    // ...
}
```

#### Updated `GetCurrentSet()` to regenerate stamina:
```go
func (s *PuzzleService) GetCurrentSet(ctx context.Context, guestID uuid.UUID) (*domain.GuestSetProgress, error) {
    // ... existing code to get/set progress ...
    
    // Try to get unlocked set
    setProgress, err := s.GuestSetProgressRepo.GetUnlockedSet(ctx, guestID)
    if err == nil && setProgress != nil {
        // ← NEW: Regenerate stamina based on time elapsed
        settings := s.configService.GetSettings()
        
        newStamina, newStaminaTime, err := s.GuestSetProgressRepo.RegenerateStamina(
            ctx,
            guestID,
            setProgress.SetID,
            setProgress.CurrentStamina,
            settings.StaminaMax,
            settings.StaminaRegenIntervalMinutes,
            settings.StaminaRegenAmount,
            setProgress.LastStaminaUpdate,
        )
        if err != nil {
            return nil, err
        }
        
        setProgress.CurrentStamina = newStamina
        setProgress.LastStaminaUpdate = newStaminaTime
        
        // ... rest of existing code ...
    }
    
    // ... rest of existing code ...
}
```

### 2. `cmd/server/main.go`

#### Reordered initialization to create `configService` before `puzzleService`:
```go
// Initialize repositories
puzzleRepo := repository.NewPuzzleRepository(db)
gameRepo := repository.NewGameRepository(db)
leaderboardRepo := repository.NewLeaderboardRepository(db)
puzzleProgressRepo := repository.NewPuzzleProgressRepository(db)
puzzleSetRepo := repository.NewPuzzleSetRepository(db)
guestSetProgressRepo := repository.NewGuestSetProgressRepository(db)
settingsRepo := repository.NewSettingsRepository(db)

// Initialize configService first (needed by PuzzleService) ← NEW
configService := service.NewConfigService(settingsRepo)

// Initialize services
puzzleService := service.NewPuzzleService(
    puzzleRepo,
    puzzleProgressRepo,
    gameRepo,
    puzzleSetRepo,
    guestSetProgressRepo,
    configService,  // ← NOW PASSES configService
)
gameService := service.NewGameService(gameRepo, puzzleService, puzzleProgressRepo, guestSetProgressRepo, configService)
leaderboardService := service.NewLeaderboardService(leaderboardRepo)
adminService := service.NewAdminService(configService, puzzleSetRepo, puzzleRepo, puzzleProgressRepo, guestSetProgressRepo)
```

## How It Works

### Before This Fix
1. User has stamina = 35, last_stamina_update = 10:00 AM
2. User plays until stamina = 0 at 10:35 AM
3. User waits 10 minutes (until 10:45 AM)
4. User refreshes page → `GetCurrentSet()` called
5. Backend returns stamina = 0 (old value from database)
6. Frontend shows stamina = 0
7. User tries to submit answer → blocked (insufficient stamina)

### After This Fix
1. User has stamina = 35, last_stamina_update = 10:00 AM
2. User plays until stamina = 0 at 10:35 AM
3. User waits 10 minutes (until 10:45 AM)
4. User refreshes page → `GetCurrentSet()` called
5. Backend calculates: 10 minutes elapsed ÷ 10 minute interval = 1 interval × 10 stamina = 10 stamina regen
6. Backend updates database: stamina = 10, last_stamina_update = 10:45 AM
7. Backend returns stamina = 10
8. Frontend shows stamina = 10 ✅
9. User can submit answers normally ✅

## Stamina Regeneration Logic

The `RegenerateStamina()` function (already existed in `guest_set_progress_repository.go`) works as follows:

1. Calculate elapsed minutes since `last_stamina_update`
2. Calculate how many regeneration intervals have passed
3. Multiply intervals by `stamina_regen_amount` to get stamina to regenerate
4. Cap at `stamina_max` to prevent exceeding maximum
5. If stamina changed, update database with new values

### Example Calculation
```
Current stamina: 35
Last update: 10:00 AM
Current time: 10:45 AM
Elapsed: 45 minutes

Config:
- Stamina regen interval: 10 minutes
- Stamina regen amount: 10 per interval
- Stamina max: 50

Calculation:
- Intervals passed: 45 ÷ 10 = 4
- Stamina to regen: 4 × 10 = 40
- New stamina: 35 + 40 = 75
- Capped at max: min(75, 50) = 50

Result: stamina = 50 (full), last_update = 10:45 AM
```

## Benefits

✅ **Automatic regeneration on page refresh** - Users always see current stamina
✅ **Better user experience** - No confusion about why stamina isn't increasing
✅ **Consistent behavior** - Stamina regens whether user plays or just refreshes
✅ **No changes to SubmitAnswer** - Existing logic for gameplay unchanged
✅ **Minimal code changes** - Only modified one function

## Testing

### Manual Testing Steps:
1. Check current stamina (e.g., 35)
2. Play until stamina runs out (0)
3. Wait for stamina to regenerate (e.g., 10-20 minutes)
4. Refresh the web page
5. Verify stamina has increased (should show regenerated value)

### Expected Result:
- After waiting 10 minutes: stamina should be at least 10
- After waiting 20 minutes: stamina should be at least 20
- After waiting enough time: stamina should reach max (50)

## Related Files

- `internal/repository/guest_set_progress_repository.go` - Contains `RegenerateStamina()` function
- `internal/service/game_service.go` - Already uses `RegenerateStamina()` in `SubmitAnswer()`
- `internal/repository/interfaces.go` - Defines `GuestSetProgressRepository` interface
- `internal/service/config_service.go` - Provides game settings (stamina config)

## Notes

- This fix only affects the `GetCurrentSet()` API
- The `GetAvailablePuzzles()` API already calls `GetCurrentSet()` internally, so it also benefits
- The stamina regeneration logic was already working correctly in `SubmitAnswer()` - this just extends it to page refreshes
- Database schema does not need to be changed
- No migration required