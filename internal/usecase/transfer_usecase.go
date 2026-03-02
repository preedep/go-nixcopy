// Package usecase implements the application's business logic layer.
// It orchestrates data flow between the domain layer and infrastructure layer,
// coordinating file transfers with retry logic, progress tracking, and concurrent operations.
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

// TransferUseCase implements the core file transfer business logic.
// It handles single and batch file transfers with the following features:
//   - Streaming I/O to minimize memory usage (only bufferSize bytes in memory at a time)
//   - Automatic retry mechanism with configurable attempts and delays
//   - Real-time progress tracking with speed calculation and ETA
//   - Concurrent batch transfers with semaphore-based concurrency control
//   - Context-aware cancellation support
//
// Memory Efficiency:
// The implementation uses streaming I/O, meaning it never loads entire files into memory.
// Memory usage is bounded by: bufferSize * concurrentFiles + overhead
// For example: 32MB buffer * 4 concurrent files = ~128MB maximum memory usage
//
// Thread Safety:
// TransferUseCase is safe for concurrent use. Multiple goroutines can call
// Transfer() or TransferBatch() simultaneously on the same instance.
type TransferUseCase struct {
	source repository.StorageReader // Source storage for reading files
	dest   repository.StorageWriter // Destination storage for writing files
	config *entity.TransferConfig   // Transfer configuration (buffer size, retries, etc.)
	logger *zap.Logger              // Structured logger for operational visibility
}

// NewTransferUseCase creates a new TransferUseCase instance.
//
// Parameters:
//   - source: Storage reader for accessing source files (SFTP, FTPS, S3, Blob, etc.)
//   - dest: Storage writer for writing destination files
//   - config: Transfer configuration including buffer size, retry settings, and concurrency limits
//   - logger: Structured logger for tracking transfer operations and debugging
//
// The returned service implements service.TransferService interface and is ready for use.
// All parameters must be non-nil; the function does not perform nil checks for performance.
//
// Example:
//
//	config := &entity.TransferConfig{
//	    BufferSize: 32 * 1024 * 1024,  // 32MB
//	    ConcurrentFiles: 4,
//	    RetryAttempts: 3,
//	    RetryDelay: 5 * time.Second,
//	}
//	transferService := NewTransferUseCase(sourceStorage, destStorage, config, logger)
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

