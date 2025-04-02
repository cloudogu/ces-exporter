# Build the manager binary
FROM golang:1.24.1-alpine AS builder

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY core core
COPY configuration configuration
COPY export export
COPY maintenance maintenance
COPY systeminfo systeminfo

# Build
RUN go mod vendor
RUN go build -mod=vendor -o target/ces-exporter

FROM alpine:3.21 AS classic
LABEL maintainer="hello@cloudogu.com" \
      NAME="ces-exporter" \
      VERSION="0.0.1"

ENV MODE=classic

WORKDIR /

COPY --from=builder /workspace/target/ces-exporter .

RUN apk update && apk upgrade && apk add --no-cache bash openssh-server rsync

EXPOSE 8080

ENTRYPOINT ["/ces-exporter"]

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot AS multinode
LABEL maintainer="hello@cloudogu.com" \
      NAME="ces-exporter" \
      VERSION="0.0.1"

ENV MODE=multinode

WORKDIR /
COPY --from=builder /workspace/target/ces-exporter .

# the linter has a problem with the valid colon-syntax
# dockerfile_lint - ignore
USER 65532:65532

EXPOSE 8080

ENTRYPOINT ["/ces-exporter"]