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
)

var transferCmd = &cobra.Command{
	Use:   "transfer",
	Short: "Transfer files between storage systems",
	Long:  `Transfer files from source to destination using streaming for memory efficiency`,
	RunE:  runTransfer,
}

func init() {
	rootCmd.AddCommand(transferCmd)

	transferCmd.Flags().StringVarP(&sourcePath, "source", "s", "", "Source file path (required)")
	transferCmd.Flags().StringVarP(&destPath, "dest", "d", "", "Destination file path (required)")
	transferCmd.MarkFlagRequired("source")
	transferCmd.MarkFlagRequired("dest")
}

func runTransfer(cmd *cobra.Command, args []string) error {
	if cfgFile == "" {
		cfgFile = "config.yaml"
	}

	viper.SetConfigFile(cfgFile)
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg config.Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
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
