SHELL := /bin/zsh

# Variables
BINARY := random_number_service
PKG := ./src
BIN_DIR := ./bin

.PHONY: all build run tidy clean dockerize help

all: build

help:
	@echo "Targets:"
	@echo "  build  - Build the service binary to $(BIN_DIR)/$(BINARY)"
	@echo "  run    - Run the service with go run"
	@echo "  dockerize - Build the Docker image 'random-number-generator:latest'"
	@echo "  tidy   - Download and tidy Go module dependencies"
	@echo "  clean  - Remove build artifacts"

$(BIN_DIR):
	@mkdir -p $(BIN_DIR)

build: tidy $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY) $(PKG)

run: tidy
	go run $(PKG)

tidy:
	go mod tidy

clean:
	rm -rf $(BIN_DIR)

# Build Docker image
dockerize:
	docker build -t random-number-generator:latest .
