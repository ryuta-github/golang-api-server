.PHONY: run test test-integration migrate-up migrate-down

run:
	go run ./cmd/api

test:
	go test ./...

test-integration:
	go test -tags=integration ./internal/infrastructure/postgres/...

migrate-up:
	docker compose run --rm migrate

migrate-down:
	docker compose run --rm migrate down 1
