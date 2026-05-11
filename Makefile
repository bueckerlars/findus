.PHONY: tidy build test run dev css docker-build db-reset

# Pinned for reproducible `make dev` (hot reload via Air).
AIR := go run github.com/air-verse/air@v1.65.1

FINDUS_DATA_DIR ?= ./data

tidy:
	go mod tidy

db-reset:
	@rm -f $(FINDUS_DATA_DIR)/findus.db $(FINDUS_DATA_DIR)/findus.db-wal $(FINDUS_DATA_DIR)/findus.db-shm
	@rm -rf $(FINDUS_DATA_DIR)/images
	@echo "Database reset at $(FINDUS_DATA_DIR)"

css:
	npm install && npx tailwindcss -i ./web/static/css/input.css -o ./web/static/css/output.css --minify

build: css
	CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/findus ./cmd/findus

test:
	go test ./...

run:
	go run ./cmd/findus

# Dev server: debug logs + rebuild/restart on Go and embedded asset changes (see .air.toml).
dev:
	@mkdir -p tmp
	FINDUS_LOG_LEVEL=debug $(AIR) -c .air.toml

docker-build:
	docker build -t findus:latest .
