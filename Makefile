# Talus Prototype Makefile

-include .env
export

.PHONY: up down migrate logs cuda

# up: start the database container
up:
	docker compose -f deployments/docker-compose.yml --env-file .env up -d

# down: stop the database container
down:
	docker compose -f deployments/docker-compose.yml down

# migrate: run golang-migrate against the database
migrate:
	migrate -path ./db/migrations -database "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=$(POSTGRES_SSLMODE)" up

# logs: tail the database container logs
logs:
	docker compose -f deployments/docker-compose.yml logs -f

# cuda: builds CUDA terrain binary
cuda:
	$(MAKE) -C cuda/terrain
