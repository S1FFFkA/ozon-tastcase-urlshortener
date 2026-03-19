.PHONY: test test-integration db-up docker-up docker-up-d docker-down docker-reset logs

test:
	go test ./...

test-integration:
	go test -tags=integration ./internal/repository -v

db-up:
	docker compose up -d db

docker-up:
	docker compose up --build

docker-up-d:
	docker compose up -d --build

docker-down:
	docker compose down

docker-reset:
	docker compose down -v

logs:
	docker compose logs -f app db
