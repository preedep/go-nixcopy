# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release of go-nixcopy
- Support for SFTP storage
- Support for FTPS storage
- Support for Azure Blob Storage
- Support for AWS S3 storage
- Streaming file transfer with minimal memory usage
- Progress tracking with real-time updates
- Concurrent file transfer support
- Configurable retry mechanism
- CLI interface with transfer and list commands
- Comprehensive configuration via YAML files
- Structured logging with Zap
- Clean Architecture implementation
- Docker support
- Example configurations for common use cases

### Features
- **Transfer Operations**
  - SFTP ↔ FTPS
  - SFTP ↔ Azure Blob Storage
  - SFTP ↔ AWS S3
  - FTPS ↔ Azure Blob Storage
  - FTPS ↔ AWS S3
  - Azure Blob Storage ↔ AWS S3

- **Performance**
  - Streaming I/O for memory efficiency
  - Configurable buffer sizes
  - Concurrent file transfers
  - Automatic retry on failures

- **Configuration**
  - YAML-based configuration
  - Support for environment variables
  - Flexible timeout settings
  - TLS/SSL support for secure connections

- **Monitoring**
  - Real-time progress tracking
  - Transfer speed calculation
  - ETA estimation
  - Structured JSON logging

## [0.1.0] - 2024-XX-XX

### Added
- Initial project structure
- Core domain entities and interfaces
- Storage implementations (SFTP, FTPS, Blob, S3)
- Transfer use case with streaming support
- CLI commands (transfer, list)
- Configuration management
- Logging infrastructure
- Example configurations
- Comprehensive Thai documentation

[Unreleased]: https://github.com/preedep/go-nixcopy/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/preedep/go-nixcopy/releases/tag/v0.1.0
