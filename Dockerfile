# Build the manager binary
FROM golang:1.24 AS builder
ARG TARGETOS
ARG TARGETARCH
# Accept version as a build argument
ARG CI_COMMIT_TAG
ARG DATE

# Set environment variable with the provided version
ENV VERSION=$CI_COMMIT_TAG
ENV DATE=$DATE
WORKDIR /workspace

# Copy the go source
COPY . .


RUN go mod tidy
RUN go mod vendor

# Build

# the GOARCH has not a default value to allow the binary be built according to the host where the command
# was called. For example, if we call make docker-build in a local env which has the Apple Silicon M1 SO
# the docker BUILDPLATFORM arg will be linux/arm64 when for Apple x86 it will be linux/amd64. Therefore,
# by leaving it empty we can ensure that the container and binary shipped on it will have the same platform.
RUN CGO_ENABLED=1 go build -ldflags "-X main.commit=${VERSION} -X main.date=${DATE}" -o /workspace/manager .

# Use distroless as minimal base image to package the manager binary
FROM gcr.io/distroless/base:latest
WORKDIR /app
COPY --from=builder /workspace/manager /app/manager


ENTRYPOINT ["/app/manager"]