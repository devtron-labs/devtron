SHELL					        ?= bash

POSTGRES_VERSION      ?= 15-alpine
POSTGRES_SUPPORTED    ?= $(POSTGRES_VERSION) 14-alpine 13-alpine 12-alpine

PROJECT               ?= upper_postgres_$(POSTGRES_VERSION)

DB_HOST               ?= 127.0.0.1
DB_PORT               ?= 5432

DB_NAME               ?= upperio
DB_USERNAME           ?= upperio_user
DB_PASSWORD           ?= upperio//s3cr37

TEST_FLAGS            ?=
PARALLEL_FLAGS        ?= --halt-on-error 2 --jobs 1

export POSTGRES_VERSION

export DB_HOST
export DB_NAME
export DB_PASSWORD
export DB_PORT
export DB_USERNAME

export TEST_FLAGS

test:
	go test -v -failfast -race -timeout 20m $(TEST_FLAGS)

test-no-race:
	go test -v -failfast $(TEST_FLAGS)

server-up: server-down
	docker-compose -p $(PROJECT) up -d && \
	sleep 10

server-down:
	docker-compose -p $(PROJECT) down

test-extended:
	parallel $(PARALLEL_FLAGS) \
		"POSTGRES_VERSION={} DB_PORT=\$$((5432+{#})) $(MAKE) server-up test server-down" ::: \
		$(POSTGRES_SUPPORTED)
