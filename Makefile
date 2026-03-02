.PHONY: build clean test install run help deps fmt lint

BINARY_NAME=nixcopy
BINARY_PATH=./bin/$(BINARY_NAME)
MAIN_PATH=./cmd/nixcopy/main.go

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

clean: ## ลบไฟล์ที่ build แล้ว
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	@echo "Clean complete"

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t go-nixcopy:latest .

docker-run: ## รัน Docker container
	docker run --rm -v $(PWD)/config.yaml:/app/config.yaml go-nixcopy:latest

example-sftp-s3: ## รันตัวอย่าง SFTP to S3
	$(BINARY_PATH) transfer -c examples/sftp-to-s3.yaml -s /remote/file.txt -d backup/file.txt

example-blob-ftps: ## รันตัวอย่าง Blob to FTPS
	$(BINARY_PATH) transfer -c examples/blob-to-ftps.yaml -s myfile.pdf -d /upload/myfile.pdf

example-s3-blob: ## รันตัวอย่าง S3 to Blob
	$(BINARY_PATH) transfer -c examples/s3-to-blob.yaml -s data/file.zip -d backups/file.zip

.DEFAULT_GOAL := help
