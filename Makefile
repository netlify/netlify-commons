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

integration-test:
	docker-compose -f Dockercompose.test.yml up --build --abort-on-container-exit --always-recreate-deps
	docker-compose -f Dockercompose.test.yml down --volumes

clean:
	docker-compose -f Dockercompose.test.yml rm -f

kafkacat:
	docker run -it --network=host confluentinc/cp-kafkacat kafkacat -b localhost:19092 -C -t gotest -J
