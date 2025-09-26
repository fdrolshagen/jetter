build:
	@go build -o bin/jetter main.go

run: build
	@bin/jetter

test:
	@go test ./...