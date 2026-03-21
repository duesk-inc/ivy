---
name: quality-workflow
description: コード品質管理の統合ワークフロー。監査→改善を自律的に実行。「品質を確認して」「問題がないか調べて」「改善点を見つけて」「品質を向上させて」といった要求で使用。
---

# コード品質管理ワークフロー

コードの品質監査と改善を段階的に実行する統合ワークフロー。

## フロー図

```
AUDIT → IMPROVE → VERIFY
   ↓        ↓        ↓
 問題特定  改善実施  品質確認
```

## アクション選択ガイド

| ユーザー要求 | 開始Phase | 詳細 |
|-------------|----------|------|
| 「品質を確認して」 | AUDIT | `references/phase-audit.md` |
| 「問題がないか調べて」 | AUDIT | `references/phase-audit.md` |
| 「改善点を見つけて」 | AUDIT | `references/phase-audit.md` |
| 「品質を向上させて」 | AUDIT → IMPROVE | `references/phase-improve.md` |
| 「セキュリティを確認」 | AUDIT (セキュリティ重点) | `references/phase-audit.md` |

## 終了条件マトリクス

| Phase | 結果 | 次のアクション |
|-------|------|--------------|
| AUDIT | SUCCESS_NO_ISSUES | 完了 |
| AUDIT | ISSUES_FOUND | → IMPROVE |
| IMPROVE | SUCCESS | → VERIFY |
| IMPROVE | PARTIAL_COMPLETE | 追加改善 |
| VERIFY | SUCCESS | 完了 |

## 出力ファイル

| Phase | 出力 |
|-------|------|
| 監査 | `docs/audit/quality-audit_{TIMESTAMP}.md` |
| 改善 | `docs/improve/quality-improve_{TIMESTAMP}.md` |

## リファレンス

- `references/phase-audit.md` - 監査フェーズ詳細
- `references/phase-improve.md` - 改善フェーズ詳細
- `references/phase-verify.md` - 検証フェーズ詳細
- `references/quality-standards.md` - 品質メトリクス基準
