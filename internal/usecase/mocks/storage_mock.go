package mocks

import (
	"context"
	"io"
	"sync"

	"github.com/preedep/go-nixcopy/internal/domain/entity"
)

// MockStorage is a mock implementation of repository.Storage for testing
type MockStorage struct {
	mu sync.Mutex

	// Tracking
	ConnectCalled    bool
	DisconnectCalled bool
	ListCalled       bool
	ReadCalled       bool
	WriteCalled      bool
	StatCalled       bool
	DeleteCalled     bool

	// Mock data
	Files       map[string]*entity.FileInfo
	FileContent map[string][]byte

	// Errors to return
	ConnectError    error
	DisconnectError error
	ListError       error
	ReadError       error
	WriteError      error
	StatError       error
	DeleteError     error
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		Files:       make(map[string]*entity.FileInfo),
		FileContent: make(map[string][]byte),
	}
}

func (m *MockStorage) Connect(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ConnectCalled = true
	return m.ConnectError
}

func (m *MockStorage) Disconnect(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.DisconnectCalled = true
	return m.DisconnectError
}

func (m *MockStorage) List(ctx context.Context, path string) ([]entity.FileInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ListCalled = true

	if m.ListError != nil {
		return nil, m.ListError
	}

	var files []entity.FileInfo
	for _, file := range m.Files {
		files = append(files, *file)
	}
	return files, nil
}

func (m *MockStorage) Read(ctx context.Context, path string) (io.ReadCloser, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ReadCalled = true

	if m.ReadError != nil {
		return nil, 0, m.ReadError
	}

	content, ok := m.FileContent[path]
	if !ok {
		return nil, 0, io.EOF
	}

	reader := &mockReadCloser{data: content}
	return reader, int64(len(content)), nil
}

func (m *MockStorage) Stat(ctx context.Context, path string) (*entity.FileInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.StatCalled = true

	if m.StatError != nil {
		return nil, m.StatError
	}

	file, ok := m.Files[path]
	if !ok {
		return nil, io.EOF
	}

	return file, nil
}

func (m *MockStorage) Write(ctx context.Context, path string, reader io.Reader, size int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.WriteCalled = true

	if m.WriteError != nil {
		return m.WriteError
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	m.FileContent[path] = data
	return nil
}

func (m *MockStorage) CreateDirectory(ctx context.Context, path string) error {
	return nil
}

func (m *MockStorage) Delete(ctx context.Context, path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.DeleteCalled = true

	if m.DeleteError != nil {
		return m.DeleteError
	}

	delete(m.Files, path)
	delete(m.FileContent, path)
	return nil
}

// Helper methods
func (m *MockStorage) AddFile(path string, content []byte, info *entity.FileInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.FileContent[path] = content
	m.Files[path] = info
}

type mockReadCloser struct {
	data []byte
	pos  int
}

func (m *mockReadCloser) Read(p []byte) (n int, err error) {
	if m.pos >= len(m.data) {
		return 0, io.EOF
	}
	n = copy(p, m.data[m.pos:])
	m.pos += n
	return n, nil
}

func (m *mockReadCloser) Close() error {
	return nil
}
