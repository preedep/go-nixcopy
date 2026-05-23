package storage

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"

	"github.com/preedep/go-nixcopy/internal/domain/entity"
	"github.com/preedep/go-nixcopy/internal/domain/repository"
	"github.com/preedep/go-nixcopy/internal/infrastructure/config"
)

type BlobStorage struct {
	config            *config.BlobConfig
	client            *azblob.Client
	containerClient   *container.Client
	uploadConcurrency int
}

func NewBlobStorage(cfg *config.BlobConfig) repository.Storage {
	return &BlobStorage{
		config:            cfg,
		uploadConcurrency: cfg.UploadConcurrency,
	}
}

func (b *BlobStorage) Connect(ctx context.Context) error {
	var serviceURL string
	if b.config.Endpoint != "" {
		serviceURL = b.config.Endpoint
	} else {
		serviceURL = fmt.Sprintf("https://%s.blob.core.windows.net/", b.config.AccountName)
	}

	var client *azblob.Client
	var err error

	switch b.config.AuthType {
	case config.BlobAuthSharedKey:
		if b.config.AccountKey == "" {
			return fmt.Errorf("account_key is required for shared_key auth type")
		}
		cred, err := azblob.NewSharedKeyCredential(b.config.AccountName, b.config.AccountKey)
		if err != nil {
			return fmt.Errorf("failed to create shared key credentials: %w", err)
		}
		client, err = azblob.NewClientWithSharedKeyCredential(serviceURL, cred, nil)
		if err != nil {
			return fmt.Errorf("failed to create blob client with shared key: %w", err)
		}

	case config.BlobAuthSASToken:
		if b.config.SASToken == "" {
			return fmt.Errorf("sas_token is required for sas_token auth type")
		}
		var sasURL string
		if !strings.Contains(b.config.SASToken, "?") {
			sasURL = serviceURL + "?" + b.config.SASToken
		} else {
			sasURL = serviceURL + b.config.SASToken
		}
		client, err = azblob.NewClientWithNoCredential(sasURL, nil)
		if err != nil {
			return fmt.Errorf("failed to create blob client with SAS token: %w", err)
		}

	case config.BlobAuthConnectionString:
		if b.config.ConnectionString == "" {
			return fmt.Errorf("connection_string is required for connection_string auth type")
		}
		client, err = azblob.NewClientFromConnectionString(b.config.ConnectionString, nil)
		if err != nil {
			return fmt.Errorf("failed to create blob client from connection string: %w", err)
		}

	case config.BlobAuthManagedIdentity:
		miOpts := &azidentity.ManagedIdentityCredentialOptions{}
		if b.config.ClientID != "" {
			miOpts.ID = azidentity.ClientID(b.config.ClientID)
		}
		cred, err := azidentity.NewManagedIdentityCredential(miOpts)
		if err != nil {
			return fmt.Errorf("failed to create managed identity credentials: %w", err)
		}
		client, err = azblob.NewClient(serviceURL, cred, nil)
		if err != nil {
			return fmt.Errorf("failed to create blob client with managed identity: %w", err)
		}

	case config.BlobAuthServicePrincipal:
		if b.config.TenantID == "" || b.config.ClientID == "" || b.config.ClientSecret == "" {
			return fmt.Errorf("tenant_id, client_id, and client_secret are required for service_principal auth type")
		}
		cred, err := azidentity.NewClientSecretCredential(
			b.config.TenantID,
			b.config.ClientID,
			b.config.ClientSecret,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to create service principal credentials: %w", err)
		}
		client, err = azblob.NewClient(serviceURL, cred, nil)
		if err != nil {
			return fmt.Errorf("failed to create blob client with service principal: %w", err)
		}

	default:
		return fmt.Errorf("unsupported auth type: %s", b.config.AuthType)
	}

	b.client = client
	b.containerClient = client.ServiceClient().NewContainerClient(b.config.ContainerName)

	return nil
}

func (b *BlobStorage) Disconnect(ctx context.Context) error {
	return nil
}

