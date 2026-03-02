package cli

import (
	"fmt"
	"time"

	"github.com/preedep/go-nixcopy/internal/infrastructure/config"
)

func applyCliFlags(cfg *config.Config) error {
	// Apply source flags
	if sourceType != "" {
		cfg.Source.Type = config.StorageType(sourceType)
	}

	switch cfg.Source.Type {
	case config.StorageTypeSFTP:
		if cfg.Source.SFTP == nil {
			cfg.Source.SFTP = &config.SFTPConfig{}
		}
		if sourceHost != "" {
			cfg.Source.SFTP.Host = sourceHost
		}
		if sourcePort != 0 {
			cfg.Source.SFTP.Port = sourcePort
		}
		if sourceUsername != "" {
			cfg.Source.SFTP.Username = sourceUsername
		}
		if sourcePassword != "" {
			cfg.Source.SFTP.Password = sourcePassword
		}
		if sourcePrivateKey != "" {
			cfg.Source.SFTP.PrivateKeyPath = sourcePrivateKey
		}
		if cfg.Source.SFTP.Timeout == 0 {
			cfg.Source.SFTP.Timeout = 30 * time.Second
		}
		if cfg.Source.SFTP.MaxPacketSize == 0 {
			cfg.Source.SFTP.MaxPacketSize = 32768
		}

	case config.StorageTypeFTPS:
		if cfg.Source.FTPS == nil {
			cfg.Source.FTPS = &config.FTPSConfig{}
		}
		if sourceHost != "" {
			cfg.Source.FTPS.Host = sourceHost
		}
		if sourcePort != 0 {
			cfg.Source.FTPS.Port = sourcePort
		}
		if sourceUsername != "" {
			cfg.Source.FTPS.Username = sourceUsername
		}
		if sourcePassword != "" {
			cfg.Source.FTPS.Password = sourcePassword
		}
		if cfg.Source.FTPS.Timeout == 0 {
			cfg.Source.FTPS.Timeout = 30 * time.Second
		}

	case config.StorageTypeS3:
		if cfg.Source.S3 == nil {
			cfg.Source.S3 = &config.S3Config{}
		}
		if sourceRegion != "" {
			cfg.Source.S3.Region = sourceRegion
		}
		if sourceBucket != "" {
			cfg.Source.S3.Bucket = sourceBucket
		}
		if sourceAuthType != "" {
			cfg.Source.S3.AuthType = config.S3AuthType(sourceAuthType)
		}
		if sourceAccessKey != "" {
			cfg.Source.S3.AccessKeyID = sourceAccessKey
		}
		if sourceSecretKey != "" {
			cfg.Source.S3.SecretAccessKey = sourceSecretKey
		}
		if cfg.Source.S3.AuthType == "" {
			cfg.Source.S3.AuthType = config.S3AuthAccessKey
		}

	case config.StorageTypeBlobStorage:
		if cfg.Source.BlobStorage == nil {
			cfg.Source.BlobStorage = &config.BlobConfig{}
		}
		if sourceAccountName != "" {
			cfg.Source.BlobStorage.AccountName = sourceAccountName
		}
		if sourceContainer != "" {
			cfg.Source.BlobStorage.ContainerName = sourceContainer
		}
		if sourceAuthType != "" {
			cfg.Source.BlobStorage.AuthType = config.BlobAuthType(sourceAuthType)
		}
		if sourceAccountKey != "" {
			cfg.Source.BlobStorage.AccountKey = sourceAccountKey
		}
		if cfg.Source.BlobStorage.AuthType == "" {
			cfg.Source.BlobStorage.AuthType = config.BlobAuthSharedKey
		}
	}

	// Apply destination flags
	if destType != "" {
		cfg.Destination.Type = config.StorageType(destType)
	}

	switch cfg.Destination.Type {
	case config.StorageTypeSFTP:
		if cfg.Destination.SFTP == nil {
			cfg.Destination.SFTP = &config.SFTPConfig{}
		}
		if destHost != "" {
			cfg.Destination.SFTP.Host = destHost
		}
		if destPort != 0 {
			cfg.Destination.SFTP.Port = destPort
		}
		if destUsername != "" {
			cfg.Destination.SFTP.Username = destUsername
		}
		if destPassword != "" {
			cfg.Destination.SFTP.Password = destPassword
		}
		if destPrivateKey != "" {
			cfg.Destination.SFTP.PrivateKeyPath = destPrivateKey
		}
		if cfg.Destination.SFTP.Timeout == 0 {
			cfg.Destination.SFTP.Timeout = 30 * time.Second
		}
		if cfg.Destination.SFTP.MaxPacketSize == 0 {
			cfg.Destination.SFTP.MaxPacketSize = 32768
		}

	case config.StorageTypeFTPS:
		if cfg.Destination.FTPS == nil {
			cfg.Destination.FTPS = &config.FTPSConfig{}
		}
		if destHost != "" {
			cfg.Destination.FTPS.Host = destHost
		}
		if destPort != 0 {
			cfg.Destination.FTPS.Port = destPort
		}
		if destUsername != "" {
			cfg.Destination.FTPS.Username = destUsername
		}
		if destPassword != "" {
			cfg.Destination.FTPS.Password = destPassword
		}
		if cfg.Destination.FTPS.Timeout == 0 {
			cfg.Destination.FTPS.Timeout = 30 * time.Second
		}

	case config.StorageTypeS3:
		if cfg.Destination.S3 == nil {
			cfg.Destination.S3 = &config.S3Config{}
		}
		if destRegion != "" {
			cfg.Destination.S3.Region = destRegion
		}
		if destBucket != "" {
			cfg.Destination.S3.Bucket = destBucket
		}
		if destAuthType != "" {
			cfg.Destination.S3.AuthType = config.S3AuthType(destAuthType)
		}
		if destAccessKey != "" {
			cfg.Destination.S3.AccessKeyID = destAccessKey
		}
		if destSecretKey != "" {
			cfg.Destination.S3.SecretAccessKey = destSecretKey
		}
		if cfg.Destination.S3.AuthType == "" {
			cfg.Destination.S3.AuthType = config.S3AuthAccessKey
		}

	case config.StorageTypeBlobStorage:
		if cfg.Destination.BlobStorage == nil {
			cfg.Destination.BlobStorage = &config.BlobConfig{}
		}
		if destAccountName != "" {
			cfg.Destination.BlobStorage.AccountName = destAccountName
		}
		if destContainer != "" {
			cfg.Destination.BlobStorage.ContainerName = destContainer
		}
		if destAuthType != "" {
			cfg.Destination.BlobStorage.AuthType = config.BlobAuthType(destAuthType)
		}
		if destAccountKey != "" {
			cfg.Destination.BlobStorage.AccountKey = destAccountKey
		}
		if cfg.Destination.BlobStorage.AuthType == "" {
			cfg.Destination.BlobStorage.AuthType = config.BlobAuthSharedKey
		}
	}

	// Apply transfer flags
	if bufferSize > 0 {
		cfg.Transfer.BufferSize = bufferSize
	}
	if concurrentFiles > 0 {
		cfg.Transfer.ConcurrentFiles = concurrentFiles
	}
	if retryAttempts > 0 {
		cfg.Transfer.RetryAttempts = retryAttempts
	}

	// Set defaults if not set
	if cfg.Transfer.BufferSize == 0 {
		cfg.Transfer.BufferSize = 32 * 1024 * 1024
	}
	if cfg.Transfer.ConcurrentFiles == 0 {
		cfg.Transfer.ConcurrentFiles = 4
	}
	if cfg.Transfer.RetryAttempts == 0 {
		cfg.Transfer.RetryAttempts = 3
	}
	if cfg.Transfer.RetryDelay == 0 {
		cfg.Transfer.RetryDelay = 5 * time.Second
	}
	if cfg.Transfer.Timeout == 0 {
		cfg.Transfer.Timeout = 30 * time.Minute
	}

	return nil
}

