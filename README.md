# Flash2fy Card Service

Simple flashcard CRUD API built in Go using a hexagonal architecture.

## Requirements

- Go 1.23+ installed locally
- PostgreSQL 13+ running locally (default DSN assumes `postgres:postgres@localhost:5432`)

## Setup

```sh
git clone <repo>
cd flash2fy
go mod download
```

## Make Targets

```sh
make build   # compile the project
make test    # run unit tests
make run     # start the API server on :8080
make tidy    # tidy go.mod/go.sum
make fmt     # format Go sources
```

## Database

Set up the `cards` table in your PostgreSQL database:

```sql
CREATE TABLE IF NOT EXISTS cards (
  id         TEXT PRIMARY KEY,
  front      TEXT NOT NULL,
  back       TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);
```

The server reads the connection string from `DATABASE_URL`; if not provided it defaults to `postgres://postgres:postgres@localhost:5432/flash2fy?sslmode=disable`.

## Telegram Bot

Set the `TELEGRAM_BOT_TOKEN` environment variable with the token issued by BotFather to enable Telegram integration. When active, any text message you send to the bot becomes the front of a newly created card (back is stored empty).

The bot replies with the generated card identifier and echoes the stored content. Omit `TELEGRAM_BOT_TOKEN` to disable the bot.

## Manual Testing

Run the server and exercise the endpoints:

```sh
DATABASE_URL=postgres://postgres:postgres@localhost:5432/flash2fy?sslmode=disable make run
# in another shell
curl -s -X POST http://localhost:8080/v1/cards \
  -H 'Content-Type: application/json' \
  -d '{"front":"What is Go?","back":"A programming language"}'

curl -s http://localhost:8080/v1/cards

curl -s http://localhost:8080/v1/cards/<id>

curl -s -X PUT http://localhost:8080/v1/cards/<id> \
  -H 'Content-Type: application/json' \
  -d '{"front":"Updated question","back":"Updated answer"}'

curl -i -X DELETE http://localhost:8080/v1/cards/<id>
```

Replace `<id>` with the identifier returned from the create response.

> Tests use the in-memory repository adapter, so `make test` does not require a running PostgreSQL instance.
