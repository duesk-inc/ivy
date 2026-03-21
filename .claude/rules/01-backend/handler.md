---
paths: backend/internal/handler/**/*.go
---

# Handler層実装規約

## 基本構造

### Handler構造体の定義

```go
type engineerHandler struct {
    engineerService service.EngineerService  // インターフェース型
    handlerUtil     *HandlerUtil
    logger          *zap.Logger
}

func NewEngineerHandler(
    engineerService service.EngineerService,
    logger *zap.Logger,
) *engineerHandler {
    return &engineerHandler{
        engineerService: engineerService,
        handlerUtil:     NewHandlerUtil(logger),
        logger:          logger,
    }
}
```

### Handlerメソッドの基本形

```go
func (h *engineerHandler) GetEngineer(c *gin.Context) {
    // 1. コンテキスト取得
    ctx := c.Request.Context()

    // 2. ログ出力（処理開始）
    h.logger.Info("GetEngineer開始", zap.String("endpoint", c.Request.URL.Path))

    // 3. 認証ユーザー取得（必要な場合）
    userID, ok := h.handlerUtil.GetAuthenticatedUserID(c)
    if !ok {
        return  // エラーレスポンスは既に返却済み
    }

    // 4. パラメータ取得
    id := c.Param("id")

    // 5. サービス呼び出し
    engineer, err := h.engineerService.GetEngineerByID(ctx, id)
    if err != nil {
        h.handleError(c, http.StatusInternalServerError, "エンジニア取得失敗", err,
            zap.String("engineer_id", id))
        return
    }

    // 6. ログ出力（処理完了）
    h.logger.Info("GetEngineer完了", zap.String("engineer_id", id))

    // 7. レスポンス返却
    c.JSON(http.StatusOK, engineer)
}
```

---

## 認証ユーザー情報の取得

### 標準パターン

```go
// ユーザーID取得（失敗時は自動で401レスポンス）
userID, ok := h.handlerUtil.GetAuthenticatedUserID(c)
if !ok {
    return
}

// UUID変換が必要な場合
userUUID, err := uuid.Parse(userID)
if err != nil {
    h.handleError(c, http.StatusBadRequest, "無効なユーザーID", err)
    return
}
```

### 管理者権限チェック

```go
isAdmin, err := h.handlerUtil.IsAdmin(c)
if err != nil {
    h.handleError(c, http.StatusInternalServerError, "権限確認失敗", err)
    return
}
if !isAdmin {
    utils.RespondForbidden(c, "管理者権限が必要です")
    return
}
```

---

## リクエストバリデーション

### JSONボディのバインド

```go
var req dto.CreateEngineerRequest
if err := c.ShouldBindJSON(&req); err != nil {
    h.handleValidationError(c, err)
    return
}
```

### クエリパラメータのバインド

```go
var query dto.EngineerListQuery
if err := c.ShouldBindQuery(&query); err != nil {
    h.handleValidationError(c, err)
    return
}
```

### パスパラメータの取得

```go
id := c.Param("id")
if id == "" {
    utils.RespondError(c, http.StatusBadRequest, "IDは必須です")
    return
}
```

---

## エラーレスポンス

### 関数一覧

| 関数 | 用途 | 例 |
|------|------|-----|
| `RespondError` | シンプルなエラー | `RespondError(c, http.StatusBadRequest, "メッセージ")` |
| `RespondErrorWithCode` | エラーコード付き | `RespondErrorWithCode(c, message.ErrCodeInvalidRequest, "メッセージ", nil)` |
| `HandleError` | ログ出力 + エラー返却 | `HandleError(c, 500, "メッセージ", h.logger, err, "key", "value")` |
| `RespondValidationError` | バリデーション詳細 | `RespondValidationError(c, errorMap)` |
| `RespondNotFound` | 404専用 | `RespondNotFound(c, "リソースが見つかりません")` |
| `RespondUnauthorized` | 401専用 | `RespondUnauthorized(c, "認証が必要です")` |
| `RespondForbidden` | 403専用 | `RespondForbidden(c, "権限がありません")` |

### HTTPステータスコードの使い分け

```go
// 400 Bad Request - リクエスト不正
RespondError(c, http.StatusBadRequest, "無効なリクエストです")

// 401 Unauthorized - 認証エラー
RespondUnauthorized(c, message.MsgUnauthorized)

// 403 Forbidden - 権限エラー
RespondForbidden(c, "この操作を行う権限がありません")

// 404 Not Found - リソース未発見
RespondNotFound(c, message.MsgEngineerNotFound)

// 409 Conflict - 状態競合（既に提出済み等）
RespondError(c, http.StatusConflict, message.MsgCannotEditSubmitted)

// 500 Internal Server Error - サーバーエラー
HandleError(c, http.StatusInternalServerError, "内部エラーが発生しました", h.logger, err)
```

### エラーメッセージのパターン

```go
func (h *leaveHandler) CancelLeaveRequest(c *gin.Context) {
    // ...
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

---

## 成功レスポンス

### 取得（GET）

```go
// 単一リソース
c.JSON(http.StatusOK, engineer)

// リスト（ページネーション付き）
c.JSON(http.StatusOK, gin.H{
    "items": engineers,
    "total": total,
    "page":  page,
    "limit": limit,
})
```

### 作成（POST）

```go
c.JSON(http.StatusCreated, createdEntity)
```

### 更新（PUT/PATCH）

```go
c.JSON(http.StatusOK, updatedEntity)
```

### 削除（DELETE）

```go
c.JSON(http.StatusOK, gin.H{"message": message.MsgDeleteSuccess})
```

---

## ログ出力

### ログレベルの使い分け

```go
// Info: 正常処理の開始・完了
h.logger.Info("処理開始",
    zap.String("endpoint", "GetEngineers"),
    zap.String("user_id", userID))

h.logger.Info("処理完了",
    zap.Int("count", len(engineers)))

// Warn: 警告（処理は継続）
h.logger.Warn("バリデーションエラー",
    zap.String("endpoint", c.Request.URL.Path),
    zap.Any("errors", errorDetails))

// Error: エラー発生
h.logger.Error("処理失敗",
    zap.Error(err),
    zap.String("user_id", userID),
    zap.String("engineer_id", engineerID))
```

### HandleError でのログ統合

```go
// HandleError は自動でログ出力 + エラーレスポンスを行う
HandleError(c, http.StatusInternalServerError, "エンジニア取得失敗", h.logger, err,
    "user_id", userID,
    "engineer_id", engineerID)
```

---

## サービス層の呼び出し

### 基本パターン

```go
// 必ずコンテキストを渡す
result, err := h.engineerService.GetEngineerByID(c.Request.Context(), id)
if err != nil {
    h.handleError(c, http.StatusInternalServerError, "取得失敗", err)
    return
}
```

### 複数サービスの連携

```go
type expenseHandler struct {
    expenseService service.ExpenseService
    s3Service      service.S3Service  // ファイル操作用
    logger         *zap.Logger
    handlerUtil    *HandlerUtil
}
```

---

## チェックリスト

新しいHandlerメソッドを作成する際：

- [ ] `ctx := c.Request.Context()` でコンテキスト取得
- [ ] 認証が必要な場合は `GetAuthenticatedUserID` を使用
- [ ] `ShouldBindJSON` または `ShouldBindQuery` でバリデーション
- [ ] サービス呼び出しにctxを渡している
- [ ] エラー時は適切なHTTPステータスコードを返している
- [ ] `HandleError` でログ出力 + エラーレスポンス
- [ ] 処理開始・完了のログを出力している
- [ ] メッセージ定数（`message` パッケージ）を使用している
