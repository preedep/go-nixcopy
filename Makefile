.PHONY: build clean test install run help deps fmt lint docker-build docker-buildx docker-run integration-test-local integration-test-down

BINARY_NAME=nixcopy
BINARY_PATH=./bin/$(BINARY_NAME)
MAIN_PATH=./cmd/nixcopy/main.go
IMAGE_NAME ?= nickmsft/gonixcopy
VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE  ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

help: ## แสดงความช่วยเหลือ
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

deps: ## ดาวน์โหลด dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

build: ## Build binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	go build -o $(BINARY_PATH) $(MAIN_PATH)
	@echo "Build complete: $(BINARY_PATH)"

build-all: ## Build สำหรับทุก platform
	@echo "Building for all platforms..."
	@mkdir -p bin
	GOOS=linux GOARCH=amd64 go build -o bin/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=linux GOARCH=arm64 go build -o bin/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 go build -o bin/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 go build -o bin/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 go build -o bin/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "Build complete for all platforms"

install: ## ติดตั้งไปยัง $GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	go install $(MAIN_PATH)
	@echo "Installation complete"

run: ## รันโปรแกรม
	go run $(MAIN_PATH)

test: ## รัน unit tests
	go test ./...

test-verbose: ## รัน unit tests พร้อม verbose mode
	go test -v ./...

test-coverage: ## รัน tests พร้อม coverage report
	go test -cover ./...
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test-unit: ## รัน unit tests
	go test -short ./...

test-race: ## รัน tests พร้อม race detector
	go test -race ./...

fmt: ## Format โค้ด
	@echo "Formatting code..."
	go fmt ./...
	gofmt -s -w .

lint: ## รัน linter
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Run: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin" && exit 1)
	golangci-lint run ./...

clean: ## ลบไฟล์ที่ build
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -rf bin/
	@echo "Clean complete"

clean-dist: ## ลบ release artifacts
	@echo "Cleaning dist directory..."
	rm -rf dist/
	@echo "Clean complete"

release: ## Build release version (current platform)
	@echo "Building release version..."
	./build-release.sh current

release-all: ## Build release version (all platforms)
	@echo "Building release for all platforms..."
	./build-release.sh all

release-linux: ## Build release for Linux AMD64
	./build-release.sh linux amd64

release-darwin: ## Build release for macOS ARM64
	./build-release.sh darwin arm64

release-windows: ## Build release for Windows AMD64
	./build-release.sh windows amd64

release-upx: ## Build release + UPX compression, current platform (requires upx)
	@echo "Building release with UPX compression..."
	UPX=1 ./build-release.sh current

release-all-upx: ## Build release + UPX compression, all platforms (requires upx)
	@echo "Building release with UPX compression for all platforms..."
	UPX=1 ./build-release.sh all

goreleaser-check: ## Validate .goreleaser.yaml config (requires goreleaser)
	goreleaser check

goreleaser-snapshot: ## Build snapshot locally via GoReleaser — no publish, no tag required (requires goreleaser)
	goreleaser release --snapshot --clean

docker-build: ## Build Docker image for current platform with OCI labels
	@echo "Building Docker image $(IMAGE_NAME):$(VERSION)..."
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-t $(IMAGE_NAME):$(VERSION) \
		-t $(IMAGE_NAME):latest \
		.

docker-buildx: ## Build and push multi-arch image (linux/amd64 + linux/arm64) — requires buildx and a registry
	@echo "Building multi-arch image $(IMAGE_NAME):$(VERSION)..."
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-t $(IMAGE_NAME):$(VERSION) \
		-t $(IMAGE_NAME):latest \
		--push \
		.

docker-run: ## Run Docker container using NIXCOPY_* env vars (no config file needed)
	docker run --rm \
		-e NIXCOPY_SOURCE_TYPE \
		-e NIXCOPY_DEST_TYPE \
		$(IMAGE_NAME):latest

docker-run-config: ## Run Docker container mounting a local config.yaml
	docker run --rm -v $(PWD)/config.yaml:/config.yaml $(IMAGE_NAME):latest transfer --config /config.yaml

COMPOSE_INT = docker-compose.integration.yml

integration-test-local: ## Spin up SFTP + FTPS + MinIO via Compose and run all integration tests locally
	docker compose -f $(COMPOSE_INT) up -d --wait --timeout 60
	docker compose -f $(COMPOSE_INT) wait minio-setup || true
	@S3_ENDPOINT=http://localhost:9000 S3_ACCESS_KEY=minioadmin S3_SECRET_KEY=minioadmin \
	 S3_BUCKET=test-bucket S3_REGION=us-east-1 \
	 SFTP_HOST=localhost SFTP_PORT=2222 SFTP_USERNAME=testuser SFTP_PASSWORD=testpass \
	 FTPS_HOST=localhost FTPS_PORT=21 FTPS_USERNAME=testuser FTPS_PASSWORD=testpass \
	 FTPS_TLS_MODE=explicit FTPS_SKIP_VERIFY=true \
	 go test -tags=integration -race -count=1 -v ./internal/infrastructure/storage/...; \
	 EXIT=$$?; \
	 docker compose -f $(COMPOSE_INT) down -v; \
	 exit $$EXIT

integration-test-down: ## Stop and remove integration test containers and volumes
	docker compose -f $(COMPOSE_INT) down -v

example-sftp-s3: ## รันตัวอย่าง SFTP to S3
	$(BINARY_PATH) transfer -c examples/sftp-to-s3.yaml -s /remote/file.txt -d backup/file.txt

example-blob-ftps: ## รันตัวอย่าง Blob to FTPS
	$(BINARY_PATH) transfer -c examples/blob-to-ftps.yaml -s myfile.pdf -d /upload/myfile.pdf

example-s3-blob: ## รันตัวอย่าง S3 to Blob
	$(BINARY_PATH) transfer -c examples/s3-to-blob.yaml -s data/file.zip -d backups/file.zip

.DEFAULT_GOAL := help
