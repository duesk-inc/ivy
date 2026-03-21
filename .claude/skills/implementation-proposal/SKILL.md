---
name: implementation-proposal
description: 実装提案ワークフロー（Codex設計レビュー自動実行付き）。複数の実装アプローチを比較検討し最適案を提案、Codexが自動レビュー。「どう実装すべきか」「実装方法を提案して」「アーキテクチャを検討して」といった要求で使用。
---

# 実装提案ワークフロー（Codex設計レビュー自動実行）

要件に対する複数の実装アプローチを検討し、最適な実装案を提案する統合ワークフロー。
**推奨案の提示後にCodex設計レビューが自動実行される。**

## フロー図

```
ANALYZE → DESIGN → EVALUATE → PROPOSE → [自動] CODEX_DESIGN_REVIEW ←→ FIX → LGTM
    ↓        ↓         ↓          ↓                                              ↓
  要件分析  複数案設計  比較評価   推奨案提示                                   確定
```

## アクション選択ガイド

| ユーザー要求 | 開始Phase | 詳細 |
|-------------|----------|------|
| 「どう実装すべきか」 | ANALYZE | `references/phase-analyze.md` |
| 「実装方法を提案して」 | ANALYZE | `references/phase-analyze.md` |
| 「アーキテクチャを検討して」 | DESIGN | `references/phase-design.md` |
| 「パフォーマンス重視で」 | EVALUATE | `references/phase-evaluate.md` |

## 必要な入力

- 実装したい機能の要件
- 制約条件（期限、リソース、技術的制約）
- 優先順位（パフォーマンス重視、保守性重視など）

## 評価軸

| 評価項目 | 標準重み | 説明 |
|---------|---------|------|
| 実装工数 | 20% | 開発にかかる時間 |
| パフォーマンス | 25% | 応答速度、スループット |
| 保守性 | 20% | 可読性、変更容易性 |
| 拡張性 | 15% | 将来の機能追加への対応 |
| セキュリティ | 10% | 脆弱性リスク |
| テスト容易性 | 10% | テストの書きやすさ |

## Phase概要

### Phase 1: ANALYZE（要件分析）
- 要件の明確化と制約条件の整理
- 詳細: `references/phase-analyze.md`

### Phase 2: DESIGN（複数案設計）
- 2〜3の実装アプローチを設計
- 詳細: `references/phase-design.md`

### Phase 3: EVALUATE（比較評価）
- 評価軸に基づく各案の比較
- 詳細: `references/phase-evaluate.md`

### Phase 4: PROPOSE（推奨案提示）
- 推奨案をユーザーに提示
- 詳細: `references/proposal-templates.md`

### Phase 5: CODEX_DESIGN_REVIEW（自動実行）
- **PROPOSEが完了すると自動的にCodex設計レビューを実行する**
- Codexは `AGENT.md` + `.claude/rules/` を参照して推奨案をレビュー
- LGTMを取得するまで修正ループを繰り返す
- 詳細: `references/phase-codex-review.md`
- プロンプト: `references/review-prompt-templates.md`

## 重要ルール

1. **Codex設計レビューは自動**: PROPOSE完了後、確認なく実行
2. **セッション管理**: `mcp__codex__codex` で開始、`mcp__codex__codex-reply` で継続
3. **LGTMまで繰り返し**: 設計レビューでLGTMを取得するまで修正→再レビューを繰り返す
4. **LGTM判定**: Codex応答に「LGTM」等の承認表現が含まれれば合格
5. **⚠️ レビュー結果の待機（最重要）**: Codexにレビューを送信した後、レビュー結果（LGTM or 指摘事項）を含むレスポンスを**必ず受け取ってから**次のフェーズに進むこと。Codexが行動計画の確認のみを返した場合は `mcp__codex__codex-reply` で承認し、実際のレビュー結果が返るまで待機する。**レビュー結果なしで次フェーズに進むことは絶対に禁止。**

## 終了条件マトリクス

| Phase | 結果 | 次のアクション |
|-------|------|--------------|
| ANALYZE | SUCCESS | → DESIGN |
| ANALYZE | NEED_CLARIFICATION | ユーザーに質問 |
| DESIGN | SUCCESS | → EVALUATE |
| EVALUATE | SUCCESS | → PROPOSE |
| PROPOSE | SUCCESS | → CODEX_DESIGN_REVIEW（自動） |
| CODEX_DESIGN_REVIEW | LGTM | 確定（実装計画へ進行可能） |
| CODEX_DESIGN_REVIEW | NEEDS_FIX | 提案修正 → 再レビュー（LGTMまで繰り返す） |

## リファレンス

- `references/phase-analyze.md` - 要件分析フェーズ詳細
- `references/phase-design.md` - 設計フェーズ詳細
- `references/phase-evaluate.md` - 評価フェーズ詳細
- `references/proposal-templates.md` - 提案書テンプレート
- `references/phase-codex-review.md` - Codex設計レビューフェーズ詳細
- `references/review-prompt-templates.md` - Codexへ送るプロンプトテンプレート
