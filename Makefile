.PHONY: build run test clean deps

# Build the application
build:
	go build -o bin/silentrig main.go

# Run the application
run:
	go run main.go

# Run with development mode
dev:
	LOG_LEVEL=debug go run main.go

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf data/

# Create necessary directories
setup:
	mkdir -p bin
	mkdir -p data

# Build for different platforms
build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/silentrig-linux-amd64 main.go

build-windows:
	GOOS=windows GOARCH=amd64 go build -o bin/silentrig-windows-amd64.exe main.go

build-darwin:
	GOOS=darwin GOARCH=amd64 go build -o bin/silentrig-darwin-amd64 main.go

# Build all platforms
build-all: build-linux build-windows build-darwin

# Install the application
install: build
	cp bin/silentrig /usr/local/bin/silentrig

# Uninstall the application
uninstall:
	rm -f /usr/local/bin/silentrig 