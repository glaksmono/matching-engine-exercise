.PHONY: build test vendor 

build:
		@echo "Installing dependencies..."
		@go mod tidy && go mod vendor
		@echo "Building the application..."
		@go build -o ./engine main.go 

test:
		@echo "generating huge orders json"
		@go run generator/generator.go
		@echo "Running unit test..."
		@go test -v

vendor:
		@echo "Installing dependencies..."
		@go mod tidy
		@go mod vendor

.DEFAULT_GOAL := build
