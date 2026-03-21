---
paths: backend/internal/repository/**/*.go
---

# Repository層実装規約

## 基本構造

### インターフェース定義

```go
// ExpenseRepository 経費申請に関するデータアクセスのインターフェース
type ExpenseRepository interface {
    // 基本CRUD
    Create(ctx context.Context, expense *model.Expense) error
    GetByID(ctx context.Context, id string) (*model.Expense, error)
    Update(ctx context.Context, expense *model.Expense) error
    Delete(ctx context.Context, id string) error

    // 一覧・検索
    List(ctx context.Context, filter *dto.ExpenseFilterRequest) ([]model.Expense, int64, error)

    // 集計
    GetMonthlyTotal(ctx context.Context, userID string, year, month int) (int64, error)
}
```

### 命名規則

| 対象 | 規則 | 例 |
|------|------|-----|
| インターフェース | `[Domain]Repository` | `ExpenseRepository` |
| 実装構造体 | `[Domain]RepositoryImpl` | `ExpenseRepositoryImpl` |
| コンストラクタ | `New[Domain]Repository` | `NewExpenseRepository` |
| ファイル名 | `[domain]_repository.go` | `expense_repository.go` |

---

## 実装構造体

### 標準パターン

```go
type ExpenseRepositoryImpl struct {
    db     *gorm.DB
    logger *zap.Logger
}

func NewExpenseRepository(db *gorm.DB, logger *zap.Logger) ExpenseRepository {
    return &ExpenseRepositoryImpl{
        db:     db,
        logger: logger,
    }
}
```

### BaseRepository 埋め込みパターン

```go
type leaveRepository struct {
    repository.BaseRepository
    logger *zap.Logger
}

func NewLeaveRepository(db *gorm.DB, logger *zap.Logger) LeaveRepository {
    return &leaveRepository{
        BaseRepository: repository.BaseRepository{DB: db},
        logger:         logger,
    }
}
```

---

## メソッドシグネチャ

### 基本ルール

- 第1引数は必ず `ctx context.Context`
- 戻り値は `(T, error)` または `([]T, int64, error)`

### パターン一覧

```go
// 単一取得
func (r *ExpenseRepositoryImpl) GetByID(ctx context.Context, id string) (*model.Expense, error)

// 条件付き取得
func (r *ExpenseRepositoryImpl) GetByUserAndMonth(ctx context.Context, userID string, year, month int) (*model.Expense, error)

// 一覧取得（ページネーション付き）
func (r *ExpenseRepositoryImpl) List(ctx context.Context, filter *dto.ExpenseFilterRequest) ([]model.Expense, int64, error)

// 存在チェック
func (r *ExpenseRepositoryImpl) ExistsByID(ctx context.Context, id string) (bool, error)

// 作成
func (r *ExpenseRepositoryImpl) Create(ctx context.Context, expense *model.Expense) error

// 更新
func (r *ExpenseRepositoryImpl) Update(ctx context.Context, expense *model.Expense) error

// 削除
func (r *ExpenseRepositoryImpl) Delete(ctx context.Context, id string) error

// カウント
func (r *ExpenseRepositoryImpl) Count(ctx context.Context, filter *dto.ExpenseFilterRequest) (int64, error)
```

---

## メソッド組織化

### セクションコメントで分類

```go
// ========================================
// 基本CRUD操作
// ========================================

func (r *ExpenseRepositoryImpl) Create(ctx context.Context, expense *model.Expense) error {
    // ...
}

func (r *ExpenseRepositoryImpl) GetByID(ctx context.Context, id string) (*model.Expense, error) {
    // ...
}

// ========================================
// 一覧・検索機能
// ========================================

func (r *ExpenseRepositoryImpl) List(ctx context.Context, filter *dto.ExpenseFilterRequest) ([]model.Expense, int64, error) {
    // ...
}

// ========================================
// 集計機能
// ========================================

func (r *ExpenseRepositoryImpl) GetMonthlyTotal(ctx context.Context, userID string, year, month int) (int64, error) {
    // ...
}

// ========================================
// チェック機能
// ========================================

func (r *ExpenseRepositoryImpl) ExistsByID(ctx context.Context, id string) (bool, error) {
    // ...
}
```

---

## GORMの使用パターン

### 基本的なクエリ

```go
// 単一取得
func (r *ExpenseRepositoryImpl) GetByID(ctx context.Context, id string) (*model.Expense, error) {
    var expense model.Expense
    err := r.db.WithContext(ctx).
        Where("id = ?", id).
        First(&expense).Error

    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, nil  // または専用エラーを返す
        }
        return nil, fmt.Errorf("経費取得失敗: %w", err)
    }
    return &expense, nil
}
```

### Preload（関連データの取得）

```go
func (r *ExpenseRepositoryImpl) GetByIDWithReceipts(ctx context.Context, id string) (*model.Expense, error) {
    var expense model.Expense
    err := r.db.WithContext(ctx).
        Preload("Receipts").
        Preload("User").
        Where("id = ?", id).
        First(&expense).Error

    if err != nil {
        return nil, fmt.Errorf("経費取得失敗: %w", err)
    }
    return &expense, nil
}
```

### 一覧取得（ページネーション）

