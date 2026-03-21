# Phase 1: 調査 (INVESTIGATE)

## 目的
既存システムへの影響と実装可能性を評価する。

## 実行手順

### 1. Serenaメモリ確認
- `read_memory("project_overview")` - プロジェクト概要
- `read_memory("coding_conventions")` - コーディング規約

### 2. 類似機能調査
```bash
search_for_pattern("類似機能パターン")
get_symbols_overview("関連ディレクトリ")
find_symbol("関連クラス/関数")
```

### 3. 調査チェックリスト
- [ ] 既存システムとの統合ポイント
- [ ] データモデルへの影響
- [ ] API設計の方針
- [ ] 認証・認可の要件
- [ ] 既存機能への影響範囲

### 4. 出力
`docs/investigate/new-feature-investigate_{TIMESTAMP}.md`

## 終了条件
- SUCCESS → Phase 2へ
- NEED_CLARIFICATION → ユーザーに質問
- TECHNICAL_CHALLENGE → 代替案検討
