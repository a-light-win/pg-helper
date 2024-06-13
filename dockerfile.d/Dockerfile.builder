ARG GO_VERSION=1.22.4
FROM docker.io/library/golang:${GO_VERSION}

# Install just
COPY dockerfile.d/scripts/install-just.sh /usr/local/bin/install-just.sh
ARG JUST_VERSION=1.28.0
RUN bash /usr/local/bin/install-just.sh ${JUST_VERSION}

# Setup build environment
COPY Justfile /setup/Justfile
COPY just.d/ /setup/just.d/
COPY go.mod /setup/go.mod

WORKDIR /setup
ENV SUDO=''
ENV GOBIN=/usr/local/bin
ENV GOCACHE=/go/cache
RUN just --unstable setup build-env && rm -rf /go

WORKDIR /pg-helper
