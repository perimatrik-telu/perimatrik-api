.PHONY: run build test tidy

run:
	go run ./cmd/server

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/perimatrik-api ./cmd/server

test:
	go test ./...

tidy:
	go mod tidy
