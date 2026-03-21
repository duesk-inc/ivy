# Phase 2: 計画 (PLAN)

## 目的
安全で効果的なバグ修正計画を策定する。

## 実行手順

### 1. 既存パターン確認
```bash
read_memory("common_pitfalls_*")
find_referencing_symbols("修正対象", depth=2)
search_for_pattern("test.*修正対象")
```

### 2. 修正方針決定
- パッチ修正 vs 根本修正
- 適用する修正パターン
- エラーハンドリングの改善

### 3. 段階的修正計画

**Phase A: 緊急対応（ホットフィックス）**
- エラー再現テストケース作成
- 暫定的な修正実装
- 影響範囲の限定的テスト

**Phase B: 根本修正**
- 詳細な影響調査
- コア修正の実装
- 関連箇所の修正

**Phase C: 予防措置**
- 入力検証の強化
- エラーメッセージの改善
- 同様パターンの検索と修正

### 4. リスク評価
- 修正による副作用のリスク
- デグレーションのリスク
- パフォーマンスへの影響

### 5. 出力
`docs/plan/bug-plan_{TIMESTAMP}.md`

## 終了条件
- SUCCESS → Phase 3へ
- URGENT_FIX → 即座にPhase 3へ
- NEED_ANALYSIS → 追加分析
