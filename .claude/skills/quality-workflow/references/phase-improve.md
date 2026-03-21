# Phase 2: 改善 (IMPROVE)

## 目的
監査で特定された問題を優先順位に従って改善する。

## 実行手順

### 1. ブランチ作成
```bash
git checkout -b quality/improvement-description
```

### 2. セキュリティ改善
```go
// Before: 入力検証なし
func Handler(c *gin.Context) {
    id := c.Param("id")
    // 直接使用
}

// After: 入力検証追加
func Handler(c *gin.Context) {
    id := c.Param("id")
    if _, err := uuid.Parse(id); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }
}
```

### 3. パフォーマンス改善
```go
// Before: N+1クエリ
for _, user := range users {
    orders, _ := repo.FindByUserID(user.ID)
}

// After: プリロード
users, _ := repo.FindAllWithOrders()
```

### 4. 保守性改善
- 複雑な関数の分割
- 重複コードの共通化
- 命名の統一

### 5. テスト追加
- 改善箇所のテスト追加
- リグレッションテスト

### 6. 出力
`docs/improve/quality-improve_{TIMESTAMP}.md`

## 改善チェックリスト
- [ ] セキュリティ問題を解決したか
- [ ] パフォーマンス問題を解決したか
- [ ] 既存テストが通るか
- [ ] 新しいテストを追加したか
- [ ] コードレビュー基準を満たすか

## 終了条件
- SUCCESS → Phase 3へ
- PARTIAL_COMPLETE → 追加改善へ
