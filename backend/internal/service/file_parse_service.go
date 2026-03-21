package service

import (
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
)

const maxFileSize = 10 * 1024 * 1024 // 10MB

// FileParseService ファイルパースサービスインターフェース
type FileParseService interface {
	ParseFile(file *multipart.FileHeader) (text string, warnings []string, err error)
	ValidateFile(file *multipart.FileHeader) error
}

type fileParseService struct {
	logger *zap.Logger
}

// NewFileParseService ファイルパースサービスを作成
func NewFileParseService(logger *zap.Logger) FileParseService {
	return &fileParseService{logger: logger}
}

// ValidateFile ファイルのバリデーション（設計書セクション4.9）
func (s *fileParseService) ValidateFile(file *multipart.FileHeader) error {
	// サイズチェック
	if file.Size > maxFileSize {
		return fmt.Errorf("ファイルサイズが大きすぎます（上限10MB）")
	}

	// 拡張子チェック
	ext := strings.ToLower(filepath.Ext(file.Filename))
	switch ext {
	case ".xlsx", ".xls", ".pdf":
		// OK
	default:
		return fmt.Errorf("対応していないファイル形式です。Excel(.xlsx/.xls)またはPDF(.pdf)をアップロードしてください。")
	}

	// マジックバイト検証（Content-Type + マジックバイト）
	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("ファイルが破損しています")
	}
	defer src.Close()

	header := make([]byte, 8)
	n, err := src.Read(header)
	if err != nil || n < 4 {
		return fmt.Errorf("ファイルが破損しています")
	}

	switch ext {
	case ".xlsx":
		// ZIP format (PK magic bytes): 50 4B 03 04
		if header[0] != 0x50 || header[1] != 0x4B || header[2] != 0x03 || header[3] != 0x04 {
			return fmt.Errorf("対応していないファイル形式です。Excel(.xlsx/.xls)またはPDF(.pdf)をアップロードしてください。")
		}
	case ".xls":
		// OLE2 Compound Document: D0 CF 11 E0
		if header[0] != 0xD0 || header[1] != 0xCF || header[2] != 0x11 || header[3] != 0xE0 {
			return fmt.Errorf("対応していないファイル形式です。Excel(.xlsx/.xls)またはPDF(.pdf)をアップロードしてください。")
		}
	case ".pdf":
		// PDF magic: %PDF
		if string(header[:4]) != "%PDF" {
			return fmt.Errorf("対応していないファイル形式です。Excel(.xlsx/.xls)またはPDF(.pdf)をアップロードしてください。")
		}
	}

	return nil
}

// ParseFile ファイルからテキストを抽出
func (s *fileParseService) ParseFile(file *multipart.FileHeader) (string, []string, error) {
	ext := strings.ToLower(filepath.Ext(file.Filename))

	switch ext {
	case ".xlsx", ".xls":
		return s.parseExcel(file)
	case ".pdf":
		return s.parsePDF(file)
	default:
		return "", nil, fmt.Errorf("対応していないファイル形式です")
	}
}

// parseExcel Excelファイルからテキストを抽出
func (s *fileParseService) parseExcel(file *multipart.FileHeader) (string, []string, error) {
	src, err := file.Open()
	if err != nil {
		return "", nil, fmt.Errorf("ファイルを開けません: %w", err)
	}
	defer src.Close()

	f, err := excelize.OpenReader(src)
	if err != nil {
		return "", nil, fmt.Errorf("Excelファイルの読み取りに失敗しました: %w", err)
	}
	defer f.Close()

	var sb strings.Builder
	var warnings []string

	for _, sheet := range f.GetSheetList() {
		rows, err := f.GetRows(sheet)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("シート '%s' の読み取りに一部失敗しました", sheet))
			continue
		}

		sb.WriteString(fmt.Sprintf("=== %s ===\n", sheet))
		for _, row := range rows {
			nonEmpty := make([]string, 0)
			for _, cell := range row {
				trimmed := strings.TrimSpace(cell)
				if trimmed != "" {
					nonEmpty = append(nonEmpty, trimmed)
				}
			}
			if len(nonEmpty) > 0 {
				sb.WriteString(strings.Join(nonEmpty, " | "))
				sb.WriteString("\n")
			}
		}
		sb.WriteString("\n")
	}

	text := strings.TrimSpace(sb.String())
	if text == "" {
		warnings = append(warnings, "テキストを抽出できませんでした。テキスト入力で補完してください")
	}

	return text, warnings, nil
}

// parsePDF PDFファイルからテキストを抽出
func (s *fileParseService) parsePDF(file *multipart.FileHeader) (string, []string, error) {
	src, err := file.Open()
	if err != nil {
		return "", nil, fmt.Errorf("ファイルを開けません: %w", err)
	}
	defer src.Close()

	// PDFのバイトを読み込み
	data, err := io.ReadAll(src)
	if err != nil {
		return "", nil, fmt.Errorf("ファイルの読み取りに失敗しました: %w", err)
	}

	// マジックバイトでPDFか確認
	if len(data) < 4 || string(data[:4]) != "%PDF" {
		return "", nil, fmt.Errorf("ファイルが破損しています")
	}

	// 簡易テキスト抽出（Phase 1ではpdfcpuの完全統合は後回し）
	// テキストストリームからの基本的な抽出を試みる
	text := extractPDFText(data)
	var warnings []string

	if text == "" {
		warnings = append(warnings, "テキストを抽出できませんでした。スキャン画像のPDFには対応していません。テキスト入力で補完してください")
	}

	return text, warnings, nil
}

// extractPDFText PDFからテキストを簡易抽出
func extractPDFText(data []byte) string {
	// Phase 1: 基本的なテキストストリーム抽出
	// テキストベースのPDFから文字列を抽出する簡易実装
	content := string(data)
	var sb strings.Builder

	// BT...ET（テキストブロック）内のTj/TJ演算子からテキストを抽出
	inText := false
	for i := 0; i < len(content)-1; i++ {
		if i+1 < len(content) && content[i] == 'B' && content[i+1] == 'T' {
			inText = true
			continue
		}
		if i+1 < len(content) && content[i] == 'E' && content[i+1] == 'T' {
			inText = false
			sb.WriteString("\n")
			continue
		}
		if inText {
			// (text) Tj パターンを探す
			if content[i] == '(' {
				j := i + 1
				for j < len(content) && content[j] != ')' {
					sb.WriteByte(content[j])
					j++
				}
				i = j
			}
		}
	}

	return strings.TrimSpace(sb.String())
}