func (b *BlobStorage) List(ctx context.Context, path string) ([]entity.FileInfo, error) {
	if b.containerClient == nil {
		return nil, fmt.Errorf("blob client not connected")
	}

	prefix := strings.TrimPrefix(path, "/")
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	pager := b.containerClient.NewListBlobsFlatPager(&container.ListBlobsFlatOptions{
		Prefix: &prefix,
	})

	fileInfos := []entity.FileInfo{}

	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list blobs: %w", err)
		}

		for _, blobItem := range resp.Segment.BlobItems {
			fileInfos = append(fileInfos, entity.FileInfo{
				Path:         *blobItem.Name,
				Name:         filepath.Base(*blobItem.Name),
				Size:         *blobItem.Properties.ContentLength,
				ModifiedTime: *blobItem.Properties.LastModified,
				IsDirectory:  false,
			})
		}
	}

	return fileInfos, nil
}

func (b *BlobStorage) Read(ctx context.Context, path string) (io.ReadCloser, int64, error) {
	if b.client == nil {
		return nil, 0, fmt.Errorf("blob client not connected")
	}

	blobName := strings.TrimPrefix(path, "/")
	blobClient := b.containerClient.NewBlobClient(blobName)

	props, err := blobClient.GetProperties(ctx, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get blob properties: %w", err)
	}

	resp, err := blobClient.DownloadStream(ctx, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to download blob: %w", err)
	}

	return resp.Body, *props.ContentLength, nil
}

func (b *BlobStorage) Stat(ctx context.Context, path string) (*entity.FileInfo, error) {
	if b.client == nil {
		return nil, fmt.Errorf("blob client not connected")
	}

	blobName := strings.TrimPrefix(path, "/")
	blobClient := b.containerClient.NewBlobClient(blobName)

	props, err := blobClient.GetProperties(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get blob properties: %w", err)
	}

	return &entity.FileInfo{
		Path:         path,
		Name:         filepath.Base(path),
		Size:         *props.ContentLength,
		ModifiedTime: *props.LastModified,
		IsDirectory:  false,
	}, nil
}

const (
	blobMinBlockSize      = int64(1 * 1024 * 1024)  // 1 MiB — Azure SDK minimum
	blobDefaultBlockSize  = int64(16 * 1024 * 1024) // 16 MiB — good default for most blob sizes
	blobMaxBlocks         = int64(50000)            // Azure limit: max 50,000 blocks per blob
	blobUploadConcurrency = 5                       // concurrent block uploads within one blob
)

// blobBlockSizeFor returns the optimal block size for a blob of the given size.
// Rules: >= blobMinBlockSize, and size/blockSize <= blobMaxBlocks.
// When size <= 0 (unknown), the default is returned.
func blobBlockSizeFor(size int64) int64 {
	blockSize := blobDefaultBlockSize
	if size > 0 {
		required := (size + blobMaxBlocks - 1) / blobMaxBlocks
		if required > blockSize {
			blockSize = required
		}
	}
	if blockSize < blobMinBlockSize {
		blockSize = blobMinBlockSize
	}
	return blockSize
}

func (b *BlobStorage) Write(ctx context.Context, path string, reader io.Reader, size int64) error {
	if b.client == nil {
		return fmt.Errorf("blob client not connected")
	}

	blobName := strings.TrimPrefix(path, "/")
	blobClient := b.containerClient.NewBlockBlobClient(blobName)

	concurrency := b.uploadConcurrency
	if concurrency <= 0 {
		concurrency = blobUploadConcurrency
	}
	_, err := blobClient.UploadStream(ctx, reader, &azblob.UploadStreamOptions{
		BlockSize:   blobBlockSizeFor(size),
		Concurrency: concurrency,
	})
	if err != nil {
		return fmt.Errorf("failed to upload blob: %w", err)
	}

	return nil
}

func (b *BlobStorage) CreateDirectory(ctx context.Context, path string) error {
	return nil
}

func (b *BlobStorage) Delete(ctx context.Context, path string) error {
	if b.client == nil {
		return fmt.Errorf("blob client not connected")
	}

	blobName := strings.TrimPrefix(path, "/")
	blobClient := b.containerClient.NewBlobClient(blobName)

	_, err := blobClient.Delete(ctx, &blob.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete blob: %w", err)
	}

	return nil
}
