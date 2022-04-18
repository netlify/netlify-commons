.PHONY: build deps test

export GO111MODULE=on

arch = $(shell uname -p)
gotags=
ifeq ($(arch),arm)
gotags = -tags dynamic
endif

build:
	go build $(gotags) ./...

deps:
	go mod verify
	go mod tidy

test:
	go test -race $(gotags) ./...
