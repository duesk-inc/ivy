---
paths: backend/internal/model/**/*.go
description: Model層のENUM、フック、メソッド定義パターン
---

# Model層 パターン規約

## ステータスENUM定義

### 基本パターン

```go
// 1. カスタム型を定義
type ExpenseStatus string

// 2. 定数を定義（日本語コメント必須）
const (
    // ExpenseStatusDraft 下書き
    ExpenseStatusDraft ExpenseStatus = "draft"
    // ExpenseStatusSubmitted 申請中
    ExpenseStatusSubmitted ExpenseStatus = "submitted"
    // ExpenseStatusApproved 承認済み
    ExpenseStatusApproved ExpenseStatus = "approved"
    // ExpenseStatusRejected 却下
    ExpenseStatusRejected ExpenseStatus = "rejected"
)

// 3. フィールドで使用
type Expense struct {
    Status ExpenseStatus `gorm:"type:varchar(20);not null;default:'draft'" json:"status"`
}
```

### 正規化関数（移行期間用）

```go
// NormalizeExpenseStatus 文字列をExpenseStatusに変換
func NormalizeExpenseStatus(status string) ExpenseStatus {
    switch status {
    case "draft":
        return ExpenseStatusDraft
    case "submitted":
        return ExpenseStatusSubmitted
    case "approved":
        return ExpenseStatusApproved
    case "rejected":
        return ExpenseStatusRejected
    default:
        return ExpenseStatusDraft
    }
}
```

### ファイル内の定義順序（重要）

```go
// 1. カスタム型定義
type XxxStatus string

// 2. 定数定義（型の直後）
const (
    XxxStatusDraft XxxStatus = "draft"
    XxxStatusSubmitted XxxStatus = "submitted"
)

// 3. 正規化/パース関数（定数の直後）
func NormalizeXxxStatus(status string) XxxStatus { ... }

// 4. 構造体定義
type Xxx struct {
    Status XxxStatus `gorm:"..." json:"status"`
}

// 5. 構造体のメソッド（TableName, BeforeCreate等）
func (x *Xxx) BeforeCreate(tx *gorm.DB) error { ... }
```

---

## GORMフック

### BeforeCreate（UUID生成）

```go
import "github.com/google/uuid"

// BeforeCreate UUIDを生成
func (e *Expense) BeforeCreate(tx *gorm.DB) error {
    if e.ID == "" {
        e.ID = uuid.New().String()
    }
    return nil
}
```

### TableName（テーブル名オーバーライド）

```go
// TableName テーブル名を指定
func (LeaveRequest) TableName() string {
    return "leave_requests"
}
```

**注意**: GORMのデフォルト命名で問題ない場合は不要

---

## ビジネスロジックメソッド

### 配置ガイドライン

| メソッド種類 | Model層 | Service層 |
|------------|---------|----------|
| 状態チェック（Can*, Is*） | ✅ | - |
| 名前取得（FullName等） | ✅ | - |
| 権限チェック（HasRole等） | ✅ | - |
| データ変換・加工 | - | ✅ |
| 複数エンティティ操作 | - | ✅ |
| 外部サービス連携 | - | ✅ |

### 状態チェックメソッド

```go
// CanEdit 編集可能かチェック
func (e *Expense) CanEdit() bool {
    return e.Status == ExpenseStatusDraft
}

// CanSubmit 提出可能かチェック
func (e *Expense) CanSubmit() bool {
    return e.Status == ExpenseStatusDraft
}

// IsActive アクティブかチェック
func (u *User) IsActive() bool {
    return u.Active && u.Status == "active" && u.DeletedAt.Time.IsZero()
}
```

### 名前取得メソッド

```go
// FullName 氏名を取得
func (u *User) FullName() string {
    if u.Name != "" {
        return u.Name
    }
    return u.LastName + " " + u.FirstName
}
```

### 権限チェックメソッド

```go
// IsAdmin 管理者権限を持っているかチェック
func (u *User) IsAdmin() bool {
    return u.Role == RoleSystemAdmin || u.Role == RoleAdmin
}

// HasRole 指定されたロールを持っているかチェック
func (u *User) HasRole(role Role) bool {
    return u.Role == role
}
```

---

## カスタムENUM型（高度）

### Scanner/Valuer実装

データベースとの値変換が必要な場合に実装。

```go
import (
    "database/sql/driver"
    "fmt"
)

type WeeklyReportStatus string

// Scan sql.Scannerインターフェース実装
func (s *WeeklyReportStatus) Scan(value interface{}) error {
    if value == nil {
        *s = WeeklyReportStatus("")
        return nil
    }

    switch v := value.(type) {
    case []byte:
        *s = WeeklyReportStatus(string(v))
        return nil
    case string:
        *s = WeeklyReportStatus(v)
        return nil
    default:
        return fmt.Errorf("cannot scan type %T into WeeklyReportStatus", value)
    }
}

// Value driver.Valuerインターフェース実装
func (s WeeklyReportStatus) Value() (driver.Value, error) {
    return string(s), nil
}
```

---

## 派生構造体パターン

### WithDetails パターン

詳細情報を含む拡張構造体が必要な場合：

```go
// ExpenseWithDetails 詳細情報付き経費申請モデル（API応答用）
type ExpenseWithDetails struct {
    Expense
    Approvals      []ExpenseApproval      `json:"approvals,omitempty"`
    CategoryMaster *ExpenseCategoryMaster `json:"category_master,omitempty"`
}
```

---

## チェックリスト

### 新しいModelを作成する際

- [ ] ファイル名が `[domain].go` 形式か
- [ ] 構造体直前に日本語コメントがあるか
- [ ] GORMタグの順序が `gorm → json` か
- [ ] プライマリキーが `primaryKey`（camelCase）か
- [ ] UUIDフィールドが `varchar(36)` か
- [ ] 文字列フィールドが `type:varchar()` 形式か
- [ ] DeletedAtが `gorm.DeletedAt` で `json:"-"` か
- [ ] BeforeCreateフックでUUID生成を実装しているか

### ステータスENUMを追加する際

- [ ] カスタム型を定義しているか（`type XxxStatus string`）
- [ ] 定数に日本語コメントがあるか
- [ ] 正規化関数が必要か検討したか

### リレーションを定義する際

- [ ] foreignKeyを明示しているか
- [ ] オプショナルな場合はポインタ型 + `omitempty` か
- [ ] カスケード削除が必要か検討したか

---

## 既存コードとの差異（参考）

> **2024年12月 統一完了**: 以下の非推奨パターンは全て修正済みです。

| 項目 | 非推奨（修正済み） | 推奨（現在の標準） |
|-----|------------------|-------------------|
| プライマリキー | `primary_key` | `primaryKey` |
| UUIDサイズ | `varchar(255)` | `varchar(36)` ※Cognito Sub関連は除く |
| 文字列サイズ | `size:255` | `type:varchar(255)` |
| DeletedAt | `*time.Time` | `gorm.DeletedAt` |
| DeletedAt JSON | `json:"deleted_at"` | `json:"-"` |
| タグ順序 | `json → gorm` | `gorm → json` |

---

## 関連規約

- 基本構造・GORMタグ → [model-structure.md](./model-structure.md)
