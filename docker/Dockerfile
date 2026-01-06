FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git make gcc musl-dev

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the node binary
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo \
    -ldflags="-w -s" \
    -o /bin/podoru-node ./cmd/node

# Build the keygen tool
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo \
    -ldflags="-w -s" \
    -o /bin/keygen ./cmd/tools/keygen

FROM alpine:latest

RUN apk --no-cache add ca-certificates

# Copy binaries from builder
COPY --from=builder /bin/podoru-node /bin/podoru-node
COPY --from=builder /bin/keygen /bin/keygen

# Create data directory
RUN mkdir -p /data

VOLUME ["/data"]

EXPOSE 8545 9000

ENTRYPOINT ["/bin/podoru-node"]
CMD ["--config", "/data/config.yaml"]
