# Codex設計レビュー プロンプトテンプレート

## 設計レビュー開始用プロンプト

`mcp__codex__codex` の `prompt` パラメータに使用する。
`{PLAN_CONTENT}` を実際の計画内容で置換すること。

```
以下の実装計画をレビューしてください。

レビューの前に、以下のファイルを読んでプロジェクトの規約を把握してください:
1. .claude/rules/05-pitfalls/known-issues.md（必須: 既知の落とし穴）
2. 設計対象に応じた .claude/rules/ 配下のファイル

## 実装計画

{PLAN_CONTENT}

AGENT.md の回答形式に従って回答してください。
問題がなければ「LGTM」、問題があれば [必須]/[推奨]/[提案] で指摘してください。
```

### mcp__codex__codex 呼び出し例

```
mcp__codex__codex(
  prompt: "以下の実装計画をレビューしてください。\n\nレビューの前に...(上記テンプレート)",
  cwd: "/Users/.../monstera",
  sandbox: "read-only"
)
```

## 修正後の再レビュー用プロンプト

`mcp__codex__codex-reply` の `prompt` パラメータに使用する。

```
前回の指摘に基づき設計を修正しました。

## 対応した指摘
{ADDRESSED_ISSUES}

## 修正後の設計
{UPDATED_PLAN}

## 未対応の指摘と理由（ある場合）
{SKIP_REASONS}

再度レビューをお願いします。問題がなければ「LGTM」と回答してください。
```

### mcp__codex__codex-reply 呼び出し例

```
mcp__codex__codex-reply(
  conversationId: "<Phase 2で取得したID>",
  prompt: "前回の指摘に基づき設計を修正しました。\n\n...(上記テンプレート)"
)
```
