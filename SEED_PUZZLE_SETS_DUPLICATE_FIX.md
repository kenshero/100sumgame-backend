# Fix for Duplicate Key Error in Seed Script

## Problem

When running `seed_puzzle_sets.go` with sets that already exist in database:
```bash
go run seed_puzzle_sets.go -sets=5

ERROR: duplicate key value violates unique constraint "puzzle_sets_set_order_key" (SQLSTATE 23505)
```

## Root Cause

The `puzzle_sets` table has `UNIQUE` constraint on `set_order`:

```sql
CREATE TABLE puzzle_sets (
    set_order INT NOT NULL UNIQUE,  -- ← UNIQUE constraint
    ...
);
```

**Scenario:**
1. Database already has puzzle_sets with set_order = 1, 2, 3, 4, 5
2. Script tries to insert set_order = 1, 2, 3, 4, 5
3. **Conflict!** set_order = 1 already exists → Duplicate key error

## Solution

Added check before INSERT to detect duplicates and skip them:

### **New Logic:**

```go
for setOrder := 1; setOrder <= *numSets; setOrder++ {
    // Check if set_order already exists
    var exists bool
    err := db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM puzzle_sets WHERE set_order = $1)", setOrder).Scan(&exists)
    
    if exists {
        fmt.Printf("⚠️  Set %d already exists, skipping...\n", setOrder)
        continue  // Skip if exists
    }
    
    // Insert only if doesn't exist
    _, err := db.Exec(ctx, "INSERT INTO puzzle_sets ...", ...)
}
```

### **Code Changes:**

**Before:**
```go
for setOrder := 1; setOrder <= *numSets; setOrder++ {
    // ❌ Directly insert - no check
    _, err := db.Exec(ctx, "INSERT INTO puzzle_sets (id, set_order, ...) VALUES ($1, $2, ...)", ...)
}
```

**After:**
```go
for setOrder := 1; setOrder <= *numSets; setOrder++ {
    // ✅ Check if exists first
    var exists bool
    err := db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM puzzle_sets WHERE set_order = $1)", setOrder).Scan(&exists)
    if err != nil {
        log.Fatalf("Failed to check if set %d exists: %v", setOrder, err)
    }

    if exists {
        fmt.Printf("⚠️  Set %d already exists, skipping...\n", setOrder)
        continue
    }

    // ✅ Only insert if doesn't exist
    _, err := db.Exec(ctx, "INSERT INTO puzzle_sets ...", ...)
}
```

## Example Output

### **Scenario: Database already has sets 1-3, running script for sets 1-5**

```bash
go run seed_puzzle_sets.go -sets=5

Connected to database successfully
🎯 Creating 5 puzzle sets with 10 puzzles each...
   Cell value range: 1-80
   Prefilled cells: 3-5

[1/5] Creating Set 1...
⚠️  Set 1 already exists, skipping...
[2/5] Creating Set 2...
⚠️  Set 2 already exists, skipping...
[3/5] Creating Set 3...
⚠️  Set 3 already exists, skipping...
[4/5] Creating Set 4...
✓ Set 4 completed with 10 puzzles
[5/5] Creating Set 5...
✓ Set 5 completed with 10 puzzles

✅ All puzzle sets and puzzles created successfully!
📊 Total: 5 sets × 10 puzzles = 50 puzzles
🎲 Cell values: 1-80
📍 Prefilled cells: 3-5
```

### **Scenario: Fresh database, creating 5 sets**

```bash
go run seed_puzzle_sets.go -sets=5

Connected to database successfully
🎯 Creating 5 puzzle sets with 10 puzzles each...
   Cell value range: 1-80
   Prefilled cells: 3-5

[1/5] Creating Set 1...
✓ Set 1 completed with 10 puzzles
[2/5] Creating Set 2...
✓ Set 2 completed with 10 puzzles
[3/5] Creating Set 3...
✓ Set 3 completed with 10 puzzles
[4/5] Creating Set 4...
✓ Set 4 completed with 10 puzzles
[5/5] Creating Set 5...
✓ Set 5 completed with 10 puzzles

✅ All puzzle sets and puzzles created successfully!
📊 Total: 5 sets × 10 puzzles = 50 puzzles
🎲 Cell values: 1-80
📍 Prefilled cells: 3-5
```

