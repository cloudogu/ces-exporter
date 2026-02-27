# syntax=docker/dockerfile:1

# Comments are provided throughout this file to help you get started.
# If you need more help, visit the Dockerfile reference guide at
# https://docs.docker.com/go/dockerfile-reference/

# Want to help us make this template better? Share your feedback here: https://forms.gle/ybq9Krt8jtBL3iCk7

################################################################################
# Create a stage for building the application.
ARG GO_VERSION=1.26.0
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION} AS build
WORKDIR /workspace

# Download dependencies as a separate step to take advantage of Docker's caching.
# Leverage a cache mount to /go/pkg/mod/ to speed up subsequent builds.
# Leverage bind mounts to go.sum and go.mod to avoid having to copy them into
# the container.
ENV GOMODCACHE=/go/pkg/mod
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download -x

# This is the architecture you’re building for, which is passed in by the builder.
# Placing it here allows the previous steps to be cached across architectures.
ARG TARGETARCH

# Build the application.
# Leverage a cache mount to /go/pkg/mod/ to speed up subsequent builds.
# Leverage a bind mount to the current directory to avoid having to copy the
# source code into the container.
ENV GOCACHE=/root/.cache/go-build
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=bind,target=. \
    CGO_ENABLED=0 GOARCH=$TARGETARCH go build -o /target/ces-exporter ./

FROM alpine:3.21 AS classic
LABEL maintainer="hello@cloudogu.com" \
      NAME="ces-exporter" \
      VERSION="2.0.0"

ENV MODE=classic

# Install any runtime dependencies that are needed to run your application.
# Leverage a cache mount to /var/cache/apk/ to speed up subsequent builds.
RUN --mount=type=cache,target=/var/cache/apk \
    apk update && \
    apk upgrade && \
    apk add \
        ca-certificates \
        tzdata \
        bash \
        openssh \
        rsync \
        nfs-utils \
        btrfs-progs \
        && \
        update-ca-certificates

# Configure SSH
RUN \
   ssh-keygen -A && \
   sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config && \
   mkdir -p /root/.ssh && \
   chmod -R 700 /root && \
   chmod -R 600 /root/.ssh

COPY resources /

WORKDIR /

COPY --from=build /target/ces-exporter .

EXPOSE 8080

ENTRYPOINT ["/startup.sh"]

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot AS multinode
LABEL maintainer="hello@cloudogu.com" \
      NAME="ces-exporter" \
      VERSION="2.0.0"

ENV MODE=multinode

WORKDIR /
COPY --from=build /target/ces-exporter .

# the linter has a problem with the valid colon-syntax
# dockerfile_lint - ignore
USER 65532:65532

EXPOSE 8080

ENTRYPOINT ["/ces-exporter"]