.PHONY: build test clean install fmt vet run

# Binary name
BINARY=cliprouter

# Build the application
build:
	go build -o $(BINARY) .

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -cover ./...

# Clean build artifacts
clean:
	rm -f $(BINARY)
	go clean

# Install to /usr/local/bin
install: build
	sudo mv $(BINARY) /usr/local/bin/

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Run the application
run: build
	./$(BINARY)

# Run with verbose logging
run-verbose: build
	./$(BINARY) -v

# Run in dry-run mode
dry-run: build
	./$(BINARY) --dry-run

# Download dependencies
deps:
	go mod download
	go mod tidy

# Run all checks (fmt, vet, test)
check: fmt vet test

all: check build
