package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/preedep/go-nixcopy/internal/domain/entity"
	"github.com/preedep/go-nixcopy/internal/domain/repository"
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
	sourceType            string
	sourceHost            string
	sourcePort            int
	sourceUsername        string
	sourcePassword        string
	sourcePrivateKey      string
	sourceRegion          string
	sourceBucket          string
	sourceAccessKey       string
	sourceSecretKey       string
	sourceAuthType        string
	sourceAccountName     string
	sourceAccountKey      string
	sourceContainer       string
	sourceTLSMode         string
	sourceGCSProject      string
	sourceCredentialsFile string
	sourceImpersonateSA   string
	sourceAccessToken     string

	// Destination flags
	destType            string
	destHost            string
	destPort            int
	destUsername        string
	destPassword        string
	destPrivateKey      string
	destRegion          string
	destBucket          string
	destAccessKey       string
	destSecretKey       string
	destAuthType        string
	destAccountName     string
	destAccountKey      string
	destContainer       string
	destTLSMode         string
	destGCSProject      string
	destCredentialsFile string
	destImpersonateSA   string
	destAccessToken     string

	// Transfer flags
	bufferSize      int
	concurrentFiles int
	retryAttempts   int
	enableResume    bool
	skipExisting    bool
	bandwidthLimit  string
	compress        string
)

// storageFactory abstracts storage construction so tests can inject fakes.
type storageFactory interface {
	NewSource(cfg *config.SourceConfig) (repository.Storage, error)
	NewDest(cfg *config.DestinationConfig) (repository.Storage, error)
}

type realStorageFactory struct{}

func (realStorageFactory) NewSource(cfg *config.SourceConfig) (repository.Storage, error) {
	return storage.NewStorageFromSourceConfig(cfg)
}

func (realStorageFactory) NewDest(cfg *config.DestinationConfig) (repository.Storage, error) {
	return storage.NewStorageFromDestConfig(cfg)
}

// activeStorageFactory is replaced by tests to inject mock storage.
var activeStorageFactory storageFactory = realStorageFactory{}

var transferCmd = &cobra.Command{
	Use:   "transfer",
	Short: "Transfer files between storage systems",
	Long:  `Transfer files from source to destination using streaming for memory efficiency`,
	RunE:  runTransfer,
}

// storageVarSet holds the flag variable pointers for one storage side (source or dest).
// registerStorageFlags binds all storage flags for a given prefix using this set.
type storageVarSet struct {
	storageType     *string
	host            *string
	port            *int
	username        *string
	password        *string
	privateKey      *string
	region          *string
	bucket          *string
	accessKey       *string
	secretKey       *string
	authType        *string
	accountName     *string
	accountKey      *string
	container       *string
	tlsMode         *string
	gcsProject      *string
	credentialsFile *string
	impersonateSA   *string
	accessToken     *string
}

