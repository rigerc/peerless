# Build stage
FROM --platform=$BUILDPLATFORM golang:1.22-alpine AS builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} GOARM=${TARGETVARIANT#v} \
    go build \
    -ldflags='-w -s -X main.version={{.Version}} -X main.commit={{.Commit}} -X main.date={{.Date}}' \
    -a -installsuffix cgo \
    -o peerless main.go

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/peerless .

# Copy documentation
COPY --from=builder /app/README.md .
COPY --from=builder /app/LICENSE .

# Add non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup
USER appuser

ENTRYPOINT ["./peerless"]
CMD ["--help"]