PWD := $(shell cd "$(CURDIR)" && pwd)

## help: print this help message.
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## check-docker: Checks if docker is available in the system
.PHONY: check-docker
check-docker:
	@command -v docker >/dev/null 2>&1 || { echo >&2 "Docker is not installed or not in PATH."; exit 1; }
	@docker info >/dev/null 2>&1 || { echo >&2 "Docker is not running or not accessible."; exit 1; }

## lint: Applies linters and static analysis tools
.PHONY: lint
lint: check-docker
	docker run --rm -t \
		-v $(PWD):/app -w /app \
		--user $(shell id -u):$(shell id -g) \
		-v $(shell go env GOCACHE):/.cache/go-build -e GOCACHE=/.cache/go-build \
		-v $(shell go env GOMODCACHE):/.cache/mod -e GOMODCACHE=/.cache/mod \
		-v $(HOME)/.cache/golangci-lint:/.cache/golangci-lint -e GOLANGCI_LINT_CACHE=/.cache/golangci-lint \
		golangci/golangci-lint:v2.1.6 golangci-lint run


## build: Build the executable for multiple platforms
.PHONY: build

## test: Execute all the tests and collect coverage information.
.PHONY: test
test:
	export GOEXPERIMENT=synctest
	go test -v -race -timeout=10s -cover ./...

## cover: Execute all the tests and show the coverage report.
.PHONY: cover
cover:
	go test -v -race -timeout=10s -cover -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	open coverage.html

.PHONY: benchmark
benchmark:
	go test -bench=./... -benchmem -v -run=^$
	