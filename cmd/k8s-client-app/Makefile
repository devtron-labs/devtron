all: build

TAG?=$(shell bash -c 'git log --pretty=format:'%h' -n 1')
FLAGS=
ENVVAR=
GOOS?=darwin
GOARCH?=arm64
OUTPUT_NAME?=devtron-k8s-client-$(GOOS)-$(GOARCH)
REGISTRY?=686244538589.dkr.ecr.us-east-2.amazonaws.com
BASEIMAGE?=alpine:3.9
#BUILD_NUMBER=$$(date +'%Y%m%d-%H%M%S')
BUILD_NUMBER := $(shell bash -c 'echo $$(date +'%Y%m%d-%H%M%S')')
GIT_COMMIT =$(shell sh -c 'git log --pretty=format:'%h' -n 1')
BUILD_TIME= $(shell sh -c 'date -u '+%Y-%m-%dT%H:%M:%SZ'')
SERVER_MODE_FULL= FULL
SERVER_MODE_EA_ONLY=EA_ONLY
export

build: clean wire
	$(ENVVAR) GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(OUTPUT_NAME) \
			-ldflags="-X 'github.com/devtron-labs/devtron/util.GitCommit=${GIT_COMMIT}' \
			-X 'github.com/devtron-labs/devtron/util.BuildTime=${BUILD_TIME}' \
			-X 'github.com/devtron-labs/devtron/util.ServerMode=${SERVER_MODE_EA_ONLY}' \
			-X 'github.com/devtron-labs/devtron/util.IsDesktopApp=true'"

wire:
	wire

clean:
	rm -f $(OUTPUT_NAME)

run: build
	./$(OUTPUT_NAME)

.PHONY: build
docker-build-image:  build
	 docker build -t devtron:$(TAG) .

.PHONY: build, all, wire, clean, run, set-docker-build-env, docker-build-push, devtron,
docker-build-push: docker-build-image
	docker tag devtron:${TAG}  ${REGISTRY}/devtron:${TAG}
	docker push ${REGISTRY}/devtron:${TAG}