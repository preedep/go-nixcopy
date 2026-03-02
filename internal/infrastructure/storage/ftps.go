package storage

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/jlaffaye/ftp"
	"github.com/preedep/go-nixcopy/internal/domain/entity"
	"github.com/preedep/go-nixcopy/internal/domain/repository"
	"github.com/preedep/go-nixcopy/internal/infrastructure/config"
)

type FTPSStorage struct {
	config    *config.FTPSConfig
	ftpClient *ftp.ServerConn
}

func NewFTPSStorage(cfg *config.FTPSConfig) repository.Storage {
	return &FTPSStorage{
		config: cfg,
	}
}

func (f *FTPSStorage) Connect(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", f.config.Host, f.config.Port)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: f.config.SkipVerify,
	}

	var conn *ftp.ServerConn
	var err error

	if f.config.TLSMode == "explicit" {
		conn, err = ftp.Dial(addr,
			ftp.DialWithTimeout(f.config.Timeout),
			ftp.DialWithExplicitTLS(tlsConfig),
		)
	} else {
		conn, err = ftp.Dial(addr,
			ftp.DialWithTimeout(f.config.Timeout),
			ftp.DialWithTLS(tlsConfig),
		)
	}

	if err != nil {
		return fmt.Errorf("failed to connect to FTP server: %w", err)
	}

	if err := conn.Login(f.config.Username, f.config.Password); err != nil {
		conn.Quit()
		return fmt.Errorf("failed to login: %w", err)
	}

	f.ftpClient = conn
	return nil
}

func (f *FTPSStorage) Disconnect(ctx context.Context) error {
	if f.ftpClient != nil {
		return f.ftpClient.Quit()
	}
	return nil
}

func (f *FTPSStorage) List(ctx context.Context, path string) ([]entity.FileInfo, error) {
	if f.ftpClient == nil {
		return nil, fmt.Errorf("FTP client not connected")
	}

	entries, err := f.ftpClient.List(path)
	if err != nil {
		return nil, fmt.Errorf("failed to list directory: %w", err)
	}

	fileInfos := make([]entity.FileInfo, 0, len(entries))
	for _, entry := range entries {
		fileInfos = append(fileInfos, entity.FileInfo{
			Path:         filepath.Join(path, entry.Name),
			Name:         entry.Name,
			Size:         int64(entry.Size),
			ModifiedTime: entry.Time,
			IsDirectory:  entry.Type == ftp.EntryTypeFolder,
		})
	}

	return fileInfos, nil
}

func (f *FTPSStorage) Read(ctx context.Context, path string) (io.ReadCloser, int64, error) {
	if f.ftpClient == nil {
		return nil, 0, fmt.Errorf("FTP client not connected")
	}

	size, err := f.ftpClient.FileSize(path)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get file size: %w", err)
	}

	resp, err := f.ftpClient.Retr(path)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve file: %w", err)
	}

	return resp, int64(size), nil
}

func (f *FTPSStorage) Stat(ctx context.Context, path string) (*entity.FileInfo, error) {
	if f.ftpClient == nil {
		return nil, fmt.Errorf("FTP client not connected")
	}

	size, err := f.ftpClient.FileSize(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get file size: %w", err)
	}

	modTime, err := f.ftpClient.GetTime(path)
	if err != nil {
		modTime = time.Now()
	}

	return &entity.FileInfo{
		Path:         path,
		Name:         filepath.Base(path),
		Size:         int64(size),
		ModifiedTime: modTime,
		IsDirectory:  false,
	}, nil
}

func (f *FTPSStorage) Write(ctx context.Context, path string, reader io.Reader, size int64) error {
	if f.ftpClient == nil {
		return fmt.Errorf("FTP client not connected")
	}

	dir := filepath.Dir(path)
	if dir != "." && dir != "/" {
		if err := f.ftpClient.MakeDir(dir); err != nil {
		}
	}

	if err := f.ftpClient.Stor(path, reader); err != nil {
		return fmt.Errorf("failed to store file: %w", err)
	}

	return nil
}

func (f *FTPSStorage) CreateDirectory(ctx context.Context, path string) error {
	if f.ftpClient == nil {
		return fmt.Errorf("FTP client not connected")
	}

	return f.ftpClient.MakeDir(path)
}

func (f *FTPSStorage) Delete(ctx context.Context, path string) error {
	if f.ftpClient == nil {
		return fmt.Errorf("FTP client not connected")
	}

	return f.ftpClient.Delete(path)
}
