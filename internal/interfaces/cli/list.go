package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/preedep/go-nixcopy/internal/domain/repository"
	"github.com/preedep/go-nixcopy/internal/infrastructure/config"
	"github.com/preedep/go-nixcopy/internal/infrastructure/logger"
	"github.com/preedep/go-nixcopy/internal/infrastructure/storage"
)

var (
	listPath   string
	listSource bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List files in storage",
	Long:  `List files and directories in the configured storage system`,
	RunE:  runList,
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().StringVarP(&listPath, "path", "p", "/", "Path to list")
	listCmd.Flags().BoolVar(&listSource, "source", true, "List source storage (default: true)")
}

func runList(cmd *cobra.Command, args []string) error {
	if cfgFile == "" {
		cfgFile = "config.yaml"
	}

	cfg, err := config.LoadFile(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config file: %w", err)
	}

	appID := os.Getenv("NIXCOPY_APP_ID")
	if appID == "" {
		appID = "go-nixcopy"
	}
	logLevel := logger.LogLevelInfo
	if verbose {
		logLevel = logger.LogLevelDebug
	}
	log := logger.NewStandardLogger(
		logger.WithAppID(appID),
		logger.WithAppVersion(os.Getenv("NIXCOPY_APP_VERSION")),
		logger.WithPodName(os.Getenv("POD_NAME")),
		logger.WithMinLevel(logLevel),
	)

	log.Debug("config loaded",
		logger.F("config_file", cfgFile),
		logger.F("source_type", string(cfg.Source.Type)),
		logger.F("dest_type", string(cfg.Destination.Type)),
	)

	ctx := context.Background()

	log.InfoReqEx("Listing storage", logger.F("path", listPath))

	var storageSystem repository.Storage
	var storageType string

	if listSource {
		var createErr error
		storageSystem, createErr = storage.NewStorageFromSourceConfig(&cfg.Source)
		if createErr != nil {
			return fmt.Errorf("failed to create storage: %w", createErr)
		}
		storageType = string(cfg.Source.Type)
	} else {
		var createErr error
		storageSystem, createErr = storage.NewStorageFromDestConfig(&cfg.Destination)
		if createErr != nil {
			return fmt.Errorf("failed to create storage: %w", createErr)
		}
		storageType = string(cfg.Destination.Type)
	}

	if err := storageSystem.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to storage: %w", err)
	}
	defer func() { _ = storageSystem.Disconnect(ctx) }()

	files, err := storageSystem.List(ctx, listPath)
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}

	fmt.Printf("Listing %s storage at path: %s\n\n", storageType, listPath)
	fmt.Printf("%-50s %-15s %-20s %s\n", "NAME", "SIZE", "MODIFIED", "TYPE")
	fmt.Println("---------------------------------------------------------------------------------------------------")

	for _, file := range files {
		fileType := "file"
		if file.IsDirectory {
			fileType = "dir"
		}
		sizeStr := formatSize(file.Size)
		modTime := file.ModifiedTime.Format(time.RFC3339)

		fmt.Printf("%-50s %-15s %-20s %s\n", file.Name, sizeStr, modTime, fileType)
	}

	fmt.Printf("\nTotal: %d items\n", len(files))

	return nil
}

func formatSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/KB)
	default:
		return fmt.Sprintf("%d B", size)
	}
}
