package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/sftp"
	"github.com/preedep/go-nixcopy/internal/domain/entity"
	"github.com/preedep/go-nixcopy/internal/domain/repository"
	"github.com/preedep/go-nixcopy/internal/infrastructure/config"
	"golang.org/x/crypto/ssh"
)

type SFTPStorage struct {
	config     *config.SFTPConfig
	sshClient  *ssh.Client
	sftpClient *sftp.Client
}

func NewSFTPStorage(cfg *config.SFTPConfig) repository.Storage {
	return &SFTPStorage{
		config: cfg,
	}
}

func (s *SFTPStorage) Connect(ctx context.Context) error {
	authMethods := []ssh.AuthMethod{}

	if s.config.Password != "" {
		authMethods = append(authMethods, ssh.Password(s.config.Password))
	}

	if s.config.PrivateKeyPath != "" {
		key, err := os.ReadFile(s.config.PrivateKeyPath)
		if err != nil {
			return fmt.Errorf("failed to read private key: %w", err)
		}

		var signer ssh.Signer
		if s.config.PrivateKeyPass != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(s.config.PrivateKeyPass))
		} else {
			signer, err = ssh.ParsePrivateKey(key)
		}
		if err != nil {
			return fmt.Errorf("failed to parse private key: %w", err)
		}

		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	sshConfig := &ssh.ClientConfig{
		User:            s.config.Username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         s.config.Timeout,
	}

	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	sshClient, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to SSH server: %w", err)
	}

	sftpClient, err := sftp.NewClient(sshClient, sftp.MaxPacket(s.config.MaxPacketSize))
	if err != nil {
		sshClient.Close()
		return fmt.Errorf("failed to create SFTP client: %w", err)
	}

	s.sshClient = sshClient
	s.sftpClient = sftpClient

	return nil
}

func (s *SFTPStorage) Disconnect(ctx context.Context) error {
	if s.sftpClient != nil {
		s.sftpClient.Close()
	}
	if s.sshClient != nil {
		s.sshClient.Close()
	}
	return nil
}

func (s *SFTPStorage) List(ctx context.Context, path string) ([]entity.FileInfo, error) {
	if s.sftpClient == nil {
		return nil, fmt.Errorf("SFTP client not connected")
	}

	entries, err := s.sftpClient.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to list directory: %w", err)
	}

	fileInfos := make([]entity.FileInfo, 0, len(entries))
	for _, entry := range entries {
		fileInfos = append(fileInfos, entity.FileInfo{
			Path:         filepath.Join(path, entry.Name()),
			Name:         entry.Name(),
			Size:         entry.Size(),
			ModifiedTime: entry.ModTime(),
			IsDirectory:  entry.IsDir(),
		})
	}

	return fileInfos, nil
}

func (s *SFTPStorage) Read(ctx context.Context, path string) (io.ReadCloser, int64, error) {
	if s.sftpClient == nil {
		return nil, 0, fmt.Errorf("SFTP client not connected")
	}

	file, err := s.sftpClient.Open(path)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to open file: %w", err)
	}

	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, 0, fmt.Errorf("failed to stat file: %w", err)
	}

	return file, stat.Size(), nil
}

func (s *SFTPStorage) Stat(ctx context.Context, path string) (*entity.FileInfo, error) {
	if s.sftpClient == nil {
		return nil, fmt.Errorf("SFTP client not connected")
	}

	stat, err := s.sftpClient.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	return &entity.FileInfo{
		Path:         path,
		Name:         stat.Name(),
		Size:         stat.Size(),
		ModifiedTime: stat.ModTime(),
		IsDirectory:  stat.IsDir(),
	}, nil
}

func (s *SFTPStorage) Write(ctx context.Context, path string, reader io.Reader, size int64) error {
	if s.sftpClient == nil {
		return fmt.Errorf("SFTP client not connected")
	}

	dir := filepath.Dir(path)
	if err := s.sftpClient.MkdirAll(dir); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := s.sftpClient.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, reader)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (s *SFTPStorage) CreateDirectory(ctx context.Context, path string) error {
	if s.sftpClient == nil {
		return fmt.Errorf("SFTP client not connected")
	}

	return s.sftpClient.MkdirAll(path)
}

func (s *SFTPStorage) Delete(ctx context.Context, path string) error {
	if s.sftpClient == nil {
		return fmt.Errorf("SFTP client not connected")
	}

	return s.sftpClient.Remove(path)
}
