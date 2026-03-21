---
paths: backend/**/*_test.go, backend/test/**/*.go
description: Go バックエンドテスト規約
---

# Backend テスト規約（Go）

## 基本原則

- **TDD実践**: テストファーストを推奨
- **カバレッジ目標**: 80%以上
- **テストの独立性**: 各テストは他のテストに依存しない
- **テストの可読性**: テスト名から何をテストしているか明確にわかる

---

## ディレクトリ構成

```
backend/
├── test/
│   ├── unit/                    # ユニットテスト
│   │   ├── expense_service_test.go
│   │   └── leave_service_test.go
│   ├── handler/                 # Handlerテスト
│   │   ├── expense_handler_test.go
│   │   └── leave_handler_test.go
│   ├── integration/             # 統合テスト
│   │   └── expense_flow_test.go
│   ├── fixtures/                # テストデータ
│   │   └── users.go
│   └── testhelper/              # テストヘルパー
│       └── cognito_helper.go
└── internal/
    └── service/
        └── expense_service_test.go  # 同一パッケージ内テストも可
```

## ファイル命名規則

| パターン | 例 |
|---------|-----|
| `*_test.go` | `expense_service_test.go` |
| `*_unit_test.go` | `expense_service_unit_test.go`（明示的にunit） |
| `*_integration_test.go` | `expense_flow_integration_test.go` |

---

## テスト関数の構造

```go
func TestExpenseService_Create(t *testing.T) {
    // テストケースを t.Run でグループ化
    t.Run("正常系: 経費申請の作成", func(t *testing.T) {
        // Arrange（準備）
        ctx := context.Background()
        userID := uuid.New().String()
        input := CreateExpenseInput{
            Title:    "営業会議費",
            Category: "entertainment",
            Amount:   5000,
        }

        // モック設定
        mockRepo := new(MockExpenseRepository)
        mockRepo.On("Create", ctx, mock.AnythingOfType("*model.Expense")).Return(nil)

        service := NewExpenseService(mockRepo, logger)

        // Act（実行）
        result, err := service.Create(ctx, userID, input)

        // Assert（検証）
        assert.NoError(t, err)
        assert.NotNil(t, result)
        assert.Equal(t, input.Title, result.Title)
        assert.Equal(t, input.Amount, result.Amount)

        // モック呼び出し検証
        mockRepo.AssertExpectations(t)
    })

    t.Run("異常系: 金額が0以下の場合エラー", func(t *testing.T) {
        // ...
    })
}
```

---

## アサーション（testify/assert）

```go
import (
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// 基本アサーション
assert.Equal(t, expected, actual)
assert.NotEqual(t, expected, actual)
assert.Nil(t, value)
assert.NotNil(t, value)
assert.True(t, condition)
assert.False(t, condition)

// エラーチェック
assert.NoError(t, err)
assert.Error(t, err)
assert.ErrorContains(t, err, "expected message")

// 数値比較
assert.Greater(t, actual, expected)
assert.Less(t, actual, expected)
assert.GreaterOrEqual(t, actual, expected)

// スライス
assert.Len(t, slice, expectedLength)
assert.Empty(t, slice)
assert.Contains(t, slice, element)

// require: 失敗時に即座にテスト終了
require.NoError(t, err)  // これが失敗したら以降のコードは実行されない
```

---

## モック（testify/mock）

```go
// モック定義
type MockExpenseRepository struct {
    mock.Mock
}

func (m *MockExpenseRepository) Create(ctx context.Context, expense *model.Expense) error {
    args := m.Called(ctx, expense)
    return args.Error(0)
}

func (m *MockExpenseRepository) GetByID(ctx context.Context, id string) (*model.Expense, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*model.Expense), args.Error(1)
}

// モック使用
mockRepo := new(MockExpenseRepository)
mockRepo.On("Create", ctx, mock.AnythingOfType("*model.Expense")).Return(nil)
mockRepo.On("GetByID", ctx, "test-id").Return(&model.Expense{ID: "test-id"}, nil)

// 検証
mockRepo.AssertExpectations(t)
mockRepo.AssertCalled(t, "Create", ctx, mock.AnythingOfType("*model.Expense"))
mockRepo.AssertNumberOfCalls(t, "Create", 1)
```

---

## テストデータ（fixtures）

```go
// test/fixtures/users.go
package fixtures

func CreateTestUser() *model.User {
    return &model.User{
        ID:        uuid.New().String(),
        Email:     "test@example.com",
        FirstName: "Test",
        LastName:  "User",
        Role:      model.RoleEngineer,
    }
}

func CreateTestExpense(userID string) *model.Expense {
    return &model.Expense{
        ID:          uuid.New().String(),
        UserID:      userID,
        Title:       "テスト経費",
        Category:    model.ExpenseCategoryEntertainment,
        Amount:      5000,
        Status:      model.ExpenseStatusDraft,
        ExpenseDate: time.Now(),
    }
}
```

---

## テスト命名規約

### 日本語を使用

```go
// ✅ 推奨: 日本語で何をテストしているか明確に
t.Run("正常系: 経費申請の作成", func(t *testing.T) { })
t.Run("異常系: 金額が0以下の場合エラー", func(t *testing.T) { })
t.Run("境界値: 上限金額での申請", func(t *testing.T) { })
```

### カテゴリ分け

| カテゴリ | 説明 |
|---------|------|
| 正常系 | 期待通りの入力で期待通りの結果 |
| 異常系 | 不正な入力でエラーが発生 |
| 境界値 | 境界条件でのテスト |
| 権限 | 権限に基づくアクセス制御 |

---

## テスト実行コマンド

```bash
# 全テスト実行
cd backend && go test ./... -v

# 特定パッケージのテスト
cd backend && go test ./internal/service/... -v

# カバレッジ付き
cd backend && go test ./... -v -cover

# カバレッジレポート生成
cd backend && go test ./... -coverprofile=coverage.out
cd backend && go tool cover -html=coverage.out -o coverage.html

# 特定のテスト関数のみ
cd backend && go test ./... -v -run TestExpenseService_Create
```

---

## チェックリスト

- [ ] `*_test.go` ファイル名になっているか
- [ ] `t.Run` でテストケースをグループ化しているか
- [ ] `testify/assert` を使用しているか
- [ ] モックは `testify/mock` を使用しているか
- [ ] テストデータは fixtures に分離しているか
- [ ] テスト名は日本語で意図が明確か
- [ ] テストは独立して実行可能か
- [ ] テストの実行順序に依存していないか
- [ ] カバレッジ目標（80%）を達成しているか
