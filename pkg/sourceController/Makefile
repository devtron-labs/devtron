all: build

TAG?=latest
FLAGS=
ENVVAR=
GOOS?=darwin
REGISTRY?=686244538589.dkr.ecr.us-east-2.amazonaws.com
BASEIMAGE?=alpine:3.9
#BUILD_NUMBER=$$(date +'%Y%m%d-%H%M%S')
#BUILD_NUMBER := $(shell bash -c 'echo $$(date +'%Y%m%d-%H%M%S')')
include $(ENV_FILE)
export

build: clean wire
	$(ENVVAR) GOOS=$(GOOS) go build -o source-controller

wire:
	wire

clean:
	rm -rf source-controller

run: build
	./source-controller

.PHONY: build
docker-build-image:  build
	 docker build -t source-controller:$(TAG) .