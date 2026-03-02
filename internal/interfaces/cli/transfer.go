package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/preedep/go-nixcopy/internal/domain/entity"
	"github.com/preedep/go-nixcopy/internal/infrastructure/config"
	"github.com/preedep/go-nixcopy/internal/infrastructure/logger"
	"github.com/preedep/go-nixcopy/internal/infrastructure/storage"
	"github.com/preedep/go-nixcopy/internal/usecase"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	sourcePath string
	destPath   string

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
	transferCmd.Flags().StringVarP(&sourcePath, "source", "s", "", "Source file path (required)")
	transferCmd.Flags().StringVarP(&destPath, "dest", "d", "", "Destination file path (required)")
	transferCmd.MarkFlagRequired("source")
	transferCmd.MarkFlagRequired("dest")

	// Source storage flags
	transferCmd.Flags().StringVar(&sourceType, "source-type", "", "Source storage type (sftp, ftps, blob, s3)")
	transferCmd.Flags().StringVar(&sourceHost, "source-host", "", "Source host")
	transferCmd.Flags().IntVar(&sourcePort, "source-port", 0, "Source port")
	transferCmd.Flags().StringVar(&sourceUsername, "source-username", "", "Source username")
	transferCmd.Flags().StringVar(&sourcePassword, "source-password", "", "Source password")
	transferCmd.Flags().StringVar(&sourcePrivateKey, "source-private-key", "", "Source private key path")
	transferCmd.Flags().StringVar(&sourceRegion, "source-region", "", "Source S3 region")
	transferCmd.Flags().StringVar(&sourceBucket, "source-bucket", "", "Source S3 bucket")
	transferCmd.Flags().StringVar(&sourceAccessKey, "source-access-key", "", "Source access key")
	transferCmd.Flags().StringVar(&sourceSecretKey, "source-secret-key", "", "Source secret key")
	transferCmd.Flags().StringVar(&sourceAuthType, "source-auth-type", "", "Source auth type")
	transferCmd.Flags().StringVar(&sourceAccountName, "source-account-name", "", "Source Azure account name")
	transferCmd.Flags().StringVar(&sourceAccountKey, "source-account-key", "", "Source Azure account key")
	transferCmd.Flags().StringVar(&sourceContainer, "source-container", "", "Source Azure container")

	// Destination storage flags
	transferCmd.Flags().StringVar(&destType, "dest-type", "", "Destination storage type (sftp, ftps, blob, s3)")
	transferCmd.Flags().StringVar(&destHost, "dest-host", "", "Destination host")
	transferCmd.Flags().IntVar(&destPort, "dest-port", 0, "Destination port")
	transferCmd.Flags().StringVar(&destUsername, "dest-username", "", "Destination username")
	transferCmd.Flags().StringVar(&destPassword, "dest-password", "", "Destination password")
	transferCmd.Flags().StringVar(&destPrivateKey, "dest-private-key", "", "Destination private key path")
	transferCmd.Flags().StringVar(&destRegion, "dest-region", "", "Destination S3 region")
	transferCmd.Flags().StringVar(&destBucket, "dest-bucket", "", "Destination S3 bucket")
	transferCmd.Flags().StringVar(&destAccessKey, "dest-access-key", "", "Destination access key")
	transferCmd.Flags().StringVar(&destSecretKey, "dest-secret-key", "", "Destination secret key")
	transferCmd.Flags().StringVar(&destAuthType, "dest-auth-type", "", "Destination auth type")
	transferCmd.Flags().StringVar(&destAccountName, "dest-account-name", "", "Destination Azure account name")
	transferCmd.Flags().StringVar(&destAccountKey, "dest-account-key", "", "Destination Azure account key")
	transferCmd.Flags().StringVar(&destContainer, "dest-container", "", "Destination Azure container")

	// Transfer flags
	transferCmd.Flags().IntVar(&bufferSize, "buffer-size", 0, "Buffer size in bytes (default: 32MB)")
	transferCmd.Flags().IntVar(&concurrentFiles, "concurrent-files", 0, "Number of concurrent file transfers")
	transferCmd.Flags().IntVar(&retryAttempts, "retry-attempts", 0, "Number of retry attempts")
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

	// Override config with CLI flags
	if err := applyCliFlags(&cfg); err != nil {
		return fmt.Errorf("failed to apply CLI flags: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&cfg); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	log, err := logger.NewLogger(&cfg.Logging)
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}
	defer log.Sync()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Info("Received interrupt signal, canceling transfer...")
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

	log.Info("Connecting to source storage", zap.String("type", string(cfg.Source.Type)))
	if err := sourceStorage.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to source: %w", err)
	}
	defer sourceStorage.Disconnect(ctx)

	log.Info("Connecting to destination storage", zap.String("type", string(cfg.Destination.Type)))
	if err := destStorage.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to destination: %w", err)
	}
	defer destStorage.Disconnect(ctx)

	transferConfig := &entity.TransferConfig{
		BufferSize:      cfg.Transfer.BufferSize,
		ConcurrentFiles: cfg.Transfer.ConcurrentFiles,
		RetryAttempts:   cfg.Transfer.RetryAttempts,
		RetryDelay:      cfg.Transfer.RetryDelay,
		Timeout:         cfg.Transfer.Timeout,
		VerifyChecksum:  cfg.Transfer.VerifyChecksum,
	}

	transferUseCase := usecase.NewTransferUseCase(sourceStorage, destStorage, transferConfig, log)

	progressChan := make(chan entity.TransferProgress, 10)
	go func() {
		for progress := range progressChan {
			if progress.Status == entity.TransferStatusInProgress {
				percentage := float64(progress.TransferredBytes) / float64(progress.TotalBytes) * 100
				speedMB := progress.Speed / (1024 * 1024)
				fmt.Printf("\r[%s] %.2f%% | %.2f MB/s | ETA: %s",
					progress.FileName,
					percentage,
					speedMB,
					progress.EstimatedTime.Round(time.Second),
				)
			} else if progress.Status == entity.TransferStatusCompleted {
				speedMB := progress.Speed / (1024 * 1024)
				fmt.Printf("\r[%s] ✓ Completed | %.2f MB/s\n",
					progress.FileName,
					speedMB,
				)
			} else if progress.Status == entity.TransferStatusFailed {
				fmt.Printf("\r[%s] ✗ Failed: %v\n",
					progress.FileName,
					progress.Error,
				)
			}
		}
	}()

	result, err := transferUseCase.Transfer(ctx, sourcePath, destPath, progressChan)
	if err != nil {
		return fmt.Errorf("transfer failed: %w", err)
	}

	fmt.Printf("\nTransfer Summary:\n")
	fmt.Printf("  Source: %s\n", result.SourcePath)
	fmt.Printf("  Destination: %s\n", result.DestinationPath)
	fmt.Printf("  Bytes Transferred: %d (%.2f MB)\n", result.BytesTransferred, float64(result.BytesTransferred)/(1024*1024))
	fmt.Printf("  Duration: %s\n", result.Duration.Round(time.Millisecond))
	fmt.Printf("  Average Speed: %.2f MB/s\n", float64(result.BytesTransferred)/(1024*1024)/result.Duration.Seconds())
	fmt.Printf("  Status: %s\n", result.Status)

	return nil
}
