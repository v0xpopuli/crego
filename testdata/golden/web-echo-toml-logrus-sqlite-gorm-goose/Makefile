.PHONY: test run build

APP_NAME ?= orders-api
BUILD_DIR ?= ./build
MAIN_PACKAGE ?= ./cmd/orders-api
HTTP_PORT ?= 8080


test:
	go test ./...

run:
	go run $(MAIN_PACKAGE)

build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PACKAGE)
