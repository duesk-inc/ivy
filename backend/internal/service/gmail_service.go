package service

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/duesk/ivy/internal/config"
	"github.com/duesk/ivy/internal/dto"
	"github.com/duesk/ivy/internal/model"
	"github.com/duesk/ivy/internal/repository"
	"go.uber.org/zap"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// GmailService Gmail連携サービスインターフェース
type GmailService interface {
	SyncEmails(ctx context.Context) (*dto.EmailSyncResponse, error)
	GetSyncState(ctx context.Context) (*model.GmailSyncState, error)
}

type gmailService struct {
	gmailSyncStateRepo  repository.GmailSyncStateRepository
	processedEmailRepo  repository.ProcessedEmailRepository
	emailAIService      EmailAIService
	jobRepo             repository.JobRepository
	engineerProfileRepo repository.EngineerProfileRepository
	settingsRepo        repository.SettingsRepository
	fileParseService    FileParseService
	config              *config.Config
	logger              *zap.Logger
}

// NewGmailService GmailServiceを作成
func NewGmailService(
	gmailSyncStateRepo repository.GmailSyncStateRepository,
	processedEmailRepo repository.ProcessedEmailRepository,
	emailAIService EmailAIService,
	jobRepo repository.JobRepository,
	engineerProfileRepo repository.EngineerProfileRepository,
	settingsRepo repository.SettingsRepository,
	fileParseService FileParseService,
	cfg *config.Config,
	logger *zap.Logger,
) GmailService {
	return &gmailService{
		gmailSyncStateRepo:  gmailSyncStateRepo,
		processedEmailRepo:  processedEmailRepo,
		emailAIService:      emailAIService,
		jobRepo:             jobRepo,
		engineerProfileRepo: engineerProfileRepo,
		settingsRepo:        settingsRepo,
		fileParseService:    fileParseService,
		config:              cfg,
		logger:              logger,
	}
}

// SyncEmails Gmailからメールを同期し、案件・人材情報を抽出
func (s *gmailService) SyncEmails(ctx context.Context) (*dto.EmailSyncResponse, error) {
	if !s.config.Gmail.Enabled {
		return nil, fmt.Errorf("Gmail連携が無効です")
	}

	syncState, err := s.gmailSyncStateRepo.GetOrCreate(ctx)
	if err != nil {
		return nil, fmt.Errorf("同期状態の取得に失敗しました: %w", err)
	}

	gmailClient, err := s.createGmailClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("Gmail APIクライアントの作成に失敗しました: %w", err)
	}

	query := s.buildSearchQuery(syncState.LastSyncedAt)
	s.logger.Info("Gmail同期開始", zap.String("query", query))

	messages, err := s.listMessages(ctx, gmailClient, query)
	if err != nil {
		return nil, fmt.Errorf("メール一覧の取得に失敗しました: %w", err)
	}

	s.logger.Info("同期対象メール件数", zap.Int("count", len(messages)))

	response := &dto.EmailSyncResponse{
		SyncedAt: time.Now(),
	}

	for _, msg := range messages {
		response.TotalProcessed++

		processErr := s.processMessage(ctx, gmailClient, msg.Id, response)
		if processErr != nil {
			s.logger.Warn("メール処理失敗",
				zap.String("message_id", msg.Id),
				zap.Error(processErr),
			)
			response.Errors++
		}
	}

	now := time.Now()
	if err := s.gmailSyncStateRepo.UpdateLastSync(ctx, syncState.LastHistoryID, now); err != nil {
		s.logger.Error("同期状態の更新に失敗しました", zap.Error(err))
	}

	s.logger.Info("Gmail同期完了",
		zap.Int("total_processed", response.TotalProcessed),
		zap.Int("new_jobs", response.NewJobs),
		zap.Int("new_engineers", response.NewEngineers),
		zap.Int("duplicates_skipped", response.DuplicatesSkipped),
		zap.Int("other_skipped", response.OtherSkipped),
		zap.Int("errors", response.Errors),
	)

	return response, nil
}

