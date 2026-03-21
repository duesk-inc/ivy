# Phase 1: 監査 (AUDIT)

## 目的
コード品質の問題を総合的に特定し、改善の優先順位を決定する。

## 入力情報
- 監査対象（モジュール、機能、ファイル）
- 監査の重点項目（セキュリティ/パフォーマンス/保守性）

## 実行手順

### 1. 静的解析
```bash
# Go
cd backend && go vet ./...
cd backend && golangci-lint run

# TypeScript
cd frontend && npm run lint
cd frontend && npm run type-check
```

### 2. Serenaでの構造分析
```bash
get_symbols_overview("対象ディレクトリ")
find_symbol("対象クラス/関数", depth=2)
find_referencing_symbols("重要シンボル")
```

### 3. 監査カテゴリ

**セキュリティ監査**
- [ ] 入力検証の不備
- [ ] 認証・認可の問題
- [ ] SQLインジェクションの可能性
- [ ] XSSの可能性
- [ ] 機密情報のハードコーディング

**パフォーマンス監査**
- [ ] N+1クエリの存在
- [ ] 不要なDB呼び出し
- [ ] メモリリークの可能性
- [ ] 不適切なキャッシュ使用

**保守性監査**
- [ ] 複雑度の高い関数
- [ ] 重複コード
- [ ] 命名の不一致
- [ ] テストカバレッジ不足

**アーキテクチャ監査**
- [ ] 循環依存
- [ ] 責務の不適切な配置
- [ ] 層の境界違反

### 4. 問題の優先順位付け

| 優先度 | 条件 |
|-------|------|
| Critical | セキュリティ脆弱性、データ損失リスク |
| High | 本番パフォーマンス問題 |
| Medium | 保守性の問題 |
| Low | コードスタイル |

### 5. 出力
`docs/audit/quality-audit_{TIMESTAMP}.md`

## 終了条件
- SUCCESS_NO_ISSUES → 問題なし
- ISSUES_FOUND → Phase 2へ
