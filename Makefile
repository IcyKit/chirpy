-include .env
export

MIGRATIONS_DIR := sql/schema

.PHONY: build run test vet fmt sqlc migrate-up migrate-down

build:
	go build -o bin/chirpy ./cmd/chirpy

run:
	go run ./cmd/chirpy

test:
	go test -race ./...

vet:
	go vet ./...

fmt:
	gofmt -w .

sqlc:
	sqlc generate

migrate-up:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" up

migrate-down:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" down
