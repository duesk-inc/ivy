---
paths: frontend/src/components/**/*.tsx
description: Ivy共通コンポーネント一覧（参照用）。実装後に更新すること。
---

# Ivy コンポーネント一覧

**新しいコンポーネントを作成する前に、既存の共通コンポーネントを確認すること**
**Monsteraの共通コンポーネントを参考にしてよいが、Ivy用に簡略化すること**

## Pages
| ページ | ファイル | 用途 |
|--------|---------|------|
| LoginPage | `pages/LoginPage.tsx` | ログイン |
| MatchingPage | `pages/MatchingPage.tsx` | メイン画面（1:1マッチング） |
| HistoryPage | `pages/HistoryPage.tsx` | マッチング履歴 |
| SettingsPage | `pages/SettingsPage.tsx` | 設定（マージン・AIモデル） |

## Components
| コンポーネント | ディレクトリ | 用途 |
|--------------|------------|------|
| JobInputPanel | `matching/` | 案件情報テキスト入力 |
| EngineerInputPanel | `matching/` | エンジニア情報テキスト入力 + ファイルアップロード |
| SupplementForm | `matching/` | 補足情報フォーム（所属・単価・国籍等） |
| MatchingResult | `matching/` | マッチング結果表示（スコア・判定・アドバイス等） |
| ScoreCard | `matching/` | 各基準のスコア表示カード |
| Header | `common/` | アプリヘッダー |
| FileUpload | `common/` | ファイルアップロードコンポーネント |
| Loading | `common/` | ローディング表示 |
| AppLayout | `layout/` | 全体レイアウト |

## Hooks
| フック | 用途 |
|--------|------|
| useMatching | マッチング実行・結果取得 |
| useHistory | 履歴一覧取得 |
| useSettings | 設定取得・更新 |
| useAuth | 認証状態管理 |

## 関連規約
- コンポーネント作成パターン → [components-patterns.md](./components-patterns.md)
- ActionButton 使用規約 → [action-button.md](./action-button.md)
- セレクトボックス規約 → [select-components.md](./select-components.md)
- テキストフィールド規約 → [text-field-components.md](./text-field-components.md)

> **重要**: 共通コンポーネント/フックを新規作成した場合は、このファイルの一覧に追加してください。
