---
paths: "**/*"
description: 全体コーディング規約（全ファイル適用）
---

# コーディング規約

## 再利用優先の原則

**新規実装より既存コードの再利用を最優先すること**

これはMonsteraプロジェクトで最も重要な原則の一つです。

### 確認順序

新しいコードを書く前に、以下の順序で確認すること：

1. **共通コンポーネント/関数が既に存在しないか確認**
   - `frontend/src/components/common/` を確認
   - `frontend/src/hooks/common/` を確認
   - `backend/internal/utils/` を確認

2. **類似機能が別の場所に実装されていないか確認**
   - 同じ機能を持つコンポーネントがfeature別に存在していないか
   - 同様のロジックがserviceやhandlerに実装されていないか

3. **既存コードを拡張して対応できないか検討**
   - 既存コンポーネントにpropsを追加する
   - 既存関数にオプション引数を追加する

4. **上記すべてで対応不可の場合のみ新規作成**
   - 新規作成する場合は、他で再利用できる汎用的な設計を心がける

### 禁止事項

- 同じ機能を持つコンポーネント/関数の重複作成
- 既存の共通コンポーネントを無視した独自実装
- ユーティリティ関数の再発明
- 「似ているが微妙に違う」コンポーネントの乱立

### MUIコンポーネント直接使用禁止

以下のMUIコンポーネントは共通コンポーネントが用意されているため、直接使用禁止：

| MUIコンポーネント | 代わりに使用 | 規約ファイル |
|-----------------|------------|------------|
| `Select`, `MenuItem` | `FormSelect`, `SimpleSelect`, `InlineSelect` | `02-frontend/select-components.md` |
| `Button` | `ActionButton` | `02-frontend/action-button.md` |

```typescript
// ❌ 禁止
import { Select, MenuItem } from '@mui/material';

// ✅ 正しい
import { FormSelect, SimpleSelect } from '@/components/common/forms';
import { InlineSelect } from '@/components/common';
```

---

## 全般原則

- **保守性・可読性を最優先**
- **コメントは最小限**（コードで意図を表現）
- **絵文字は使用禁止**
- **既存コンポーネント・関数を優先利用**
- **過度な抽象化を避ける**

---

## Go（Backend）

### 命名規則

| 対象 | 規則 | 例 |
|------|------|-----|
| パッケージ名 | 小文字、単数形 | `handler`, `service`, `repository` |
| 公開関数/型 | PascalCase | `GetEngineerByID`, `EngineerService` |
| 非公開関数/変数 | camelCase | `validateInput`, `userID` |
| 定数 | PascalCase または UPPER_SNAKE_CASE | `MaxRetryCount`, `DEFAULT_TIMEOUT` |
| インターフェース | 動詞 + er（可能な場合） | `Reader`, `EngineerService` |

### ディレクトリ構成

```
backend/internal/
├── handler/      # HTTPハンドラー（リクエスト/レスポンス処理）
├── service/      # ビジネスロジック
├── repository/   # データアクセス（GORM）
├── model/        # データモデル（エンティティ）
├── dto/          # データ転送オブジェクト（API入出力）
├── middleware/   # ミドルウェア（認証、ログ等）
├── errors/       # エラー定義
├── message/      # メッセージ定数
├── constants/    # 定数定義
├── utils/        # ユーティリティ関数
└── config/       # 設定
```

### 層の責務

| 層 | 責務 | 依存先 |
|----|------|--------|
| Handler | HTTPリクエスト処理、バリデーション、レスポンス | Service |
| Service | ビジネスロジック、トランザクション管理 | Repository |
| Repository | データアクセス、クエリ実行 | Model |

---

## TypeScript（Frontend）

### 命名規則

| 対象 | 規則 | 例 |
|------|------|-----|
| コンポーネント | PascalCase | `EngineerList`, `ExpenseForm` |
| 関数・変数 | camelCase | `handleSubmit`, `userName` |
| 定数 | UPPER_SNAKE_CASE | `API_BASE_URL`, `MAX_FILE_SIZE` |
| 型・インターフェース | PascalCase（Iプレフィックス不要） | `Engineer`, `ExpenseListProps` |
| カスタムフック | use + PascalCase | `useExpenses`, `useAuth` |
| イベントハンドラ（Props） | on + Action | `onClick`, `onSubmit` |
| イベントハンドラ（実装） | handle + Action | `handleClick`, `handleSubmit` |