## Use Cases

### **1. Fill Gaps in Existing Data**
```bash
# Database has sets 1-10, want to add sets 11-50
go run seed_puzzle_sets.go -sets=50
# Result: Sets 1-10 skipped, sets 11-50 created
```

### **2. Run Script Multiple Times**
```bash
# First run - creates sets 1-50
go run seed_puzzle_sets.go -sets=50

# Second run - skips all (already exists)
go run seed_puzzle_sets.go -sets=50
# Result: All 50 sets skipped with warnings
```

### **3. Add More Sets After Testing**
```bash
# Testing with 5 sets initially
go run seed_puzzle_sets.go -sets=5

# Later, want to expand to 50 sets
go run seed_puzzle_sets.go -sets=50
# Result: Sets 1-5 skipped, sets 6-50 created
```

## Benefits

✅ **No data loss** - Doesn't delete existing puzzle sets  
✅ **Fill gaps** - Only creates missing sets  
✅ **Idempotent** - Can run script multiple times safely  
✅ **User-friendly** - Clear warnings when skipping  
✅ **No manual cleanup** - No need to delete before seeding  

## Implementation Details

### **SQL Query Used:**
```sql
SELECT EXISTS(SELECT 1 FROM puzzle_sets WHERE set_order = $1)
```

This query returns:
- `true` if set_order exists
- `false` if set_order doesn't exist

### **Skip Logic:**
```go
if exists {
    fmt.Printf("⚠️  Set %d already exists, skipping...\n", setOrder)
    continue  // Skip to next iteration
}
```

## Testing

### **Test Case 1: Fresh Database**
```bash
# Clean database first
docker-compose exec -T db psql -U postgres -d sum100game < scripts/reset_puzzle_sets.sql

# Run seed script
go run seed_puzzle_sets.go -sets=5

# Expected: All 5 sets created, no warnings
```

### **Test Case 2: Database with Existing Sets**
```bash
# Run seed script twice
go run seed_puzzle_sets.go -sets=5
go run seed_puzzle_sets.go -sets=5

# Expected: Second run skips all sets with warnings
```

### **Test Case 3: Fill Gaps**
```bash
# Create sets 1-3
go run seed_puzzle_sets.go -sets=3

# Run for 5 sets (should create 4-5)
go run seed_puzzle_sets.go -sets=5

# Expected: Sets 1-3 skipped, sets 4-5 created
```

## Related Files

- `scripts/seed_puzzle_sets.go` - Added duplicate check logic
- `scripts/SEED_PUZZLE_SETS_README.md` - Updated documentation
- `internal/database/migrations/009_add_puzzle_sets.sql` - Database schema

## Alternatives Considered

### **Alternative 1: Force Flag (--force)**
```bash
go run seed_puzzle_sets.go -sets=5 --force  # Delete and recreate
```
**Pros:** Clean slate, predictable  
**Cons:** Data loss, destructive  

### **Alternative 2: Continue from Max Order**
```bash
# Check max set_order, continue from there
```
**Pros:** No duplicates  
**Cons:** Can't fill gaps, only append  

### **Chosen: Check and Skip**
**Pros:** Safe, flexible, idempotent  
**Cons:** Slightly more code (worth it!)  

## Summary

This fix allows the seed script to handle existing data gracefully by:
1. Checking if each `set_order` exists before inserting
2. Skipping sets that already exist
3. Only creating new sets that are missing
4. Providing clear feedback about skipped sets

This makes the script safe to run multiple times and useful for filling gaps in existing data without losing any puzzles or progress.