package usecase

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/preedep/go-nixcopy/internal/domain/entity"
	"github.com/preedep/go-nixcopy/internal/domain/repository"
	"github.com/preedep/go-nixcopy/internal/domain/service"
	"go.uber.org/zap"
)

type TransferUseCase struct {
	source repository.StorageReader
	dest   repository.StorageWriter
	config *entity.TransferConfig
	logger *zap.Logger
}

func NewTransferUseCase(
	source repository.StorageReader,
	dest repository.StorageWriter,
	config *entity.TransferConfig,
	logger *zap.Logger,
) service.TransferService {
	return &TransferUseCase{
		source: source,
		dest:   dest,
		config: config,
		logger: logger,
	}
}

func (t *TransferUseCase) Transfer(
	ctx context.Context,
	sourcePath, destPath string,
	progressChan chan<- entity.TransferProgress,
) (*entity.TransferResult, error) {
	startTime := time.Now()
	result := &entity.TransferResult{
		SourcePath:      sourcePath,
		DestinationPath: destPath,
		Status:          entity.TransferStatusPending,
	}

	if progressChan != nil {
		defer close(progressChan)
	}

	t.logger.Info("Starting transfer",
		zap.String("source", sourcePath),
		zap.String("destination", destPath),
	)

	stat, err := t.source.Stat(ctx, sourcePath)
	if err != nil {
		result.Status = entity.TransferStatusFailed
		result.Error = fmt.Errorf("failed to stat source file: %w", err)
		return result, result.Error
	}

	if stat.IsDirectory {
		return nil, fmt.Errorf("directory transfer not supported in single file mode")
	}

	var lastErr error
	for attempt := 0; attempt <= t.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			t.logger.Warn("Retrying transfer",
				zap.Int("attempt", attempt),
				zap.String("source", sourcePath),
			)
			time.Sleep(t.config.RetryDelay)
		}

		reader, size, err := t.source.Read(ctx, sourcePath)
		if err != nil {
			lastErr = fmt.Errorf("failed to read source file: %w", err)
			continue
		}

		progressReader := &progressReader{
			reader:       reader,
			total:        size,
			progressChan: progressChan,
			fileName:     stat.Name,
			startTime:    startTime,
			bufferSize:   t.config.BufferSize,
		}

		err = t.dest.Write(ctx, destPath, progressReader, size)
		reader.Close()

		if err != nil {
			lastErr = fmt.Errorf("failed to write destination file: %w", err)
			continue
		}

		result.BytesTransferred = size
		result.Duration = time.Since(startTime)
		result.Status = entity.TransferStatusCompleted

		t.logger.Info("Transfer completed",
			zap.String("source", sourcePath),
			zap.String("destination", destPath),
			zap.Int64("bytes", size),
			zap.Duration("duration", result.Duration),
		)

		if progressChan != nil {
			progressChan <- entity.TransferProgress{
				FileName:         stat.Name,
				TotalBytes:       size,
				TransferredBytes: size,
				Speed:            float64(size) / result.Duration.Seconds(),
				StartTime:        startTime,
				Status:           entity.TransferStatusCompleted,
			}
		}

		return result, nil
	}

	result.Status = entity.TransferStatusFailed
	result.Error = lastErr
	result.Duration = time.Since(startTime)

	if progressChan != nil {
		progressChan <- entity.TransferProgress{
			FileName:  stat.Name,
			Status:    entity.TransferStatusFailed,
			Error:     lastErr,
			StartTime: startTime,
		}
	}

	return result, lastErr
}

func (t *TransferUseCase) TransferBatch(
	ctx context.Context,
	sourcePaths []string,
	destBasePath string,
	progressChan chan<- entity.TransferProgress,
) ([]*entity.TransferResult, error) {
	results := make([]*entity.TransferResult, len(sourcePaths))
	semaphore := make(chan struct{}, t.config.ConcurrentFiles)
	resultChan := make(chan struct {
		index  int
		result *entity.TransferResult
	}, len(sourcePaths))

	t.logger.Info("Starting batch transfer",
		zap.Int("total_files", len(sourcePaths)),
		zap.Int("concurrent", t.config.ConcurrentFiles),
	)

	for i, sourcePath := range sourcePaths {
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		case semaphore <- struct{}{}:
		}

		go func(index int, src string) {
			defer func() { <-semaphore }()

			// Construct destination path
			stat, err := t.source.Stat(ctx, src)
			var destPath string
			if err == nil && stat != nil {
				destPath = filepath.Join(destBasePath, stat.Name)
			} else {
				destPath = filepath.Join(destBasePath, filepath.Base(src))
			}

			// Create individual progress channel for this file
			fileProgressChan := make(chan entity.TransferProgress, 10)
			go func() {
				for progress := range fileProgressChan {
					if progressChan != nil {
						progressChan <- progress
					}
				}
			}()

			result, err := t.Transfer(ctx, src, destPath, fileProgressChan)
			if err != nil {
				t.logger.Error("Transfer failed",
					zap.String("source", src),
					zap.String("destination", destPath),
					zap.Error(err),
				)
			} else {
				t.logger.Info("Transfer succeeded",
					zap.String("source", src),
					zap.String("destination", destPath),
					zap.Int64("bytes", result.BytesTransferred),
				)
			}

			resultChan <- struct {
				index  int
				result *entity.TransferResult
			}{index: index, result: result}
		}(i, sourcePath)
	}

	// Wait for all transfers to complete
	for i := 0; i < len(sourcePaths); i++ {
		res := <-resultChan
		results[res.index] = res.result
	}

	// Drain semaphore
	for i := 0; i < cap(semaphore); i++ {
		semaphore <- struct{}{}
	}

	return results, nil
}

type progressReader struct {
	reader         io.Reader
	total          int64
	transferred    int64
	progressChan   chan<- entity.TransferProgress
	fileName       string
	startTime      time.Time
	lastUpdateTime time.Time
	bufferSize     int
}

func (p *progressReader) Read(buf []byte) (int, error) {
	n, err := p.reader.Read(buf)
	p.transferred += int64(n)

	if p.progressChan != nil && time.Since(p.lastUpdateTime) > 500*time.Millisecond {
		elapsed := time.Since(p.startTime).Seconds()
		speed := float64(p.transferred) / elapsed
		remaining := p.total - p.transferred
		estimatedTime := time.Duration(float64(remaining)/speed) * time.Second

		p.progressChan <- entity.TransferProgress{
			FileName:         p.fileName,
			TotalBytes:       p.total,
			TransferredBytes: p.transferred,
			Speed:            speed,
			StartTime:        p.startTime,
			EstimatedTime:    estimatedTime,
			Status:           entity.TransferStatusInProgress,
		}

		p.lastUpdateTime = time.Now()
	}

	return n, err
}

func (p *progressReader) Close() error {
	if closer, ok := p.reader.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
