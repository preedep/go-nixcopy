package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/preedep/go-nixcopy/internal/domain/entity"
	"github.com/preedep/go-nixcopy/internal/infrastructure/config"
	"github.com/preedep/go-nixcopy/internal/infrastructure/logger"
	"github.com/preedep/go-nixcopy/internal/infrastructure/storage"
	"github.com/preedep/go-nixcopy/internal/usecase"
)

// transferSummary is the structured JSON line written to stdout on completion.
// Downstream tools (Airflow log parsers, Loki, CloudWatch Insights) can filter
// on event="transfer_summary" to extract transfer metrics without parsing human text.
type transferSummary struct {
	Event            string       `json:"event"`
	TotalFiles       int          `json:"total_files"`
	Successful       int          `json:"successful"`
	Skipped          int          `json:"skipped"`
	Failed           int          `json:"failed"`
	BytesTransferred int64        `json:"bytes_transferred"`
	DurationMs       int64        `json:"duration_ms"`
	AverageSpeedMBps float64      `json:"average_speed_mbps,omitempty"`
	FailedFiles      []failedFile `json:"failed_files,omitempty"`
}

type failedFile struct {
	Path  string `json:"path"`
	Error string `json:"error"`
}

var (
	sourcePath  string
	sourcePaths []string
	destPath    string

	// Source flags
	sourceType        string
	sourceHost        string
	sourcePort        int
	sourceUsername    string
	sourcePassword    string
	sourcePrivateKey  string
	sourceRegion      string
	sourceBucket      string
	sourceAccessKey   string
	sourceSecretKey   string
	sourceAuthType    string
	sourceAccountName string
	sourceAccountKey  string
	sourceContainer   string

	// Destination flags
	destType        string
	destHost        string
	destPort        int
	destUsername    string
	destPassword    string
	destPrivateKey  string
	destRegion      string
	destBucket      string
	destAccessKey   string
	destSecretKey   string
	destAuthType    string
	destAccountName string
	destAccountKey  string
	destContainer   string

	// Transfer flags
	bufferSize      int
	concurrentFiles int
	retryAttempts   int
	enableResume    bool
	skipExisting    bool
	bandwidthLimit  string
	compress        string
)

var transferCmd = &cobra.Command{
	Use:   "transfer",
	Short: "Transfer files between storage systems",
	Long:  `Transfer files from source to destination using streaming for memory efficiency`,
	RunE:  runTransfer,
}