// registerStorageFlags registers the 18 storage flags for a given side prefix ("source" or "dest").
// The env var suffix is the upper-cased prefix (e.g. NIXCOPY_SOURCE_HOST).
func registerStorageFlags(cmd *cobra.Command, prefix string, v *storageVarSet) {
	up := strings.ToUpper(prefix)
	cmd.Flags().StringVar(v.storageType, prefix+"-type", "", "Storage type (sftp, ftps, blob, s3) (env: NIXCOPY_"+up+"_TYPE)")
	cmd.Flags().StringVar(v.host, prefix+"-host", "", "Host (env: NIXCOPY_"+up+"_HOST)")
	cmd.Flags().IntVar(v.port, prefix+"-port", 0, "Port (env: NIXCOPY_"+up+"_PORT)")
	cmd.Flags().StringVar(v.username, prefix+"-username", "", "Username (env: NIXCOPY_"+up+"_USERNAME)")
	cmd.Flags().StringVar(v.password, prefix+"-password", "", "Password (env: NIXCOPY_"+up+"_PASSWORD)")
	cmd.Flags().StringVar(v.privateKey, prefix+"-private-key", "", "Private key path (env: NIXCOPY_"+up+"_PRIVATE_KEY)")
	cmd.Flags().StringVar(v.region, prefix+"-region", "", "S3 region (env: NIXCOPY_"+up+"_REGION)")
	cmd.Flags().StringVar(v.bucket, prefix+"-bucket", "", "S3/GCS bucket (env: NIXCOPY_"+up+"_BUCKET)")
	cmd.Flags().StringVar(v.accessKey, prefix+"-access-key", "", "Access key (env: NIXCOPY_"+up+"_ACCESS_KEY)")
	cmd.Flags().StringVar(v.secretKey, prefix+"-secret-key", "", "Secret key (env: NIXCOPY_"+up+"_SECRET_KEY)")
	cmd.Flags().StringVar(v.authType, prefix+"-auth-type", "", "Auth type (env: NIXCOPY_"+up+"_AUTH_TYPE)")
	cmd.Flags().StringVar(v.accountName, prefix+"-account-name", "", "Azure account name (env: NIXCOPY_"+up+"_ACCOUNT_NAME)")
	cmd.Flags().StringVar(v.accountKey, prefix+"-account-key", "", "Azure account key (env: NIXCOPY_"+up+"_ACCOUNT_KEY)")
	cmd.Flags().StringVar(v.container, prefix+"-container", "", "Azure container (env: NIXCOPY_"+up+"_CONTAINER)")
	cmd.Flags().StringVar(v.tlsMode, prefix+"-tls-mode", "", "FTPS TLS mode: explicit (STARTTLS, port 21) or implicit (TLS-first, port 990) (env: NIXCOPY_"+up+"_TLS_MODE)")
	cmd.Flags().StringVar(v.gcsProject, prefix+"-gcs-project", "", "GCS project ID (env: NIXCOPY_"+up+"_GCS_PROJECT)")
	cmd.Flags().StringVar(v.credentialsFile, prefix+"-credentials-file", "", "GCS service account JSON key file path (env: NIXCOPY_"+up+"_CREDENTIALS_FILE)")
	cmd.Flags().StringVar(v.impersonateSA, prefix+"-impersonate-service-account", "", "GCS service account to impersonate (env: NIXCOPY_"+up+"_IMPERSONATE_SA)")
	cmd.Flags().StringVar(v.accessToken, prefix+"-access-token", "", "GCS short-lived OAuth2 access token (env: NIXCOPY_"+up+"_ACCESS_TOKEN)")
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

	registerStorageFlags(transferCmd, "source", &storageVarSet{
		storageType:     &sourceType,
		host:            &sourceHost,
		port:            &sourcePort,
		username:        &sourceUsername,
		password:        &sourcePassword,
		privateKey:      &sourcePrivateKey,
		region:          &sourceRegion,
		bucket:          &sourceBucket,
		accessKey:       &sourceAccessKey,
		secretKey:       &sourceSecretKey,
		authType:        &sourceAuthType,
		accountName:     &sourceAccountName,
		accountKey:      &sourceAccountKey,
		container:       &sourceContainer,
		tlsMode:         &sourceTLSMode,
		gcsProject:      &sourceGCSProject,
		credentialsFile: &sourceCredentialsFile,
		impersonateSA:   &sourceImpersonateSA,
		accessToken:     &sourceAccessToken,
	})

	registerStorageFlags(transferCmd, "dest", &storageVarSet{
		storageType:     &destType,
		host:            &destHost,
		port:            &destPort,
		username:        &destUsername,
		password:        &destPassword,
		privateKey:      &destPrivateKey,
		region:          &destRegion,
		bucket:          &destBucket,
		accessKey:       &destAccessKey,
		secretKey:       &destSecretKey,
		authType:        &destAuthType,
		accountName:     &destAccountName,
		accountKey:      &destAccountKey,
		container:       &destContainer,
		tlsMode:         &destTLSMode,
		gcsProject:      &destGCSProject,
		credentialsFile: &destCredentialsFile,
		impersonateSA:   &destImpersonateSA,
		accessToken:     &destAccessToken,
	})

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

	if cfgFile != "" {
		loaded, err := config.LoadFile(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config file: %w", err)
		}
		cfg = *loaded
	} else {
		cfg = *config.DefaultConfig()
	}

	// Overlay environment variables (NIXCOPY_*) — env wins over config file, loses to CLI flags.
	// LoadFromEnv sets types from NIXCOPY_SOURCE/DEST_TYPE, then calls ApplyBackendEnv internally.
	config.LoadFromEnv(&cfg)

	// If --source-type or --dest-type was given, it overrides NIXCOPY_SOURCE/DEST_TYPE.
	// Re-apply backend env vars so backend-specific vars (e.g. NIXCOPY_SOURCE_HOST) are
	// picked up for the CLI-supplied type even when NIXCOPY_SOURCE_TYPE was not set.
	if sourceType != "" {
		cfg.Source.Type = config.StorageType(sourceType)
	}
	if destType != "" {
		cfg.Destination.Type = config.StorageType(destType)
	}
	config.ApplyBackendEnv(&cfg)

	// Override config with CLI flags (individual field values win over env vars)
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

	appID := os.Getenv("NIXCOPY_APP_ID")
	if appID == "" {
		appID = "go-nixcopy"
	}
	appVersion := os.Getenv("NIXCOPY_APP_VERSION")
	if appVersion == "" {
		appVersion = "1.0.0"
	}

	logLevel := logger.LogLevelInfo
	if verbose {
		logLevel = logger.LogLevelDebug
	}
	log := logger.NewStandardLogger(
		logger.WithAppID(appID),
		logger.WithAppVersion(appVersion),
		logger.WithServiceID(fmt.Sprintf("%s-to-%s", cfg.Source.Type, cfg.Destination.Type)),
		logger.WithPodName(os.Getenv("POD_NAME")),
		logger.WithMinLevel(logLevel),
	)

	corrID := os.Getenv("NIXCOPY_CORRELATION_ID")
	if corrID == "" {
		corrID = logger.GenerateID()
	}
	log = log.WithCorrelation(corrID, logger.GenerateID())

	log.Debug("config resolved",
		logger.F("source_type", string(cfg.Source.Type)),
		logger.F("dest_type", string(cfg.Destination.Type)),
		logger.F("buffer_size", cfg.Transfer.BufferSize),
		logger.F("concurrent_files", cfg.Transfer.ConcurrentFiles),
		logger.F("retry_attempts", cfg.Transfer.RetryAttempts),
		logger.F("retry_delay", cfg.Transfer.RetryDelay.String()),
		logger.F("verify_checksum", cfg.Transfer.VerifyChecksum),
		logger.F("enable_resume", cfg.Transfer.EnableResume),
		logger.F("skip_existing", cfg.Transfer.SkipExisting),
		logger.F("bandwidth_limit", cfg.Transfer.BandwidthLimit),
		logger.F("compression", cfg.Transfer.Compression),
	)

	// Validate configuration
	if err := validateConfig(&cfg); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Config is valid — suppress usage for all subsequent runtime errors.
	cmd.SilenceUsage = true

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Warn("Received interrupt signal, canceling transfer")
		cancel()
	}()

	sourceStorage, err := activeStorageFactory.NewSource(&cfg.Source)
	if err != nil {
		log.ErrorReqEx("failed to initialize source storage",
			logger.F("type", string(cfg.Source.Type)),
			logger.FError(err),
		)
		return fmt.Errorf("failed to create source storage: %w", err)
	}

	destStorage, err := activeStorageFactory.NewDest(&cfg.Destination)
	if err != nil {
		log.ErrorReqEx("failed to initialize destination storage",
			logger.F("type", string(cfg.Destination.Type)),
			logger.FError(err),
		)
		return fmt.Errorf("failed to create destination storage: %w", err)
	}

	log.InfoReqEx("Connecting to source storage", logger.F("type", string(cfg.Source.Type)))
	if err := sourceStorage.Connect(ctx); err != nil {
		log.ErrorReqEx("failed to connect to source storage",
			logger.F("type", string(cfg.Source.Type)),
			logger.FError(err),
		)
		return fmt.Errorf("failed to connect to source: %w", err)
	}
	defer func() { _ = sourceStorage.Disconnect(ctx) }()

	log.InfoReqEx("Connecting to destination storage", logger.F("type", string(cfg.Destination.Type)))
	if err := destStorage.Connect(ctx); err != nil {
		log.ErrorReqEx("failed to connect to destination storage",
			logger.F("type", string(cfg.Destination.Type)),
			logger.FError(err),
		)
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

	expandTransferPaths()

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
	for _, f := range filesToTransfer {
		log.Debug("queued file", logger.F("path", f))
	}

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
		resolvedDest := resolveDestPath(filesToTransfer[0], destPath)
		result, err := transferUseCase.Transfer(ctx, filesToTransfer[0], resolvedDest, progressChan)
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

// resolveDestPath returns the effective destination path for a single-file transfer.
// If destPath ends with "/" it is treated as a directory and the source filename is
// appended, mirroring Unix cp behaviour: cp file.txt /dir/ → /dir/file.txt.
func resolveDestPath(srcPath, destPath string) string {
	if strings.HasSuffix(destPath, "/") {
		return destPath + filepath.Base(srcPath)
	}
	return destPath
}

// expandTransferPaths expands ${ENV_VAR} references in the path flags so users
// can write --source "${PWD}/data" or --dest "${OUTDIR}/result" in scripts.
func expandTransferPaths() {
	sourcePath = os.ExpandEnv(sourcePath)
	for i, p := range sourcePaths {
		sourcePaths[i] = os.ExpandEnv(p)
	}
	destPath = os.ExpandEnv(destPath)
}
