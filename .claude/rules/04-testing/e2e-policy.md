---
paths: frontend/e2e/**/*.ts, frontend/e2e/**/*.spec.ts, tests/**/*.spec.ts, **/*.cy.ts
description: E2Eテストにおける標準ワークフローとデバッグ方針
---

# E2Eテストポリシー

## 基本原則

UIテストの実装・修正・実行時は、以下のツールを使用する。

| 分類 | ツール | 用途 | 必須度 |
|------|--------|------|--------|
| 必須 | **Playwright** | E2Eテストランナー | ✅ |
| 推奨 | **report/trace** | 失敗時の一次解析 | ✅ (FAIL時) |
| 条件付き | **Chrome DevTools** (MCP) | DOM/console深掘り | report/traceで不足時 |
| 条件付き | **Claude in Chrome** | 視覚・動線確認 | UI変更時推奨 |

> **方針**: 基本はPlaywright成果物（report/trace）で解析し、MCP/Claude in Chromeは「必要時のみ」使用する。

---

## 標準フロー

```
RUN → CHECK
  │
  ├─ [PASS] → DONE（UI変更時はVISUAL推奨）
  │
  └─ [FAIL] → ANALYZE (report/trace)
               │
               ├─ 原因判明 → FIX → RE-RUN
               │
               └─ 不明 → INSPECT (MCP) → FIX → RE-RUN
```

### 失敗時の調査（ANALYZE → 必要ならINSPECT）

1. **ANALYZE（一次解析）**【必須】
   - HTML report で失敗箇所・スクリーンショットを確認
   - trace で操作の時系列・ネットワークを確認
   - **多くの場合、ここで原因が判明する**

2. **INSPECT（MCP調査）**【条件付き】
   - report/trace で不足する場合のみ
   - DOM状態の詳細確認
   - コンソールログの詳細確認

### 成功時の視覚検証（VISUAL）【条件付き】

**テストPASS時、毎回の視覚検証は不要。**

視覚検証が推奨されるケース:
- UI変更を含むPRの最終確認
- レイアウト・スタイルに関するバグ修正後
- レスポンシブ対応の確認

---

## なぜPlaywright成果物が先なのか

| 理由 | 説明 |
|------|------|
| 効率性 | report/traceは即座に確認可能、MCPは起動・操作のオーバーヘッドがある |
| 情報量 | traceには操作履歴・ネットワーク・DOM状態がすべて記録されている |
| 再現性 | traceは失敗時の状態を正確に保存、MCPは現在の状態しか見えない |
| 運用負荷 | MCP/Chromeを常時使用するとプロセスが増殖しやすい |

---

## 禁止事項

### 推測に基づく修正

```typescript
// ❌ 禁止: report/traceを確認せず修正
// 「たぶんボタンが表示されていないはず」
await page.waitForSelector('.submit-button');

// ❌ 禁止: 「セレクタが間違っていそう」で変更
await page.click('[data-testid="submit"]'); // 確認なしで書いた
```

### 確認なしでの繰り返し修正

```typescript
// ❌ 禁止: 失敗 → セレクタ変更 → 失敗 → セレクタ変更...の繰り返し
// report/traceを見れば1回で正しいセレクタが判明する
```

---

## 正しいデバッグフロー

### Step 1: ANALYZE（report/trace確認）【必須】

```bash
# HTML report を開く
.claude/skills/run-ui-test/scripts/03-open-report.sh

# trace を開く（あれば）
.claude/skills/run-ui-test/scripts/04-open-trace.sh
```

確認項目:
- どのテストが失敗したか
- どのステップで失敗したか
- 失敗時のスクリーンショット
- ネットワークリクエスト（trace）

### Step 2: INSPECT（MCP調査）【条件付き】

以下の場合のみ使用:
- [ ] DOM状態の詳細確定が必要
- [ ] コンソールログ/エラーの詳細確認が必要
- [ ] traceが記録されていなかった
- [ ] 再現が難しく実ブラウザで確認したい

```
mcp__chrome-devtools__screenshot
mcp__chrome-devtools__getPageContent
mcp__chrome-devtools__getConsoleLogs
mcp__chrome-devtools__evaluate
```

### Step 3: 事実に基づく修正

```typescript
// ✅ 正しい: report/traceで確認した後
// 確認結果: ボタンは存在するが disabled 状態
await page.waitForSelector('button[type="submit"]:not([disabled])');

// ✅ 正しい: コンソールで API エラーを確認した後
// 確認結果: 401 Unauthorized が発生していた
// → 認証トークンの設定を修正
```

---

## スキルとの連携

E2Eテスト実行時は `run-ui-test` スキルを使用すること。

スキルには以下が含まれる:
- `scripts/01-run.sh` - テスト実行
- `scripts/03-open-report.sh` - report確認
- `scripts/04-open-trace.sh` - trace確認
- `scripts/02-inspect.sh` - MCP検証ガイド

---

## エージェントへの指示

### frontend-developer エージェント使用時

E2Eテストに関わる作業では:

1. テスト失敗時、**まず report/trace を確認**
2. report/trace で不足する場合のみ MCP を使用
3. 「事実」を報告した上で修正案を提示
4. UI変更がある場合は視覚検証を推奨

### bug-investigator エージェント使用時

UIに関連するバグ調査では:

1. 再現手順を **report/trace** で記録
2. 必要な場合のみ MCP でスクリーンショット取得
3. エラー発生タイミングのログを確認

---

## チェックリスト

E2Eテスト作業時:

### 常に必要

- [ ] `frontend` ディレクトリに `playwright.config.ts` が存在するか
- [ ] テスト対象のアプリケーションが起動しているか（`localhost:3000`）

### 失敗時（FAIL）

- [ ] HTML report で失敗箇所を確認したか
- [ ] trace で操作の時系列を確認したか（あれば）
- [ ] report/trace で原因が判明したか
  - Yes → FIX へ
  - No → INSPECT（MCP）へ
- [ ] 「推測」ではなく「事実」に基づいて修正したか
- [ ] 修正後にテストを再実行して PASS を確認したか

### 成功時（PASS）

- [ ] UI変更がある場合は視覚検証を検討したか

---

## Visual Verification Policy（条件付き）

### 概要

テストPASS = 完了 とする。ただし、UI変更がある場合は視覚検証を推奨。

### 視覚検証が推奨されるケース

| ケース | 推奨度 |
|--------|--------|
| UI変更を含むPR | ✅ 推奨 |
| レイアウト・スタイル変更 | ✅ 推奨 |
| レスポンシブ対応 | ✅ 推奨 |
| ロジックのみの変更 | ❌ 不要 |
| バックエンドのみの変更 | ❌ 不要 |

### 禁止事項

```
❌ テストPASS時に「毎回必須」で視覚検証を実行
❌ 「念のため」での視覚検証
❌ MCP/Claude in Chromeの常時起動
```

### 推奨事項

```
✅ UI変更時のみ Claude in Chrome で確認
✅ 確認後は即座にブラウザを閉じる
✅ 反復実行時はPlaywright成果物で確認
```
