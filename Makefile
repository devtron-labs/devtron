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
	docker run --env-file=wireNilChecker.env  --privileged -d --name dind-test -v $(PWD)/:/wirenil/:ro -v $(PWD)/temp/:/tempfile docker:dind
	docker exec dind-test sh -c "mkdir test && cp -r wirenil/* test/"
	docker exec dind-test sh -c "cd test && ./tests/integrationTesting/create-test-env.sh"
	docker exec dind-test sh -c "cd test && ./tests/integrationTesting/run-integration-test.sh"
	docker exec dind-test sh -c "cd test && touch output.env"
	docker exec dind-test sh -c 'NODE_IP_ADDRESS=$$(kubectl get node  --no-headers  -o custom-columns=INTERNAL-IP:status.addresses[0].address) PG_ADDR=$$NODE_IP_ADDRESS NATS_SERVER_HOST=nats://$$NODE_IP_ADDRESS:30236 sh -c "cd test && go run ."'
	docker exec dind-test sh -c "cp ./test/output.env ./tempfile"
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
