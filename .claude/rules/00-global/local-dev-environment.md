# ローカル開発環境設定

## 認証（AWS Cognito）

ローカル開発では **dev環境のAWS Cognitoプール** を使用する。cognito-localは使用しない。

### dev Cognito プール情報

| 項目 | 値 |
|------|-----|
| プール名 | `monstera-dev-user-pool-v2` |
| プールID | `ap-northeast-1_T9q5pMSr4` |
| リージョン | `ap-northeast-1` |
| Client ID | `7g5b27lcugmmr131vd30obtcle` |
| Client Name | `monstera-backend-client` |
| ARN | `arn:aws:cognito-idp:ap-northeast-1:307031016432:userpool/ap-northeast-1_T9q5pMSr4` |

### .env 設定

```env
COGNITO_REGION=ap-northeast-1
COGNITO_USER_POOL_ID=ap-northeast-1_T9q5pMSr4
COGNITO_CLIENT_ID=7g5b27lcugmmr131vd30obtcle
COGNITO_CLIENT_SECRET=<.envファイルを参照>
COGNITO_ENDPOINT=
AWS_DEFAULT_REGION=ap-northeast-1
```

`COGNITO_ENDPOINT` を空にすることで、docker-compose.ymlのデフォルト（cognito-local）を上書きし、AWS Cognitoに直接接続する。

### AWS SSO ログイン

dev環境のAWSリソースにアクセスするには、事前にSSOログインが必要:

```bash
aws sso login --profile monstera-dev
```

---

## テストユーザー

ローカルDBのテストユーザーIDは **dev CognitoのSub ID** と一致させてある。

### ユーザーシード

```bash
# DB全クリア後に実行
bash docker/seed/create_dev_users.sh
```

このスクリプトは以下を実行:
1. cognito-localにユーザー作成（Cognito Local使用時のフォールバック）
2. PostgreSQLにユーザー + テストデータ挿入

### テストユーザー一覧

| Email | Role | JobCategory | Password |
|-------|------|-------------|----------|
| sysadmin-eng@test.local | SystemAdmin | engineer | TestPass123! |
| sysadmin-sales@test.local | SystemAdmin | sales | TestPass123! |
| sysadmin-back@test.local | SystemAdmin | back_office | TestPass123! |
| admin-eng@test.local | Admin | engineer | TestPass123! |
| admin-sales@test.local | Admin | sales | TestPass123! |
| admin-back@test.local | Admin | back_office | TestPass123! |
| user-eng@test.local | User | engineer | TestPass123! |
| user-sales@test.local | User | sales | TestPass123! |
| user-back@test.local | User | back_office | TestPass123! |
| daichiro.uesaka@duesk.co.jp | SystemAdmin | engineer | - |

---

## DB リセット手順

```bash
# 1. ボリューム削除 + 再起動
docker compose down -v && docker compose up -d

# 2. マイグレーション完了を待つ（自動実行）

# 3. テストユーザー + テストデータ投入
bash docker/seed/create_dev_users.sh
```
