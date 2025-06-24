DB_HOST ?= localhost
DB_PORT ?= 5555
DB_NAME ?= ai_poc_db
DB_USER ?= user
DB_PASSWORD ?= password
MIGRATIONS_DIR = ./internal/db/migrations

# Build connection string
DB_URL = postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

# Migrate commands
.PHONY: migrate-create migrate-up migrate-down migrate-version migrate-force

build:
	@echo "Building the Go application..."
	GOOS=linux GOARCH=amd64 go build -o main ./cmd
# Create a new migration
# Usage: make migrate-create name=create_users_table
migrate-create:
	@echo "Creating migration $(name)..."
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

# Run all pending migrations
migrate-up:
	@echo "Running all migrations..."
	migrate -database "$(DB_URL)" -path $(MIGRATIONS_DIR) up

# Run specific number of migrations
# Usage: make migrate-up-step step=1
migrate-up-step:
	@echo "Running $(step) migrations..."
	migrate -database "$(DB_URL)" -path $(MIGRATIONS_DIR) up $(step)

# Rollback all migrations
migrate-down:
	@echo "Rolling back all migrations..."
	migrate -database "$(DB_URL)" -path $(MIGRATIONS_DIR) down

# Rollback specific number of migrations
# Usage: make migrate-down-step step=1
migrate-down-step:
	@echo "Rolling back $(step) migrations..."
	migrate -database "$(DB_URL)" -path $(MIGRATIONS_DIR) down $(step)

# Check current migration version
migrate-version:
	@echo "Checking current migration version..."
	migrate -database "$(DB_URL)" -path $(MIGRATIONS_DIR) version

# Force migration to specific version
# Usage: make migrate-force version=1
migrate-force:
	@echo "Forcing migration to version $(version)..."
	migrate -database "$(DB_URL)" -path $(MIGRATIONS_DIR) force $(version)

# Create database if not exists
postgres-create-db:
	@echo "Creating database $(DB_NAME) if not exists..."
	PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d postgres -c "CREATE DATABASE $(DB_NAME) WITH ENCODING='UTF8' LC_COLLATE='en_US.UTF-8' LC_CTYPE='en_US.UTF-8' TEMPLATE=template0;" || true

# Seed sample data
# Note: This requires having a seed.sql file in your migrations directory
postgres-seed-data:
	@echo "Seeding sample data..."
	PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) -f $(MIG
