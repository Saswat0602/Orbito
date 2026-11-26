.PHONY: all build clean dev test test-coverage lint fmt vet run tidy generate \
	db-generate db-migrate db-migrate-deploy db-status db-seed db-seed-admin \
	db-seed-demo db-seed-clear db-seed-reset db-reset ctl

export GOTOOLCHAIN = local

APP_NAME = api
CTL_NAME = orbitoctl
BUILD_DIR = bin
CMD_DIR = cmd/api
CTL_DIR = cmd/orbitoctl

all: build

build:
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="-w -s" -o $(BUILD_DIR)/$(APP_NAME) ./$(CMD_DIR)

ctl:
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="-w -s" -o $(BUILD_DIR)/$(CTL_NAME) ./$(CTL_DIR)

generate:
	sqlc generate -f db/sqlc.yaml

# Database commands (interacting via standard Prisma migration authority)
db-generate: generate

db-migrate:
	@npm --prefix ../.. run db:migrate

db-migrate-deploy:
	@npm --prefix ../.. run db:migrate:deploy

db-status:
	@npm --prefix ../.. run prisma -- migrate status

db-seed:
	@npm --prefix ../.. run db:seed

db-seed-admin:
	@npm --prefix ../.. run db:seed:admin

db-seed-demo:
	@npm --prefix ../.. run db:seed:demo

db-seed-clear:
	@npm --prefix ../.. run db:seed:clear

db-seed-reset:
	@npm --prefix ../.. run db:seed:reset

db-reset:
	@npm --prefix ../.. run db:reset

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
