HADOOP_COMMON_PROTOS = $(shell find protocol/hadoop_common -name '*.proto')
HADOOP_HDFS_PROTOS = $(shell find protocol/hadoop_hdfs -name '*.proto')
GENERATED_PROTOS = $(shell echo "$(HADOOP_HDFS_PROTOS) $(HADOOP_COMMON_PROTOS)" | sed 's/\.proto/\.pb\.go/g')
SOURCES = $(shell find . -name '*.go') $(GENERATED_PROTOS)

# Protobuf needs one of these for every 'import "foo.proto"' in .protoc files.
PROTO_MAPPING = MSecurity.proto=github.com/colinmarc/hdfs/protocol/hadoop_common

TRAVIS_TAG ?= $(shell git rev-parse HEAD)
ARCH = $(shell go env GOOS)-$(shell go env GOARCH)
RELEASE_NAME = gohdfs-$(TRAVIS_TAG)-$(ARCH)

all: hdfs

%.pb.go: $(HADOOP_HDFS_PROTOS) $(HADOOP_COMMON_PROTOS)
	protoc --go_out='$(PROTO_MAPPING):protocol/hadoop_common' -Iprotocol/hadoop_common -Iprotocol/hadoop_hdfs $(HADOOP_COMMON_PROTOS)
	protoc --go_out='$(PROTO_MAPPING):protocol/hadoop_hdfs' -Iprotocol/hadoop_common -Iprotocol/hadoop_hdfs $(HADOOP_HDFS_PROTOS)

clean-protos:
	find . -name *.pb.go | xargs rm

hdfs: clean $(SOURCES)
	go build -ldflags "-X main.version=$(TRAVIS_TAG)" ./cmd/hdfs

test: hdfs
	go test -v -race ./...
	bats ./cmd/hdfs/test/*.bats

clean:
	rm -f ./hdfs
	rm -rf gohdfs-*

release: hdfs
	mkdir -p $(RELEASE_NAME)
	cp hdfs README.md LICENSE.txt cmd/hdfs/bash_completion $(RELEASE_NAME)/
	tar -cvzf $(RELEASE_NAME).tar.gz $(RELEASE_NAME)

.PHONY: clean clean-protos install test release
