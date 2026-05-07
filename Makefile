.PHONY: build clean fmt fmt-check vet test tests lint release-snapshot

APP_NAME ?= crego
BUILD_DIR ?= ./build/app
MAIN_PACKAGE ?= ./cmd/crego
GO_PACKAGES ?= ./...
GO_FILES := $(shell find . -name '*.go' -not -path './dist/*' -not -path './build/*')

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

fmt:
	@echo "Formatting Go files..."
	@gofmt -w $(GO_FILES)
	@echo "Formatting completed!"

fmt-check:
	@echo "Checking Go formatting..."
	@files="$$(gofmt -l $(GO_FILES))"; \
	if [ -n "$$files" ]; then \
		echo "$$files"; \
		exit 1; \
	fi
	@echo "Formatting check completed!"

vet:
	@echo "Running go vet..."
	@go vet $(GO_PACKAGES)
	@echo "Vet completed!"

test:
	@echo "Running tests..."
	@go test $(GO_PACKAGES) -v
	@echo "Tests completed!"

tests: test

lint: fmt-check vet test

release-snapshot:
	@goreleaser release --snapshot --clean
