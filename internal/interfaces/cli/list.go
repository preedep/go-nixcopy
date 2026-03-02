package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/preedep/go-nixcopy/internal/infrastructure/config"
	"github.com/preedep/go-nixcopy/internal/infrastructure/logger"
	"github.com/preedep/go-nixcopy/internal/infrastructure/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

	ctx := context.Background()

	var storageSystem storage.Storage
	var storageType string

	if listSource {
		storageSystem, err = storage.NewStorageFromSourceConfig(&cfg.Source)
		storageType = string(cfg.Source.Type)
	} else {
		storageSystem, err = storage.NewStorageFromDestConfig(&cfg.Destination)
		storageType = string(cfg.Destination.Type)
	}

	if err != nil {
		return fmt.Errorf("failed to create storage: %w", err)
	}

	if err := storageSystem.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to storage: %w", err)
	}
	defer storageSystem.Disconnect(ctx)

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
