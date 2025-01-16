FROM golang:1.21 AS build-env

RUN echo $GOPATH && \
    apt update && \
    apt install git gcc musl-dev make -y && \
    go install github.com/google/wire/cmd/wire@latest
    
WORKDIR /go/src/github.com/devtron-labs/devtron

ADD . /go/src/github.com/devtron-labs/devtron/

ADD ./vendor/github.com/Microsoft/ /go/src/github.com/devtron-labs/devtron/vendor/github.com/microsoft/

RUN GOOS=linux make build

# uncomment this post build arg
FROM ubuntu:22.04@sha256:1b8d8ff4777f36f19bfe73ee4df61e3a0b789caeff29caa019539ec7c9a57f95 as  devtron-all

RUN apt update && \
    apt install ca-certificates git curl -y && \
    apt clean autoclean && \
    apt autoremove -y && \
    rm -rf /var/lib/apt/lists/* && \
    useradd -ms /bin/bash devtron

COPY  --chown=devtron:devtron --from=build-env  /go/src/github.com/devtron-labs/devtron/devtron .

COPY  --chown=devtron:devtron --from=build-env  /go/src/github.com/devtron-labs/devtron/auth_model.conf .

COPY --chown=devtron:devtron --from=build-env  /go/src/github.com/devtron-labs/devtron/argocd-assets/ /go/src/github.com/devtron-labs/devtron/vendor/github.com/argoproj/argo-cd/assets

COPY --chown=devtron:devtron --from=build-env  /go/src/github.com/devtron-labs/devtron/scripts/devtron-reference-helm-charts scripts/devtron-reference-helm-charts

COPY --chown=devtron:devtron --from=build-env  /go/src/github.com/devtron-labs/devtron/scripts/sql scripts/sql

COPY --chown=devtron:devtron --from=build-env  /go/src/github.com/devtron-labs/devtron/scripts/casbin scripts/casbin

COPY --chown=devtron:devtron --from=build-env  /go/src/github.com/devtron-labs/devtron/scripts/argo-assets/APPLICATION_TEMPLATE.tmpl scripts/argo-assets/APPLICATION_TEMPLATE.tmpl

COPY  --chown=devtron:devtron ./git-ask-pass.sh /git-ask-pass.sh

RUN chmod +x /git-ask-pass.sh 

USER devtron

CMD ["./devtron"]