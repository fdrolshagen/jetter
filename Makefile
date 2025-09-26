build:
	@go build -o bin/jetter main.go

run: build
	@bin/jetter -f ./examples/example.http

test:
	@go test ./...