func validateConfig(cfg *config.Config) error {
	// Validate source
	if cfg.Source.Type == "" {
		return fmt.Errorf("source type is required")
	}

	switch cfg.Source.Type {
	case config.StorageTypeSFTP:
		if cfg.Source.SFTP == nil {
			return fmt.Errorf("SFTP source configuration is required")
		}
		if cfg.Source.SFTP.Host == "" {
			return fmt.Errorf("source SFTP host is required")
		}
		if cfg.Source.SFTP.Username == "" {
			return fmt.Errorf("source SFTP username is required")
		}

	case config.StorageTypeFTPS:
		if cfg.Source.FTPS == nil {
			return fmt.Errorf("FTPS source configuration is required")
		}
		if cfg.Source.FTPS.Host == "" {
			return fmt.Errorf("source FTPS host is required")
		}

	case config.StorageTypeS3:
		if cfg.Source.S3 == nil {
			return fmt.Errorf("S3 source configuration is required")
		}
		if cfg.Source.S3.Region == "" {
			return fmt.Errorf("source S3 region is required")
		}
		if cfg.Source.S3.Bucket == "" {
			return fmt.Errorf("source S3 bucket is required")
		}

	case config.StorageTypeBlobStorage:
		if cfg.Source.BlobStorage == nil {
			return fmt.Errorf("Blob Storage source configuration is required")
		}
		if cfg.Source.BlobStorage.AccountName == "" {
			return fmt.Errorf("source Blob Storage account name is required")
		}
		if cfg.Source.BlobStorage.ContainerName == "" {
			return fmt.Errorf("source Blob Storage container name is required")
		}
	}

	// Validate destination
	if cfg.Destination.Type == "" {
		return fmt.Errorf("destination type is required")
	}

	switch cfg.Destination.Type {
	case config.StorageTypeSFTP:
		if cfg.Destination.SFTP == nil {
			return fmt.Errorf("SFTP destination configuration is required")
		}
		if cfg.Destination.SFTP.Host == "" {
			return fmt.Errorf("destination SFTP host is required")
		}
		if cfg.Destination.SFTP.Username == "" {
			return fmt.Errorf("destination SFTP username is required")
		}

	case config.StorageTypeFTPS:
		if cfg.Destination.FTPS == nil {
			return fmt.Errorf("FTPS destination configuration is required")
		}
		if cfg.Destination.FTPS.Host == "" {
			return fmt.Errorf("destination FTPS host is required")
		}

	case config.StorageTypeS3:
		if cfg.Destination.S3 == nil {
			return fmt.Errorf("S3 destination configuration is required")
		}
		if cfg.Destination.S3.Region == "" {
			return fmt.Errorf("destination S3 region is required")
		}
		if cfg.Destination.S3.Bucket == "" {
			return fmt.Errorf("destination S3 bucket is required")
		}

	case config.StorageTypeBlobStorage:
		if cfg.Destination.BlobStorage == nil {
			return fmt.Errorf("Blob Storage destination configuration is required")
		}
		if cfg.Destination.BlobStorage.AccountName == "" {
			return fmt.Errorf("destination Blob Storage account name is required")
		}
		if cfg.Destination.BlobStorage.ContainerName == "" {
			return fmt.Errorf("destination Blob Storage container name is required")
		}
	}

	return nil
}
