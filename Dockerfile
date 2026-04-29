# Build stage — always runs on the host machine's native arch for fast compilation
FROM --platform=$BUILDPLATFORM golang:1.24-alpine3.21 AS builder

ARG TARGETOS=linux
ARG TARGETARCH=amd64

WORKDIR /build

RUN apk add --no-cache git make

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -a -trimpath -ldflags="-s -w" -o nixcopy ./cmd/nixcopy/main.go

# Runtime stage — distroless/static-debian12: no shell, no package manager, no apk CVEs.
# Includes CA certificates and tzdata. Runs as nonroot (UID 65532) by default.
FROM --platform=$TARGETPLATFORM gcr.io/distroless/static-debian12:nonroot

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

COPY --from=builder /build/nixcopy /nixcopy

ENTRYPOINT ["/nixcopy"]
CMD ["--help"]
