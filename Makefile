.PHONY: all build backend-build frontend-install frontend-build frontend-dev frontend-lint gen-db gen-mocks test test-storage test-business test-coverage migrate run-local run-docker install-deps install-deps-backend install-deps-frontend

# Default: build everything
all: build

# Dependencies
install-deps: install-deps-backend install-deps-frontend

install-deps-backend:
	go mod download
	go install github.com/vektra/mockery/v2@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

install-deps-frontend:
	cd frontend && bun install

build: frontend-build backend-build

# Backend (depends on frontend-build for embedded static files and gen-db for database code)
backend-build: frontend-build gen-db gen-mocks
	go build -o bin/ ./cmd/...

# Frontend
frontend-install:
	cd frontend && bun install

frontend-build: frontend-install
	cd frontend && bun run build

frontend-dev: frontend-install
	cd frontend && bun run dev

frontend-lint:
	cd frontend && bun run lint

# Database
gen-db:
	sqlc -f db/sqlc.yaml generate

# Migrations
migrate:
	@mkdir -p bin/migrations
	@cp db/migrations/*.sql bin/migrations/
	@echo "Running migrations..."
	@for file in bin/migrations/*.sql; do \
		echo "Applying $$file..."; \
		sqlite3 goodreads.db < $$file; \
	done
	@echo "Migrations complete"

# Testing
gen-mocks:
	mockery

test: gen-mocks
	go test -race ./...

coverage: gen-mocks
	go test $$(go list ./... | grep -v /mocks | grep -v /db | grep -v /cmd | grep -v /static) -coverprofile=coverage.out 
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run
run-local: build migrate
	./bin/goodreads

run-docker:
	docker-compose up --build
