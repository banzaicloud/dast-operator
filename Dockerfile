FROM golang:1.14 as builder

WORKDIR /workspace
# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/
COPY webhooks/ webhooks/
COPY pkg/ pkg/
# Copy the Go Modules manifests
COPY go.mod go.mod
Copy go.sum go.sum

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/manager .
USER nonroot:nonroot

ENTRYPOINT ["/manager"]