// GetSyncState 同期状態を取得
func (s *gmailService) GetSyncState(ctx context.Context) (*model.GmailSyncState, error) {
	state, err := s.gmailSyncStateRepo.GetOrCreate(ctx)
	if err != nil {
		return nil, fmt.Errorf("同期状態の取得に失敗しました: %w", err)
	}
	return state, nil
}

// createGmailClient サービスアカウントを使用してGmail APIクライアントを作成
func (s *gmailService) createGmailClient(ctx context.Context) (*gmail.Service, error) {
	keyFile := s.config.Gmail.ServiceAccountKeyFile
	if keyFile == "" {
		return nil, fmt.Errorf("サービスアカウントキーファイルが設定されていません")
	}

	creds, err := google.JWTConfigFromJSON([]byte(keyFile), gmail.GmailReadonlyScope)
	if err != nil {
		// ファイルパスとして読み込む場合
		return nil, fmt.Errorf("サービスアカウント認証情報の読み込みに失敗しました: %w", err)
	}

	targetEmail := s.config.Gmail.TargetEmail
	if targetEmail == "" {
		return nil, fmt.Errorf("対象メールアドレスが設定されていません")
	}
	creds.Subject = targetEmail

	client := creds.Client(ctx)
	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("Gmail APIサービスの作成に失敗しました: %w", err)
	}

	return srv, nil
}

// buildSearchQuery 検索クエリを構築
func (s *gmailService) buildSearchQuery(lastSyncedAt time.Time) string {
	if lastSyncedAt.IsZero() {
		// 初回同期: 過去7日間
		since := time.Now().AddDate(0, 0, -7)
		return fmt.Sprintf("after:%d", since.Unix())
	}
	return fmt.Sprintf("after:%d", lastSyncedAt.Unix())
}

// listMessages メール一覧を取得
func (s *gmailService) listMessages(ctx context.Context, srv *gmail.Service, query string) ([]*gmail.Message, error) {
	var allMessages []*gmail.Message
	pageToken := ""

	for {
		req := srv.Users.Messages.List("me").Q(query).MaxResults(100)
		if pageToken != "" {
			req = req.PageToken(pageToken)
		}

		resp, err := req.Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("メール一覧API呼び出し失敗: %w", err)
		}

		allMessages = append(allMessages, resp.Messages...)

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return allMessages, nil
}

// processMessage 個別メールを処理
func (s *gmailService) processMessage(ctx context.Context, srv *gmail.Service, messageID string, response *dto.EmailSyncResponse) error {
	msg, err := srv.Users.Messages.Get("me", messageID).Format("full").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("メール詳細取得失敗: %w", err)
	}

	sender := extractHeader(msg.Payload, "From")
	subject := extractHeader(msg.Payload, "Subject")
	body := extractBody(msg.Payload)

	contentHash := computeContentHash(sender, subject, body)

	exists, err := s.processedEmailRepo.ExistsByContentHash(ctx, contentHash)
	if err != nil {
		return fmt.Errorf("重複チェック失敗: %w", err)
	}
	if exists {
		response.DuplicatesSkipped++
		return nil
	}

	classification, err := s.emailAIService.ClassifyEmail(ctx, subject, body, sender)
	if err != nil {
		return fmt.Errorf("メール分類失敗: %w", err)
	}

	emailText := fmt.Sprintf("件名: %s\n送信者: %s\n\n%s", subject, sender, body)

	switch classification.Classification {
	case "job":
		if err := s.processJobEmail(ctx, srv, msg, emailText, contentHash); err != nil {
			return fmt.Errorf("案件メール処理失敗: %w", err)
		}
		response.NewJobs++

	case "engineer":
		attachmentText := s.extractAttachmentText(ctx, srv, msg)
		if err := s.processEngineerEmail(ctx, emailText, attachmentText, contentHash, messageID); err != nil {
			return fmt.Errorf("人材メール処理失敗: %w", err)
		}
		response.NewEngineers++

	default:
		response.OtherSkipped++
	}

	processedEmail := &model.ProcessedEmail{
		ContentHash:    contentHash,
		GmailMessageID: messageID,
		Classification: model.EmailClassification(classification.Classification),
		ProcessedAt:    time.Now(),
	}
	if err := s.processedEmailRepo.Create(ctx, processedEmail); err != nil {
		s.logger.Warn("処理済みメール記録失敗", zap.Error(err))
	}

	return nil
}

