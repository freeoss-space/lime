BINARY     := lime
MODULE     := github.com/freeoss-space/lime
CMD        := ./cmd/lime
VERSION    ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT     ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE       ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS    := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"
GOFLAGS    := -trimpath

COVERAGE_THRESHOLD := 70

.PHONY: all build clean test test-race lint fmt vet coverage install help

all: build

## build: Build the binary
build:
	go build $(GOFLAGS) $(LDFLAGS) -o bin/$(BINARY) $(CMD)

## install: Install to GOPATH/bin
install:
	go install $(GOFLAGS) $(LDFLAGS) $(CMD)

## test: Run tests
test:
	go test ./... -timeout 60s

## test-race: Run tests with race detector
test-race:
	go test -race ./... -timeout 60s

## test-verbose: Run tests verbosely
test-verbose:
	go test -v ./... -timeout 60s

## coverage: Run tests with coverage
coverage:
	go test -coverprofile=coverage.out ./... -timeout 60s
	go tool cover -func=coverage.out
	@COVERAGE=$$(go tool cover -func=coverage.out | grep total | awk '{print substr($$3, 1, length($$3)-1)}'); \
		echo "Total coverage: $$COVERAGE%"; \
		if [ $$(echo "$$COVERAGE < $(COVERAGE_THRESHOLD)" | bc -l) -eq 1 ]; then \
			echo "Coverage below threshold ($(COVERAGE_THRESHOLD)%)"; exit 1; \
		fi

## coverage-html: Generate HTML coverage report
coverage-html: coverage
	go tool cover -html=coverage.out -o coverage.html

## lint: Run golangci-lint
lint:
	golangci-lint run ./...

## fmt: Format code
fmt:
	gofmt -w -s .
	goimports -w -local $(MODULE) .

## vet: Run go vet
vet:
	go vet ./...

## clean: Remove build artifacts
clean:
	rm -rf bin/ dist/ coverage.out coverage.html

## tidy: Tidy dependencies
tidy:
	go mod tidy

## help: Show this help
help:
	@grep -E '^## ' Makefile | sed 's/## //'

bin/:
	mkdir -p bin
