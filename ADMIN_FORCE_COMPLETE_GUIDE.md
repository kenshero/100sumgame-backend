# Admin Force Complete Puzzles Guide

## Overview

The `adminForceCompletePuzzles` mutation allows administrators to force-complete all puzzles in the current unlocked set for a guest. This is useful for testing purposes without having to play through all puzzles manually.

## Features

- **Protected by Admin Auth**: Requires the admin secret token in the Authorization header (checked at resolver level)
- **Current Set Only**: Completes only the currently unlocked set (is_unlocked = true)
- **No Auto-Unlock**: Does NOT automatically unlock the next set (user must unlock manually via frontend)
- **Safe Completion**: Marks puzzles as completed with empty solved positions (admin-forced)
- **Resolver-Level Security**: Token validation happens inside GraphQL resolvers via `checkAdminToken()` function in `helpers.go`
- **HTTP Request Injection**: The HTTP request is injected into GraphQL context for admin token validation

## How to Use

### 1. Using GraphQL Playground

Open GraphQL Playground at `http://localhost:8080` and use the following mutation:

```graphql
mutation {
  adminForceCompletePuzzles(guestId: "YOUR_GUEST_UUID_HERE") {
    guestId
    setId
    puzzlesCompleted
    isUnlocked
    isCompleted
    unlockedAt
    completedAt
    currentStamina
    currentScore
  }
}
```

**Important**: Set the Authorization header:
```
Authorization: Bearer 9yoNTrMFJGR16VZ3EU75Mwji8QUSjvukKirrEpTqjpHEaMobLtJD40juYsAND7oJKfJJxSNZ6NqnIxI8v8s63r
```

### 2. Using cURL

```bash
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer 9yoNTrMFJGR16VZ3EU75Mwji8QUSjvukKirrEpTqjpHEaMobLtJD40juYsAND7oJKfJJxSNZ6NqnIxI8v8s63r" \
  -d '{
    "query": "mutation { adminForceCompletePuzzles(guestId: \"YOUR_GUEST_UUID_HERE\") { guestId setId puzzlesCompleted isCompleted } }"
  }'
```

### 3. Using JavaScript/Fetch

```javascript
const response = await fetch('http://localhost:8080/graphql', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer 9yoNTrMFJGR16VZ3EU75Mwji8QUSjvukKirrEpTqjpHEaMobLtJD40juYsAND7oJKfJJxSNZ6NqnIxI8v8s63r'
  },
  body: JSON.stringify({
    query: `mutation {
      adminForceCompletePuzzles(guestId: "YOUR_GUEST_UUID_HERE") {
        guestId
        setId
        puzzlesCompleted
        isCompleted
        currentStamina
        currentScore
      }
    }`
  })
});

const data = await response.json();
console.log(data);
```

## Example Response

```json
{
  "data": {
    "adminForceCompletePuzzles": {
      "guestId": "123e4567-e89b-12d3-a456-426614174000",
      "setId": "987e6543-e21b-43d9-a765-526614174999",
      "puzzlesCompleted": 10,
      "isUnlocked": true,
      "isCompleted": true,
      "unlockedAt": "2026-03-01T10:00:00Z",
      "completedAt": "2026-03-01T17:30:00Z",
      "currentStamina": 35,
      "currentScore": 500
    }
  }
}
```

## Error Responses

### 1. No Unlocked Set Found
```json
{
  "data": null,
  "errors": [
    {
      "message": "no unlocked set found for this guest",
      "path": ["adminForceCompletePuzzles"]
    }
  ]
}
```

### 2. Set Already Completed
```json
{
  "data": null,
  "errors": [
    {
      "message": "set is already completed",
      "path": ["adminForceCompletePuzzles"]
    }
  ]
}
```

### 3. Invalid Admin Token
```json
{
  "data": null,
  "errors": [
    {
      "message": "Invalid admin token",
      "path": ["adminForceCompletePuzzles"]
    }
  ]
}
```

## What Happens Internally

1. **Fetch Unlocked Set**: Gets the guest's current unlocked set (`is_unlocked = true`)
2. **Validate Status**: Checks that the set is not already completed
3. **Get All Puzzles**: Retrieves all 10 puzzles in the set
4. **Mark Puzzles Complete**: For each puzzle, adds an entry to `guest_puzzle_progress` with status = COMPLETED
5. **Update Set Progress**: Updates `guest_set_progress`:
   - `puzzles_completed` = 10
   - `is_completed` = true
   - `completed_at` = current timestamp
6. **Return Result**: Returns the updated set progress

## Security Notes

- The admin secret token is the same as used for `refreshGameConfig`
- This is a **development/testing-only** feature
- In production, consider implementing proper admin user authentication
- The endpoint is protected by `AdminAuthMiddleware` but should be used carefully

## Testing Workflow

1. Start the server: `go run cmd/server/main.go`
2. Get a guest UUID from your frontend or database
3. Call `adminForceCompletePuzzles` mutation with admin token
4. Verify the set is completed by calling `getCurrentSet` query
5. Use frontend to unlock the next set (admin won't do this automatically)

## Related Endpoints

- `getCurrentSet(guestId)`: Check guest's current set progress
- `getAvailablePuzzles(guestId)`: See puzzles available for the guest
- `unlockPuzzlesAfterAd(guestId)`: Unlock the next set (frontend usage)
- `refreshGameConfig`: Admin-only config refresh