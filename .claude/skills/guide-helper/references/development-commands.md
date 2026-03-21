# 開発コマンドリファレンス

## Docker Compose（フルスタック）

```bash
# 全サービス起動（PostgreSQL, Redis, Backend, Frontend）
docker compose up -d

# 全サービス停止
docker compose down

# ログ確認
docker compose logs -f [service-name]

# 特定サービスの再ビルド＆再起動
docker compose up -d --build [service-name]

# サービス一覧確認
docker compose ps
```

## Backend (Go)

### 実行

```bash
# ローカルでサーバー起動
cd backend && go run cmd/server/main.go

# マイグレーションのみ実行
cd backend && ./entrypoint.sh migrate

# マイグレーション＋サーバー起動
cd backend && ./entrypoint.sh start

# マイグレーションスキップで直接起動
cd backend && ./entrypoint.sh direct
```

### テスト

```bash
# 全テスト実行
cd backend && go test ./... -v

# カバレッジ付きテスト
cd backend && go test ./... -v -cover

# 特定パッケージのテスト
cd backend && go test ./internal/service/... -v

# 特定テスト関数のみ実行
cd backend && go test -v -run TestFunctionName ./...
```

### その他

```bash
# コードフォーマット
cd backend && go fmt ./...

# 静的解析
cd backend && go vet ./...

# 依存関係整理
cd backend && go mod tidy
```

## Frontend (Next.js)

### 実行

```bash
# 開発サーバー起動
cd frontend && npm run dev

# プロダクションビルド
cd frontend && npm run build

# プロダクションサーバー起動
cd frontend && npm run start
```

### テスト・品質チェック

```bash
# 型チェック
cd frontend && npm run type-check

# Linting
cd frontend && npm run lint

# ユニットテスト
cd frontend && npm run test

# ウォッチモードでテスト
cd frontend && npm run test:watch

# カバレッジ付きテスト
cd frontend && npm run test:coverage
```

### E2E テスト（Playwright）

```bash
# 全E2Eテスト実行
cd frontend && npm run test:e2e

# Smokeテストのみ
cd frontend && npm run test:e2e:smoke

# UI モードでテスト
cd frontend && npm run test:e2e:ui

# デバッグモード
cd frontend && npm run test:e2e:headed

# Playwrightインストール
cd frontend && npm run test:e2e:install
```

## データベース

### マイグレーション

```bash
# Docker経由でマイグレーション実行
docker compose exec backend ./entrypoint.sh migrate

# マイグレーションファイル作成（手動）
touch backend/migrations/XXXXXX_description.up.sql
touch backend/migrations/XXXXXX_description.down.sql

# ローカルでマイグレーション実行（直接）
cd backend && migrate -path ./migrations -database "postgresql://user:pass@localhost:5432/dbname?sslmode=disable" up

# マイグレーションロールバック
cd backend && migrate -path ./migrations -database "postgresql://user:pass@localhost:5432/dbname?sslmode=disable" down 1
```

### PostgreSQL直接接続

```bash
# Docker経由でpsql接続
docker compose exec postgres psql -U monstera -d monstera

# テーブル一覧
\dt

# 特定テーブルの構造確認
\d table_name
```

## 環境変数

### 設定ファイル

| ファイル | 用途 |
|---------|------|
| `.env` | ルート（Docker Compose用） |
| `backend/.env` | Backend用 |
| `frontend/.env.local` | Frontend用 |

### 主要な環境変数

**Database:**
- `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD`

**Redis:**
- `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`

**Storage:**
- `USE_MOCK_S3` (ローカルはtrue)
- `AWS_S3_BUCKET_NAME`, `AWS_REGION`

**Cognito:**
- `COGNITO_USER_POOL_ID`, `COGNITO_CLIENT_ID`, `COGNITO_REGION`

**Frontend:**
- `NEXT_PUBLIC_API_URL`, `NEXT_SERVER_API_URL`
