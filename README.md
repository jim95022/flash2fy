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

### Configuration

Runtime configuration comes from environment variables. For local development you can drop a `.env` file in the repo root (not committed) with entries such as:

```
SERVER_ADDR=:8080
DATABASE_URL=postgres://postgres:postgres@localhost:5432/flash2fy?sslmode=disable
TELEGRAM_BOT_TOKEN=<bot-token>
TELEGRAM_WEBHOOK_URL=https://<public-host>
TELEGRAM_WEBHOOK_SECRET=<optional-secret>
TELEGRAM_WEBHOOK_PATH=/telegram/webhook
```

Values from `.env` override the defaults baked into the app; you can also export these variables directly in your shell.

## Make Targets

```sh
make build   # compile the project
make test    # run unit tests
make run     # start the API server on :8080
make tidy    # tidy go.mod/go.sum
make fmt     # format Go sources
make db      # start a local PostgreSQL container on 5432
```

## Database

Set up the `cards` table in your PostgreSQL database:

```sql
CREATE TABLE IF NOT EXISTS cards (
  id         TEXT PRIMARY KEY,
  front      TEXT NOT NULL,
  back       TEXT NOT NULL,
  owner_id   TEXT,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS users (
  id        TEXT PRIMARY KEY,
  nickname  TEXT NOT NULL
);
```

The server reads the connection string from `DATABASE_URL`; if not provided it defaults to `postgres://postgres:postgres@localhost:5432/flash2fy?sslmode=disable`.

## Telegram Bot

Set the `TELEGRAM_BOT_TOKEN` environment variable with the token issued by BotFather to enable Telegram integration. When active, any text message you send to the bot becomes the front of a newly created card (back is stored empty).

To run the bot via webhook you must also provide:

```
TELEGRAM_WEBHOOK_URL=https://<public-host>                   # must be HTTPS and reachable by Telegram
TELEGRAM_WEBHOOK_SECRET=<optional-secret>                     # optional secret for X-Telegram-Bot-Api-Secret-Token
TELEGRAM_WEBHOOK_PATH=/telegram/webhook                       # local route served by the app
```

If `TELEGRAM_WEBHOOK_URL` already carries a path segment, that path is used automatically and `TELEGRAM_WEBHOOK_PATH` is ignored. Otherwise the path value (default `/telegram/webhook`) is appended to the base URL so the registered endpoint and local route remain in sync.

For local testing you can expose your server using a tunnel (e.g. `ngrok http 8080`) and point `TELEGRAM_WEBHOOK_URL` to the generated HTTPS endpoint.

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
