# Phase 3: 検証 (VERIFY)

## 目的
改善後の品質を検証する。

## 実行手順

### 1. 静的解析の再実行
```bash
cd backend && golangci-lint run
cd frontend && npm run lint
```

### 2. テスト実行
```bash
cd backend && go test ./... -v -cover
cd frontend && npm run test
```

### 3. パフォーマンステスト
- ベンチマーク実行
- 改善前後の比較

### 4. セキュリティテスト
- 入力検証テスト
- 認証・認可テスト

## 検証チェックリスト
- [ ] 静的解析でエラーがないか
- [ ] 全テストがパスするか
- [ ] カバレッジが基準を満たすか
- [ ] パフォーマンスが改善されたか

## 終了条件
- SUCCESS → 完了
