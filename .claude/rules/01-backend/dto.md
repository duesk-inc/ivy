---
paths: backend/internal/dto/**/*.go
---

# DTO層実装規約

## 基本原則

DTOはAPI層とService層のデータ受け渡しを担当する。
Model（データベースエンティティ）とは分離し、API仕様の変更がModelに影響しないようにする。

---

## ファイル構成

### 命名規則

```
internal/dto/
├── expense_dto.go          # 機能ドメイン単位
├── leave_dto.go
├── engineer_dto.go
├── weekly_report_dto.go
└── common_dto.go           # 共通DTO（ページネーション等）
```

### ファイルサイズ制限

- **1ファイル300行以下を推奨**
- 超過する場合は機能別に分割

```
expense_dto.go           # 基本CRUD
expense_filter_dto.go    # フィルター・検索
expense_approval_dto.go  # 承認フロー
expense_receipt_dto.go   # 領収書関連
```

---

## 構造体命名規則

### リクエスト系

| パターン | 用途 | 例 |
|---------|------|-----|
| `Create[Domain]Request` | 新規作成 | `CreateExpenseRequest` |
| `Update[Domain]Request` | 更新 | `UpdateExpenseRequest` |
| `[Domain]FilterRequest` | フィルター条件 | `ExpenseFilterRequest` |
| `[Domain]ListRequest` | 一覧取得パラメータ | `ExpenseListRequest` |
| `[Domain]ActionRequest` | 特定アクション | `ExpenseApproveRequest` |

### レスポンス系

| パターン | 用途 | 例 |
|---------|------|-----|
| `[Domain]Response` | 単一レスポンス | `ExpenseResponse` |
| `[Domain]ListResponse` | 一覧レスポンス | `ExpenseListResponse` |
| `[Domain]DetailResponse` | 詳細レスポンス | `ExpenseDetailResponse` |
| `[Domain]SummaryResponse` | 集計レスポンス | `ExpenseSummaryResponse` |

---

## 構造体定義パターン

### リクエスト構造体

```go
// CreateExpenseRequest 経費申請の作成リクエスト
type CreateExpenseRequest struct {
    // 必須フィールド
    Title       string `json:"title" binding:"required,min=1,max=255"`
    Category    string `json:"category" binding:"required,oneof=transport entertainment supplies other"`
    Amount      int    `json:"amount" binding:"required,min=1,max=10000000"`
    ExpenseDate string `json:"expense_date" binding:"required"`

    // オプショナルフィールド（ポインタ型）
    Description *string  `json:"description,omitempty" binding:"omitempty,max=1000"`
    ProjectID   *string  `json:"project_id,omitempty" binding:"omitempty,uuid"`

    // スライス
    ReceiptURLs []string `json:"receipt_urls" binding:"omitempty,dive,url,max=10"`
}
```

### 更新リクエスト（部分更新対応）

```go
// UpdateExpenseRequest 経費申請の更新リクエスト
// すべてのフィールドがオプショナル（部分更新対応）
type UpdateExpenseRequest struct {
    Title       *string  `json:"title,omitempty" binding:"omitempty,min=1,max=255"`
    Category    *string  `json:"category,omitempty" binding:"omitempty,oneof=transport entertainment supplies other"`
    Amount      *int     `json:"amount,omitempty" binding:"omitempty,min=1,max=10000000"`
    Description *string  `json:"description,omitempty" binding:"omitempty,max=1000"`
    ExpenseDate *string  `json:"expense_date,omitempty"`
}
```

### レスポンス構造体

```go
// ExpenseResponse 経費申請のレスポンス
type ExpenseResponse struct {
    ID          string    `json:"id"`
    UserID      string    `json:"user_id"`
    Title       string    `json:"title"`
    Category    string    `json:"category"`
    Amount      int       `json:"amount"`
    Status      string    `json:"status"`
    Description string    `json:"description,omitempty"`
    ExpenseDate string    `json:"expense_date"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`

    // 関連データ（オプショナル）
    User     *UserMinimalResponse `json:"user,omitempty"`
    Receipts []ReceiptResponse    `json:"receipts,omitempty"`
}

// ExpenseListResponse 経費申請一覧のレスポンス
type ExpenseListResponse struct {
    Items      []ExpenseResponse `json:"items"`
    Total      int64             `json:"total"`
    Page       int               `json:"page"`
    Limit      int               `json:"limit"`
    TotalPages int               `json:"total_pages"`
}
```

