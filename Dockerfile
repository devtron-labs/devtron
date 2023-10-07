# Devtron binary build stage.
FROM golang:1.20-bullseye as build-env

# Install the required dependencies and wire go binary
# for compile time dependency injection.
RUN echo $GOPATH && \
    apt update && \
    apt install git gcc musl-dev make -y && \
    go install github.com/google/wire/cmd/wire@latest

# Copy the project files into the image and
# build the Devtron Go binary.
WORKDIR /go/src/github.com/devtron-labs/devtron
COPY . /go/src/github.com/devtron-labs/devtron/
RUN GOOS=linux make build-all

# Final stage consisting of the devtron binary and
# other required artifacts
FROM ubuntu as devtron-all

# Install the required dependencies for the final stage.
RUN apt update && \
    apt install ca-certificates git curl -y && \
    apt clean autoclean && \
    apt autoremove -y && rm -rf /var/lib/apt/lists/*

# Copy the Devtron binary from the build stage alongwith auth_model.conf in the current working directory.
COPY --from=build-env \
    /go/src/github.com/devtron-labs/devtron/devtron \
	/go/src/github.com/devtron-labs/devtron/auth_model.conf ./

# Copy ArgoCD assets into the docker image.
COPY --from=build-env /go/src/github.com/devtron-labs/devtron/vendor/github.com/argoproj/argo-cd/assets/ /go/src/github.com/devtron-labs/devtron/vendor/github.com/argoproj/argo-cd/assets

# Copy other required scripts into the docker image in the "scripts" directory.
COPY --from=build-env /go/src/github.com/devtron-labs/devtron/scripts/devtron-reference-helm-charts \
    /go/src/github.com/devtron-labs/devtron/scripts/sql \
    /go/src/github.com/devtron-labs/devtron/scripts/casbin \
    /go/src/github.com/devtron-labs/devtron/scripts/argo-assets scripts/

# Copy git-ask-pass.sh to the image and make it executable.
COPY ./git-ask-pass.sh /git-ask-pass.sh
RUN chmod +x /git-ask-pass.sh

# Configuring the user and configure its access to the required files.
RUN useradd -ms /bin/bash devtron && \
    chown -R devtron:devtron ./devtron ./git-ask-pass.sh ./auth_model.conf ./scripts

# Configure the user.
USER devtron

# Specify the command to execute the devtron binary in the container.
CMD ["./devtron"]

#FROM alpine:3.15.0 as  devtron-ea

#RUN apk add --no-cache ca-certificates
#COPY --from=build-env  /go/src/github.com/devtron-labs/devtron/auth_model.conf .
#COPY --from=build-env  /go/src/github.com/devtron-labs/devtron/cmd/external-app/devtron-ea .

#COPY --from=build-env  /go/src/github.com/devtron-labs/devtron/vendor/github.com/argoproj/argo-cd/assets/ /go/src/github.com/devtron-labs/devtron/vendor/github.com/argoproj/argo-cd/assets
#COPY --from=build-env  /go/src/github.com/devtron-labs/devtron/scripts/devtron-reference-helm-charts scripts/devtron-reference-helm-charts
#COPY --from=build-env  /go/src/github.com/devtron-labs/devtron/scripts/argo-assets/APPLICATION_TEMPLATE.JSON scripts/argo-assets/APPLICATION_TEMPLATE.JSON

#CMD ["./devtron-ea"]
