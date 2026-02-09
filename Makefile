include .env
export

test-server-run:
	go run cmd/testServer/main.go

test-server-build:
	go build -o .bin/testServer  cmd/testServer/main.go

test-server-start:
	.bin/testServer

test-server-build-and-start:
	$(MAKE) test-server-build
	$(MAKE) test-server-start

test-server-deploy:
	docker compose up --build

test-server-undeploy:
	docker compose down

test-server-deploy-postgres:
	docker compose up -d pg_service

postgres-migrate-up:
	migrate -path migrations -database "postgres://${PG_USER_NAME}:${PG_USER_PASSWORD}@localhost:${PG_PORT}/${PG_DB}?sslmode=disable" up

postgres-migrate-down:
	migrate -path migrations -database "postgres://${PG_USER_NAME}:${PG_USER_PASSWORD}@localhost:${PG_PORT}/${PG_DB}?sslmode=disable" down