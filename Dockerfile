FROM golang:1.21 AS build-env

RUN echo $GOPATH
RUN apt update
RUN apt install git gcc musl-dev make -y
RUN go install github.com/google/wire/cmd/wire@latest
WORKDIR /go/src/github.com/devtron-labs/devtron
ADD . /go/src/github.com/devtron-labs/devtron/
ADD ./vendor/github.com/Microsoft/ /go/src/github.com/devtron-labs/devtron/vendor/github.com/microsoft/
RUN GOOS=linux make build-all

# uncomment this post build arg
FROM ubuntu:22.04@sha256:1b8d8ff4777f36f19bfe73ee4df61e3a0b789caeff29caa019539ec7c9a57f95 as  devtron-all

RUN apt update
RUN apt install ca-certificates git curl -y
RUN apt clean autoclean
RUN apt autoremove -y && rm -rf /var/lib/apt/lists/*
COPY --from=build-env  /go/src/github.com/devtron-labs/devtron/devtron .
COPY --from=build-env  /go/src/github.com/devtron-labs/devtron/auth_model.conf .
#COPY --from=build-env  /go/src/github.com/devtron-labs/devtron/vendor/github.com/argoproj/argo-cd/assets/ /go/src/github.com/devtron-labs/devtron/vendor/github.com/argoproj/argo-cd/assets
COPY --from=build-env  /go/src/github.com/devtron-labs/devtron/argocd-assets/ /go/src/github.com/devtron-labs/devtron/vendor/github.com/argoproj/argo-cd/assets
COPY --from=build-env  /go/src/github.com/devtron-labs/devtron/scripts/devtron-reference-helm-charts scripts/devtron-reference-helm-charts
COPY --from=build-env  /go/src/github.com/devtron-labs/devtron/scripts/sql scripts/sql
COPY --from=build-env  /go/src/github.com/devtron-labs/devtron/scripts/casbin scripts/casbin
COPY --from=build-env  /go/src/github.com/devtron-labs/devtron/scripts/argo-assets/APPLICATION_TEMPLATE.tmpl scripts/argo-assets/APPLICATION_TEMPLATE.tmpl

COPY ./git-ask-pass.sh /git-ask-pass.sh
RUN chmod +x /git-ask-pass.sh

RUN useradd -ms /bin/bash devtron
RUN chown -R devtron:devtron ./devtron
RUN chown -R devtron:devtron ./git-ask-pass.sh
RUN chown -R devtron:devtron ./auth_model.conf 
RUN chown -R devtron:devtron ./scripts

USER devtron

CMD ["./devtron"]


#FROM alpine:3.15.0 as  devtron-ea

#RUN apk add --no-cache ca-certificates
#COPY --from=build-env  /go/src/github.com/devtron-labs/devtron/auth_model.conf .
#COPY --from=build-env  /go/src/github.com/devtron-labs/devtron/cmd/external-app/devtron-ea .

#COPY --from=build-env  /go/src/github.com/devtron-labs/devtron/vendor/github.com/argoproj/argo-cd/assets/ /go/src/github.com/devtron-labs/devtron/vendor/github.com/argoproj/argo-cd/assets
#COPY --from=build-env  /go/src/github.com/devtron-labs/devtron/scripts/devtron-reference-helm-charts scripts/devtron-reference-helm-charts
#COPY --from=build-env  /go/src/github.com/devtron-labs/devtron/scripts/argo-assets/APPLICATION_TEMPLATE.JSON scripts/argo-assets/APPLICATION_TEMPLATE.JSON

#CMD ["./devtron-ea"]
