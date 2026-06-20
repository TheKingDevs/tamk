.PHONY: build build-linux build-arm install run test test-short lint vet fmt check tidy clean coverage pre-commit dev setup docker-build docker-up bench mocks help

APP_NAME = tamk
BIN_DIR = bin
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git log -1 --format=%h 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS = -ldflags="-X 'github.com/Shadw-Developer/tamk/config.Version=$(VERSION)' -X 'github.com/Shadw-Developer/tamk/config.Commit=$(COMMIT)' -X 'github.com/Shadw-Developer/tamk/config.Date=$(DATE)'"

build:
	go build $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME) ./cmd/$(APP_NAME)

build-linux:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME)-linux-amd64 ./cmd/$(APP_NAME)

build-arm:
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME)-linux-arm64 ./cmd/$(APP_NAME)

install:
	go install $(LDFLAGS) ./cmd/$(APP_NAME)

run: build
	./$(BIN_DIR)/$(APP_NAME)

test:
	go test -race -cover -v ./...

test-short:
	go test -cover -short ./...

lint:
	golangci-lint run ./...

vet:
	go vet ./...

fmt:
	gofmt -s -w .

check: fmt vet lint test

tidy:
	go mod tidy && go mod verify

clean:
	rm -rf $(BIN_DIR)
	rm -rf tmp/
	rm -f coverage.out
	rm -f coverage.html

coverage:
	go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html

pre-commit: fmt vet test-short

dev:
	go run ./cmd/$(APP_NAME) --create

setup:
	go mod tidy
	go mod download

docker-build:
	docker build -t $(APP_NAME) .

docker-up:
	docker-compose up -d

bench:
	go test -bench=. -benchmem ./...

mocks:
	go install github.com/vektra/mockery/v2@latest
	mockery --dir=internal/domain --output=internal/domain/mocks --outpkg=mocks --all

help:
	@echo 'Usage: make <target>'
	@echo ''
	@echo 'Development:'
	@echo '  build          Build the binary (default)'
	@echo '  build-linux    Build for linux/amd64'
	@echo '  build-arm      Build for linux/arm64'
	@echo '  install        go install'
	@echo '  run            Build and run'
	@echo '  dev            Run in dev mode'
	@echo ''
	@echo 'Quality:'
	@echo '  fmt            Format all Go code'
	@echo '  vet            Run go vet'
	@echo '  lint           Run golangci-lint'
	@echo '  check          Full CI pipeline (fmt + vet + lint + test)'
	@echo '  pre-commit     Fast pre-commit checks (fmt + vet + test-short)'
	@echo ''
	@echo 'Testing:'
	@echo '  test           Run all tests with race detector'
	@echo '  test-short     Run tests without race flag'
	@echo '  coverage       Generate coverage report'
	@echo '  bench          Run benchmarks'
	@echo ''
	@echo 'Housekeeping:'
	@echo '  tidy           go mod tidy && verify'
	@echo '  clean          Remove build artifacts'
	@echo '  setup          Install dependencies'
	@echo ''
	@echo 'Docker:'
	@echo '  docker-build   Build Docker image'
	@echo '  docker-up      Start Docker Compose'
	@echo ''
	@echo 'Tools:'
	@echo '  mocks          Generate mocks with mockery'
