.PHONY: all deps test

PROJECT = netlify-commons

help: ## Show this help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

all: deps test

deps: ## Install dependencies.
	go get -u github.com/Masterminds/glide && glide install

test: ## Run tests.
	go test -v `go list ./... | grep -v /vendor/`
