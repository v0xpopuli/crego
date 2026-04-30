.PHONY: build clean tests

APP_NAME ?= crego
BUILD_DIR ?= ./build/app
MAIN_PACKAGE ?= ./cmd/crego

VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILT ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS ?= -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.built=$(BUILT)

build:
	@$(MAKE) clean
	@echo "Building $(APP_NAME) binary..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PACKAGE)
	@echo "Binary built successfully: $(BUILD_DIR)/$(APP_NAME)"

clean:
	@echo "Cleaning up..."
	@rm -rf ./build
	@echo "Clean up completed!"

tests:
	@echo "Running tests..."
	@go test ./...
	@echo "Tests completed!"
