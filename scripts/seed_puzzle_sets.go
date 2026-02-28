package main

import (
	"context"
	"encoding/json"
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

// Generate a valid 5x5 puzzle grid that sums to 100
func generatePuzzleGrid() [5][5]int {
	// Create grid
	grid := [5][5]int{}

	// Fill with random values
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			grid[i][j] = rand.Intn(99) + 1
		}
	}

	// Calculate current sum
	sum := 0
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			sum += grid[i][j]
		}
	}

	// Adjust the last cell to make sum = 100
	grid[4][4] = 100 - (sum - grid[4][4])

	// Ensure the adjusted value is valid (1-99)
	for grid[4][4] < 1 {
		// If too low, redistribute from other cells
		grid[4][4] += 10
		for i := 0; i < 4; i++ {
			for j := 0; j < 5; j++ {
				if grid[i][j] > 10 {
					grid[i][j] -= 10
					break
				}
			}
		}
	}

	for grid[4][4] > 99 {
		// If too high, redistribute to other cells
		grid[4][4] -= 10
		for i := 0; i < 4; i++ {
			for j := 0; j < 5; j++ {
				if grid[i][j] < 90 {
					grid[i][j] += 10
					break
				}
			}
		}
	}

	return grid
}

// Get random prefilled positions (3-5 cells)
func getPrefilledPositions() []map[string]int {
	// Get 3-5 random cells to prefill
	numPrefilled := rand.Intn(3) + 3 // 3, 4, or 5 cells
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

func main() {
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

	// Create puzzle sets (3 sets for now)
	numSets := 3
	puzzlesPerSet := 10

	fmt.Printf("Creating %d puzzle sets with %d puzzles each...\n", numSets, puzzlesPerSet)

	for setOrder := 1; setOrder <= numSets; setOrder++ {
		// Insert puzzle set
		setID := uuid.New()
		setName := fmt.Sprintf("Set %d", setOrder)

		_, err := db.Exec(ctx, `
			INSERT INTO puzzle_sets (id, set_order, name, puzzles_count)
			VALUES ($1, $2, $3, $4)
		`, setID, setOrder, setName, puzzlesPerSet)

		if err != nil {
			log.Fatalf("Failed to insert puzzle set %d: %v", setOrder, err)
		}

		fmt.Printf("Created puzzle set: %s (ID: %s)\n", setName, setID)

		// Generate and insert puzzles for this set
		for puzzleNum := 1; puzzleNum <= puzzlesPerSet; puzzleNum++ {
			// Generate puzzle
			grid := generatePuzzleGrid()
			prefilledPositions := getPrefilledPositions()

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

			// Calculate difficulty based on prefilled positions
			difficulty := "MEDIUM"
			if len(prefilledPositions) <= 3 {
				difficulty = "HARD"
			} else if len(prefilledPositions) >= 5 {
				difficulty = "EASY"
			}

			// Insert puzzle
			puzzleID := uuid.New()
			_, err = db.Exec(ctx, `
				INSERT INTO puzzle_pool (id, set_id, grid_solution, prefilled_positions, difficulty, created_at)
				VALUES ($1, $2, $3, $4, $5, $6)
			`, puzzleID, setID, gridJSON, positionsJSON, difficulty, time.Now())

			if err != nil {
				log.Fatalf("Failed to insert puzzle %d for set %d: %v", puzzleNum, setOrder, err)
			}

			fmt.Printf("  - Puzzle %d created (ID: %s, Difficulty: %s)\n", puzzleNum, puzzleID, difficulty)
		}

		fmt.Printf("Completed Set %d with %d puzzles\n\n", setOrder, puzzlesPerSet)
	}

	fmt.Println("✅ All puzzle sets and puzzles created successfully!")
	fmt.Printf("Total: %d sets with %d puzzles each (%d total puzzles)\n", numSets, puzzlesPerSet, numSets*puzzlesPerSet)
}
