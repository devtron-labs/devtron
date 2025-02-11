SHELL					        ?= bash

MYSQL_VERSION         ?= 8.1
MYSQL_SUPPORTED       ?= $(MYSQL_VERSION) 5.7
PROJECT               ?= upper_mysql_$(MYSQL_VERSION)

DB_HOST               ?= 127.0.0.1
DB_PORT               ?= 3306

DB_NAME               ?= upperio
DB_USERNAME           ?= upperio_user
DB_PASSWORD           ?= upperio//s3cr37

TEST_FLAGS            ?=
PARALLEL_FLAGS        ?= --halt-on-error 2 --jobs 1

export MYSQL_VERSION

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
	sleep 15

server-down:
	docker-compose -p $(PROJECT) down

test-extended:
	parallel $(PARALLEL_FLAGS) \
		"MYSQL_VERSION={} DB_PORT=\$$((3306+{#})) $(MAKE) server-up test server-down" ::: \
		$(MYSQL_SUPPORTED)
