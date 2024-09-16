
include .envrc
#psql --host=localhost --dbname=greenlight-db --username=greenlight-admin

# ==================================================================================== #
 # HELPERS 
# ==================================================================================== #



.PHONY: help
help:
	@echo 'Usage'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

# ==================================================================================== #
 # DEVELOPMENT 
# ==================================================================================== #

## run/api: run the cmd/api application
.PHONY: run/api
run/api:
	@go run ./cmd/api -db-dsn=${GREENLIGHT_DB_DSN}

## db/psql: connect to the database using psql
.PHONY: db/psql
db/psql:
	psql ${GREENLIGHT_DB_DSN}

## db/migrations/up: apply all up database migrations 
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo 'Running up migration...'
	migrate -path ./migration -database ${GREENLIGHT_DB_DSN} up

## db/migrations/new name=$1: create a new database migration
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext sql -dir migration ${name}

# ==================================================================================== #
 # QUALITY CONTROL 
# ==================================================================================== #

## audit: tidy dependencies and format, vet and test all code
.PHONY: audit
audit: vendor
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...

## vendor: tidy and vendor dependency
.PHONY: vendor
vendor:
	@echo 'Tidying and verifiying module dependencies'
	go mod tidy
	go mod verify
	@echo 'Vendoring dependencies'
	go mod vendor

# ==================================================================================== #
 # BUILD 
# ==================================================================================== #

current_time = $(shell date --iso-8601=seconds) 
linker_flags = '-s'

## build/api: build the cmd/api application
.PHONY: build/api
build/api:
	@echo 'Building cmd/api...'
	go build  -o=./bin/api -ldflags='${linker_flags}' ./cmd/api
	GOOS=linux GOARCH=amd64 go build -o=./bin/linux_amd64/api -ldflags='${linker_flags}' ./cmd/api
