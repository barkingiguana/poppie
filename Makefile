# Poppie — run `make help` to see all targets
.DEFAULT_GOAL := help
SHELL := /bin/bash
GREEN  := \033[0;32m
YELLOW := \033[0;33m
CYAN   := \033[0;36m
RESET  := \033[0m

BINARY := poppie
BUILD_DIR := bin
GO_MODULE := github.com/BarkingIguana/poppie
VERSION := $(shell cat VERSION)

# ─── Quick Start ──────────────────────────────────────────────────────────────

.PHONY: install
install: ## Set up local development environment
	go mod download
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/cucumber/godog/cmd/godog@latest
	@echo -e "$(GREEN)Done! Run 'make check' to verify.$(RESET)"

.PHONY: check
check: lint test bdd sdk-go-test sdk-python-test ## Run all quality checks
	@echo -e "$(GREEN)All checks passed!$(RESET)"

# ─── Quality ──────────────────────────────────────────────────────────────────

.PHONY: lint
lint: ## Run golangci-lint
	golangci-lint run ./...

.PHONY: format
format: ## Auto-format code
	gofmt -w .
	goimports -w .

# ─── Protobuf ────────────────────────────────────────────────────────────────

.PHONY: proto
proto: sdk-python-generate ## Regenerate protobuf/gRPC code (Go + Python)
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/**/*.proto
	@echo -e "$(GREEN)Protobuf code regenerated.$(RESET)"

.PHONY: sdk-python-generate
sdk-python-generate: ## Regenerate Python protobuf stubs
	python3 -m grpc_tools.protoc \
		-Iproto \
		--python_out=sdk/python/src/poppie/_generated \
		--grpc_python_out=sdk/python/src/poppie/_generated \
		poppie/poppie.proto
	@# Fix import path for package-relative imports
	sed -i '' 's/^from poppie import/from . import/' \
		sdk/python/src/poppie/_generated/poppie/poppie_pb2_grpc.py
	@echo -e "$(GREEN)Python stubs regenerated.$(RESET)"

# ─── Testing ──────────────────────────────────────────────────────────────────

.PHONY: test
test: ## Run Go unit tests (excludes BDD)
	go test -race -count=1 $(shell go list ./... | grep -v /features/)

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -func=coverage.out
	@echo -e "$(CYAN)HTML report: go tool cover -html=coverage.out$(RESET)"

.PHONY: bdd
bdd: ## Run BDD/Cucumber feature specs
	go test -race -count=1 -v ./features/steps/

# ─── SDKs ─────────────────────────────────────────────────────────────────────

.PHONY: sdk-go-test
sdk-go-test: ## Run Go SDK tests
	go test -race -count=1 ./sdk/go/

.PHONY: sdk-python-test
sdk-python-test: ## Run Python SDK tests
	cd sdk/python && python3 -m pytest tests/ -v

# ─── Development ──────────────────────────────────────────────────────────────

.PHONY: run
run: build ## Run poppie server locally
	./$(BUILD_DIR)/$(BINARY) server start

.PHONY: build
build: ## Build poppie binary
	go build -ldflags "-X $(GO_MODULE)/internal/server.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY) ./cmd/poppie/
	@echo -e "$(GREEN)Built: $(BUILD_DIR)/$(BINARY)$(RESET)"

# ─── Deployment ───────────────────────────────────────────────────────────────

.PHONY: release
release: ## Build release binaries for all platforms
	GOOS=darwin GOARCH=arm64 go build -ldflags "-X $(GO_MODULE)/internal/server.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY)-darwin-arm64 ./cmd/poppie/
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X $(GO_MODULE)/internal/server.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY)-darwin-amd64 ./cmd/poppie/
	GOOS=linux GOARCH=amd64 go build -ldflags "-X $(GO_MODULE)/internal/server.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY)-linux-amd64 ./cmd/poppie/
	GOOS=linux GOARCH=arm64 go build -ldflags "-X $(GO_MODULE)/internal/server.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY)-linux-arm64 ./cmd/poppie/
	@echo -e "$(GREEN)Release binaries built in $(BUILD_DIR)/$(RESET)"

# ─── Utilities ────────────────────────────────────────────────────────────────

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf $(BUILD_DIR)/ coverage.out
	@echo -e "$(GREEN)Cleaned.$(RESET)"

# ─── Help ─────────────────────────────────────────────────────────────────────

.PHONY: help
help: ## Show this help message
	@echo -e "$(CYAN)Available targets:$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*## "}; {printf "  $(GREEN)%-20s$(RESET) %s\n", $$1, $$2}'
