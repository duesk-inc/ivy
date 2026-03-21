# guide-helper スキル

開発コマンドと開発手順のガイドを提供するスキルです。

## 使用条件

- 「コマンドを教えて」「どうやって実行する？」などの質問
- 開発環境のセットアップ方法を知りたいとき
- テスト、ビルド、マイグレーションの実行方法を確認したいとき

## 提供する情報

### 1. 開発コマンド

詳細は `references/development-commands.md` を参照:

- Docker Compose操作
- Backend (Go) コマンド
- Frontend (Next.js) コマンド

### 2. 開発タスク手順

詳細は `references/development-tasks.md` を参照:

- 新規APIエンドポイントの追加手順
- 新規フロントエンド機能の追加手順
- データベースマイグレーション手順

## クイックリファレンス

### よく使うコマンド

```bash
# 全サービス起動
docker compose up -d

# ログ確認
docker compose logs -f backend

# Frontendテスト
cd frontend && npm run test

# Backendテスト
cd backend && go test ./... -v

# 型チェック
cd frontend && npm run type-check

# マイグレーション実行
docker compose exec backend ./entrypoint.sh migrate
```

## 注意

このスキルはガイド提供のみを行います。実際の実装は `backend-developer` または `frontend-developer` エージェントに委譲してください。
