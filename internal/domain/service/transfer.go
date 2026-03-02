package service

import (
	"context"

	"github.com/preedep/go-nixcopy/internal/domain/entity"
)

type TransferService interface {
	Transfer(ctx context.Context, sourcePath, destPath string, progressChan chan<- entity.TransferProgress) (*entity.TransferResult, error)
	TransferBatch(ctx context.Context, sourcePaths []string, destBasePath string, progressChan chan<- entity.TransferProgress) ([]*entity.TransferResult, error)
}

type ProgressCallback func(progress entity.TransferProgress)