func init() {
	rootCmd.AddCommand(transferCmd)

	// Path flags
	transferCmd.Flags().StringVarP(&sourcePath, "source", "s", "", "Source file path or pattern (supports wildcards: *.pdf, **/*.txt)")
	transferCmd.Flags().StringSliceVar(&sourcePaths, "sources", []string{}, "Multiple source file paths (comma-separated)")
	transferCmd.Flags().StringVarP(&destPath, "dest", "d", "", "Destination path (required)")
	if err := transferCmd.MarkFlagRequired("dest"); err != nil {
		panic(err)
	}

	// Source storage flags
	transferCmd.Flags().StringVar(&sourceType, "source-type", "", "Source storage type (sftp, ftps, blob, s3) (env: NIXCOPY_SOURCE_TYPE)")
	transferCmd.Flags().StringVar(&sourceHost, "source-host", "", "Source host (env: NIXCOPY_SOURCE_HOST)")
	transferCmd.Flags().IntVar(&sourcePort, "source-port", 0, "Source port (env: NIXCOPY_SOURCE_PORT)")
	transferCmd.Flags().StringVar(&sourceUsername, "source-username", "", "Source username (env: NIXCOPY_SOURCE_USERNAME)")
	transferCmd.Flags().StringVar(&sourcePassword, "source-password", "", "Source password (env: NIXCOPY_SOURCE_PASSWORD)")
	transferCmd.Flags().StringVar(&sourcePrivateKey, "source-private-key", "", "Source private key path (env: NIXCOPY_SOURCE_PRIVATE_KEY)")
	transferCmd.Flags().StringVar(&sourceRegion, "source-region", "", "Source S3 region (env: NIXCOPY_SOURCE_REGION)")
	transferCmd.Flags().StringVar(&sourceBucket, "source-bucket", "", "Source S3 bucket (env: NIXCOPY_SOURCE_BUCKET)")
	transferCmd.Flags().StringVar(&sourceAccessKey, "source-access-key", "", "Source access key (env: NIXCOPY_SOURCE_ACCESS_KEY)")
	transferCmd.Flags().StringVar(&sourceSecretKey, "source-secret-key", "", "Source secret key (env: NIXCOPY_SOURCE_SECRET_KEY)")
	transferCmd.Flags().StringVar(&sourceAuthType, "source-auth-type", "", "Source auth type (env: NIXCOPY_SOURCE_AUTH_TYPE)")
	transferCmd.Flags().StringVar(&sourceAccountName, "source-account-name", "", "Source Azure account name (env: NIXCOPY_SOURCE_ACCOUNT_NAME)")
	transferCmd.Flags().StringVar(&sourceAccountKey, "source-account-key", "", "Source Azure account key (env: NIXCOPY_SOURCE_ACCOUNT_KEY)")
	transferCmd.Flags().StringVar(&sourceContainer, "source-container", "", "Source Azure container (env: NIXCOPY_SOURCE_CONTAINER)")

	// Destination storage flags
	transferCmd.Flags().StringVar(&destType, "dest-type", "", "Destination storage type (sftp, ftps, blob, s3) (env: NIXCOPY_DEST_TYPE)")
	transferCmd.Flags().StringVar(&destHost, "dest-host", "", "Destination host (env: NIXCOPY_DEST_HOST)")
	transferCmd.Flags().IntVar(&destPort, "dest-port", 0, "Destination port (env: NIXCOPY_DEST_PORT)")
	transferCmd.Flags().StringVar(&destUsername, "dest-username", "", "Destination username (env: NIXCOPY_DEST_USERNAME)")
	transferCmd.Flags().StringVar(&destPassword, "dest-password", "", "Destination password (env: NIXCOPY_DEST_PASSWORD)")
	transferCmd.Flags().StringVar(&destPrivateKey, "dest-private-key", "", "Destination private key path (env: NIXCOPY_DEST_PRIVATE_KEY)")
	transferCmd.Flags().StringVar(&destRegion, "dest-region", "", "Destination S3 region (env: NIXCOPY_DEST_REGION)")
	transferCmd.Flags().StringVar(&destBucket, "dest-bucket", "", "Destination S3 bucket (env: NIXCOPY_DEST_BUCKET)")
	transferCmd.Flags().StringVar(&destAccessKey, "dest-access-key", "", "Destination access key (env: NIXCOPY_DEST_ACCESS_KEY)")
	transferCmd.Flags().StringVar(&destSecretKey, "dest-secret-key", "", "Destination secret key (env: NIXCOPY_DEST_SECRET_KEY)")
	transferCmd.Flags().StringVar(&destAuthType, "dest-auth-type", "", "Destination auth type (env: NIXCOPY_DEST_AUTH_TYPE)")
	transferCmd.Flags().StringVar(&destAccountName, "dest-account-name", "", "Destination Azure account name (env: NIXCOPY_DEST_ACCOUNT_NAME)")
	transferCmd.Flags().StringVar(&destAccountKey, "dest-account-key", "", "Destination Azure account key (env: NIXCOPY_DEST_ACCOUNT_KEY)")
	transferCmd.Flags().StringVar(&destContainer, "dest-container", "", "Destination Azure container (env: NIXCOPY_DEST_CONTAINER)")

	// Transfer flags
	transferCmd.Flags().IntVar(&bufferSize, "buffer-size", 0, "Buffer size in bytes (default: 32MB) (env: NIXCOPY_BUFFER_SIZE)")
	transferCmd.Flags().IntVar(&concurrentFiles, "concurrent-files", 0, "Number of concurrent file transfers (env: NIXCOPY_CONCURRENT_FILES)")
	transferCmd.Flags().IntVar(&retryAttempts, "retry-attempts", 0, "Number of retry attempts (env: NIXCOPY_RETRY_ATTEMPTS)")
	transferCmd.Flags().BoolVar(&enableResume, "resume", false, "Resume interrupted transfer if destination has a partial file (local and SFTP only) (env: NIXCOPY_ENABLE_RESUME)")
	transferCmd.Flags().BoolVar(&skipExisting, "skip-existing", false, "Skip transfer if destination already has a file with the same size (idempotent retries) (env: NIXCOPY_SKIP_EXISTING)")
	transferCmd.Flags().StringVar(&bandwidthLimit, "bandwidth-limit", "", "Max bandwidth per file (e.g. 10MB, 1GB, 512KB); 0 or empty = unlimited (env: NIXCOPY_BANDWIDTH_LIMIT)")
	transferCmd.Flags().StringVar(&compress, "compress", "", "Compress data stream before writing (gzip or zstd); empty = no compression (env: NIXCOPY_COMPRESSION)")
}

