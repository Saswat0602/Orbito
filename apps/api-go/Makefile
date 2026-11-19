.PHONY: all build clean dev test test-coverage lint fmt vet run tidy

APP_NAME = api
BUILD_DIR = bin
CMD_DIR = cmd/api

all: build

build:
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="-w -s" -o $(BUILD_DIR)/$(APP_NAME) ./$(CMD_DIR)

run:
	go run ./$(CMD_DIR)

dev:
	air

test:
	go test -v -race ./...

test-coverage:
	@mkdir -p coverage
	go test -race -coverprofile=coverage/coverage.out ./...
	go tool cover -html=coverage/coverage.out -o coverage/coverage.html

lint:
	golangci-lint run ./...

fmt:
	gofmt -s -w .

vet:
	go vet ./...

clean:
	rm -rf $(BUILD_DIR) coverage

tidy:
	go mod tidy
