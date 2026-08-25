.PHONY: run build test swagger seed migrate-create migrate-up migrate-down migrate-version

run:
	go run ./cmd

build:
	go build ./...

test:
	go test ./...

swagger:
	go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/main.go

migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

migrate-up:
	go run ./cmd/migrate up

migrate-down:
	go run ./cmd/migrate down 1

migrate-version:
	go run ./cmd/migrate version

seed:
	go run ./cmd/seed
