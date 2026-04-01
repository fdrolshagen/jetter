# Makefile for Jetter

BINARY = bin/jetter
EXAMPLES = ./examples

.PHONY: build install run test local-setup help

help:
	@echo "📖 Available commands:"
	@echo "  🚀 build         - Build the project binary ($(BINARY))"
	@echo "  📦 install       - Install the binary to ~/bin"
	@echo "  🏃 run           - Run Jetter with example HTTP file"
	@echo "  🧪 test          - Run all Go tests"
	@echo "  🐳 local-setup   - Start local environment with Docker Compose"
	@echo "  📊 coverage      - Generate test coverage report"

build:
	@echo "🚀 Building the project..."
	@go build -o $(BINARY) main.go
	@echo "✅ Build complete: $(BINARY)"

install: build
	@echo "📦 Installing $(BINARY) to ~/bin..."
	@cp $(BINARY) ~/bin
	@echo "✅ Installation complete!"

run: build
	@echo "🏃 Running Jetter with example HTTP file..."
	@$(BINARY) -f $(EXAMPLES)/example.http -e $(EXAMPLES)/http-client.env.json:local
	@echo "✅ Run finished!"

test:
	@echo "🧪 Running tests..."
	@go test ./...
	@echo "✅ Tests finished!"

local-setup:
	@echo "🐳 Starting local setup with Docker Compose..."
	@docker-compose -f testing/docker-compose.yml up --remove-orphans
	@echo "✅ Local setup complete!"

coverage:
	@echo "📊 Generating coverage report..."
	@go test ./...  -coverpkg=./... -coverprofile ./bin/coverage.out
	@go tool cover -func ./coverage.out
	@echo "✅ Coverage report generated"