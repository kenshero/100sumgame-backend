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
	maxVal        = 80 // Maximum value per cell (1-80)
	minPrefill    = 3  // Minimum prefilled cells
	maxPrefill    = 5  // Maximum prefilled cells
	puzzlesPerSet = 10
)

// Generate a valid 5x5 puzzle grid that sums to 100
// Each cell can have value from 1 to maxVal (default 80)
func generatePuzzleGrid(maxVal int) [5][5]int {
	grid := [5][5]int{}

	// Fill first 24 cells with random values (1 to maxVal)
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			if i == 4 && j == 4 {
				continue // Skip last cell - will be calculated
			}
			grid[i][j] = rand.Intn(maxVal) + 1
		}
	}

	// Calculate sum of first 24 cells
	sum := 0
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			if i == 4 && j == 4 {
				continue
			}
			sum += grid[i][j]
		}
	}

	// Calculate last cell value to make sum = 100
	lastVal := 100 - sum

	// Ensure last cell is within valid range (1 to maxVal)
	if lastVal < 1 {
		// If too low, redistribute from other cells
		needed := 1 - lastVal
		for i := 0; i < 5 && needed > 0; i++ {
			for j := 0; j < 5 && needed > 0; j++ {
				if i == 4 && j == 4 {
					continue
				}
				if grid[i][j] > 1 {
					transfer := min(grid[i][j]-1, needed)
					grid[i][j] -= transfer
					needed -= transfer
				}
			}
		}
		lastVal = 1
	} else if lastVal > maxVal {
		// If too high, redistribute to other cells
		excess := lastVal - maxVal
		for i := 0; i < 5 && excess > 0; i++ {
			for j := 0; j < 5 && excess > 0; j++ {
				if i == 4 && j == 4 {
					continue
				}
				if grid[i][j] < maxVal {
					transfer := min(maxVal-grid[i][j], excess)
					grid[i][j] += transfer
					excess -= transfer
				}
			}
		}
		lastVal = maxVal
	}

	grid[4][4] = lastVal
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

	// Delete existing puzzle sets and puzzles (uncomment to clean slate)
	// _, err = db.Exec(ctx, "DELETE FROM guest_set_progress")
	// if err != nil {
	// 	log.Fatalf("Failed to delete guest set progress: %v", err)
	// }
	// _, err = db.Exec(ctx, "DELETE FROM puzzle_sets CASCADE")
	// if err != nil {
	// 	log.Fatalf("Failed to delete puzzle sets: %v", err)
	// }

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
