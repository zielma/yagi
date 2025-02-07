.PHONY: run test cover watch

run:
	go run cmd/app/main.go

watch:
	air

test:
	go test ./... 

test-verbose:
	go test ./... -v

cover:
	go test ./... -coverprofile=c.out
	go tool cover -html=c.out