// processJobEmail 案件メールを処理してJobを作成
func (s *gmailService) processJobEmail(ctx context.Context, srv *gmail.Service, msg *gmail.Message, emailText, contentHash string) error {
	parsed, err := s.emailAIService.ParseJobFromEmail(ctx, emailText)
	if err != nil {
		return fmt.Errorf("案件情報パース失敗: %w", err)
	}

	var parsedData model.ParsedJobData
	if err := json.Unmarshal(parsed, &parsedData); err != nil {
		s.logger.Warn("パース結果のデコード失敗", zap.Error(err))
	}

	expiresAt := computeJobExpiresAt(parsedData.StartMonth)

	job := &model.Job{
		ContentHash:   contentHash,
		SourceEmailID: msg.Id,
		RawText:       emailText,
		Parsed:        parsed,
		StartMonth:    parsedData.StartMonth,
		Status:        model.JobStatusActive,
		ExpiresAt:     expiresAt,
	}

	if err := s.jobRepo.Create(ctx, job); err != nil {
		return fmt.Errorf("案件作成失敗: %w", err)
	}

	s.logger.Info("案件作成成功", zap.String("job_id", job.ID))
	return nil
}

// processEngineerEmail 人材メールを処理してEngineerProfileを作成
func (s *gmailService) processEngineerEmail(ctx context.Context, emailText, attachmentText, contentHash, messageID string) error {
	parsed, err := s.emailAIService.ParseEngineerFromEmail(ctx, emailText, attachmentText)
	if err != nil {
		return fmt.Errorf("人材情報パース失敗: %w", err)
	}

	var parsedData model.ParsedEngineerData
	if err := json.Unmarshal(parsed, &parsedData); err != nil {
		s.logger.Warn("パース結果のデコード失敗", zap.Error(err))
	}

	expiresAt := computeEngineerExpiresAt(parsedData.StartMonth)

	profile := &model.EngineerProfile{
		ContentHash:   contentHash,
		SourceEmailID: messageID,
		RawText:       emailText,
		Parsed:        parsed,
		StartMonth:    parsedData.StartMonth,
		Status:        model.EngineerProfileStatusActive,
		ExpiresAt:     expiresAt,
	}

	if err := s.engineerProfileRepo.Create(ctx, profile); err != nil {
		return fmt.Errorf("人材プロファイル作成失敗: %w", err)
	}

	s.logger.Info("人材プロファイル作成成功", zap.String("profile_id", profile.ID))
	return nil
}

// extractAttachmentText メール添付ファイルからテキストを抽出
func (s *gmailService) extractAttachmentText(ctx context.Context, srv *gmail.Service, msg *gmail.Message) string {
	if msg.Payload == nil || msg.Payload.Parts == nil {
		return ""
	}

	var texts []string
	for _, part := range msg.Payload.Parts {
		if part.Body == nil || part.Body.AttachmentId == "" {
			continue
		}

		filename := part.Filename
		if filename == "" {
			continue
		}

		lowerName := strings.ToLower(filename)
		if !strings.HasSuffix(lowerName, ".xlsx") &&
			!strings.HasSuffix(lowerName, ".xls") &&
			!strings.HasSuffix(lowerName, ".pdf") {
			continue
		}

		att, err := srv.Users.Messages.Attachments.Get("me", msg.Id, part.Body.AttachmentId).Context(ctx).Do()
		if err != nil {
			s.logger.Warn("添付ファイル取得失敗",
				zap.String("filename", filename),
				zap.Error(err),
			)
			continue
		}

		data, err := base64.URLEncoding.DecodeString(att.Data)
		if err != nil {
			s.logger.Warn("添付ファイルデコード失敗",
				zap.String("filename", filename),
				zap.Error(err),
			)
			continue
		}

		text := s.parseAttachmentData(filename, data)
		if text != "" {
			texts = append(texts, text)
		}
	}

	return strings.Join(texts, "\n\n")
}

