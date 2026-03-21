# Ivy（アイビー）— SES マッチングツール

## プロジェクト概要
SES営業向けの案件×エンジニアマッチングWebアプリ。
Claude APIを使ってマッチ度をスコアリングし、営業の判断を支援する。
自社エンジニアの案件探し・BP人材の案件マッチングの両方に対応。

## 設計書
**実装前に必ず読むこと:**
- 設計書: `/Users/daichirouesaka/Documents/duesk-company/products/matching-tool/DESIGN.md`
- AIプロンプト: `/Users/daichirouesaka/Documents/duesk-company/products/matching-tool/matching_prompt.md`

## 参考プロジェクト（流用元）
- Monstera: `/Users/daichirouesaka/dev/monstera`
  - 流用対象: config, middleware(cognito_auth, rate_limit, security, request_logger), common/logger, s3_service
  - 流用しない: freee連携, 週報, スキルシート生成, Slack通知

## 技術スタック
- Backend: Go 1.24 + Gin + GORM + PostgreSQL 16 + Redis 7
- Frontend: React + Vite + TypeScript + MUI v7 + TanStack React Query
- AI: Claude API (Anthropic)
- Auth: AWS Cognito (MonsteraとUser Pool共有)
- Infra: Docker + AWS ECS Fargate

## アーキテクチャ
```
Backend: Handler → Service → Repository（Monsteraと同じ3層構造）
Frontend: Pages → Components → Hooks → API Client
```

## ポート（Monsteraと競合しないよう分離）
| サービス | Ivy | Monstera |
|---------|-----|---------|
| Backend | 8081 | 8080 |
| Frontend | 5173 | 3000 |
| PostgreSQL | 5434 | 5432 |
| Redis | 6380 | 6379 |
| Cognito Local | 9230 | 9229 |

## 開発コマンド
```bash
make up          # Docker起動
make down        # Docker停止
make logs        # バックエンドログ
make db-shell    # DB接続
make backend-test # テスト実行
```

## 環境変数
`.env` に定義（gitignore済み）。設計書セクション9を参照。

## AI Service インターフェース
```go
type AIService interface {
    Match(ctx context.Context, req MatchRequest) (*MatchResponse, error)
}
```
- 本番: ClaudeAIService（Claude API呼び出し）
- テスト: MockAIService（固定JSON返却）
- 切替: `USE_MOCK_AI=true/false`

## 認証
- MonsteraのCognito User Poolを共有
- JITプロビジョニング: 初回ログイン時にIvy usersテーブルに自動作成
- engineerロール → 403拒否
- admin/sales → アクセス許可

## Phase 1 実装タスク
1. ✅ プロジェクト初期セットアップ
2. ✅ go.sum生成 + 依存解決
3. ✅ config（Monstera参考）
4. ✅ logger（Monstera流用）
5. ✅ middleware（cognito_auth, role_auth, rate_limit, security — Monstera流用）
6. ✅ model（User, Matching, Setting）
7. ✅ repository（matching, settings）
8. ✅ service（claude_service, mock_ai_service, matching_service, file_parse_service）
9. ✅ handler（auth, matching, file, settings）
10. ✅ routes（Monstera参考のルーティングパターン）
11. ✅ main.go（完成版）
12. ✅ フロントエンド（React SPA）
13. ✅ テスト
14. ✅ デプロイ設定

## コーディング規約
- Monsteraの `.claude/rules/` を参考にする
- Handler: リクエストバリデーション + レスポンス構築のみ。ビジネスロジックはServiceに書く
- Service: ビジネスロジック。DBアクセスはRepositoryに委譲
- Repository: GORMによるDB操作のみ
- エラーハンドリング: Ginの `c.JSON()` でステータスコード + エラーメッセージを返す
- ログ: zap（構造化ログ）
