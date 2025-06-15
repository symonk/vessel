## help: print this help message.
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## build: Build the executable for multiple platforms

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
	