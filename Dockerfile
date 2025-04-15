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
COPY *.go .
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

COPY resources /

WORKDIR /

COPY --from=builder /workspace/target/ces-exporter .

RUN apk update && apk upgrade && \
  apk --no-cache add bash openssh rsync nfs-utils && \
  ssh-keygen -A && sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config && \
  mkdir -p /root/.ssh && chmod -R 700 /root && chmod -R 600 /root/.ssh

EXPOSE 8080

ENTRYPOINT ["/startup.sh"]

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