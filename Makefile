#===================#
#== Env Variables ==#
#===================#
DOCKER_COMPOSE_FILE ?= docker-compose.dev.yml
RUN_IN_DOCKER ?= docker compose -f ${DOCKER_COMPOSE_FILE} exec -T app
WITH_TTY ?= -t

#===============#
#== App Build ==#
#===============#

build-native: RUN_IN_DOCKER=
build-native: build

build: ## Build application
build:
	@echo "==========================="
	@echo "Building binary"
	@echo "==========================="
	${RUN_IN_DOCKER} sh -c "go build -mod=vendor -o bin/app cmd/main.go"


run-server: ## Run HTTP server
run-server: build
	docker compose -f ${DOCKER_COMPOSE_FILE} exec -t app sh -c "./scripts/run_server.sh"

clean: # clean executables
	rm -rf bin/*

#=======================#
#== ENVIRONMENT SETUP ==#
#=======================#

create-env-file:
	cp .env.sample .env

docker-start:
	@echo "=========================="
	@echo "Starting Docker Containers"
	@echo "=========================="
	docker compose -f ${DOCKER_COMPOSE_FILE} up -d --build --remove-orphans
	docker compose -f ${DOCKER_COMPOSE_FILE} ps

docker-stop:
	@echo "=========================="
	@echo "Stopping Docker Containers"
	@echo "=========================="
	docker compose -f ${DOCKER_COMPOSE_FILE} stop
	docker compose -f ${DOCKER_COMPOSE_FILE} ps

docker-clean: docker-stop
	@echo "=========================="
	@echo "Removing Docker Containers"
	@echo "=========================="
	docker compose -f ${DOCKER_COMPOSE_FILE} rm -v -f

docker-restart: docker-stop docker-start

environment: ## The only command needed to start a working environment
environment: create-env-file docker-restart build

environment-clean: ## The only command needed to clean the environment
environment-clean: docker-clean clean
	rm -rf postgres-data

migrate-up: ## Run migrations against non test DB
	APP_ENV=development docker compose -f ${DOCKER_COMPOSE_FILE} run --rm ${WITH_TTY} migrate up

migrate-down: ## Rollback migrations against non test DB
	APP_ENV=development docker compose -f ${DOCKER_COMPOSE_FILE} run --rm ${WITH_TTY} migrate down 1

migrate-test-up: ## Run migrations against test DB
	APP_ENV=test docker compose -f ${DOCKER_COMPOSE_FILE} run --rm ${WITH_TTY} migrate up

migrate-test-down: ## Rollback migrations against test DB
	APP_ENV=test docker compose -f ${DOCKER_COMPOSE_FILE} run --rm ${WITH_TTY} migrate down 1

migrate-create: ## Create a DB migration. You need to pass the file name, e.g. `make migrate-create name=migration-name`
	docker compose -f ${DOCKER_COMPOSE_FILE} run --rm migrate create -ext sql -dir /migrations $(name)

#====================#
#== QUALITY CHECKS ==#
#====================#

static-analysis: ## Run all enabled linters
static-analysis: create-env-file
	@echo "======================="
	@echo "Running static analysis"
	@echo "======================="
	${RUN_IN_DOCKER} golangci-lint run

tests-suite: ## Run all the tests suite
tests-suite: tests-unit tests-integration

tests-unit: ## Run unit tests
tests-unit: create-env-file
	@echo "=================="
	@echo "Running unit tests"
	@echo "=================="
	${RUN_IN_DOCKER} sh -c "./scripts/unit_test.sh"

tests-integration: ## Run integration tests
tests-integration: migrate-test-up
	@echo "========================"
	@echo "Running integration tests"
	@echo "========================"
	${RUN_IN_DOCKER} sh -c ./scripts/integration_test.sh

validate-swagger: ## Validate swagger
validate-swagger:
	which redocly || npm install @redocly/cli -g
	redocly lint docs/api.yaml

prepare-ci: ## Prepare CI environment
prepare-ci: create-env-file docker-restart build migrate-test-up

ci: ## Run all CI checks
ci: static-analysis tests-unit tests-integration

ci-native: RUN_IN_DOCKER=
ci-native: ci

#==========#
#== DOCS ==#
#==========#

api-docs: ## Generate API docs with swaggo
	@echo "========================="
	@echo "Generate Swagger API Docs"
	@echo "========================="
	go install github.com/swaggo/swag/cmd/swag@latest
	swag init --parseDependency --parseInternal -g cmd/main.go -ot "json"
