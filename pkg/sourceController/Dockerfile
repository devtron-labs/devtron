FROM golang:1.20-alpine3.18  AS build-env

RUN echo $GOPATH

RUN apk add --no-cache git gcc musl-dev
RUN apk add --update make
WORKDIR /go/src/github.com/devtron-labs/devtron/source-controller
ADD . /go/src/github.com/devtron-labs/devtron/source-controller
RUN GOOS=linux make

FROM alpine:3.18

RUN apk add --update ca-certificates

RUN adduser -D devtron

COPY --from=build-env  /go/src/github.com/devtron-labs/devtron/source-controller .

RUN chown devtron:devtron ./devtron/source-controller

RUN chmod +x .devtron/source-controller

USER devtron

ENTRYPOINT [".devtron/source-controller"]