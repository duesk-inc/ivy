# バグ修正パターン集

## 入力検証エラー

### パターン: バリデーション追加
```go
// Before
func Handler(c *gin.Context) {
    id := c.Param("id")
    // 直接使用
}

// After
func Handler(c *gin.Context) {
    id := c.Param("id")
    if _, err := uuid.Parse(id); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }
}
```

## Null/Undefined エラー

### パターン: 存在チェック追加
```typescript
// Before
const name = user.profile.name;

// After
const name = user?.profile?.name ?? '';
```

## N+1クエリ

### パターン: プリロード
```go
// Before
for _, user := range users {
    orders, _ := repo.FindByUserID(user.ID)
}

// After
users, _ := repo.FindAllWithOrders()
```

## 状態競合

### パターン: トランザクション追加
```go
// Before
user, _ := repo.FindByID(id)
user.Status = "active"
repo.Update(user)

// After
err := db.Transaction(func(tx *gorm.DB) error {
    user, _ := repo.WithTx(tx).FindByID(id)
    user.Status = "active"
    return repo.WithTx(tx).Update(user)
})
```

## 型不一致

### パターン: 型変換の明示化
```typescript
// Before
const id = params.id; // string | string[]

// After
const id = Array.isArray(params.id) ? params.id[0] : params.id;
```
