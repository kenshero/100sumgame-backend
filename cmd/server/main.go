package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

	"github.com/kenshero/100sumgame/internal/config"
	"github.com/kenshero/100sumgame/internal/database"
	"github.com/kenshero/100sumgame/internal/graphql/generated"
	"github.com/kenshero/100sumgame/internal/graphql/resolver"
	"github.com/kenshero/100sumgame/internal/middleware"
	"github.com/kenshero/100sumgame/internal/repository"
	"github.com/kenshero/100sumgame/internal/service"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Get secret key from environment or use default
	secretKey := os.Getenv("SESSION_SECRET_KEY")
	if secretKey == "" {
		secretKey = "change-this-secret-key-in-production-use-at-least-32-characters-long"
		log.Println("WARNING: Using default session secret key. Set SESSION_SECRET_KEY in production!")
	}

	// Check if secure mode (HTTPS)
	isSecure := os.Getenv("ENVIRONMENT") == "production"

	// Connect to database
	db, err := database.NewPostgresConnection(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Connected to database successfully")

	// Initialize repositories
	puzzleRepo := repository.NewPuzzleRepository(db)
	gameRepo := repository.NewGameRepository(db)
	leaderboardRepo := repository.NewLeaderboardRepository(db)
	puzzleProgressRepo := repository.NewPuzzleProgressRepository(db)
	puzzleSetRepo := repository.NewPuzzleSetRepository(db)
	guestSetProgressRepo := repository.NewGuestSetProgressRepository(db)
	settingsRepo := repository.NewSettingsRepository(db)

	// Initialize configService first (needed by PuzzleService)
	configService := service.NewConfigService(settingsRepo)

	// Initialize services
	puzzleService := service.NewPuzzleService(puzzleRepo, puzzleProgressRepo, gameRepo, puzzleSetRepo, guestSetProgressRepo, configService)
	gameService := service.NewGameService(gameRepo, puzzleService, puzzleProgressRepo, guestSetProgressRepo, configService)
	leaderboardService := service.NewLeaderboardService(leaderboardRepo)
	adminService := service.NewAdminService(configService, puzzleSetRepo, puzzleRepo, puzzleProgressRepo, guestSetProgressRepo)
	// aiService := service.NewAIService(cfg.GeminiAPIKey) // Will be implemented later

	// Load game configuration from database on startup
	if err := configService.LoadSettings(); err != nil {
		log.Printf("WARNING: Failed to load game settings from database: %v", err)
	} else {
		log.Println("Game configuration loaded successfully")
	}

	// Initialize security middlewares
	sessionManager := middleware.NewSessionManager(secretKey)
	rateLimiter := middleware.NewRateLimiter(100, 1*time.Minute) // 100 requests per minute per IP

	// Initialize GraphQL resolver
	resolverDeps := &resolver.Resolver{
		GameService:        gameService,
		PuzzleService:      puzzleService,
		LeaderboardService: leaderboardService,
		AdminService:       adminService,
		// AIService:          aiService,
	}

	// Create GraphQL server
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{
		Resolvers: resolverDeps,
	}))

	// Custom middleware to inject HTTP request into GraphQL context
	// Note: Using the contextKey type from resolver package
	srvWithReqContext := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Import the requestContextKey from resolver package
		ctx := context.WithValue(r.Context(), resolver.RequestContextKey, r)
		srv.ServeHTTP(w, r.WithContext(ctx))
	})

	// Setup router
	r := chi.NewRouter()

	// Middleware
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Timeout(60 * time.Second))

	// Security middleware
	r.Use(middleware.RateLimitMiddleware(rateLimiter))
	r.Use(middleware.AuthMiddleware(sessionManager, isSecure))

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:3000", "*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Routes
	r.Handle("/", playground.Handler("Sum-100 Puzzle", "/graphql"))
	r.Handle("/graphql", srvWithReqContext)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("🚀 Server starting on http://localhost:%s", cfg.Port)
		log.Printf("📊 GraphQL Playground: http://localhost:%s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
