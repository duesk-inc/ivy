# Phase 3: 実装 (IMPLEMENT) + Phase 4: コードレビュー (CODE_REVIEW)

## 目的
Codex LGTMを得た設計に基づき実装し、code-reviewerでコードレビューを行う。

---

## Phase 3: 実装

### 1. 実装前確認
- LGTM済み設計書を再確認
- Codexからのフィードバックが反映済みか確認

### 2. サブエージェント活用

適切なサブエージェントに実装を委譲する：

| 対象 | サブエージェント |
|------|----------------|
| Go API（Handler/Service/Repository/DB） | `backend-developer` |
| Next.js/React（コンポーネント/フック） | `frontend-developer` |

```
Task(
  subagent_type="backend-developer",
  prompt="[LGTM済み設計に基づく実装指示]"
)
```

### 3. 実装順序
```
データベース → Repository → Service → Handler → Frontend
```

### 4. 品質確認（実装完了後）
```bash
# バックエンド
cd backend && go build ./...
cd backend && go test ./... -v

# フロントエンド
cd frontend && npm run lint
cd frontend && npm run type-check
```

### 5. 終了条件
- SUCCESS → CODE_REVIEW へ
- NEED_REDESIGN → Phase 1（PLAN）へ戻る

---

## Phase 4: コードレビュー（code-reviewer）

### コードレビューはClaude Codeのcode-reviewerが担当

Codexではなく、プロジェクト規約を熟知した `code-reviewer` サブエージェントを使う。

### 実行方法

```
Task(
  subagent_type="code-reviewer",
  prompt="以下の実装をレビューしてください。

## 設計背景
[Codex LGTMを得た設計の概要]

## 変更ファイル
[変更されたファイル一覧]

## レビュー重点
- 設計通りに実装されているか
- プロジェクト規約に準拠しているか
- セキュリティ・パフォーマンスの問題がないか",
  run_in_background=true
)
```

### レビュー結果の処理

| code-reviewer判定 | アクション |
|-------------------|----------|
| 必須修正 0件 | 完了 |
| 必須修正 あり | 指摘に基づき修正 → 再レビュー |
| 推奨修正のみ | ユーザーに対応要否を確認 |

### ユーザーへの報告

```markdown
## 開発完了サマリー

### 設計レビュー（Codex）
- 判定: LGTM
- レビュー回数: N回（修正M回）

### コードレビュー（code-reviewer）
- 必須修正: X件
- 推奨修正: Y件
- 提案: Z件

### 最終ステータス: [完了 / 修正対応中]
```
