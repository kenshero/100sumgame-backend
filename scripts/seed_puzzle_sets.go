package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	dbURL = "postgres://postgres:postgres@localhost:5432/sum100game?sslmode=disable"
)

// Constants for puzzle generation
const (
	maxVal        = 40 // Maximum value per cell (1-40) - adjusted for better puzzle generation
	minPrefill    = 3  // Minimum prefilled cells
	maxPrefill    = 5  // Maximum prefilled cells
	puzzlesPerSet = 10
)

// Generate a valid 5x5 puzzle grid where each row and column sums to 100
// Each cell can have value from 1 to maxVal (default 40)
func generatePuzzleGrid(maxVal int) [5][5]int {
	// Start with a valid grid (all 20s)
	grid := [5][5]int{}
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			grid[i][j] = 20
		}
	}

	// Add random variations using 4-cell cycles
	// A cycle (r1,c1)->(r1,c2)->(r2,c2)->(r2,c1)->back maintains all constraints
	for iteration := 0; iteration < 200; iteration++ {
		// Pick two different rows and two different columns
		r1 := rand.Intn(5)
		r2 := rand.Intn(5)
		for r1 == r2 {
			r2 = rand.Intn(5)
		}

		c1 := rand.Intn(5)
		c2 := rand.Intn(5)
		for c1 == c2 {
			c2 = rand.Intn(5)
		}

		// Calculate maximum transfer amount that keeps all 4 cells in valid range
		// Transfer: +X to (r1,c1), -X to (r1,c2), +X to (r2,c2), -X to (r2,c1)
		maxTransfer := min4(
			maxVal-grid[r1][c1], // Can't exceed max at (r1,c1)
			grid[r1][c2]-1,      // Can't go below 1 at (r1,c2)
			maxVal-grid[r2][c2], // Can't exceed max at (r2,c2)
			grid[r2][c1]-1,      // Can't go below 1 at (r2,c1)
		)

		if maxTransfer < 1 {
			continue
		}

		// Apply transfer
		transfer := rand.Intn(maxTransfer) + 1
		grid[r1][c1] += transfer
		grid[r1][c2] -= transfer
		grid[r2][c2] += transfer
		grid[r2][c1] -= transfer
	}

	// Verify grid is valid (it should be)
	if !verifyGrid(grid) {
		// Fallback to all 20s if something went wrong
		for i := 0; i < 5; i++ {
			for j := 0; j < 5; j++ {
				grid[i][j] = 20
			}
		}
	}

	return grid
}

// Helper function for recursive generation with backtracking
func generatePuzzleHelper(grid *[5][5]int, rowTarget, colTarget [5]int, row, col, maxVal int) bool {
	// Base case: all cells filled
	if row == 5 {
		return true
	}

	// Calculate next cell position
	nextRow := (col + 1) / 5
	nextCol := (col + 1) % 5

	// Calculate min and max possible value for this cell
	// We need: rowTarget[row] - remainingInRow >= 1
	// And: colTarget[col] - remainingInCol >= 1

	// Calculate remaining needed for this row
	remainingInRow := 0
	for j := col; j < 5; j++ {
		remainingInRow += grid[row][j]
	}

	// Calculate remaining needed for this column
	remainingInCol := 0
	for i := row; i < 5; i++ {
		remainingInCol += grid[i][col]
	}

	// Min value: ensure we can reach 1 for both row and col
	minOption1 := rowTarget[row] - remainingInRow - maxVal*(4-col)
	minOption2 := colTarget[col] - remainingInCol - maxVal*(4-row)
	minVal := max(max(1, minOption1), minOption2)

	// Max value: ensure we don't exceed targets
	maxOption1 := rowTarget[row] - remainingInRow - (4 - col)
	maxOption2 := colTarget[col] - remainingInCol - (4 - row)
	maxValForCell := min(min(maxVal, maxOption1), maxOption2)

	if minVal > maxValForCell {
		return false // No valid value possible
	}

	// Try values in random order for variety
	valueRange := make([]int, maxValForCell-minVal+1)
	for i := range valueRange {
		valueRange[i] = minVal + i
	}

	rand.Shuffle(len(valueRange), func(i, j int) {
		valueRange[i], valueRange[j] = valueRange[j], valueRange[i]
	})

	// Try each possible value
	for _, val := range valueRange {
		grid[row][col] = val

		// Recursively fill remaining cells
		if generatePuzzleHelper(grid, rowTarget, colTarget, nextRow, nextCol, maxVal) {
			return true
		}

		// Backtrack
		grid[row][col] = 0
	}

	return false
}

