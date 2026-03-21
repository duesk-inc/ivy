---
paths: backend/**/*.go
---

# バックエンド エラーハンドリング規約

## メッセージ定数の活用

**エラーメッセージは `internal/message` パッケージの定数を使用すること**

### 定義場所

```go
// internal/message/error_codes.go
const (
    ErrCodeInvalidRequest    = "INVALID_REQUEST"
    ErrCodeUnauthorized      = "UNAUTHORIZED"
    ErrCodeForbidden         = "FORBIDDEN"
    ErrCodeNotFound          = "NOT_FOUND"
    ErrCodeConflict          = "CONFLICT"
    ErrCodeInternalError     = "INTERNAL_ERROR"
)

// internal/message/leave.go
const (
    MsgReasonRequired            = "理由が必要な休暇種別です"
    MsgHourlyLeaveNotAllowed     = "時間単位取得不可の休暇種別です"
    MsgLeaveBalanceExceededFormat = "残日数（%.1f日）を超えています"
)
```

### 使用方法

```go
import "github.com/duesk/monstera/internal/message"

// 定数をそのまま使用
return nil, errors.New(message.MsgReasonRequired)

// フォーマット文字列として使用
return nil, fmt.Errorf(message.MsgLeaveBalanceExceededFormat, balance.RemainingDays)
```

---

## エラーラッピング

### 基本パターン: fmt.Errorf

```go
// コンテキスト情報を追加してラッピング
user, err := s.userRepo.FindByID(ctx, id)
if err != nil {
    return nil, fmt.Errorf("ユーザー取得失敗 (id=%s): %w", id, err)
}
```

### 推奨パターン: logger.LogAndWrapError

```go
import "github.com/duesk/monstera/internal/logger"

// ログ出力 + ラッピングを同時に行う
if err != nil {
    return nil, logger.LogAndWrapError(s.logger, err, "休暇種別の取得に失敗しました",
        zap.String("leave_type_id", req.LeaveTypeID),
        zap.String("user_id", req.UserID))
}
```

---

## gorm エラー処理

### ErrRecordNotFound

```go
import "gorm.io/gorm"

user, err := s.userRepo.FindByID(ctx, id)
if err != nil {
    if err == gorm.ErrRecordNotFound {
        return nil, fmt.Errorf(message.MsgUserNotFound+": %w", err)
    }
    return nil, fmt.Errorf("ユーザー取得エラー: %w", err)
}
```

### 重複エラー

```go
if err := s.db.Create(&entity).Error; err != nil {
    if strings.Contains(err.Error(), "duplicate key") {
        return nil, errors.New("このデータは既に登録されています")
    }
    return nil, fmt.Errorf("登録エラー: %w", err)
}
```

---

## HTTPステータスコード対応

### Handler層でのマッピング

```go
func (h *leaveHandler) CancelLeaveRequest(c *gin.Context) {
    err := h.leaveService.CancelLeaveRequest(ctx, requestID, userUUID)
    if err != nil {
        // エラーメッセージに応じてステータスコードを決定
        statusCode := http.StatusBadRequest

        switch err.Error() {
        case message.MsgCannotEditOthersRequest:
            statusCode = http.StatusForbidden
        case message.MsgLeaveRequestNotFound:
            statusCode = http.StatusNotFound
        case message.MsgCannotCancelNonPendingRequest:
            statusCode = http.StatusConflict
        }

        h.handleError(c, statusCode, message.MsgLeaveRequestCancelFailed, err)
        return
    }
    // ...
}
```

### ステータスコード一覧

| ステータス | 用途 | 例 |
|-----------|------|-----|
| 400 Bad Request | リクエスト不正、バリデーションエラー | 必須項目が未入力 |
| 401 Unauthorized | 認証エラー | トークン無効、未ログイン |
| 403 Forbidden | 権限エラー | 他人のデータにアクセス |
| 404 Not Found | リソース未発見 | 指定IDのデータが存在しない |
| 409 Conflict | 状態競合 | 既に提出済み、重複登録 |
| 422 Unprocessable Entity | ビジネスルール違反 | 残日数超過 |
| 500 Internal Server Error | サーバー内部エラー | DB接続エラー等 |

---

## Service層でのエラーハンドリング

### ビジネスルール検証

