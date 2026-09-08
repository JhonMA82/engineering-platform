.PHONY: build test vet cover

build:
	go build ./...

test:
	go test ./...

vet:
	go vet ./...

cover:
	go test -coverprofile=cover.out ./... && go tool cover -func=cover.out
