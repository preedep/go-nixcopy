package repository

import (
	"context"
	"io"

	"github.com/preedep/go-nixcopy/internal/domain/entity"
)

type StorageReader interface {
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	List(ctx context.Context, path string) ([]entity.FileInfo, error)
	Read(ctx context.Context, path string) (io.ReadCloser, int64, error)
	Stat(ctx context.Context, path string) (*entity.FileInfo, error)
}

type StorageWriter interface {
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	Write(ctx context.Context, path string, reader io.Reader, size int64) error
	CreateDirectory(ctx context.Context, path string) error
	Delete(ctx context.Context, path string) error
}

type Storage interface {
	StorageReader
	StorageWriter
}
