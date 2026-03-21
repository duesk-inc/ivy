# Phase 1: 調査 (INVESTIGATE)

## 目的
バグの根本原因を特定し、影響範囲を把握する。

## 入力情報
- バグの症状・エラーメッセージ
- 再現手順
- 発生環境（開発/本番）
- 関連するログ

## 実行手順

### 1. バグ再現の試行
- 報告された手順で再現を確認
- 再現できない場合は追加情報を要求

### 2. Serenaでの調査
```bash
search_for_pattern("エラーメッセージの一部")
find_symbol("エラー発生クラス/関数")
find_referencing_symbols("問題シンボル")
read_memory("common_pitfalls_*")
```

### 3. データフロー追跡
- 入力から出力までの処理経路を把握
- 関連コンポーネント（DB、API、状態管理）を確認

### 4. 調査チェックリスト
- [ ] エラーの直接的な原因
- [ ] エラー発生条件
- [ ] 影響を受ける機能・ユーザー
- [ ] データ整合性への影響
- [ ] セキュリティへの影響

### 5. 出力
`docs/investigate/bug-investigate_{TIMESTAMP}.md`

## 終了条件
- SUCCESS → Phase 2へ
- CANNOT_REPRODUCE → 追加情報要求
- NEED_DEEP_ANALYSIS → 根本原因分析へ
