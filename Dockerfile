FROM golang:1.18-alpine3.14  AS build-env

RUN echo $GOPATH

RUN apk add --no-cache git gcc musl-dev
RUN apk add --update make
RUN go install github.com/google/wire/cmd/wire@latest
WORKDIR /go/src/github.com/devtron-labs/devtron
ADD . /go/src/github.com/devtron-labs/devtron/
RUN GOOS=linux make build-all

# uncomment this post build arg
FROM alpine:3.15.0 as  devtron-all
RUN apk add --no-cache ca-certificates
RUN apk update
RUN apk add git
RUN apk add curl
COPY --from=build-env  /go/src/github.com/devtron-labs/devtron/devtron .
COPY --from=build-env  /go/src/github.com/devtron-labs/devtron/auth_model.conf .
COPY --from=build-env  /go/src/github.com/devtron-labs/devtron/vendor/github.com/argoproj/argo-cd/assets/ /go/src/github.com/devtron-labs/devtron/vendor/github.com/argoproj/argo-cd/assets
COPY --from=build-env  /go/src/github.com/devtron-labs/devtron/scripts/devtron-reference-helm-charts scripts/devtron-reference-helm-charts
COPY --from=build-env  /go/src/github.com/devtron-labs/devtron/scripts/argo-assets/APPLICATION_TEMPLATE.JSON scripts/argo-assets/APPLICATION_TEMPLATE.JSON

COPY ./git-ask-pass.sh /git-ask-pass.sh
RUN chmod +x /git-ask-pass.sh

CMD ["./devtron"]


#FROM alpine:3.15.0 as  devtron-ea

#RUN apk add --no-cache ca-certificates
#COPY --from=build-env  /go/src/github.com/devtron-labs/devtron/auth_model.conf .
#COPY --from=build-env  /go/src/github.com/devtron-labs/devtron/cmd/external-app/devtron-ea .

#COPY --from=build-env  /go/src/github.com/devtron-labs/devtron/vendor/github.com/argoproj/argo-cd/assets/ /go/src/github.com/devtron-labs/devtron/vendor/github.com/argoproj/argo-cd/assets
#COPY --from=build-env  /go/src/github.com/devtron-labs/devtron/scripts/devtron-reference-helm-charts scripts/devtron-reference-helm-charts
#COPY --from=build-env  /go/src/github.com/devtron-labs/devtron/scripts/argo-assets/APPLICATION_TEMPLATE.JSON scripts/argo-assets/APPLICATION_TEMPLATE.JSON

#CMD ["./devtron-ea"]
