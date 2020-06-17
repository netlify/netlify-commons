.PHONY: build deps test

export GO111MODULE=on

build:
	go build ./...

deps:
	go mod verify
	go mod tidy

test:
	go test -v ./...
