GO ?= go

.PHONY: build run test tidy fmt db

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

db:
	docker rm -f flash2fy-postgres 2>/dev/null || true
	docker run -d --name flash2fy-postgres \
		-e POSTGRES_DB=flash2fy \
		-e POSTGRES_USER=postgres \
		-e POSTGRES_PASSWORD=postgres \
		-p 5432:5432 \
		postgres:15-alpine
