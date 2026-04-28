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

// Resumer is an optional interface that storage backends can implement to
// support resuming interrupted transfers. Backends that do not implement it
// will fall back to full re-transfer on retry.
type Resumer interface {
	// ReadFrom returns a reader starting at offset bytes into the file.
	// The returned int64 is the number of bytes remaining (total size - offset).
	ReadFrom(ctx context.Context, path string, offset int64) (io.ReadCloser, int64, error)
	// AppendWrite writes reader content into the destination file starting at
	// offset, extending an existing partial file.
	AppendWrite(ctx context.Context, path string, reader io.Reader, size, offset int64) error
}
