GO ?= go

.PHONY: build run test tidy fmt

build:
	$(GO) build ./...

test:
	$(GO) test ./...

run:
	$(GO) run ./cmd/server

tidy:
	$(GO) mod tidy

fmt:
	gofmt -w ./cmd ./internal
