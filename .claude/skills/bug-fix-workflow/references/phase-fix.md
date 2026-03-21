# Phase 3: 修正 (FIX)

## 目的
バグを確実に修正し、デグレを防止する。

## 実行手順

### 1. 修正前の影響確認（必須）
```bash
find_referencing_symbols("修正対象")  # 全参照把握
# 既存テストの確認
```

### 2. ブランチ作成
```bash
git checkout -b bugfix/issue-description
```

### 3. バグ修正の実装
- 最小限の変更で確実に修正
- エラーハンドリングの改善
- 入力検証の強化

### 4. テストの追加・修正
- バグ再現テストケース追加
- 修正検証テストケース追加
- 既存テストの更新

### 5. リグレッションテスト
```bash
# バックエンド
cd backend && go test ./... -v

# フロントエンド
cd frontend && npm run lint
cd frontend && npm run type-check
```

### 6. 出力
`docs/fix/bug-fix_{TIMESTAMP}.md`

## 修正チェックリスト
- [ ] 根本原因が解決されているか
- [ ] 影響範囲をすべて修正したか
- [ ] テストケースを追加したか
- [ ] 既存のテストが通るか
- [ ] エラーハンドリングは適切か

## 終了条件
- SUCCESS → テストフェーズへ
- PARTIAL_FIX → 関連箇所の修正継続
- NEED_REFACTORING → リファクタリングへ