```go
func (s *leaveService) CreateLeaveRequest(ctx context.Context, req dto.LeaveRequestRequest) (dto.LeaveRequestResponse, error) {
    // 休暇種別の検証
    leaveType, err := s.leaveRepo.GetLeaveTypeByID(ctx, req.LeaveTypeID)
    if err != nil {
        return dto.LeaveRequestResponse{}, logger.LogAndWrapError(
            s.logger, err, "休暇種別が見つかりません",
            zap.String("leave_type_id", req.LeaveTypeID))
    }

    // ビジネスルール: 理由必須チェック
    if leaveType.ReasonRequired && req.Reason == "" {
        return dto.LeaveRequestResponse{}, errors.New(message.MsgReasonRequired)
    }

    // ビジネスルール: 時間単位取得可否
    if req.IsHourlyBased && !leaveType.IsHourlyAvailable {
        return dto.LeaveRequestResponse{}, errors.New(message.MsgHourlyLeaveNotAllowed)
    }

    // ビジネスルール: 残日数チェック
    if leaveType.RequiresBalance {
        balance, err := s.leaveRepo.GetUserLeaveBalanceByType(ctx, req.UserID, req.LeaveTypeID)
        if err != nil {
            return dto.LeaveRequestResponse{}, logger.LogAndWrapError(
                s.logger, err, "残日数取得失敗")
        }
        if balance.RemainingDays < req.TotalDays {
            return dto.LeaveRequestResponse{}, fmt.Errorf(
                message.MsgLeaveBalanceExceededFormat, balance.RemainingDays)
        }
    }

    // 処理続行...
}
```

### 重複チェック

```go
// 重複チェック
exists, err := s.engineerRepo.ExistsByEmail(ctx, input.Email)
if err != nil {
    return nil, fmt.Errorf("メールアドレスの確認中にエラー: %w", err)
}
if exists {
    return nil, errors.New("このメールアドレスは既に使用されています")
}
```

---

## Handler層でのエラーハンドリング

### HandleError 関数

```go
// ログ出力 + エラーレスポンスを一括で行う
HandleError(c, http.StatusInternalServerError, "エンジニア取得失敗", h.logger, err,
    "user_id", userID,
    "engineer_id", engineerID)
```

### バリデーションエラー

```go
var req dto.CreateEngineerRequest
if err := c.ShouldBindJSON(&req); err != nil {
    // バリデーション詳細をマップに変換
    errorMap := h.handlerUtil.CreateValidationErrorMap(err)
    RespondValidationError(c, errorMap)
    return
}
```

### レスポンス形式

```go
// エラーレスポンスの標準形式
type ErrorResponse struct {
    Error   string            `json:"error"`
    Code    string            `json:"code,omitempty"`
    Details map[string]string `json:"details,omitempty"`
}
```

---

## ログ出力

### ログレベルの使い分け

```go
// Info: 正常処理の記録
s.logger.Info("休暇申請作成成功",
    zap.String("user_id", userID),
    zap.String("request_id", request.ID))

// Warn: 警告（処理は継続）
s.logger.Warn("残日数が少なくなっています",
    zap.String("user_id", userID),
    zap.Float64("remaining", balance.RemainingDays))

// Error: エラー発生
s.logger.Error("休暇申請作成失敗",
    zap.Error(err),
    zap.String("user_id", userID))
```

### 構造化ログ

```go
// zap.Field を使用して構造化
s.logger.Error("処理失敗",
    zap.Error(err),                        // エラー詳細
    zap.String("user_id", userID),         // 文字列
    zap.Int("count", len(items)),          // 数値
    zap.Bool("is_admin", isAdmin),         // 真偽値
    zap.Any("request", req),               // 任意の型
    zap.Duration("elapsed", elapsed))      // 時間
```

---

## トランザクション内のエラー

```go
err := s.db.Transaction(func(tx *gorm.DB) error {
    txRepo := repository.NewEngineerRepository(tx, s.logger)

    if err := txRepo.Create(ctx, user); err != nil {
        // トランザクション内のエラーはそのまま返す
        // 自動的にロールバックされる
        return fmt.Errorf("ユーザー作成失敗: %w", err)
    }

    if err := txRepo.CreateHistory(ctx, history); err != nil {
        return fmt.Errorf("履歴作成失敗: %w", err)
    }

    return nil
})

if err != nil {
    // 外部APIのロールバック処理（必要な場合）
    if cognitoUserCreated {
        s.cognitoAuth.DeleteUser(ctx, email)
    }
    return nil, err
}
```

---

## チェックリスト

### エラー定義

- [ ] エラーメッセージは `message` パッケージの定数を使用しているか
- [ ] 新しいメッセージは適切なファイルに定数として追加したか

### エラーラッピング

- [ ] `%w` を使用してエラーチェーンを維持しているか
- [ ] `logger.LogAndWrapError` または `fmt.Errorf` でラッピングしているか
- [ ] 必要なコンテキスト情報を含めているか

### Handler層

- [ ] 適切なHTTPステータスコードを返しているか
- [ ] `HandleError` でログ出力とレスポンスを行っているか
- [ ] バリデーションエラーは詳細情報を含めているか

### ログ出力

- [ ] 適切なログレベル（Info/Warn/Error）を使用しているか
- [ ] 構造化ログ（zap.Field）を使用しているか
- [ ] ユーザーID等の追跡可能な情報を含めているか