// parseAttachmentData 添付ファイルのバイトデータからテキストを抽出
func (s *gmailService) parseAttachmentData(filename string, data []byte) string {
	lowerName := strings.ToLower(filename)

	switch {
	case strings.HasSuffix(lowerName, ".xlsx"), strings.HasSuffix(lowerName, ".xls"):
		return s.parseExcelBytes(data)
	case strings.HasSuffix(lowerName, ".pdf"):
		return s.parsePDFBytes(data)
	default:
		return ""
	}
}

// parseExcelBytes Excelバイトデータからテキストを抽出
func (s *gmailService) parseExcelBytes(data []byte) string {
	reader := strings.NewReader(string(data))
	_ = reader
	s.logger.Debug("Excel添付ファイルのパースはfileParseServiceの拡張が必要です")
	return ""
}

// parsePDFBytes PDFバイトデータからテキストを抽出
func (s *gmailService) parsePDFBytes(data []byte) string {
	_ = data
	s.logger.Debug("PDF添付ファイルのパースはfileParseServiceの拡張が必要です")
	return ""
}

// ========================================
// ヘルパー関数
// ========================================

// computeContentHash メール内容のハッシュを計算（重複検知用）
func computeContentHash(sender, subject, body string) string {
	truncatedBody := body
	if len(truncatedBody) > 500 {
		truncatedBody = truncatedBody[:500]
	}
	h := sha256.Sum256([]byte(sender + subject + truncatedBody))
	return hex.EncodeToString(h[:])
}

// extractHeader メールヘッダーから指定キーの値を取得
func extractHeader(payload *gmail.MessagePart, name string) string {
	if payload == nil {
		return ""
	}
	for _, header := range payload.Headers {
		if strings.EqualFold(header.Name, name) {
			return header.Value
		}
	}
	return ""
}

// extractBody メール本文を抽出（text/plain優先、なければtext/html）
func extractBody(payload *gmail.MessagePart) string {
	if payload == nil {
		return ""
	}

	// 単一パートの場合
	if payload.MimeType == "text/plain" && payload.Body != nil && payload.Body.Data != "" {
		decoded, err := base64.URLEncoding.DecodeString(payload.Body.Data)
		if err == nil {
			return string(decoded)
		}
	}

	// マルチパートの場合
	var plainText, htmlText string
	for _, part := range payload.Parts {
		switch part.MimeType {
		case "text/plain":
			if part.Body != nil && part.Body.Data != "" {
				decoded, err := base64.URLEncoding.DecodeString(part.Body.Data)
				if err == nil {
					plainText = string(decoded)
				}
			}
		case "text/html":
			if part.Body != nil && part.Body.Data != "" {
				decoded, err := base64.URLEncoding.DecodeString(part.Body.Data)
				if err == nil {
					htmlText = string(decoded)
				}
			}
		case "multipart/alternative", "multipart/mixed", "multipart/related":
			// ネストされたマルチパートを再帰的に処理
			nested := extractBody(part)
			if nested != "" && plainText == "" {
				plainText = nested
			}
		}
	}

	if plainText != "" {
		return plainText
	}
	return htmlText
}

// computeJobExpiresAt 案件の有効期限を計算（start_monthの3ヶ月後）
func computeJobExpiresAt(startMonth string) *time.Time {
	if startMonth == "" {
		// start_monthがない場合は作成日から90日後
		t := time.Now().AddDate(0, 0, 90)
		return &t
	}

	parsed, err := time.Parse("2006-01", startMonth)
	if err != nil {
		t := time.Now().AddDate(0, 0, 90)
		return &t
	}

	// start_monthの3ヶ月後
	t := parsed.AddDate(0, 3, 0)
	return &t
}

// computeEngineerExpiresAt 人材プロファイルの有効期限を計算
func computeEngineerExpiresAt(startMonth string) *time.Time {
	if startMonth == "" {
		t := time.Now().AddDate(0, 0, 90)
		return &t
	}

	parsed, err := time.Parse("2006-01", startMonth)
	if err != nil {
		t := time.Now().AddDate(0, 0, 90)
		return &t
	}

	t := parsed.AddDate(0, 3, 0)
	return &t
}
