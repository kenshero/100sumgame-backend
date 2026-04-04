package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// Constants for puzzle generation
const (
	maxVal        = 40 // Maximum value per cell (1-40) - adjusted for better puzzle generation
	minPrefill    = 3  // Minimum prefilled cells
	maxPrefill    = 5  // Maximum prefilled cells
	puzzlesPerSet = 10
)

// getDBUrl ดึงค่า URL จาก Environment Variable
// โดยจะโหลดจากไฟล์ devops/.env.prod ก่อน (หากมี)
func getDBUrl() string {
	// หา path ของไฟล์ .env.prod จาก root ย้อนกลับไป 1 หรือหลายระดับ
	// เนื่องจากมักจะรันสคริปต์นี้จาก root หรือ folder scripts
	var envPath string
	if _, err := os.Stat("devops/.env.prod"); err == nil {
		envPath = "devops/.env.prod"
	} else if _, err := os.Stat("../devops/.env.prod"); err == nil {
		envPath = "../devops/.env.prod"
	}

	if envPath != "" {
		err := godotenv.Load(envPath)
		if err != nil {
			log.Printf("Warning: Failed to load %s: %v", envPath, err)
		} else {
			fmt.Printf("Loaded environment from %s\n", envPath)
		}
	}

	url := os.Getenv("DATABASE_URL")
	if url == "" {
		log.Fatal("Error: DATABASE_URL is not set in environment or .env.prod file")
	}
	return url
}

// Generate a valid 5x5 puzzle grid where each row and column sums to 100
func generatePuzzleGrid(maxVal int) [5][5]int {
	grid := [5][5]int{}
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			grid[i][j] = 20
		}
	}

	for iteration := 0; iteration < 200; iteration++ {
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

		maxTransfer := min4(
			maxVal-grid[r1][c1],
			grid[r1][c2]-1,
			maxVal-grid[r2][c2],
			grid[r2][c1]-1,
		)

		if maxTransfer < 1 {
			continue
		}

		transfer := rand.Intn(maxTransfer) + 1
		grid[r1][c1] += transfer
		grid[r1][c2] -= transfer
		grid[r2][c2] += transfer
		grid[r2][c1] -= transfer
	}

	if !verifyGrid(grid) {
		for i := 0; i < 5; i++ {
			for j := 0; j < 5; j++ {
				grid[i][j] = 20
			}
		}
	}

	return grid
}

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

func verifyGrid(grid [5][5]int) bool {
	for i := 0; i < 5; i++ {
		sum := 0
		for j := 0; j < 5; j++ {
			sum += grid[i][j]
		}
		if sum != 100 {
			return false
		}
	}

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

func getPrefilledPositions(minPrefill, maxPrefill int) []map[string]int {
	numPrefilled := rand.Intn(maxPrefill-minPrefill+1) + minPrefill
	positions := make([]map[string]int, 0)

	allPositions := make([][2]int, 0)
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			allPositions = append(allPositions, [2]int{i, j})
		}
	}

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
	numSets := flag.Int("sets", 50, "Number of puzzle sets to generate (default: 50)")
	help := flag.Bool("help", false, "Show help message")
	flag.Parse()

	if *help {
		fmt.Println("Puzzle Sets Seed Script (Production Ready - reads devops/.env.prod)")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  go run scripts/seed_puzzle_sets_prod.go [options]")
		fmt.Println()
		fmt.Println("Options:")
		fmt.Println("  -sets int")
		fmt.Println("        Number of puzzle sets to generate (default: 50)")
		fmt.Println("  -help")
		fmt.Println("        Show this help message")
		return
	}

	if *numSets < 1 {
		log.Fatal("Error: -sets must be at least 1")
	}
	if *numSets > 1000 {
		log.Fatal("Error: -sets cannot exceed 1000")
	}

	ctx := context.Background()

	// เรียกใช้ URL ที่โหลดมาจาก devops/.env.prod
	dbURL := getDBUrl()
	db, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// พิมพ์ออกมาให้ชัดเจน
	fmt.Println("Connected to database successfully")

	fmt.Printf("🎯 Creating %d puzzle sets with %d puzzles each...\n", *numSets, puzzlesPerSet)
	fmt.Printf("   Cell value range: 1-%d\n", maxVal)
	fmt.Printf("   Prefilled cells: %d-%d\n", minPrefill, maxPrefill)
	fmt.Println()

	for setOrder := 1; setOrder <= *numSets; setOrder++ {
		fmt.Printf("[%d/%d] Creating Set %d...\n", setOrder, *numSets, setOrder)

		var exists bool
		err := db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM puzzle_sets WHERE set_order = $1)", setOrder).Scan(&exists)
		if err != nil {
			log.Fatalf("Failed to check if set %d exists: %v", setOrder, err)
		}

		if exists {
			fmt.Printf("⚠️  Set %d already exists, skipping...\n", setOrder)
			continue
		}

		setID := uuid.New()

		_, err = db.Exec(ctx, `
			INSERT INTO puzzle_sets (id, set_order, difficulty)
			VALUES ($1, $2, $3)
		`, setID, setOrder, "MEDIUM")

		if err != nil {
			log.Fatalf("Failed to insert puzzle set %d: %v", setOrder, err)
		}

		for puzzleNum := 1; puzzleNum <= puzzlesPerSet; puzzleNum++ {
			grid := generatePuzzleGrid(maxVal)
			prefilledPositions := getPrefilledPositions(minPrefill, maxPrefill)

			gridJSON, err := json.Marshal(grid)
			if err != nil {
				log.Fatalf("Failed to marshal grid: %v", err)
			}

			positionsJSON, err := json.Marshal(prefilledPositions)
			if err != nil {
				log.Fatalf("Failed to marshal prefilled positions: %v", err)
			}

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
}