### ディレクトリ構成

```
frontend/src/
├── components/
│   ├── common/       # 共通コンポーネント（再利用必須）
│   ├── features/     # 機能別コンポーネント
│   ├── ui/           # UI専用コンポーネント
│   └── admin/        # 管理者向けコンポーネント
├── hooks/
│   ├── common/       # 共通フック（再利用必須）
│   └── [feature]/    # 機能別フック
├── lib/
│   └── api/          # APIクライアント
├── types/            # 型定義
├── utils/            # ユーティリティ関数
├── constants/        # 定数
└── context/          # Reactコンテキスト
```

---

## コード品質

### コメント規約

```typescript
// ✅ 必要な場合のみコメント
// TODO: 一時的な回避策。Issue #123 で対応予定
const workaround = ...;

// ❌ 不要なコメント
// ユーザーIDを取得する
const userId = getUserId();
```

### インポート順序

```typescript
// 1. 外部ライブラリ
import React from 'react';
import { Box, Button } from '@mui/material';

// 2. 内部モジュール（絶対パス）
import { createPresetApiClient } from '@/lib/api';
import { useToast } from '@/hooks/common/useToast';

// 3. 型定義
import type { Engineer } from '@/types/engineer';
```

---

## セキュリティ規約

- **APIエンドポイントは認証必須**（ホワイトリスト方式で除外）
- **入力検証は両層（Frontend/Backend）で実施**
- **RBACによる権限管理を徹底**
- **機密情報はコードに含めない**（環境変数を使用）
- **SQLインジェクション対策**（GORMのパラメータバインディング使用）

---

## テスト規約

- **TDD実践**（テストファースト推奨）
- **ユニットテストは必須**
- **テストファイル名**: `*_test.go`, `*.test.ts`, `*.test.tsx`
- **カバレッジ目標**: 80%以上

---

## ルールのメンテナンス

**新しい共通機能を作成したら、必ず対応するルールファイルを更新すること**

### 更新が必要なケース

| 作成したもの | 更新するルールファイル |
|-------------|----------------------|
| 共通コンポーネント | `02-frontend/components.md` の一覧に追加 |
| 共通フック | `02-frontend/components.md` の共通フック一覧に追加 |
| APIクライアント関数 | `02-frontend/api-client.md` にパターンを追加（必要に応じて） |
| バックエンドユーティリティ | `01-backend/` 配下の該当ファイルに追加 |
| 新しいエラーパターン | `01-backend/error-handling.md` または `02-frontend/error-handling.md` |
| 発見した落とし穴 | `05-pitfalls/known-issues.md` に追加 |

### 更新内容の例

**共通コンポーネント追加時**（`02-frontend/components.md`）:

```markdown
| `NewComponent` | `common/` | 用途の説明 |
```

**共通フック追加時**（`02-frontend/components.md`）:

```markdown
| `useNewHook` | `common/` | 用途の説明 |
```

### 更新を忘れないために

1. **PRレビュー時に確認**: 共通機能の追加があればルール更新も含まれているか
2. **コミットメッセージに明記**: `feat: add CommonButton component (update rules)`
3. **チェックリストに追加**: 下記の「新しいコードを書く前に」に含める

---

## チェックリスト

新しいコードを書く前に：

- [ ] 既存の共通コンポーネント/関数で対応できないか確認したか
- [ ] 類似機能が他の場所に実装されていないか確認したか
- [ ] 命名規則に従っているか
- [ ] 適切なディレクトリに配置しているか
- [ ] 不要なコメントを書いていないか
- [ ] セキュリティ要件を満たしているか

共通機能を新規作成した後：

- [ ] 対応するルールファイル（`.claude/rules/`）を更新したか
- [ ] 一覧表に追加したか（コンポーネント名、場所、用途）
