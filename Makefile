# Every target that talks to the database reads DB_URL from the environment,
# falling back to the docker-compose default.
GATOR_DB_PORT ?= 5432
DB_URL ?= postgres://postgres:postgres@localhost:$(GATOR_DB_PORT)/gator?sslmode=disable
export GATOR_DB_PORT

GOOSE := goose -dir sql/schema postgres "$(DB_URL)"

.PHONY: build test test-integration lint fmt sqlc db-up db-down migrate-up migrate-down migrate-status psql demo up down

build: ## Build bin/gator
	go build -o bin/gator ./cmd/gator

test: ## Unit tests (no Docker needed)
	go test -race ./...

test-integration: ## Unit + Postgres integration tests (needs Docker)
	go test -race -tags=integration ./...

lint: ## gofmt check, go vet, and golangci-lint if installed
	@test -z "$$(gofmt -l .)" || { echo "gofmt needed:"; gofmt -l .; exit 1; }
	go vet ./...
	go vet -tags=integration ./...
	@if command -v golangci-lint >/dev/null; then golangci-lint run; else echo "golangci-lint not installed; skipping"; fi

fmt: ## Format all Go files
	gofmt -w .

sqlc: ## Regenerate internal/database from sql/
	sqlc generate

db-up: ## Start Postgres in Docker and wait until it is healthy
	docker compose up -d --wait db

db-down: ## Stop Postgres (data is kept in the gator-db volume)
	docker compose down

migrate-up: ## Apply all migrations to DB_URL
	$(GOOSE) up

migrate-down: ## Roll back the latest migration on DB_URL
	$(GOOSE) down

migrate-status: ## Show applied migrations on DB_URL
	$(GOOSE) status

psql: ## Open a psql shell on DB_URL
	psql "$(DB_URL)"

DEMO_DB_PORT ?= 55434
DEMO_DB_URL := postgres://postgres:postgres@localhost:$(DEMO_DB_PORT)/gator?sslmode=disable

demo: ## Render docs/demo.gif with VHS against a throwaway database
	CGO_ENABLED=0 go build -o bin/gator ./cmd/gator
	GATOR_DB_PORT=$(DEMO_DB_PORT) docker compose -p gator-demo up -d --wait db
	goose -dir sql/schema postgres "$(DEMO_DB_URL)" up
	DB_URL="$(DEMO_DB_URL)" vhs demo.tape; status=$$?; \
		GATOR_DB_PORT=$(DEMO_DB_PORT) docker compose -p gator-demo down -v; exit $$status

# Old target names, kept so existing habits still work.
up: migrate-up
down: migrate-down