```go
func (r *ExpenseRepositoryImpl) List(ctx context.Context, filter *dto.ExpenseFilterRequest) ([]model.Expense, int64, error) {
    var expenses []model.Expense
    var total int64

    query := r.db.WithContext(ctx).Model(&model.Expense{})

    // フィルター適用
    if filter.UserID != "" {
        query = query.Where("user_id = ?", filter.UserID)
    }
    if filter.Status != "" {
        query = query.Where("status = ?", filter.Status)
    }
    if filter.Category != "" {
        query = query.Where("category = ?", filter.Category)
    }
    if !filter.StartDate.IsZero() {
        query = query.Where("expense_date >= ?", filter.StartDate)
    }
    if !filter.EndDate.IsZero() {
        query = query.Where("expense_date <= ?", filter.EndDate)
    }

    // 総件数取得
    if err := query.Count(&total).Error; err != nil {
        return nil, 0, fmt.Errorf("件数取得失敗: %w", err)
    }

    // ページネーション
    offset := (filter.Page - 1) * filter.Limit
    query = query.Offset(offset).Limit(filter.Limit)

    // ソート
    if filter.SortBy != "" {
        order := filter.SortBy
        if filter.SortOrder == "desc" {
            order += " DESC"
        }
        query = query.Order(order)
    } else {
        query = query.Order("created_at DESC")
    }

    // 実行
    if err := query.Find(&expenses).Error; err != nil {
        return nil, 0, fmt.Errorf("一覧取得失敗: %w", err)
    }

    return expenses, total, nil
}
```

### 作成・更新

```go
// 作成
func (r *ExpenseRepositoryImpl) Create(ctx context.Context, expense *model.Expense) error {
    if err := r.db.WithContext(ctx).Create(expense).Error; err != nil {
        return fmt.Errorf("経費作成失敗: %w", err)
    }
    return nil
}

// 更新
func (r *ExpenseRepositoryImpl) Update(ctx context.Context, expense *model.Expense) error {
    if err := r.db.WithContext(ctx).Save(expense).Error; err != nil {
        return fmt.Errorf("経費更新失敗: %w", err)
    }
    return nil
}

// 部分更新（指定カラムのみ）
func (r *ExpenseRepositoryImpl) UpdateStatus(ctx context.Context, id string, status string) error {
    result := r.db.WithContext(ctx).
        Model(&model.Expense{}).
        Where("id = ?", id).
        Update("status", status)

    if result.Error != nil {
        return fmt.Errorf("ステータス更新失敗: %w", result.Error)
    }
    if result.RowsAffected == 0 {
        return errors.New("対象レコードが見つかりません")
    }
    return nil
}
```

### 削除

```go
// 物理削除
func (r *ExpenseRepositoryImpl) Delete(ctx context.Context, id string) error {
    result := r.db.WithContext(ctx).
        Where("id = ?", id).
        Delete(&model.Expense{})

    if result.Error != nil {
        return fmt.Errorf("経費削除失敗: %w", result.Error)
    }
    return nil
}

// 論理削除（deleted_atを使用）
func (r *ExpenseRepositoryImpl) SoftDelete(ctx context.Context, id string) error {
    return r.db.WithContext(ctx).
        Model(&model.Expense{}).
        Where("id = ?", id).
        Update("deleted_at", time.Now()).Error
}
```

### 集計クエリ

```go
func (r *ExpenseRepositoryImpl) GetMonthlyTotal(ctx context.Context, userID string, year, month int) (int64, error) {
    var total int64

    startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
    endDate := startDate.AddDate(0, 1, -1)

    err := r.db.WithContext(ctx).
        Model(&model.Expense{}).
        Where("user_id = ?", userID).
        Where("expense_date BETWEEN ? AND ?", startDate, endDate).
        Where("status != ?", "cancelled").
        Select("COALESCE(SUM(amount), 0)").
        Scan(&total).Error

    if err != nil {
        return 0, fmt.Errorf("月次合計取得失敗: %w", err)
    }
    return total, nil
}
```

---

## エラーハンドリング

### gorm.ErrRecordNotFound の処理

```go
if err == gorm.ErrRecordNotFound {
    return nil, nil  // パターン1: nilを返す（呼び出し元で判定）
}

if err == gorm.ErrRecordNotFound {
    return nil, errors.New(message.MsgExpenseNotFound)  // パターン2: 専用エラー
}
```

### 重複エラーの処理

```go
if err := r.db.Create(&entity).Error; err != nil {
    if strings.Contains(err.Error(), "duplicate key") ||
       strings.Contains(err.Error(), "UNIQUE constraint failed") {
        return errors.New("このデータは既に登録されています")
    }
    return fmt.Errorf("作成失敗: %w", err)
}
```

---

## トランザクション対応

### Service層から渡されるDBを使用

```go
// Service層でトランザクション開始時
err := s.db.Transaction(func(tx *gorm.DB) error {
    // トランザクション用のリポジトリを新規作成
    txRepo := repository.NewExpenseRepository(tx, s.logger)
    return txRepo.Create(ctx, expense)
})
```

### リポジトリ側は特別な対応不要

```go
// 通常通りr.dbを使用（Service層からtxが渡される）
func (r *ExpenseRepositoryImpl) Create(ctx context.Context, expense *model.Expense) error {
    return r.db.WithContext(ctx).Create(expense).Error
}
```

---

## チェックリスト

新しいRepositoryメソッドを作成する際：

- [ ] 第1引数が `ctx context.Context`
- [ ] 戻り値が `(T, error)` または `([]T, int64, error)` 形式
- [ ] `r.db.WithContext(ctx)` でコンテキストを渡している
- [ ] `gorm.ErrRecordNotFound` を適切に処理している
- [ ] エラーは `fmt.Errorf` でラッピングしている
- [ ] 一覧取得にはページネーション（Offset/Limit）を実装している
- [ ] セクションコメントで機能を分類している

新しいRepositoryファイルを作成する際：

- [ ] インターフェースを先に定義している
- [ ] インターフェース直前に日本語コメントがある
- [ ] コンストラクタ `New[Domain]Repository` を実装している
- [ ] `db` と `logger` をフィールドに持っている
