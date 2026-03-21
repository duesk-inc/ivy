# 開発タスク手順

## 新規 API エンドポイントの追加

### 手順

1. **DTO定義** (`backend/internal/dto/`)
   - リクエスト/レスポンス構造体を定義
   - バリデーションタグを追加

2. **Repository** (`backend/internal/repository/`)
   - 必要なデータアクセスメソッドを追加
   - インターフェースを定義

3. **Service** (`backend/internal/service/`)
   - ビジネスロジックを実装
   - トランザクション管理を行う

4. **Handler** (`backend/internal/handler/`)
   - HTTPリクエストをハンドリング
   - DTOへのバインドとバリデーション
   - レスポンスフォーマット

5. **Routes** (`backend/internal/routes/`)
   - エンドポイントを登録

6. **テスト**
   - 各層のユニットテストを追加

### 詳細規約

各層の詳細な規約は以下のルールファイルを参照:

- `.claude/rules/01-backend/handler.md`
- `.claude/rules/01-backend/service.md`
- `.claude/rules/01-backend/repository.md`
- `.claude/rules/01-backend/dto.md`

---

## 新規フロントエンド機能の追加

### 手順

1. **型定義** (`frontend/src/types/`)
   - APIレスポンスの型を定義

2. **API関数** (`frontend/src/lib/api/`)
   - `createPresetApiClient('auth')` を使用
   - `handleApiError` でエラーハンドリング

3. **カスタムフック** (`frontend/src/hooks/`)
   - React Query を使用
   - `queryKeys` から一元管理されたキーを使用

4. **コンポーネント** (`frontend/src/components/`)
   - 共通コンポーネントを優先利用
   - feature別のディレクトリに配置

5. **ページ** (`frontend/src/app/`)
   - App Router規約に従う

### 詳細規約

各要素の詳細な規約は以下のルールファイルを参照:

- `.claude/rules/02-frontend/api-client.md`
- `.claude/rules/02-frontend/react-query.md`
- `.claude/rules/02-frontend/components-index.md`
- `.claude/rules/02-frontend/types.md`

---

## データベースマイグレーション

### 手順

1. **ファイル作成**
   ```bash
   touch backend/migrations/XXXXXX_description.up.sql
   touch backend/migrations/XXXXXX_description.down.sql
   ```
   - 番号は既存の最大値 + 1

2. **UP マイグレーション作成**
   - テーブル作成、カラム追加など

3. **DOWN マイグレーション作成**
   - UP の逆操作（ロールバック用）
   - 必ずロールバック可能な状態にする

4. **実行**
   ```bash
   docker compose exec backend ./entrypoint.sh migrate
   ```

5. **ロールバックテスト**
   ```bash
   cd backend && migrate -path ./migrations -database "..." down 1
   cd backend && migrate -path ./migrations -database "..." up
   ```

### 詳細規約

- `.claude/rules/03-database/migrations.md`

---

## 開発ワークフロー

### 新機能開発

```
1. feature ブランチ作成
   └─ git checkout -b feature/xxx

2. Backend 実装
   ├─ DTO定義
   ├─ Repository実装
   ├─ Service実装
   ├─ Handler実装
   └─ テスト追加

3. Frontend 実装
   ├─ 型定義
   ├─ API関数
   ├─ カスタムフック
   ├─ コンポーネント
   └─ テスト追加

4. 動作確認
   ├─ docker compose up -d
   ├─ go test ./... -v
   ├─ npm run type-check
   └─ npm run test

5. コミット＆PR
```

### バグ修正

```
1. 原因調査
   └─ bug-investigator エージェント使用

2. 修正実装
   └─ backend-developer or frontend-developer

3. テスト追加
   └─ 再発防止のリグレッションテスト

4. 動作確認

5. コミット＆PR
```
