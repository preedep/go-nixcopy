# Build stage — always runs on the host machine's native arch for fast compilation
FROM --platform=$BUILDPLATFORM golang:1.21-alpine3.21 AS builder

ARG TARGETOS=linux
ARG TARGETARCH=amd64

WORKDIR /build

RUN apk add --no-cache git make

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -a -installsuffix cgo -o nixcopy ./cmd/nixcopy/main.go

# Runtime stage
FROM --platform=$TARGETPLATFORM alpine:3.21

ARG VERSION=dev
ARG GIT_COMMIT=unknown
ARG BUILD_DATE=unknown

LABEL org.opencontainers.image.title="go-nixcopy" \
      org.opencontainers.image.description="Fast universal file transfer CLI" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${GIT_COMMIT}" \
      org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.source="https://github.com/preedep/go-nixcopy" \
      org.opencontainers.image.licenses="MIT"

RUN apk --no-cache add ca-certificates && \
    addgroup -S nixcopy && \
    adduser -S -G nixcopy nixcopy

WORKDIR /app

COPY --from=builder /build/nixcopy .
COPY config.example.yaml .
COPY examples/ ./examples/

RUN chown -R nixcopy:nixcopy /app

VOLUME ["/app/config"]

USER nixcopy

ENTRYPOINT ["./nixcopy"]
CMD ["--help"]
