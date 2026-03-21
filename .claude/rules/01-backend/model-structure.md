---
paths: backend/internal/model/**/*.go
description: Model層の基本構造とGORMタグ規約
---

# Model層 構造規約

## 基本原則

ModelはGORMを使用したデータベースエンティティを定義する。
ビジネスロジックは最小限に留め、主にService層で実装する。

---

## ファイル構成

### 命名規則

```
internal/model/
├── user.go                    # 単一エンティティ
├── expense.go                 # 関連する型をまとめる
├── expense_category.go        # サブドメイン
├── enum_types.go              # 共通ENUM型
├── constants.go               # 共通定数
└── role.go                    # 権限関連
```

### ファイルサイズ

- **1ファイル400行以下を推奨**
- 超過する場合は関連エンティティを別ファイルに分割

---

## GORMタグ規約

### タグ順序（重要）

```go
// ✅ 正しい順序: gorm → json
ID string `gorm:"type:varchar(36);primaryKey" json:"id"`

// ❌ 避ける: json → gorm
ID string `json:"id" gorm:"type:varchar(36);primaryKey"`
```

### プライマリキー

```go
// ✅ 正しい: primaryKey（GORM v2 スタイル、camelCase）
ID string `gorm:"type:varchar(36);primaryKey" json:"id"`

// ❌ 避ける: primary_key（GORM v1 スタイル）
ID string `gorm:"type:varchar(255);primary_key" json:"id"`
```

### UUIDフィールド

```go
// ✅ 正しい: varchar(36) - UUID標準サイズ
ID       string `gorm:"type:varchar(36);primaryKey" json:"id"`
UserID   string `gorm:"type:varchar(36);not null" json:"user_id"`

// ❌ 避ける: varchar(255) - 過剰なサイズ
ID string `gorm:"type:varchar(255);primaryKey" json:"id"`
```

### 例外規則（重要）

#### Cognito Sub ID

**User.ID および UserID を参照する外部キーは `varchar(255)` を維持する。**

```go
// ✅ 正しい: Cognito Sub 対応（User関連のみ例外）
type User struct {
    ID string `gorm:"type:varchar(255);primaryKey" json:"id"` // Cognito Sub
}

type Expense struct {
    ID     string `gorm:"type:varchar(36);primaryKey" json:"id"`      // 自己生成UUID → varchar(36)
    UserID string `gorm:"type:varchar(255);not null" json:"user_id"`  // User.IDへの外部キー → varchar(255)
}
```

#### マスタテーブルの整数ID

```go
// ✅ マスタテーブル: 整数ID + autoIncrement
type Process struct {
    ID   int32  `gorm:"primaryKey;autoIncrement" json:"id"`
    Name string `gorm:"type:varchar(100);not null;unique" json:"name"`
}
```

### 文字列・NOT NULL・インデックス

```go
// 文字列フィールド
Name        string `gorm:"type:varchar(100);not null" json:"name"`
Description string `gorm:"type:text" json:"description"`
Email       string `gorm:"type:varchar(255);not null;unique" json:"email"`

// NOT NULL・デフォルト値
Status    string    `gorm:"type:varchar(20);not null;default:'draft'" json:"status"`
IsActive  bool      `gorm:"not null;default:true" json:"is_active"`
CreatedAt time.Time `gorm:"not null;default:NOW()" json:"created_at"`

// オプショナルフィールドはポインタ型
ApproverID *string `gorm:"type:varchar(36)" json:"approver_id"`

// インデックス
UserID string `gorm:"type:varchar(36);not null;index" json:"user_id"`
Code   string `gorm:"type:varchar(50);not null;uniqueIndex" json:"code"`
```

---

## ソフトデリート

```go
// ✅ 正しい: gorm.DeletedAt（GORM v2 推奨）、APIに出力しない
DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

// ❌ 避ける
DeletedAt *time.Time `gorm:"index" json:"deleted_at"`
```

---

## リレーション定義

### 基本パターン

```go
type Expense struct {
    ID         string `gorm:"type:varchar(36);primaryKey" json:"id"`
    UserID     string `gorm:"type:varchar(36);not null" json:"user_id"`
    ApproverID *string `gorm:"type:varchar(36)" json:"approver_id"`

    // BelongsTo（必須リレーション）
    User User `gorm:"foreignKey:UserID" json:"user"`

    // BelongsTo（オプショナルリレーション）
    Approver *User `gorm:"foreignKey:ApproverID;references:ID" json:"approver,omitempty"`

    // HasMany
    Receipts []ExpenseReceipt `gorm:"foreignKey:ExpenseID" json:"receipts,omitempty"`
}
```

### カスケード制約

```go
// 親削除時に子も削除
Receipts []ExpenseReceipt `gorm:"foreignKey:ExpenseID;constraint:OnDelete:CASCADE" json:"receipts,omitempty"`
```

### 自己参照

```go
type User struct {
    ID        string `gorm:"type:varchar(36);primaryKey" json:"id"`
    ManagerID *string `gorm:"type:varchar(36)" json:"manager_id"`
    Manager   *User  `gorm:"foreignKey:ManagerID" json:"manager,omitempty"`
}
```

---

## タイムスタンプ

```go
type Expense struct {
    // ... 他のフィールド

    CreatedAt time.Time      `gorm:"not null;default:NOW()" json:"created_at"`
    UpdatedAt time.Time      `gorm:"not null;default:NOW()" json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// PostgreSQL用: timestamptz を使用
CreatedAt time.Time `gorm:"type:timestamptz;not null;default:NOW()" json:"created_at"`
```

---

## 関連規約

- ENUM・フック・メソッド → [model-patterns.md](./model-patterns.md)
