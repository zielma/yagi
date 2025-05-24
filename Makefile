.PHONY: run test cover watch lint lint-fix build

run:
	go run cmd/app/main.go

watch:
	air

test:
	go test -race ./... 

test-verbose:
	go test ./... -v

cover:
	go test -race ./... -coverprofile=c.out
	go tool cover -html=c.out

lint:
	golangci-lint run 

lint-fix:
	golangci-lint run --fix

build: lint test
	go build -o bin/yagi cmd/app/main.go
