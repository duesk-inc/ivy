package service

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// S3Service ファイルストレージサービスインターフェース
type S3Service interface {
	UploadFile(ctx context.Context, file *multipart.FileHeader) (string, error)
}

// MockS3Service ローカルファイルシステムを使ったモックS3サービス
type MockS3Service struct {
	basePath string
	logger   *zap.Logger
}

// NewMockS3Service モックS3サービスを作成
func NewMockS3Service(logger *zap.Logger) *MockS3Service {
	basePath := "./tmp/mock_s3"
	os.MkdirAll(basePath, 0755)
	return &MockS3Service{
		basePath: basePath,
		logger:   logger,
	}
}

// UploadFile ファイルをローカルに保存
func (s *MockS3Service) UploadFile(ctx context.Context, file *multipart.FileHeader) (string, error) {
	ext := filepath.Ext(file.Filename)
	now := time.Now()
	key := fmt.Sprintf("uploads/%s/%s/%s%s",
		now.Format("2006"),
		now.Format("01"),
		uuid.New().String(),
		ext,
	)

	fullPath := filepath.Join(s.basePath, key)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return "", fmt.Errorf("ディレクトリ作成失敗: %w", err)
	}

	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("ファイルを開けません: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("ファイル作成失敗: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("ファイルコピー失敗: %w", err)
	}

	s.logger.Info("MockS3: ファイルアップロード完了",
		zap.String("key", key),
		zap.String("filename", file.Filename),
	)

	return key, nil
}
