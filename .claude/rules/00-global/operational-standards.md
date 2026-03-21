---
paths: "**/*"
description: Claude Code運用における標準ルール（全ファイル適用）
---

# 運用標準規約

このドキュメントは、Claude Code運用における標準ルールを定義します。

---

## 1. メモリ・計画書の管理

### 1.1 タスク完了時のクリーンアップ（必須）

タスク完了の定義には、以下のクリーンアップが含まれます：

| 削除対象 | 例 | 理由 |
|---------|-----|------|
| 一時計画書 | `.claude/plans/*.md` | 実装完了後は不要 |
| 日付付きメモリ | `*_20250118.md`, `*_2025-01-18.md` | 一時的な調査記録 |
| 完了タスクメモリ | `*_complete.md`, `*_done.md` | 完了後は不要 |
| 進捗ログ | `*_progress.md`, `*_log.md` | 完了後は不要 |
| フェーズメモリ | `*_phase1.md`, `*_phase2.md` | 全フェーズ完了後に削除 |

### 1.2 メモリ作成時のルール

```
✅ 推奨: 永続的な知見・パターン
   - api_client_design_guidelines.md
   - common_pitfalls_*.md
   - testing-best-practices.md

❌ 避ける: 一時的な作業記録
   - task_progress_20250105.md
   - bug_investigation_log.md
   - implementation_phase1_complete.md
```

### 1.3 定期棚卸し

- **頻度**: 月1回程度
- **確認事項**:
  - 不要になったメモリの削除
  - 重複パターンの統合
  - 目標: 50個以下を維持

---

## 2. Skills vs Commands

### 2.1 Skill優先の原則

定型作業は `commands/` ではなく `skills/` として実装すること。

| 用途 | 推奨 | 理由 |
|------|------|------|
| 定型ワークフロー | `skills/` | 再利用性が高い |
| 複雑な処理 | `skills/` + `scripts/` | ロジック分離 |
| 単発の手順 | `commands/` | 簡易的な用途のみ |

### 2.2 Skill設計原則

```
skill-name/
├── SKILL.md           # 要約のみ（500行以下）
├── scripts/           # 実装ロジック（Python/Bash）
├── references/        # 詳細ドキュメント
└── assets/            # テンプレート等
```

**SKILL.mdに書くべき内容:**
- 概要（1-2段落）
- 使用条件・トリガー
- 基本的なワークフロー
- リソースへの参照

**SKILL.mdに書くべきでない内容:**
- 詳細な実装ロジック（→ scripts/へ）
- 長いリファレンス（→ references/へ）
- テンプレート（→ assets/へ）

---

## 3. MCP Server管理

### 3.1 最小化の原則

常駐させるMCPは必要最小限に保つ。

| MCP | 常駐 | 理由 |
|-----|------|------|
| serena | ✅ | コード操作の基盤 |
| context7 | ✅ | ドキュメント検索 |
| chrome-devtools | ⚠️ | UI確認時のみ |
| その他 | ❌ | タスク完了後は無効化 |

### 3.2 タスク完了時の確認

特定タスク用に有効化したMCPは、タスク完了時に以下を確認：

```
「このMCPは引き続き必要ですか？不要であれば設定から外すことを推奨します」
```

---

## 4. タスク完了チェックリスト

実装タスク完了時に確認：

- [ ] コードが正常に動作する
- [ ] テストが通過する
- [ ] 一時メモリ（日付付き、*_complete等）を削除した
- [ ] `.claude/plans/` の該当計画書を削除した
- [ ] 永続的な知見があればメモリに記録した
- [ ] 特定タスク用MCPの無効化を検討した

---

## 5. 命名規則

### 5.1 メモリファイル名

```
# 良い例（永続的な知見）
api_client_design_guidelines.md
common_pitfalls_user_roles.md
testing-best-practices.md

# 悪い例（一時的、日付付き）
bug_fix_20250105.md
task_progress_phase1.md
investigation_complete.md
```

### 5.2 計画書ファイル名

```
# 良い例（機能名のみ）
supply-chain-ui-implementation.md
user-authentication-redesign.md

# 悪い例（日付や状態を含む）
supply-chain-ui-20250105.md
user-auth-in-progress.md
```

---

## 更新履歴

- 2025-01-05: 初版作成（棚卸しレポートに基づく）
