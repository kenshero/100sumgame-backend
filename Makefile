.PHONY: help dev dev-up dev-logs down logs db-shell db-wait gql-gen migrate seed setup clean build-prod

# Colors
GREEN  := \033[0;32m
YELLOW := \033[1;33m
BLUE   := \033[0;34m
NC     := \033[0m

# Default target
help:
	@echo ""
	@echo "$(BLUE)========================================$(NC)"
	@echo "$(BLUE)  Sum-100 Puzzle Game - Commands$(NC)"
	@echo "$(BLUE)========================================$(NC)"
	@echo ""
	@echo "$(GREEN)Development:$(NC)"
	@echo "  make dev        - 🚀 Start all & show backend logs (DB background, Go foreground)"
	@echo "  make dev-up     - Start all services in background"
	@echo "  make dev-logs   - Show backend logs (attach)"
	@echo "  make down       - Stop all services"
	@echo "  make restart    - Restart backend container"
	@echo ""
	@echo "$(GREEN)Setup (first time):$(NC)"
	@echo "  make setup      - 📦 Run migrations + seed data"
	@echo ""
	@echo "$(GREEN)Database:$(NC)"
	@echo "  make db-shell   - Open PostgreSQL shell"
	@echo "  make migrate    - Run database migrations"
	@echo "  make seed       - Seed puzzle data"
	@echo ""
	@echo "$(GREEN)Code Generation:$(NC)"
	@echo "  make gql-gen    - Generate GraphQL code"
	@echo "  make gql-install- Install gqlgen tool"
	@echo ""
	@echo "$(GREEN)Production:$(NC)"
	@echo "  make build-prod - Build production Docker image"
	@echo ""
	@echo "$(GREEN)Cleanup:$(NC)"
	@echo "  make clean      - Remove all containers and volumes"
	@echo ""

# ============================================
# Development
# ============================================

# Start DB in background, then run backend with logs visible
dev: db-start db-wait
	@echo ""
	@echo "$(GREEN)========================================$(NC)"
	@echo "$(GREEN)  🚀 Starting Go Backend (logs below)$(NC)"
	@echo "$(GREEN)  Press Ctrl+C to stop$(NC)"
	@echo "$(GREEN)========================================$(NC)"
	@echo ""
	docker-compose up --build backend

# Start database only (background)
db-start:
	@echo "$(YELLOW)🐘 Starting PostgreSQL...$(NC)"
	docker-compose up -d db

# Wait for database to be ready
db-wait:
	@echo "$(YELLOW)⏳ Waiting for PostgreSQL...$(NC)"
	@for i in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20; do \
		if docker exec sum100-db pg_isready -U postgres > /dev/null 2>&1; then \
			echo "$(GREEN)✅ PostgreSQL is ready!$(NC)"; \
			break; \
		fi; \
		sleep 1; \
	done

# Start all in background
dev-up:
	@echo "$(YELLOW)🚀 Starting all services in background...$(NC)"
	docker-compose up -d --build
	@echo "$(GREEN)✅ Services started!$(NC)"
	@echo ""
	@echo "View logs: make dev-logs"
	@echo "Stop:      make down"

# Show backend logs (attach to running container)
dev-logs:
	docker-compose logs -f --tail=100 backend

logs:
	docker-compose logs -f --tail=100 backend

down:
	@echo "$(YELLOW)🛑 Stopping all services...$(NC)"
	docker-compose down
	@echo "$(GREEN)✅ All services stopped$(NC)"

restart:
	@echo "$(YELLOW)🔄 Restarting backend...$(NC)"
	docker-compose restart backend
	@echo "$(GREEN)✅ Backend restarted$(NC)"

# ============================================
# Setup (first time)
# ============================================

setup: db-start db-wait migrate seed
	@echo ""
	@echo "$(GREEN)✅ Setup completed!$(NC)"
	@echo "Run 'make dev' to start development"

# ============================================
# Database
# ============================================

db-shell:
	docker exec -it sum100-db psql -U postgres -d sum100game

migrate:
	@echo "$(YELLOW)📦 Running migrations...$(NC)"
	docker exec -i sum100-db psql -U postgres -d sum100game < internal/database/migrations/001_create_puzzle_pool.sql
	docker exec -i sum100-db psql -U postgres -d sum100game < internal/database/migrations/002_create_game_sessions.sql
	docker exec -i sum100-db psql -U postgres -d sum100game < internal/database/migrations/003_create_leaderboard.sql
	@echo "$(GREEN)✅ Migrations completed!$(NC)"

seed:
	@echo "$(YELLOW)🌱 Seeding puzzle data...$(NC)"
	docker exec -i sum100-db psql -U postgres -d sum100game < scripts/seed_puzzles.sql
	@echo "$(GREEN)✅ Seeding completed!$(NC)"

# ============================================
# Code Generation
# ============================================

gql-gen:
	@echo "$(YELLOW)⚡ Generating GraphQL code...$(NC)"
	go run github.com/99designs/gqlgen generate
	@echo "$(GREEN)✅ GraphQL code generated!$(NC)"

gql-install:
	@echo "$(YELLOW)📥 Installing gqlgen...$(NC)"
	go install github.com/99designs/gqlgen@latest
	@echo "$(GREEN)✅ gqlgen installed!$(NC)"

# ============================================
# Production
# ============================================

build-prod:
	@echo "$(YELLOW)🏗️  Building production image...$(NC)"
	docker-compose -f docker-compose.prod.yml build
	@echo "$(GREEN)✅ Production image built!$(NC)"

# ============================================
# Cleanup
# ============================================

clean:
	@echo "$(YELLOW)🧹 Cleaning up containers and volumes...$(NC)"
	docker-compose down -v --remove-orphans
	@echo "$(GREEN)✅ Cleaned up!$(NC)"
