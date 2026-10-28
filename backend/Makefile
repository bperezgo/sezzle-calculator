.PHONY: run build test test-race cover cover-html lint lint-fix fmt generate-mocks regenerate-mocks vuln check

MOCKERY := $(shell command -v mockery 2>/dev/null || echo $(HOME)/go/bin/mockery)

run:
	GIN_MODE=debug go run ./cmd

build:
	go build -o build/calculator ./cmd

test:
	go test ./... -count=1

test-race:
	go test ./... -race -count=1 -coverprofile=coverage.out

cover: test-race
	go tool cover -func=coverage.out | tail -n 1

cover-html: test-race
	go tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run

lint-fix:
	golangci-lint run --fix

fmt:
	golangci-lint fmt

generate-mocks:
	$(MOCKERY)

regenerate-mocks:
	rm -rf internal/domain/ports/mocks
	$(MAKE) generate-mocks

vuln:
	govulncheck ./...

# Everything that must be green before a commit
check: lint test-race
