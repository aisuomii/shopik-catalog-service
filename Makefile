BINARY := catalog-service
CMD := ./cmd/catalog-service
BIN_DIR := bin

LOCAL_DIR := local-deployment
ENV_FILE := $(LOCAL_DIR)/.env
ENV_EXAMPLE := $(LOCAL_DIR)/.env.example
COMPOSE := docker compose -f $(LOCAL_DIR)/docker-compose.yaml

# Local settings come from local-deployment/.env when it exists; the defaults
# below keep every target working on a fresh clone without one. Both the
# database container and the service read these same names, so the two cannot
# drift apart.
ifneq (,$(wildcard $(ENV_FILE)))
include $(ENV_FILE)
endif

POSTGRES_USER ?= catalog
POSTGRES_PASSWORD ?= catalog
POSTGRES_DB ?= catalog
POSTGRES_PORT ?= 5433
CONFIG_PATH ?= $(LOCAL_DIR)/config.yaml

export POSTGRES_USER POSTGRES_PASSWORD POSTGRES_DB POSTGRES_PORT CONFIG_PATH

.DEFAULT_GOAL := help

## help: list the available targets
help:
	@echo "shopik-catalog-service"
	@echo ""
	@grep -E '^## ' $(MAKEFILE_LIST) | sed -e 's/^## /  make /' -e 's/: /\t/' | expand -t 22
	@echo ""
	@echo "postgres: localhost:$(POSTGRES_PORT)/$(POSTGRES_DB)  config: $(CONFIG_PATH)"

## doctor: check that the required tooling is present
doctor:
	@go version || { echo "go is required"; exit 1; }
	@docker version --format 'docker {{.Server.Version}}' || { echo "docker must be installed and running"; exit 1; }
	@echo "go.mod requires: $$(go mod edit -json | sed -n 's/.*"Go": "\([^"]*\)".*/go \1/p')"

## env: create local-deployment/.env from the example if it is missing
env:
	@test -f $(ENV_FILE) || { cp $(ENV_EXAMPLE) $(ENV_FILE); echo "created $(ENV_FILE)"; }

## up: start the local postgres container and wait until it is healthy
up: env
	$(COMPOSE) up -d --wait

## down: stop the local postgres container, keeping its data
down:
	$(COMPOSE) down

## clean: stop the local postgres container and delete its volume
clean:
	$(COMPOSE) down -v
	rm -rf $(BIN_DIR)

## logs: follow the postgres container logs
logs:
	$(COMPOSE) logs -f postgres

## psql: open a psql shell inside the postgres container
psql:
	$(COMPOSE) exec postgres psql -U $(POSTGRES_USER) -d $(POSTGRES_DB)

## run: start the database if needed, then run the service against it
run: up
	go run $(CMD)

## build: compile the service into bin/
build:
	go build -o $(BIN_DIR)/$(BINARY) $(CMD)

## test: run the test suite
test:
	go test ./...

## fmt: format the code
fmt:
	gofmt -s -w .

## tidy: sync go.mod and go.sum with the imports
tidy:
	go mod tidy

## check: verify formatting, vet and tests the way CI would
check:
	@test -z "$$(gofmt -s -l .)" || { echo "gofmt needed:"; gofmt -s -l .; exit 1; }
	go vet ./...
	go test ./...

.PHONY: help doctor env up down clean logs psql run build test fmt tidy check
