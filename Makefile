.PHONY: build deps test

export GO111MODULE=on

build:
	go build ./...

deps:
	go mod verify
	go mod tidy

test:
	go test ./...

integration-test:
	docker-compose -f Dockercompose.test.yml up --build --abort-on-container-exit --always-recreate-deps
	docker-compose -f Dockercompose.test.yml down --volumes
	
clean:
	docker-compose -f Dockercompose.test.yml rm -f