// Shuffle grid values while maintaining row/column sums
func shuffleGrid(grid *[5][5]int) {
	// Collect all values
	values := make([]int, 25)
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			values[i*5+j] = grid[i][j]
		}
	}

	// Shuffle values
	rand.Shuffle(len(values), func(i, j int) {
		values[i], values[j] = values[j], values[i]
	})

	// Put values back (this will break constraints, but we'll fix it)
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			grid[i][j] = values[i*5+j]
		}
	}

	// Rebalance to fix constraints
	rebalanceGrid(grid)
}

// Rebalance grid to satisfy row/column constraints
func rebalanceGrid(grid *[5][5]int) {
	// Simple approach: adjust to make rows and columns sum to 100
	// This is a simplification - in practice you might need more sophisticated algorithms

	// For now, just use the original valid grid
	// The shuffle above breaks constraints, so we need a better approach

	// Better: swap values between cells in the same row/column
	for swap := 0; swap < 100; swap++ {
		i1, j1 := rand.Intn(5), rand.Intn(5)
		i2, j2 := rand.Intn(5), rand.Intn(5)

		// Swap
		grid[i1][j1], grid[i2][j2] = grid[i2][j2], grid[i1][j1]

		// Check if still valid
		if verifyGrid(*grid) {
			return
		}

		// Swap back
		grid[i1][j1], grid[i2][j2] = grid[i2][j2], grid[i1][j1]
	}
}

// Helper for max
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Helper for min of 4 values
func min4(a, b, c, d int) int {
	minVal := a
	if b < minVal {
		minVal = b
	}
	if c < minVal {
		minVal = c
	}
	if d < minVal {
		minVal = d
	}
	return minVal
}

// Verify that all rows and columns sum to 100
func verifyGrid(grid [5][5]int) bool {
	// Check all rows
	for i := 0; i < 5; i++ {
		sum := 0
		for j := 0; j < 5; j++ {
			sum += grid[i][j]
		}
		if sum != 100 {
			return false
		}
	}

	// Check all columns
	for j := 0; j < 5; j++ {
		sum := 0
		for i := 0; i < 5; i++ {
			sum += grid[i][j]
		}
		if sum != 100 {
			return false
		}
	}

	return true
}

// Fallback method with retry for edge cases
func generatePuzzleGridRetry(maxVal int) [5][5]int {
	for retry := 0; retry < 1000; retry++ {
		grid := [5][5]int{}
		sum := 0

		// Fill first 24 cells
		for i := 0; i < 5; i++ {
			for j := 0; j < 5; j++ {
				if i == 4 && j == 4 {
					continue
				}
				// Use smaller max for better success rate
				cellMax := min(maxVal, 20)
				grid[i][j] = rand.Intn(cellMax) + 1
				sum += grid[i][j]
			}
		}

		lastVal := 100 - sum
		if lastVal >= 1 && lastVal <= maxVal {
			grid[4][4] = lastVal
			totalSum := 0
			for i := 0; i < 5; i++ {
				for j := 0; j < 5; j++ {
					totalSum += grid[i][j]
				}
			}
			if totalSum == 100 {
				return grid
			}
		}
	}

	// Last resort: generate simple grid with all 4s (25 * 4 = 100)
	grid := [5][5]int{}
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			grid[i][j] = 4
		}
	}
	return grid
}

