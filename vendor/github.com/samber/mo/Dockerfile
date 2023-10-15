
FROM golang:1.18-bullseye

WORKDIR /go/src/github.com/samber/mo

COPY Makefile go.* ./

RUN make tools
