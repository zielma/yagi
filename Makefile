.PHONY: run test cover watch

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