// Get random prefilled positions (minPrefill to maxPrefill cells)
func getPrefilledPositions(minPrefill, maxPrefill int) []map[string]int {
	// Get random number of cells to prefill
	numPrefilled := rand.Intn(maxPrefill-minPrefill+1) + minPrefill
	positions := make([]map[string]int, 0)

	// Get all possible positions (0-4, 0-4)
	allPositions := make([][2]int, 0)
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			allPositions = append(allPositions, [2]int{i, j})
		}
	}

	// Shuffle and pick first numPrefilled
	rand.Shuffle(len(allPositions), func(i, j int) {
		allPositions[i], allPositions[j] = allPositions[j], allPositions[i]
	})

	for i := 0; i < numPrefilled; i++ {
		pos := map[string]int{
			"row": allPositions[i][0],
			"col": allPositions[i][1],
		}
		positions = append(positions, pos)
	}

	return positions
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	// Parse command-line flags
	numSets := flag.Int("sets", 50, "Number of puzzle sets to generate (default: 50)")
	help := flag.Bool("help", false, "Show help message")
	flag.Parse()

	// Show help
	if *help {
		fmt.Println("Puzzle Sets Seed Script")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  go run seed_puzzle_sets.go [options]")
		fmt.Println()
		fmt.Println("Options:")
		fmt.Println("  -sets int")
		fmt.Println("        Number of puzzle sets to generate (default: 50)")
		fmt.Println("  -help")
		fmt.Println("        Show this help message")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  go run seed_puzzle_sets.go           # Generate 50 sets (default)")
		fmt.Println("  go run seed_puzzle_sets.go -sets=10  # Generate 10 sets")
		fmt.Println("  go run seed_puzzle_sets.go -sets=1   # Generate 1 set")
		fmt.Println("  go run seed_puzzle_sets.go -sets=100 # Generate 100 sets")
		return
	}

	// Validate numSets
	if *numSets < 1 {
		log.Fatal("Error: -sets must be at least 1")
	}
	if *numSets > 1000 {
		log.Fatal("Error: -sets cannot exceed 1000")
	}

	// Connect to database
	ctx := context.Background()

	db, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("Connected to database successfully")

	// Delete existing puzzle sets and puzzles (clean slate)
	_, err = db.Exec(ctx, "DELETE FROM guest_set_progress")
	if err != nil {
		log.Fatalf("Failed to delete guest set progress: %v", err)
	}
	_, err = db.Exec(ctx, "DELETE FROM game_sessions")
	if err != nil {
		log.Fatalf("Failed to delete game sessions: %v", err)
	}
	_, err = db.Exec(ctx, "DELETE FROM puzzle_sets CASCADE")
	if err != nil {
		log.Fatalf("Failed to delete puzzle sets: %v", err)
	}
	fmt.Println("✓ Cleared existing puzzle sets, progress, and game sessions")

	fmt.Printf("🎯 Creating %d puzzle sets with %d puzzles each...\n", *numSets, puzzlesPerSet)
	fmt.Printf("   Cell value range: 1-%d\n", maxVal)
	fmt.Printf("   Prefilled cells: %d-%d\n", minPrefill, maxPrefill)
	fmt.Println()

	for setOrder := 1; setOrder <= *numSets; setOrder++ {
		// Progress indicator
		fmt.Printf("[%d/%d] Creating Set %d...\n", setOrder, *numSets, setOrder)

		// Check if set_order already exists
		var exists bool
		err := db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM puzzle_sets WHERE set_order = $1)", setOrder).Scan(&exists)
		if err != nil {
			log.Fatalf("Failed to check if set %d exists: %v", setOrder, err)
		}

		if exists {
			fmt.Printf("⚠️  Set %d already exists, skipping...\n", setOrder)
			continue
		}

		// Insert puzzle set
		setID := uuid.New()

		_, err = db.Exec(ctx, `
			INSERT INTO puzzle_sets (id, set_order, difficulty)
			VALUES ($1, $2, $3)
		`, setID, setOrder, "MEDIUM")

		if err != nil {
			log.Fatalf("Failed to insert puzzle set %d: %v", setOrder, err)
		}

		// Generate and insert puzzles for this set
		for puzzleNum := 1; puzzleNum <= puzzlesPerSet; puzzleNum++ {
			// Generate puzzle with configured parameters
			grid := generatePuzzleGrid(maxVal)
			prefilledPositions := getPrefilledPositions(minPrefill, maxPrefill)

			// Convert grid to JSON
			gridJSON, err := json.Marshal(grid)
			if err != nil {
				log.Fatalf("Failed to marshal grid: %v", err)
			}

			// Convert prefilled positions to JSON
			positionsJSON, err := json.Marshal(prefilledPositions)
			if err != nil {
				log.Fatalf("Failed to marshal prefilled positions: %v", err)
			}

			// Insert puzzle
			puzzleID := uuid.New()
			_, err = db.Exec(ctx, `
				INSERT INTO puzzle_pool (id, set_id, grid_solution, prefilled_positions, difficulty, created_at)
				VALUES ($1, $2, $3, $4, $5, $6)
			`, puzzleID, setID, gridJSON, positionsJSON, "MEDIUM", time.Now())

			if err != nil {
				log.Fatalf("Failed to insert puzzle %d for set %d: %v", puzzleNum, setOrder, err)
			}
		}

		fmt.Printf("✓ Set %d completed with %d puzzles\n", setOrder, puzzlesPerSet)
	}

	fmt.Println()
	fmt.Println("✅ All puzzle sets and puzzles created successfully!")
	fmt.Printf("📊 Total: %d sets × %d puzzles = %d puzzles\n", *numSets, puzzlesPerSet, *numSets*puzzlesPerSet)
	fmt.Printf("🎲 Cell values: 1-%d\n", maxVal)
	fmt.Printf("📍 Prefilled cells: %d-%d\n", minPrefill, maxPrefill)
}
