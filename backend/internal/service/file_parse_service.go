package service

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"unicode"

	gopdf "github.com/ledongthuc/pdf"
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

	data, err := io.ReadAll(src)
	if err != nil {
		return "", nil, fmt.Errorf("ファイルの読み取りに失敗しました: %w", err)
	}

	reader := bytes.NewReader(data)
	pdfReader, err := gopdf.NewReader(reader, int64(len(data)))
	if err != nil {
		return "", nil, fmt.Errorf("PDFファイルの読み取りに失敗しました: %w", err)
	}

	var sb strings.Builder
	var warnings []string

	for i := 1; i <= pdfReader.NumPage(); i++ {
		page := pdfReader.Page(i)
		if page.V.IsNull() {
			continue
		}
		text, err := page.GetPlainText(nil)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("ページ%dの読み取りに一部失敗しました", i))
			continue
		}
		sb.WriteString(text)
		sb.WriteString("\n")
	}

	text := sanitizeText(strings.TrimSpace(sb.String()))

	if text == "" {
		warnings = append(warnings, "テキストを抽出できませんでした。スキャン画像のPDFには対応していません。テキスト入力で補完してください")
	}

	return text, warnings, nil
}

// sanitizeText nullバイトや非表示制御文字を除去
func sanitizeText(s string) string {
	return strings.Map(func(r rune) rune {
		if r == 0 {
			return -1
		}
		if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
			return -1
		}
		return r
	}, s)
}
