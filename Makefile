#!make

all: build

TAG?=$(shell bash -c 'git log --pretty=format:'%h' -n 1')
FLAGS=
ENVVAR=
GOOS?=darwin
REGISTRY?=686244538589.dkr.ecr.us-east-2.amazonaws.com
BASEIMAGE?=alpine:3.9
#BUILD_NUMBER=$$(date +'%Y%m%d-%H%M%S')
BUILD_NUMBER := $(shell bash -c 'echo $$(date +'%Y%m%d-%H%M%S')')
ENV_FILE?=scripts/dev-conf/envfile.env
GIT_COMMIT =$(shell sh -c 'git log --pretty=format:'%h' -n 1')
BUILD_TIME= $(shell sh -c 'date -u '+%Y-%m-%dT%H:%M:%SZ'')
SERVER_MODE_FULL= FULL
SERVER_MODE_EA_ONLY=EA_ONLY
#TEST_BRANCH=PUT_YOUR_BRANCH_HERE
#LATEST_HASH=PUT_YOUR_HASH_HERE
GOFLAGS:= $(GOFLAGS) -buildvcs=false
include $(ENV_FILE)
export

build: clean wire
	$(ENVVAR) GOOS=$(GOOS) go build -o devtron \
			-ldflags="-X 'github.com/devtron-labs/devtron/util.GitCommit=${GIT_COMMIT}' \
			-X 'github.com/devtron-labs/devtron/util.BuildTime=${BUILD_TIME}' \
			-X 'github.com/devtron-labs/devtron/util.ServerMode=${SERVER_MODE_FULL}'"

wire:
	wire

clean:
	rm -rf devtron

test-all: test-unit
	echo 'test cases ran successfully'

test-unit:
	go test ./pkg/pipeline

test-integration:
	export INTEGRATION_TEST_ENV_ID=$(docker run --env TEST_BRANCH=$TEST_BRANCH --env LATEST_HASH=$LATEST_HASH --privileged -d --name dind-test -v $PWD/tests/integrationTesting/:/tmp/ docker:dind)
	docker exec ${INTEGRATION_TEST_ENV_ID} sh /tmp/create-test-env.sh
	docker exec ${INTEGRATION_TEST_ENV_ID} sh /tests/integrationTesting/run-integration-test.sh

run: build
	./devtron

.PHONY: build
docker-build-image:  build
	 docker build -t devtron:$(TAG) .

.PHONY: build, all, wire, clean, run, set-docker-build-env, docker-build-push, devtron,
docker-build-push: docker-build-image
	docker tag devtron:${TAG}  ${REGISTRY}/devtron:${TAG}
	docker push ${REGISTRY}/devtron:${TAG}

#############################################################################

build-all: build
	make --directory ./cmd/external-app build

build-ea:
	make --directory ./cmd/external-app build
