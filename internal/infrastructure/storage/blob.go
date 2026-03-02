package storage

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/preedep/go-nixcopy/internal/domain/entity"
	"github.com/preedep/go-nixcopy/internal/domain/repository"
	"github.com/preedep/go-nixcopy/internal/infrastructure/config"
)

type BlobStorage struct {
	config          *config.BlobConfig
	client          *azblob.Client
	containerClient *container.Client
}

func NewBlobStorage(cfg *config.BlobConfig) repository.Storage {
	return &BlobStorage{
		config: cfg,
	}
}

func (b *BlobStorage) Connect(ctx context.Context) error {
	var serviceURL string
	if b.config.Endpoint != "" {
		serviceURL = b.config.Endpoint
	} else {
		serviceURL = fmt.Sprintf("https://%s.blob.core.windows.net/", b.config.AccountName)
	}

	cred, err := azblob.NewSharedKeyCredential(b.config.AccountName, b.config.AccountKey)
	if err != nil {
		return fmt.Errorf("failed to create credentials: %w", err)
	}

	client, err := azblob.NewClientWithSharedKeyCredential(serviceURL, cred, nil)
	if err != nil {
		return fmt.Errorf("failed to create blob client: %w", err)
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

func (b *BlobStorage) Write(ctx context.Context, path string, reader io.Reader, size int64) error {
	if b.client == nil {
		return fmt.Errorf("blob client not connected")
	}

	blobName := strings.TrimPrefix(path, "/")
	blobClient := b.containerClient.NewBlockBlobClient(blobName)

	_, err := blobClient.UploadStream(ctx, reader, &azblob.UploadStreamOptions{})
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