### フィルター構造体

```go
// ExpenseFilterRequest 経費申請のフィルター条件
type ExpenseFilterRequest struct {
    // フィルター条件
    UserID    string    `form:"user_id"`
    Status    string    `form:"status" binding:"omitempty,oneof=draft submitted approved rejected"`
    Category  string    `form:"category"`
    StartDate time.Time `form:"start_date" time_format:"2006-01-02"`
    EndDate   time.Time `form:"end_date" time_format:"2006-01-02"`
    MinAmount *int      `form:"min_amount"`
    MaxAmount *int      `form:"max_amount"`

    // ページネーション
    Page  int `form:"page" binding:"min=1"`
    Limit int `form:"limit" binding:"min=1,max=100"`

    // ソート
    SortBy    string `form:"sort_by" binding:"omitempty,oneof=expense_date amount created_at"`
    SortOrder string `form:"sort_order" binding:"omitempty,oneof=asc desc"`
}

// SetDefaults デフォルト値を設定
func (f *ExpenseFilterRequest) SetDefaults() {
    if f.Page == 0 {
        f.Page = 1
    }
    if f.Limit == 0 {
        f.Limit = 20
    }
    if f.SortOrder == "" {
        f.SortOrder = "desc"
    }
}
```

---

## バリデーションタグ

### よく使うタグ

| タグ | 説明 | 例 |
|-----|------|-----|
| `required` | 必須 | `binding:"required"` |
| `omitempty` | 空の場合は検証スキップ | `binding:"omitempty"` |
| `min` / `max` | 長さ・値の範囲 | `binding:"min=1,max=255"` |
| `oneof` | 列挙値 | `binding:"oneof=draft submitted"` |
| `uuid` | UUID形式 | `binding:"uuid"` |
| `email` | メール形式 | `binding:"email"` |
| `url` | URL形式 | `binding:"url"` |
| `dive` | スライス要素の検証 | `binding:"dive,url"` |

### タグの記述順序

```go
// ✅ 正しい順序: json → binding
Field string `json:"field" binding:"required,min=1"`

// ❌ 避ける: binding が先
Field string `binding:"required" json:"field"`
```

---

## カスタムバリデーション

### Validate() メソッド

```go
// Validate ビジネスルールの検証
func (r *CreateExpenseLimitRequest) Validate() error {
    // スコープ別バリデーション
    switch r.LimitScope {
    case "user":
        if r.UserID == nil {
            return fmt.Errorf("個人制限の場合、ユーザーIDは必須です")
        }
    case "category":
        if r.Category == nil {
            return fmt.Errorf("カテゴリ制限の場合、カテゴリは必須です")
        }
    case "global":
        // 追加チェック不要
    default:
        return fmt.Errorf("無効なスコープです: %s", r.LimitScope)
    }

    // 日付チェック
    if r.EffectiveFrom != nil && r.EffectiveFrom.Before(time.Now()) {
        return fmt.Errorf("適用開始日時は未来の日時を指定してください")
    }

    // 金額チェック
    if r.MonthlyLimit != nil && *r.MonthlyLimit <= 0 {
        return fmt.Errorf("月次上限は正の値を指定してください")
    }

    return nil
}
```

### Handler層での呼び出し

```go
func (h *expenseHandler) CreateExpenseLimit(c *gin.Context) {
    var req dto.CreateExpenseLimitRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        h.handleValidationError(c, err)
        return
    }

    // カスタムバリデーション
    if err := req.Validate(); err != nil {
        utils.RespondError(c, http.StatusBadRequest, err.Error())
        return
    }

    // 処理続行...
}
```

---

## 変換メソッド

### To系: リクエスト → モデル

```go
// ToExpense CreateExpenseRequestをmodel.Expenseに変換
func (r *CreateExpenseRequest) ToExpense(userID string) model.Expense {
    expense := model.Expense{
        ID:          uuid.New().String(),
        UserID:      userID,
        Title:       r.Title,
        Category:    model.ExpenseCategory(r.Category),
        Amount:      r.Amount,
        Status:      model.ExpenseStatusDraft,
        ExpenseDate: r.parseExpenseDate(),
    }

    if r.Description != nil {
        expense.Description = *r.Description
    }
    if r.ProjectID != nil {
        expense.ProjectID = r.ProjectID
    }

    return expense
}

func (r *CreateExpenseRequest) parseExpenseDate() time.Time {
    t, _ := time.Parse("2006-01-02", r.ExpenseDate)
    return t
}
```

