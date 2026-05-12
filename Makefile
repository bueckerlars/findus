.PHONY: tidy build test run dev frontend-dist docker-build db-reset lint format hooks-install hooks-uninstall hooks-run

# Pinned for reproducible `make dev` (hot reload via Air).
AIR := go run github.com/air-verse/air@v1.65.1

# Pre-commit / local checks (pin versions for reproducible installs via `go run`).
LEFTHOOK := go run github.com/evilmartians/lefthook@v1.13.6
GOLANGCI_LINT := go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.5

FINDUS_DATA_DIR ?= ./data

tidy:
	go mod tidy

db-reset:
	@rm -f $(FINDUS_DATA_DIR)/findus.db $(FINDUS_DATA_DIR)/findus.db-wal $(FINDUS_DATA_DIR)/findus.db-shm
	@rm -rf $(FINDUS_DATA_DIR)/images
	@echo "Database reset at $(FINDUS_DATA_DIR)"

frontend-dist:
	cd frontend && npm install && npm run build

build: frontend-dist
	CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/findus ./backend/app

test:
	go test ./...

lint:
	go vet ./...
	$(GOLANGCI_LINT) run ./...

# Format all Go sources in the module (matches pre-commit gofmt, but whole tree).
format:
	go fmt ./...

hooks-install:
	@command -v lefthook >/dev/null 2>&1 || { \
		echo "Installing lefthook to $$(go env GOPATH)/bin ..."; \
		GOBIN="$$(go env GOPATH)/bin" go install github.com/evilmartians/lefthook@v1.13.6; \
		echo "Ensure $$(go env GOPATH)/bin is on your PATH for git hooks to run."; \
	}
	$(LEFTHOOK) install

hooks-uninstall:
	$(LEFTHOOK) uninstall

# Run the same checks as git pre-commit (gofmt + lint) on all tracked Go files.
hooks-run:
	$(LEFTHOOK) run pre-commit --all-files

run:
	go run ./backend/app

# Dev server: debug logs + rebuild/restart on Go and embedded asset changes (see .air.toml).
dev:
	@mkdir -p tmp
	FINDUS_LOG_LEVEL=debug $(AIR) -c .air.toml

docker-build:
	docker build -t findus:dev .
