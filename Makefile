.PHONY: build lint test fmt

build:
	go build -o forest .

lint:
	golangci-lint run

test:
	gotestsum ./...

fmt:
	golangci-lint run --fix