### From系: モデル → レスポンス

```go
// FromExpense model.ExpenseからExpenseResponseを生成
func ExpenseResponseFromModel(expense model.Expense) ExpenseResponse {
    return ExpenseResponse{
        ID:          expense.ID,
        UserID:      expense.UserID,
        Title:       expense.Title,
        Category:    string(expense.Category),
        Amount:      expense.Amount,
        Status:      string(expense.Status),
        Description: expense.Description,
        ExpenseDate: expense.ExpenseDate.Format("2006-01-02"),
        CreatedAt:   expense.CreatedAt,
        UpdatedAt:   expense.UpdatedAt,
    }
}

// スライス変換
func ExpenseListResponseFromModels(expenses []model.Expense, total int64, page, limit int) ExpenseListResponse {
    items := make([]ExpenseResponse, len(expenses))
    for i, e := range expenses {
        items[i] = ExpenseResponseFromModel(e)
    }

    totalPages := int(total) / limit
    if int(total)%limit > 0 {
        totalPages++
    }

    return ExpenseListResponse{
        Items:      items,
        Total:      total,
        Page:       page,
        Limit:      limit,
        TotalPages: totalPages,
    }
}
```

### Apply系: 更新リクエスト → 既存モデル

```go
// ApplyToExpense 更新リクエストを既存のExpenseに適用
func (r *UpdateExpenseRequest) ApplyToExpense(expense *model.Expense) {
    if r.Title != nil {
        expense.Title = *r.Title
    }
    if r.Category != nil {
        expense.Category = model.ExpenseCategory(*r.Category)
    }
    if r.Amount != nil {
        expense.Amount = *r.Amount
    }
    if r.Description != nil {
        expense.Description = *r.Description
    }
    if r.ExpenseDate != nil {
        t, _ := time.Parse("2006-01-02", *r.ExpenseDate)
        expense.ExpenseDate = t
    }
}
```

---

## 共通DTO

### ページネーション

```go
// PaginationRequest ページネーションの共通リクエスト
type PaginationRequest struct {
    Page  int `form:"page" binding:"min=1"`
    Limit int `form:"limit" binding:"min=1,max=100"`
}

func (p *PaginationRequest) SetDefaults() {
    if p.Page == 0 {
        p.Page = 1
    }
    if p.Limit == 0 {
        p.Limit = 20
    }
}

func (p *PaginationRequest) Offset() int {
    return (p.Page - 1) * p.Limit
}
```

### ソート

```go
// SortRequest ソートの共通リクエスト
type SortRequest struct {
    SortBy    string `form:"sort_by"`
    SortOrder string `form:"sort_order" binding:"omitempty,oneof=asc desc"`
}

func (s *SortRequest) OrderClause(allowedFields []string) string {
    // 許可されたフィールドかチェック
    allowed := false
    for _, f := range allowedFields {
        if f == s.SortBy {
            allowed = true
            break
        }
    }
    if !allowed {
        return "created_at DESC"
    }

    order := s.SortBy
    if s.SortOrder == "desc" {
        order += " DESC"
    }
    return order
}
```

---

## チェックリスト

新しいDTOを作成する際：

- [ ] 構造体名が命名規則（`Create*Request`, `*Response` 等）に従っているか
- [ ] JSONタグが `json:"field_name"` 形式で記述されているか
- [ ] バリデーションタグが `binding:"..."` で記述されているか
- [ ] タグの順序が `json` → `binding` になっているか
- [ ] オプショナルフィールドはポインタ型 + `omitempty` になっているか
- [ ] 複雑なバリデーションは `Validate()` メソッドで実装しているか
- [ ] `To*()` / `From*()` / `Apply*()` 変換メソッドを実装しているか

ファイル管理：

- [ ] ファイルサイズが300行以下か（超過時は分割を検討）
- [ ] ファイル名が `[domain]_dto.go` 形式か
- [ ] 構造体の直前に日本語コメントがあるか
