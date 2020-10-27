SHELL:=/bin/bash
.PHONY: test format

format:
	golint .
	gofmt -w -s .

test:
	go vet .
