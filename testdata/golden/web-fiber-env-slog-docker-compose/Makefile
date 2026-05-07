.PHONY: test run build docker-build docker-run compose-up compose-down compose-logs

APP_NAME ?= orders-api
BUILD_DIR ?= ./build
MAIN_PACKAGE ?= ./cmd/orders-api
HTTP_PORT ?= 8080
IMAGE_NAME ?= orders-api:local
DOCKER_RUN_ENV ?=
COMPOSE_FILE ?= deployments/docker-compose.yml


test:
	go test ./...

run:
	go run $(MAIN_PACKAGE)

build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PACKAGE)


docker-build:
	docker build -t $(IMAGE_NAME) .

docker-run:
	docker run --rm -p $(HTTP_PORT):$(HTTP_PORT) $(DOCKER_RUN_ENV) $(IMAGE_NAME)


compose-up:
	docker compose -f $(COMPOSE_FILE) up --build

compose-down:
	docker compose -f $(COMPOSE_FILE) down

compose-logs:
	docker compose -f $(COMPOSE_FILE) logs -f app