func runTransfer(cmd *cobra.Command, args []string) error {
	var cfg config.Config

	// Try to load config file if it exists
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		if err := viper.ReadInConfig(); err == nil {
			if err := viper.Unmarshal(&cfg); err != nil {
				return fmt.Errorf("failed to unmarshal config: %w", err)
			}
		}
	} else {
		// Use default config if no config file
		cfg = *config.DefaultConfig()
	}

	// Overlay environment variables (NIXCOPY_*) — env wins over config file, loses to CLI flags
	config.LoadFromEnv(&cfg)

	// Override config with CLI flags
	applyCliFlags(&cfg)

	// Parse --bandwidth-limit flag (human-readable string) and override config.
	// NIXCOPY_BANDWIDTH_LIMIT (raw int64 bytes) is already applied by LoadFromEnv above.
	if bandwidthLimit != "" {
		bw, err := usecase.ParseBandwidth(bandwidthLimit)
		if err != nil {
			return fmt.Errorf("invalid --bandwidth-limit: %w", err)
		}
		cfg.Transfer.BandwidthLimit = bw
	}

	// Validate configuration
	if err := validateConfig(&cfg); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	appID := os.Getenv("NIXCOPY_APP_ID")
	if appID == "" {
		appID = "go-nixcopy"
	}
	appVersion := os.Getenv("NIXCOPY_APP_VERSION")
	if appVersion == "" {
		appVersion = "1.0.0"
	}

	log := logger.NewStandardLogger(
		logger.WithAppID(appID),
		logger.WithAppVersion(appVersion),
		logger.WithServiceID(fmt.Sprintf("%s-to-%s", cfg.Source.Type, cfg.Destination.Type)),
		logger.WithPodName(os.Getenv("POD_NAME")),
	)

	corrID := os.Getenv("NIXCOPY_CORRELATION_ID")
	if corrID == "" {
		corrID = logger.GenerateID()
	}
	log = log.WithCorrelation(corrID, logger.GenerateID())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Warn("Received interrupt signal, canceling transfer")
		cancel()
	}()

	sourceStorage, err := storage.NewStorageFromSourceConfig(&cfg.Source)
	if err != nil {
		return fmt.Errorf("failed to create source storage: %w", err)
	}

	destStorage, err := storage.NewStorageFromDestConfig(&cfg.Destination)
	if err != nil {
		return fmt.Errorf("failed to create destination storage: %w", err)
	}

	log.InfoReqEx("Connecting to source storage", logger.F("type", string(cfg.Source.Type)))
	if err := sourceStorage.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to source: %w", err)
	}
	defer func() { _ = sourceStorage.Disconnect(ctx) }()

	log.InfoReqEx("Connecting to destination storage", logger.F("type", string(cfg.Destination.Type)))
	if err := destStorage.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to destination: %w", err)
	}
	defer func() { _ = destStorage.Disconnect(ctx) }()

	transferConfig := &entity.TransferConfig{
		BufferSize:      cfg.Transfer.BufferSize,
		ConcurrentFiles: cfg.Transfer.ConcurrentFiles,
		RetryAttempts:   cfg.Transfer.RetryAttempts,
		RetryDelay:      cfg.Transfer.RetryDelay,
		Timeout:         cfg.Transfer.Timeout,
		VerifyChecksum:  cfg.Transfer.VerifyChecksum,
		EnableResume:    cfg.Transfer.EnableResume,
		SkipExisting:    cfg.Transfer.SkipExisting,
		BandwidthLimit:  cfg.Transfer.BandwidthLimit,
		Compression:     cfg.Transfer.Compression,
	}

	transferUseCase := usecase.NewTransferUseCase(sourceStorage, destStorage, transferConfig, log)

	// Collect all source paths
	var allSourcePaths []string

	// Add single source path if provided
	if sourcePath != "" {
		allSourcePaths = append(allSourcePaths, sourcePath)
	}

	// Add multiple source paths if provided
	if len(sourcePaths) > 0 {
		allSourcePaths = append(allSourcePaths, sourcePaths...)
	}

	if len(allSourcePaths) == 0 {
		return fmt.Errorf("no source files specified, use --source or --sources")
	}

	// Expand patterns and collect all files
	var filesToTransfer []string
	patternMatcher := usecase.NewPatternMatcher(sourceStorage, log)

	for _, srcPath := range allSourcePaths {
		matchedFiles, err := patternMatcher.MatchFiles(ctx, srcPath)
		if err != nil {
			log.Warn("Failed to match pattern",
				logger.F("pattern", srcPath),
				logger.FError(err),
			)
			continue
		}
		filesToTransfer = append(filesToTransfer, matchedFiles...)
	}

	if len(filesToTransfer) == 0 {
		return fmt.Errorf("no files matched the specified patterns")
	}

	log.Info("Files to transfer",
		logger.F("count", len(filesToTransfer)),
		logger.F("concurrent", cfg.Transfer.ConcurrentFiles),
	)

	progressChan := make(chan entity.TransferProgress, 100)
	var progressWg sync.WaitGroup
	progressWg.Add(1)
	go func() {
		defer progressWg.Done()
		for progress := range progressChan {
			switch progress.Status {
			case entity.TransferStatusInProgress:
				percentage := float64(progress.TransferredBytes) / float64(progress.TotalBytes) * 100
				speedMB := progress.Speed / (1024 * 1024)
				fmt.Fprintf(os.Stderr, "\r[%s] %.2f%% | %.2f MB/s | ETA: %s",
					progress.FileName,
					percentage,
					speedMB,
					progress.EstimatedTime.Round(time.Second),
				)
			case entity.TransferStatusCompleted:
				speedMB := progress.Speed / (1024 * 1024)
				fmt.Fprintf(os.Stderr, "\r[%s] ✓ Completed | %.2f MB/s\n",
					progress.FileName,
					speedMB,
				)
			case entity.TransferStatusSkipped:
				fmt.Fprintf(os.Stderr, "\r[%s] ↷ Skipped\n", progress.FileName)
			case entity.TransferStatusFailed:
				fmt.Fprintf(os.Stderr, "\r[%s] ✗ Failed: %v\n",
					progress.FileName,
					progress.Error,
				)
			}
		}
		// Terminate any partial \r line so the shell prompt appears cleanly.
		fmt.Fprintln(os.Stderr)
	}()

	startTime := time.Now()
	var results []*entity.TransferResult

	if len(filesToTransfer) == 1 {
		// Single file transfer
		result, err := transferUseCase.Transfer(ctx, filesToTransfer[0], destPath, progressChan)
		if err != nil {
			return fmt.Errorf("transfer failed: %w", err)
		}
		results = []*entity.TransferResult{result}
	} else {
		// Batch transfer
		batchResults, err := transferUseCase.TransferBatch(ctx, filesToTransfer, destPath, progressChan)
		if err != nil {
			return fmt.Errorf("batch transfer failed: %w", err)
		}
		results = batchResults
	}

	close(progressChan)
	progressWg.Wait() // ensure all progress lines are flushed to stderr before JSON hits stdout
	totalDuration := time.Since(startTime)

	var successCount, failCount, skipCount int
	var totalBytes int64

	for _, result := range results {
		switch result.Status {
		case entity.TransferStatusCompleted:
			successCount++
			totalBytes += result.BytesTransferred
		case entity.TransferStatusSkipped:
			skipCount++
		default:
			failCount++
		}
	}

	summary := transferSummary{
		Event:            "transfer_summary",
		TotalFiles:       len(results),
		Successful:       successCount,
		Skipped:          skipCount,
		Failed:           failCount,
		BytesTransferred: totalBytes,
		DurationMs:       totalDuration.Milliseconds(),
	}

	if totalDuration.Seconds() > 0 && totalBytes > 0 {
		summary.AverageSpeedMBps = float64(totalBytes) / (1024 * 1024) / totalDuration.Seconds()
	}

	for _, result := range results {
		if result.Status == entity.TransferStatusFailed {
			errStr := ""
			if result.Error != nil {
				errStr = result.Error.Error()
			}
			summary.FailedFiles = append(summary.FailedFiles, failedFile{
				Path:  result.SourcePath,
				Error: errStr,
			})
		}
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(summary); err != nil {
		return fmt.Errorf("failed to write summary: %w", err)
	}

	if failCount > 0 {
		return fmt.Errorf("%d of %d file(s) failed to transfer", failCount, len(results))
	}

	return nil
}
