.PHONY: build proto test clean run-scheduler run-worker docker-build

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Binary names
SCHEDULER_BINARY=bin/scheduler
WORKER_BINARY=bin/worker

# Build all binaries
build: build-scheduler build-worker

# Build scheduler
build-scheduler:
	$(GOBUILD) -o $(SCHEDULER_BINARY) cmd/scheduler/main.go

# Build worker
build-worker:
	$(GOBUILD) -o $(WORKER_BINARY) cmd/worker/main.go

# Generate gRPC code from proto files
proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/proto/*.proto

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests with race detector
test-race:
	$(GOTEST) -race -v ./...

# Run integration tests
test-integration:
	$(GOTEST) -v -tags=integration ./tests

# Run benchmarks
bench:
	$(GOTEST) -bench=. -benchmem ./tests

# Run scheduler
run-scheduler:
	$(SCHEDULER_BINARY) --config config/scheduler.yaml

# Run worker
run-worker:
	$(WORKER_BINARY) --config config/worker.yaml

# Install dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f *.log

# Docker build
docker-build:
	docker build -f deployments/docker/Dockerfile.scheduler -t orchestrator-scheduler .
	docker build -f deployments/docker/Dockerfile.worker -t orchestrator-worker .

# Docker compose up
docker-up:
	docker-compose -f deployments/docker/docker-compose.yaml up

# Docker compose down
docker-down:
	docker-compose -f deployments/docker/docker-compose.yaml down
