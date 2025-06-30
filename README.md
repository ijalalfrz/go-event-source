# GO Clean Architecture with Event Sourcing Schema

Prereuisties:
- Go v1.23.4
- docker
- docker compose



# How to run
- Migrate DB using commnad `make migrate-up`
- Make environtment using command `make environment`
- Run HTTP Server using command `make run-server`


# Test
- To run all unit and integration test use command `make tests-suite`
- To run unit test use command `make tests-unit`
- To run integration test use command `make tests-integration`


### Components

Components based on `docker-compose.dev.yml`:

- app: the application itself
- postgres: the database
- migrate: responsible for running and rollback database migrations
- swagger: OpenAPI documentation

### Dependencies

Beside Go official packages, this project use following 3rd party packages:

- go-kit: the application toolkit that allow us to have loosely coupled application layers
- go-chi: HTTP router
- cobra: command line application that allow us to have subcommands and parameters
- viper: application configurations loader
- go-playground/validator: request validator
- godog: cucumber tests for integration tests
- gojsonschema: JSON schema validator
- gojsonq: JSON querier
- go-testfixtures: automating load and teardown test data
- testify: test assertion helper



# Assumptions 
- Event will be our source of truth timeline of balance movement we can replay event at specific time or scale app later with OLAP service using event data, there will be `init_balance`, `deposit_received`, `balance_debited`, `balance_credited` event
- Event also has a sequence number to capture all historical event also as a optimistic locking mechanism if there are multiple event with the same sequence number will be prevented using unique index
- Account will store latest balance per account or source of truth balance in this OLTP service
- `X-TRANSACTION-ID` will be used as our idempotent id to prevent double requests and stored in event table
- `X-TIMESTAMP` is used in account cration, transfer balance API to prevent invalid request time and replay attacks
- Signature verification: would be good to implement but in this project we don’t use it
