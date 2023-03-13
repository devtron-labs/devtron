FROM golang:1.18  AS build-env

RUN echo $GOPATH
RUN apt update
RUN apt install git gcc musl-dev make -y
RUN go install github.com/google/wire/cmd/wire@latest
WORKDIR /go/src/github.com/devtron-labs/devtron

ADD \.github/  /go/src/github.com/devtron-labs/devtron/\.github/
ADD \.git/  /go/src/github.com/devtron-labs/devtron/\.git/
ADD api/  /go/src/github.com/devtron-labs/devtron/api/
ADD assets/  /go/src/github.com/devtron-labs/devtron/assets/
ADD charts/  /go/src/github.com/devtron-labs/devtron/charts/
ADD client/  /go/src/github.com/devtron-labs/devtron/client/
ADD cmd/  /go/src/github.com/devtron-labs/devtron/cmd/
ADD contrib-chart/  /go/src/github.com/devtron-labs/devtron/contrib-chart/
ADD internal/  /go/src/github.com/devtron-labs/devtron/internal/
#ADD manifests/  /go/src/github.com/devtron-labs/devtron/manifests/
ADD otel/  /go/src/github.com/devtron-labs/devtron/otel/
ADD pkg/  /go/src/github.com/devtron-labs/devtron/pkg/
ADD scripts/  /go/src/github.com/devtron-labs/devtron/scripts/
ADD tests/  /go/src/github.com/devtron-labs/devtron/tests/
ADD util/  /go/src/github.com/devtron-labs/devtron/util/
ADD vendor/  /go/src/github.com/devtron-labs/devtron/vendor/


ADD  .deepsource.toml .gitattributes .gitbook.yaml .gitignore auth_model.conf git-ask-pass.sh go.mod go.sum main.go Wire.go wire_gen.go App.go authWire.go Makefile /go/src/github.com/devtron-labs/devtron/
# ADD . /go/src/github.com/devtron-labs/devtron/
# RUN  apt install tree
# RUN echo $(tree -L 1 -a /go/src/github.com/devtron-labs/devtron)
RUN GOOS=linux make build-all

# uncomment this post build arg
FROM ubuntu as  devtron-all

RUN apt update
RUN apt install ca-certificates git curl -y
RUN apt clean autoclean
RUN apt autoremove -y && rm -rf /var/lib/apt/lists/*
COPY --from=build-env  /go/src/github.com/devtron-labs/devtron/devtron .
COPY --from=build-env  /go/src/github.com/devtron-labs/devtron/auth_model.conf .
COPY --from=build-env  /go/src/github.com/devtron-labs/devtron/vendor/github.com/argoproj/argo-cd/assets/ /go/src/github.com/devtron-labs/devtron/vendor/github.com/argoproj/argo-cd/assets
COPY --from=build-env  /go/src/github.com/devtron-labs/devtron/scripts/devtron-reference-helm-charts scripts/devtron-reference-helm-charts
COPY --from=build-env  /go/src/github.com/devtron-labs/devtron/scripts/argo-assets/APPLICATION_TEMPLATE.JSON scripts/argo-assets/APPLICATION_TEMPLATE.JSON

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