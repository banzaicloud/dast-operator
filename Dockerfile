FROM golang:1.18.3-alpine AS builder

WORKDIR /workspace
# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/
COPY webhooks/ webhooks/
COPY pkg/ pkg/
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM alpine:3.16.0
WORKDIR /
COPY --from=builder /workspace/manager .
USER 65534:65534

ENTRYPOINT ["/manager"]
