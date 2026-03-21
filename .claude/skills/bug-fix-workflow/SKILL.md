---
name: bug-fix-workflow
description: バグ修正の統合ワークフロー（Codex設計レビュー自動実行付き）。調査→計画→Codex修正方針レビュー（自動）→修正→code-reviewerコードレビューを自律的に実行。「バグを修正して」「エラーが出る」「動かない」「不具合がある」「おかしい」といった要求で使用。
---

# バグ修正ワークフロー（Codex修正方針レビュー自動実行）

バグの調査→修正計画→Codex方針レビュー→修正→code-reviewerコードレビューを段階的に実行する統合ワークフロー。
**修正計画策定後にCodex方針レビューが自動実行される。**（緊急度Critical/Highの場合はスキップ）

## フロー図

```
INVESTIGATE → PLAN → [自動] CODEX_DESIGN_REVIEW ←→ FIX_PLAN → LGTM
                                                                 ↓
                                                               FIX
                                                                 ↓
                                                       CODE_REVIEW (code-reviewer)
                                                                 ↓
                                                               完了

※ Critical/High: INVESTIGATE → PLAN → FIX → CODE_REVIEW（Codexレビューをスキップ）
```

## アクション選択ガイド

| ユーザー要求 | 開始Phase | 詳細 |
|-------------|----------|------|
| 「バグを修正して」 | INVESTIGATE | `references/phase-investigate.md` |
| 「エラーが出る」 | INVESTIGATE | `references/phase-investigate.md` |
| 「動かない」 | INVESTIGATE | `references/phase-investigate.md` |
| 「不具合がある」 | INVESTIGATE | `references/phase-investigate.md` |

## 緊急度判定

| レベル | 条件 | Codexレビュー | 対応 |
|-------|------|-------------|------|
| Critical | 本番障害、データ損失 | **スキップ** | 即座にFIXへ |
| High | 主要機能停止 | **スキップ** | PLANを簡略化→FIX |
| Medium | 一部機能障害 | **自動実行** | 通常フロー |
| Low | 軽微なUI問題 | **自動実行** | 通常フロー |

## Phase概要

### Phase 1: INVESTIGATE（調査）
- 原因特定、再現手順確認、影響範囲分析
- 詳細: `references/phase-investigate.md`

### Phase 2: PLAN（修正計画策定）
- 根本原因に基づく修正方針を策定
- 詳細: `references/phase-plan.md`

### Phase 3: CODEX_DESIGN_REVIEW（自動実行 ※緊急時スキップ）
- **PLANが完了すると自動的にCodex方針レビューを実行する**
- Codexは `AGENT.md` + `.claude/rules/` を参照して修正方針をレビュー
- LGTMを取得するまで修正ループを繰り返す
- 詳細: `references/phase-codex-review.md`
- プロンプト: `references/review-prompt-templates.md`

### Phase 4: FIX（修正実装）
- LGTM済み方針に基づき修正
- サブエージェント活用（backend-developer / frontend-developer）
- 詳細: `references/phase-fix.md`

### Phase 5: CODE_REVIEW（code-reviewerコードレビュー）
- `code-reviewer` サブエージェントでコードレビュー
- `run_in_background: true` でバックグラウンド実行

## 重要ルール

1. **緊急度判定を最初に行う**: Critical/HighならCodexレビューをスキップ
2. **Codex方針レビューは自動**: PLAN完了後、緊急度Medium/Lowなら確認なく実行
3. **セッション管理**: `mcp__codex__codex` で開始、`mcp__codex__codex-reply` で継続
4. **LGTMまで繰り返し**: 方針レビューでLGTMを取得するまで修正→再レビューを繰り返す
5. **LGTM判定**: Codex応答に「LGTM」等の承認表現が含まれれば合格
6. **⚠️ レビュー結果の待機（最重要）**: Codexにレビューを送信した後、レビュー結果（LGTM or 指摘事項）を含むレスポンスを**必ず受け取ってから**次のフェーズに進むこと。Codexが行動計画の確認のみを返した場合は `mcp__codex__codex-reply` で承認し、実際のレビュー結果が返るまで待機する。**レビュー結果なしで次フェーズに進むことは絶対に禁止。**

## 終了条件マトリクス

| Phase | 結果 | 次のアクション |
|-------|------|--------------|
| INVESTIGATE | SUCCESS | → PLAN |
| INVESTIGATE | CANNOT_REPRODUCE | 追加情報要求 |
| PLAN | SUCCESS (Medium/Low) | → CODEX_DESIGN_REVIEW（自動） |
| PLAN | SUCCESS (Critical/High) | → FIX（Codexスキップ） |
| PLAN | URGENT_FIX | → 即座にFIX |
| CODEX_DESIGN_REVIEW | LGTM | → FIX |
| CODEX_DESIGN_REVIEW | NEEDS_FIX | 計画修正 → 再レビュー（LGTMまで繰り返す） |
| FIX | SUCCESS | → CODE_REVIEW |
| FIX | NEED_REFACTORING | リファクタリングへ |
| CODE_REVIEW | APPROVED | 完了 |
| CODE_REVIEW | NEEDS_FIX | 修正 → 再レビュー |

## 出力ファイル

| Phase | 出力 |
|-------|------|
| 調査 | `docs/investigate/bug-investigate_{TIMESTAMP}.md` |
| 計画 | `docs/plan/bug-plan_{TIMESTAMP}.md` |
| 修正 | `docs/fix/bug-fix_{TIMESTAMP}.md` |

## リファレンス

- `references/phase-investigate.md` - 調査フェーズ詳細
- `references/phase-plan.md` - 計画フェーズ詳細
- `references/phase-codex-review.md` - Codex方針レビューフェーズ詳細
- `references/review-prompt-templates.md` - Codexへ送るプロンプトテンプレート
- `references/phase-fix.md` - 修正フェーズ詳細
- `references/fix-patterns.md` - バグ修正パターン集
- `references/root-cause-analysis.md` - 根本原因分析（5 Whys等）
