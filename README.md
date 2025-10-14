# Flash2fy Card Service

Simple flashcard CRUD API built in Go using a hexagonal architecture.

## Requirements

- Go 1.23+ installed locally

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

## Manual Testing

Run the server and exercise the endpoints:

```sh
make run
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
