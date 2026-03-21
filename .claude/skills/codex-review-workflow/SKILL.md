---
name: codex-review-workflow
description: 機能開発・改修の統合ワークフロー（Codex設計レビュー自動実行付き）。新機能追加・既存機能の改修・拡張・変更に対応。調査→計画→Codex設計レビュー（自動）→実装→code-reviewerコードレビューの流れで品質を担保する。「新機能を追加したい」「〜機能を実装したい」「〜できるようにしたい」「新しい機能を開発して」「既存機能を改修したい」「〜を変更したい」「〜を拡張したい」「〜に機能を追加したい」「〜を改善したい」「〜の仕様を変えたい」「Codexレビュー付きで開発したい」「設計をCodexにチェックさせたい」といった要求で使用。
---

# 機能開発・改修ワークフロー（Codex設計レビュー自動実行）

新機能開発および既存機能の改修・拡張に対応する統合ワークフロー。
調査→計画→Codex設計レビュー→実装→code-reviewerコードレビューを段階的に実行。
**計画策定後にCodex設計レビューが自動実行される。** ユーザーの明示的な指示は不要。

## フロー図

```
INVESTIGATE → PLAN → [自動] CODEX_DESIGN_REVIEW ←→ FIX → LGTM
                                                           ↓
                                                       IMPLEMENT
                                                           ↓
                                                   CODE_REVIEW (code-reviewer)
                                                           ↓
                                                         完了
```

## アクション選択ガイド

| ユーザー要求 | 開始Phase | 詳細 |
|-------------|----------|------|
| **新機能開発** | | |
| 「新機能を追加したい」 | INVESTIGATE | `references/phase-investigate.md` |
| 「〜機能を実装したい」 | INVESTIGATE | `references/phase-investigate.md` |
| 「〜できるようにしたい」 | INVESTIGATE | `references/phase-investigate.md` |
| 「新しい機能を開発して」 | INVESTIGATE | `references/phase-investigate.md` |
| **既存機能の改修・拡張** | | |
| 「既存機能を改修したい」 | INVESTIGATE | `references/phase-investigate.md` |
| 「〜を変更したい」 | INVESTIGATE | `references/phase-investigate.md` |
| 「〜を拡張したい」 | INVESTIGATE | `references/phase-investigate.md` |
| 「〜に機能を追加したい」 | INVESTIGATE | `references/phase-investigate.md` |
| 「〜を改善したい」 | INVESTIGATE | `references/phase-investigate.md` |
| 「〜の仕様を変えたい」 | INVESTIGATE | `references/phase-investigate.md` |
| 「〜にフィルターを追加したい」 | INVESTIGATE | `references/phase-investigate.md` |
| **直接指定** | | |
| 「設計をCodexにチェックさせたい」 | PLAN | `references/phase-plan.md` |

## 実装順序の原則

```
データベース → Repository → Service → Handler → Frontend
     ↓            ↓           ↓         ↓          ↓
  マイグレ     CRUD実装   ビジネス    API定義    UI実装
```

## Phase概要

### Phase 1: INVESTIGATE（調査）
- 既存システムへの影響と実装可能性を評価
- 詳細: `references/phase-investigate.md`

### Phase 2: PLAN（計画策定）
- 調査結果を基に詳細設計と実装計画を策定
- 詳細: `references/phase-plan.md`

### Phase 3: CODEX_DESIGN_REVIEW（自動実行）
- **PLANが完了すると自動的にCodex設計レビューを実行する**
- `mcp__codex__codex` でCodexセッション開始、設計をレビュー依頼
- Codexは `AGENT.md` + `.claude/rules/` を参照してレビュー
- LGTMを取得するまで修正ループを繰り返す
- 詳細: `references/phase-codex-review.md`
- プロンプト: `references/review-prompt-templates.md`

### Phase 4: IMPLEMENT（実装）
- LGTM済み設計に基づき実装
- サブエージェント活用（backend-developer / frontend-developer）
- 詳細: `references/phase-implement.md`

### Phase 5: CODE_REVIEW（code-reviewerコードレビュー）
- `code-reviewer` サブエージェントでコードレビュー
- 詳細: `references/phase-implement.md`

## 重要ルール

1. **Codex設計レビューは自動**: PLAN完了後、ユーザーへの確認なくCodexへ送信する
2. **AGENT.md必須**: プロジェクトルートの `AGENT.md` がCodexに設計知識を与える
3. **Codexセッション管理**: `mcp__codex__codex` で開始し、conversationIdを保持。修正後は `mcp__codex__codex-reply` で継続
4. **LGTMまで繰り返し**: 設計レビューでLGTMを取得するまで修正→再レビューを繰り返す
5. **LGTM判定**: Codex応答に「LGTM」「問題なし」「approve」等が含まれれば合格
6. **コードレビュー**: `code-reviewer` サブエージェントに委譲（`run_in_background: true`）
7. **⚠️ レビュー結果の待機（最重要）**: Codexにレビューを送信した後、レビュー結果（LGTM or 指摘事項）を含むレスポンスを**必ず受け取ってから**次のフェーズに進むこと。Codexが行動計画の確認のみを返した場合は `mcp__codex__codex-reply` で承認し、実際のレビュー結果が返るまで待機する。**レビュー結果なしでIMPLEMENTに進むことは絶対に禁止。**

## 終了条件マトリクス

| Phase | 結果 | 次のアクション |
|-------|------|--------------|
| INVESTIGATE | SUCCESS | → PLAN |
| INVESTIGATE | NEED_CLARIFICATION | ユーザーに質問 |
| PLAN | SUCCESS | → CODEX_DESIGN_REVIEW（自動） |
| PLAN | NEED_REDESIGN | → INVESTIGATE |
| CODEX_DESIGN_REVIEW | LGTM | → IMPLEMENT |
| CODEX_DESIGN_REVIEW | NEEDS_FIX | 計画修正 → 再レビュー（LGTMまで繰り返す） |
| IMPLEMENT | SUCCESS | → CODE_REVIEW |
| IMPLEMENT | NEED_REDESIGN | → PLAN |
| CODE_REVIEW | APPROVED | 完了 |
| CODE_REVIEW | NEEDS_FIX | 修正 → 再レビュー |

## 出力ファイル

| Phase | 出力 |
|-------|------|
| 調査 | `docs/investigate/feature-investigate_{TIMESTAMP}.md` |
| 計画 | `docs/plan/feature-plan_{TIMESTAMP}.md` |
| 実装 | `docs/implement/feature-implement_{TIMESTAMP}.md` |

## リファレンス

- `references/phase-investigate.md` - 調査フェーズ詳細
- `references/phase-plan.md` - 計画策定フェーズ詳細
- `references/phase-codex-review.md` - Codex設計レビューフェーズ詳細
- `references/phase-implement.md` - 実装 + code-reviewerフェーズ詳細
- `references/review-prompt-templates.md` - Codexへ送るプロンプトテンプレート
- `references/implementation-patterns.md` - 実装パターン集