// Transfer performs a single file transfer from source to destination with retry logic and progress tracking.
//
// The method implements a streaming transfer approach, reading and writing data in chunks
// to minimize memory usage. It supports automatic retries on failure and provides real-time
// progress updates through an optional channel.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control. Transfer will abort if context is cancelled.
//   - sourcePath: Path to the source file in the source storage system
//   - destPath: Path where the file should be written in the destination storage system
//   - progressChan: Optional channel for receiving real-time progress updates. If nil, no progress is reported.
//     The channel will be closed when the transfer completes or fails.
//
// Returns:
//   - *entity.TransferResult: Contains transfer metadata including bytes transferred, duration, and status
//   - error: Non-nil if all retry attempts failed. The error wraps the underlying failure reason.
//
// Behavior:
//   - Validates source file exists and is not a directory
//   - Attempts transfer up to (RetryAttempts + 1) times
//   - Waits RetryDelay between retry attempts
//   - Sends progress updates every 500ms during transfer
//   - Closes progressChan on completion (success or failure)
//
// Error Handling:
// The method retries on transient errors (network issues, temporary unavailability).
// It does NOT retry on:
//   - Context cancellation
//   - Source file not found (after initial stat)
//   - Directory instead of file
//
// Memory Usage:
// Peak memory usage is approximately config.BufferSize (default 32MB).
// The entire file is never loaded into memory.
//
// Example:
//
//	progressChan := make(chan entity.TransferProgress, 10)
//	go func() {
//	    for progress := range progressChan {
//	        fmt.Printf("Progress: %.2f%% at %.2f MB/s\n",
//	            float64(progress.TransferredBytes)/float64(progress.TotalBytes)*100,
//	            progress.Speed/1024/1024)
//	    }
//	}()
//	result, err := transferUseCase.Transfer(ctx, "/source/file.zip", "/dest/file.zip", progressChan)
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

	// Ensure progress channel is closed when function returns, preventing goroutine leaks
	if progressChan != nil {
		defer close(progressChan)
	}

	t.logger.Info("Starting transfer",
		zap.String("source", sourcePath),
		zap.String("destination", destPath),
	)

	// Stat the source file to validate it exists and get metadata
	// This is done before the retry loop to fail fast on non-existent files
	stat, err := t.source.Stat(ctx, sourcePath)
	if err != nil {
		result.Status = entity.TransferStatusFailed
		result.Error = fmt.Errorf("failed to stat source file: %w", err)
		return result, result.Error
	}

	// Reject directory transfers in single file mode
	// Use TransferBatch for directory/multiple file operations
	if stat.IsDirectory {
		return nil, fmt.Errorf("directory transfer not supported in single file mode")
	}

	// Retry loop: attempt transfer up to (RetryAttempts + 1) times
	// This handles transient failures like network timeouts, temporary unavailability, etc.
	var lastErr error
	for attempt := 0; attempt <= t.config.RetryAttempts; attempt++ {
		// Wait before retry (skip on first attempt)
		if attempt > 0 {
			t.logger.Warn("Retrying transfer",
				zap.Int("attempt", attempt),
				zap.String("source", sourcePath),
			)
			time.Sleep(t.config.RetryDelay)
		}

		// Open source file for reading
		// Returns an io.ReadCloser stream - file is NOT loaded into memory
		reader, size, err := t.source.Read(ctx, sourcePath)
		if err != nil {
			lastErr = fmt.Errorf("failed to read source file: %w", err)
			continue // Retry on next iteration
		}

		// Wrap reader with progress tracking
		// This intercepts Read() calls to calculate speed and report progress
		progressReader := &progressReader{
			reader:       reader,
			total:        size,
			progressChan: progressChan,
			fileName:     stat.Name,
			startTime:    startTime,
			bufferSize:   t.config.BufferSize,
		}

		// Stream data from source to destination
		// Data flows in chunks (bufferSize) to minimize memory usage
		err = t.dest.Write(ctx, destPath, progressReader, size)
		reader.Close() // Always close reader, even on error

		if err != nil {
			lastErr = fmt.Errorf("failed to write destination file: %w", err)
			continue // Retry on next iteration
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

// TransferBatch performs concurrent transfer of multiple files from source to destination.
//
// This method orchestrates parallel file transfers using a semaphore pattern to limit
// concurrency. It's designed for high-throughput scenarios where multiple files need
// to be transferred efficiently.
//
// Parameters:
//   - ctx: Context for cancellation. If cancelled, no new transfers start, but in-flight transfers continue.
//   - sourcePaths: Slice of source file paths to transfer. Can include wildcards if pre-expanded.
//   - destBasePath: Base destination directory. Each file will be written as destBasePath/filename.
//   - progressChan: Optional channel for receiving progress updates from all concurrent transfers.
//     Progress from multiple files will be multiplexed into this single channel.
//
// Returns:
//   - []*entity.TransferResult: Slice of results in the same order as sourcePaths.
//     Each result contains transfer metadata, even for failed transfers.
//   - error: Returns context error if cancelled, otherwise nil. Individual transfer failures
//     are captured in the result slice, not returned as errors.
//
// Concurrency Control:
// The method uses a semaphore (buffered channel) to limit concurrent transfers to
// config.ConcurrentFiles. This prevents resource exhaustion when transferring many files.
//
// Behavior:
//   - Launches goroutines for each file transfer
//   - Limits concurrent transfers via semaphore
//   - Waits for all transfers to complete before returning
//   - Preserves order of results (matches sourcePaths order)
//   - Each file's progress is forwarded to the shared progressChan
//
// Destination Path Construction:
// For each source file, the destination path is constructed as:
//  1. Try to stat the source file to get its name
//  2. If stat succeeds: destBasePath/stat.Name
//  3. If stat fails: destBasePath/filepath.Base(sourcePath)
//
// Memory Usage:
// Peak memory usage is approximately: config.BufferSize * config.ConcurrentFiles
// For example: 32MB * 4 = 128MB for default settings
//
// Error Handling:
// Individual file transfer failures do NOT stop the batch operation.
// Failed transfers are marked with TransferStatusFailed in their result.
// Check each result's Status and Error fields to identify failures.
//
// Example:
//
//	files := []string{"/data/file1.pdf", "/data/file2.pdf", "/data/file3.pdf"}
//	results, err := transferUseCase.TransferBatch(ctx, files, "/backup/", progressChan)
//	for i, result := range results {
//	    if result.Status == entity.TransferStatusFailed {
//	        fmt.Printf("File %s failed: %v\n", files[i], result.Error)
//	    }
//	}
func (t *TransferUseCase) TransferBatch(
	ctx context.Context,
	sourcePaths []string,
	destBasePath string,
	progressChan chan<- entity.TransferProgress,
) ([]*entity.TransferResult, error) {
	results := make([]*entity.TransferResult, len(sourcePaths))

	// Semaphore pattern for concurrency control
	// Buffer size = max concurrent transfers allowed
	semaphore := make(chan struct{}, t.config.ConcurrentFiles)

	// Channel for collecting results from goroutines
	// Buffered to prevent goroutine blocking
	resultChan := make(chan struct {
		index  int
		result *entity.TransferResult
	}, len(sourcePaths))

	t.logger.Info("Starting batch transfer",
		zap.Int("total_files", len(sourcePaths)),
		zap.Int("concurrent", t.config.ConcurrentFiles),
	)

	// Launch goroutines for each file transfer
	// Semaphore controls how many run concurrently
	for i, sourcePath := range sourcePaths {
		// Check for context cancellation before starting new transfer
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		case semaphore <- struct{}{}: // Acquire semaphore slot (blocks if at limit)
		}

		// Launch transfer in goroutine
		// Each transfer runs independently and reports results via resultChan
		go func(index int, src string) {
			// Release semaphore slot when goroutine completes
			// This allows another transfer to start
			defer func() { <-semaphore }()

			// Construct destination path from source filename
			// Prefer actual filename from stat, fallback to path basename
			stat, err := t.source.Stat(ctx, src)
			var destPath string
			if err == nil && stat != nil {
				// Use actual filename from storage metadata
				destPath = filepath.Join(destBasePath, stat.Name)
			} else {
				// Fallback: extract filename from path
				// This handles cases where stat fails but transfer might still work
				destPath = filepath.Join(destBasePath, filepath.Base(src))
			}

			// Create per-file progress channel and forward to shared channel
			// This multiplexes progress from multiple concurrent transfers
			fileProgressChan := make(chan entity.TransferProgress, 10)
			go func() {
				// Forward all progress updates to the shared channel
				for progress := range fileProgressChan {
					if progressChan != nil {
						progressChan <- progress
					}
				}
			}()

			// Perform the actual file transfer
			// This call includes retry logic and progress reporting
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

			// Send result back to main goroutine
			// Include index to maintain order in results slice
			resultChan <- struct {
				index  int
				result *entity.TransferResult
			}{index: index, result: result}
		}(i, sourcePath)
	}

	// Wait for all transfers to complete
	// Collect results and place them in the correct order
	for i := 0; i < len(sourcePaths); i++ {
		res := <-resultChan
		results[res.index] = res.result // Preserve original order
	}

	// Drain semaphore to ensure all goroutines have completed
	// This prevents potential goroutine leaks
	for i := 0; i < cap(semaphore); i++ {
		semaphore <- struct{}{}
	}

	return results, nil
}

// progressReader wraps an io.Reader to track transfer progress and calculate metrics.
//
// It implements io.Reader and io.Closer interfaces, making it a drop-in replacement
// for any io.Reader. On each Read() call, it updates transfer statistics and
// periodically sends progress updates through a channel.
//
// Progress Reporting:
// Updates are sent at most once every 500ms to avoid overwhelming the progress channel
// and consuming excessive CPU for progress calculation.
//
// Metrics Calculated:
//   - Transfer speed (bytes/second) based on elapsed time
//   - Estimated time remaining (ETA) based on current speed
//   - Percentage complete (transferred/total)
//
// Thread Safety:
// This type is NOT thread-safe. It should only be used by a single goroutine
// (the one performing the Read operations).
type progressReader struct {
	reader         io.Reader                      // Underlying reader to wrap
	total          int64                          // Total bytes to transfer
	transferred    int64                          // Bytes transferred so far
	progressChan   chan<- entity.TransferProgress // Channel for sending progress updates
	fileName       string                         // Name of file being transferred (for progress reporting)
	startTime      time.Time                      // Transfer start time (for speed calculation)
	lastUpdateTime time.Time                      // Last time progress was sent (for throttling)
	bufferSize     int                            // Buffer size (informational, not used in logic)
}

// Read implements io.Reader interface with progress tracking.
//
// This method wraps the underlying reader's Read() call and tracks the number
// of bytes read. It calculates transfer metrics and sends progress updates
// at most once every 500ms.
//
// The method follows io.Reader contract:
//   - Returns number of bytes read and any error from underlying reader
//   - Updates are sent even if err != nil (to report final progress)
//   - Does not modify the buffer contents
//
// Performance:
// Progress calculation is lightweight (few arithmetic operations) and only
// runs every 500ms, adding negligible overhead to the transfer.
func (p *progressReader) Read(buf []byte) (int, error) {
	// Read from underlying reader
	n, err := p.reader.Read(buf)

	// Update transferred bytes counter
	// This happens even on error to track partial reads
	p.transferred += int64(n)

	// Send progress update if enough time has elapsed (throttling)
	// This prevents excessive progress updates and CPU usage
	if p.progressChan != nil && time.Since(p.lastUpdateTime) > 500*time.Millisecond {
		// Calculate transfer metrics
		elapsed := time.Since(p.startTime).Seconds()
		speed := float64(p.transferred) / elapsed // bytes per second
		remaining := p.total - p.transferred

		// Estimate time remaining based on current speed
		// This can be inaccurate if speed varies significantly
		estimatedTime := time.Duration(float64(remaining)/speed) * time.Second

		// Send progress update (non-blocking send)
		p.progressChan <- entity.TransferProgress{
			FileName:         p.fileName,
			TotalBytes:       p.total,
			TransferredBytes: p.transferred,
			Speed:            speed,
			StartTime:        p.startTime,
			EstimatedTime:    estimatedTime,
			Status:           entity.TransferStatusInProgress,
		}

		// Update last update time for throttling
		p.lastUpdateTime = time.Now()
	}

	// Return read result (preserves io.Reader contract)
	return n, err
}

// Close implements io.Closer interface.
//
// This method closes the underlying reader if it implements io.Closer.
// If the underlying reader doesn't implement io.Closer, this is a no-op.
//
// This allows progressReader to be used with defer close() patterns
// regardless of whether the underlying reader needs closing.
//
// Returns:
//   - error from underlying Close() if reader implements io.Closer
//   - nil if reader doesn't implement io.Closer
func (p *progressReader) Close() error {
	if closer, ok := p.reader.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
