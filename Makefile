VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

.PHONY: build lint test fmt

build:
	go build -ldflags "-X github.com/mhamza15/forest/cmd.Version=$(VERSION)" -o forest .

lint:
	golangci-lint run

test:
	gotestsum ./...

fmt:
	golangci-lint run --fix
