package handler

import (
	"net/http"

	"github.com/duesk/ivy/internal/dto"
	"github.com/duesk/ivy/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// FileHandler ファイルハンドラー
type FileHandler struct {
	fileParseService service.FileParseService
	s3Service        service.S3Service
	logger           *zap.Logger
}

// NewFileHandler ファイルハンドラーを作成
func NewFileHandler(fileParseService service.FileParseService, s3Service service.S3Service, logger *zap.Logger) *FileHandler {
	return &FileHandler{
		fileParseService: fileParseService,
		s3Service:        s3Service,
		logger:           logger,
	}
}

// Parse ファイルアップロード→テキスト抽出
func (h *FileHandler) Parse(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "ファイルが必要です"})
		return
	}

	// バリデーション
	if err := h.fileParseService.ValidateFile(file); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	// S3にアップロード
	fileKey, err := h.s3Service.UploadFile(c.Request.Context(), file)
	if err != nil {
		h.logger.Error("ファイルアップロード失敗", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "ファイルのアップロードに失敗しました"})
		return
	}

	// テキスト抽出
	text, warnings, err := h.fileParseService.ParseFile(file)
	if err != nil {
		h.logger.Error("ファイルパース失敗", zap.Error(err), zap.String("filename", file.Filename))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.FileParseResponse{
		Text:          text,
		FileKey:       fileKey,
		FileName:      file.Filename,
		ParseWarnings: warnings,
	})
}
