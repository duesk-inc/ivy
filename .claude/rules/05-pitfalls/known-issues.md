---
paths: "**/*"
description: 過去に発生した問題と対策（全ファイル適用）
---

# 既知の落とし穴

このドキュメントは、Monsteraプロジェクトで過去に発生した問題とその対策をまとめたものです。
同じ問題を繰り返さないために、実装前に確認してください。

---

## 1. APIクライアント関連

### 1.1 モジュールレベルでのクライアント定義

**問題**: `createPresetApiClient` をモジュールレベルで定義すると、認証状態の不整合やキャッシュ問題が発生

```typescript
// ❌ これが原因で認証エラーが発生
const client = createPresetApiClient('auth');

export const getUsers = async () => {
  return await client.get('/users');
};
```

**対策**: 関数内で毎回クライアントを生成

```typescript
// ✅ 正しい
export const getUsers = async () => {
  const client = createPresetApiClient('auth');
  return await client.get('/users');
};
```

**関連メモリ**: `api_client_best_practices`, `api_client_singleton_issue`

---

### 1.2 baseURLの上書き

**問題**: プリセットに追加設定を渡すと、プリセットの自動設定が破壊される

```typescript
// ❌ プリセット設定が無効になる
const client = createPresetApiClient('auth', {
  baseURL: API_BASE_URL
});
```

**対策**: プリセット名のみ指定し、追加設定は行わない

```typescript
// ✅ 正しい
const client = createPresetApiClient('auth');
```

**関連メモリ**: `api_client_base_url_override_issue`, `api_preset_misuse_pattern_upload`

---

### 1.3 /api/v1 のハードコーディング

**問題**: パスに `/api/v1` を含めると二重になる

```typescript
// ❌ /api/v1/api/v1/users になる
await client.get('/api/v1/users');
```

**対策**: パスのみ指定（/api/v1 は自動付与）

```typescript
// ✅ 正しい
await client.get('/users');
```

**関連メモリ**: `api_client_path_duplication_pattern`

---

### 1.4 API関数追加時の既存パターン未確認

**問題**: 新規API関数で `convertSnakeToCamel` を適用し忘れ、snake_caseのまま返却

**発生状況**:
- `notification.ts` に `getNotificationDetail` を追加
- 同ファイルの既存関数は全て `convertSnakeToCamel` を使用していた
- 新規関数だけ変換なしで実装 → 呼び出し側でsnake_caseアクセスが必要に

```typescript
// ❌ 同一ファイルの既存パターンを確認せず実装
export const getNotificationDetail = async (id: string) => {
  const response = await client.get(`/notifications/${id}`);
  return response.data;  // snake_case のまま
};

// ✅ 既存関数を見れば気づける
export const getUserNotifications = async (...) => {
  ...
  return convertSnakeToCamel<UserNotificationList>(response.data);
};
```

**対策**: 同一ファイルに関数を追加する際、既存関数のパターン（変換、エラーハンドリング、型）を確認

**関連メモリ**: なし

---

## 2. バックエンド関連

### 2.1 UserRepository Preloadエラー

**問題**: `Preload("UserRoles")` でuser_rolesテーブルが存在しない場合、ユーザー検索自体が失敗

```
ERROR: relation "user_roles" does not exist (SQLSTATE 42P01)
```

**影響**: 認証済みユーザーでも「ユーザーが見つかりません」エラーが発生

**対策**:
1. マイグレーションを確実に実行してuser_rolesテーブルを作成
2. `docker compose exec backend ./entrypoint.sh migrate` を実行

**関連メモリ**: `common_pitfalls_user_roles_preload`

---

### 2.2 トランザクション内で既存リポジトリを使用

**問題**: トランザクション外のリポジトリインスタンスを使うと別接続になり、トランザクションが効かない

```go
// ❌ s.engineerRepo は別接続
err := s.db.Transaction(func(tx *gorm.DB) error {
    return s.engineerRepo.Create(ctx, user)
})
```

**対策**: トランザクション内では新規リポジトリを生成

```go
// ✅ 正しい
err := s.db.Transaction(func(tx *gorm.DB) error {
    txRepo := repository.NewEngineerRepository(tx, s.logger)
    return txRepo.Create(ctx, user)
})
```

---

### 2.3 エラーステータスコードの不適切な使用

**問題**: すべてのエラーに500を返すと、クライアント側で適切なハンドリングができない

**対策**: エラー内容に応じたステータスコードを返す

| ステータス | 用途 |
|-----------|------|
| 400 | リクエスト不正、バリデーションエラー |
| 401 | 認証エラー |
| 403 | 権限エラー |
| 404 | リソース未発見 |
| 409 | 状態競合（既に提出済み等） |
| 500 | サーバー内部エラー |

**関連メモリ**: `expense_limit_error_http_status_fix`

---

### 2.4 Delete-All → Re-Create パターンによるデータ消失（本番障害）

**問題**: 関連データ（work_histories等）を更新する際に「全削除→再作成」パターンを使用すると、リクエストに該当データが含まれない場合に既存データが全て消失する

**発生状況（2026-03）**:
- プロフィール編集ページ（`POST /api/v1/me/profile`）で資格情報のみ更新
- `ProfileSaveRequest.WorkHistory` はフロントから送信されないため常に空スライス
- `UpdateUserProfileWithDTO` が無条件にwork_historiesを全削除→0件で再作成
- 結果: 2名のユーザーの案件経歴データが消失

```go
// ❌ 危険: リクエストデータの有無を確認せずに全削除
tx.Where("profile_id = ?", profile.ID).Delete(&model.WorkHistory{})
for _, workReq := range request.WorkHistory { // 空スライスなので何も作成されない
    // ...
}

// ✅ 正しい: データが提供されている場合のみ削除・再作成
if len(request.WorkHistory) > 0 {
    tx.Where("profile_id = ?", profile.ID).Delete(&model.WorkHistory{})
    for _, workReq := range request.WorkHistory {
        // ...
    }
}
```

**対策**:
1. Delete-All → Re-Create パターンを使用する場合、リクエストにデータが含まれるかを必ずチェック
2. 空スライスで既存データを削除しない（`len(slice) > 0` ガード）
3. 異なるフォームが同じテーブルを操作する場合、各フォームの責務範囲を明確にする

**関連ファイル**: `backend/internal/service/profile_service.go`, `backend/internal/service/skill_sheet_service.go`

---

## 3. フロントエンド関連

### 3.1 Hydration Mismatch (MUI + Next.js)

**問題**: MUIコンポーネントでサーバー/クライアントのレンダリング結果が不一致

```
Warning: Prop `className` did not match.
```

**対策**: MUIやhooksを使用するコンポーネントに `'use client'` ディレクティブを追加

```typescript
'use client';

import { Button } from '@mui/material';

export const MyComponent = () => {
  return <Button>Click</Button>;
};
```

**関連メモリ**: `hydration_mismatch_mui_nextjs`

---

### 3.2 React Query キャッシュキー不一致

**問題**: 同じデータに対して異なるキャッシュキーを使用すると、重複フェッチや不整合が発生

```typescript
// ❌ キーが統一されていない
useQuery({ queryKey: ['engineers'], ... });
useQuery({ queryKey: ['admin-engineers'], ... });
```

**対策**: `queryKeys` オブジェクトから一元管理

```typescript
// ✅ 正しい
import { queryKeys } from '@/lib/tanstack-query';

useQuery({ queryKey: queryKeys.adminEngineers(params), ... });
```

---

### 3.3 React Hooks の依存配列不足

**問題**: useEffect/useCallback/useMemo の依存配列が不完全だと、古い値を参照したり無限ループが発生

```typescript
// ❌ filters が依存配列にない
useEffect(() => {
  fetchData(filters);
}, []);  // filters変更時に再実行されない
```

**対策**: ESLint の react-hooks/exhaustive-deps ルールに従う

```typescript
// ✅ 正しい
useEffect(() => {
  fetchData(filters);
}, [filters]);
```

**関連メモリ**: `react-hooks-common-errors`

---

### 3.4 Server Component で ISR を使用したユーザー固有データのキャッシュ（情報漏えいリスク）

**問題**: Next.js の ISR（Incremental Static Regeneration）や `revalidate` オプションを使用して、ユーザー固有のデータ（通知、ダッシュボード等）をServer Componentでフェッチ・キャッシュすると、**他のユーザーのデータが別のユーザーに表示される情報漏えいリスク**が発生

**発生状況**:
- `lib/api/server/dashboard.ts` で `revalidate: 60` を設定
- ダッシュボードデータはユーザーIDに基づいて取得される
- ISRはユーザー単位ではなくルート単位でキャッシュするため、ユーザーAのデータがユーザーBに表示される可能性

```typescript
// ❌ 禁止: ユーザー固有データをISRでキャッシュ
export async function getAdminDashboardData(options: { revalidate?: number }) {
  const cookieStore = await cookies();
  const token = cookieStore.get('authToken');

  // このデータはユーザー固有なのに、revalidateでキャッシュされる
  const response = await fetch(`${API_URL}/admin/dashboard`, {
    headers: { Authorization: `Bearer ${token?.value}` },
    next: { revalidate: options.revalidate }, // ❌ 危険
  });
  return response.json();
}

// page.tsx (Server Component)
export default async function DashboardPage() {
  const initialData = await getAdminDashboardData({ revalidate: 60 }); // ❌
  return <DashboardClient initialData={initialData} />;
}
```

**対策**: ユーザー固有のデータは必ずクライアントサイドでフェッチ

```typescript
// ✅ 正しい: Client ComponentでReact Queryを使用
// page.tsx
export default function DashboardPage() {
  return (
    <Suspense fallback={<LoadingSkeleton />}>
      <DashboardClient />  {/* initialData は渡さない */}
    </Suspense>
  );
}

// DashboardClient.tsx
'use client';
export function DashboardClient() {
  // クライアントサイドでユーザー固有データをフェッチ
  const { data, loading } = useDashboard();
  // ...
}
```

**ISR を使って良いケース（ユーザー固有でないデータのみ）**:
- 祝日マスタ
- 設定マスタ
- 公開ニュース・お知らせ
- 静的コンテンツ

**ISR を使ってはいけないケース（ユーザー固有データ）**:
- ダッシュボード統計
- 通知一覧
- 自分の経費・週報
- プロフィール情報

**関連メモリ**: なし

---

### 3.5 ログアウト時にReact Queryキャッシュが残留（データ漏えいリスク）

**問題**: ログアウト時にReact Queryのキャッシュをクリアしないと、次にログインしたユーザーに前のユーザーのデータが表示される

**発生状況**:
- ユーザーAでログインし、提出書類データがキャッシュされる（キー: `["mySubmissions", 2026, 2]`）
- ログアウト → Auth状態はクリアされるが、React Queryキャッシュはそのまま
- ユーザーBでログイン → 同じクエリキーでキャッシュヒットし、ユーザーAのデータが表示される
- `gcTime`が2時間のため、キャッシュが長時間残留する

```typescript
// ❌ Auth状態だけクリアしてキャッシュを放置
const logout = useCallback(async () => {
  await apiLogout();
  setUser(null);
  setIsAuthenticated(false);
  router.push("/login");
}, [router]);

// ✅ キャッシュも必ずクリア
import { cacheUtils } from "@/lib/query-client";

const logout = useCallback(async () => {
  await apiLogout();
  setUser(null);
  setIsAuthenticated(false);
  cacheUtils.clearAll();
  router.push("/login");
}, [router]);
```

**対策**: ログアウト処理で必ず `cacheUtils.clearAll()` を呼び出す

**関連メモリ**: なし

---

## 4. データベース関連

### 4.1 マイグレーションの整合性

**問題**: up.sql と down.sql の不整合、または実行順序の問題

**対策**:
1. 連番を正しく付ける（000XXX形式）
2. down.sql でロールバック可能か確認
3. 既存データへの影響を考慮

```bash
# マイグレーション状態確認
docker compose exec backend migrate -path ./migrations -database "..." version
```

**関連メモリ**: `migration_integrity_fix_20241214`

---

### 4.2 外部キー制約違反

**問題**: 関連テーブルのデータを先に削除せずに親テーブルを削除しようとするとエラー

**対策**:
1. ON DELETE CASCADE を適切に設定
2. または削除順序を制御（子 → 親の順）

---

## 5. 認証関連

### 5.1 Cognito と DB の同期ずれ

**問題**: Cognitoにはユーザーが存在するがDBにはない、またはその逆

**対策**:
1. ユーザー作成時はトランザクションで両方を同時に
2. Cognito作成後にDB作成が失敗したらCognitoからも削除

```go
// Cognito作成成功後、DB作成失敗時のロールバック
if err != nil {
    if cognitoSub != "" {
        s.cognitoAuth.client.AdminDeleteUser(ctx, &cognitoidentityprovider.AdminDeleteUserInput{
            UserPoolId: aws.String(s.config.Cognito.UserPoolID),
            Username:   aws.String(input.Email),
        })
    }
    return nil, err
}
```

**関連メモリ**: `cognito_db_sync_issue`

---

## 6. 外部API連携関連

### 6.1 外部APIが会社全体のデータを返す場合のユーザー分離漏れ（情報漏えいリスク）

**問題**: freee APIの `paid_holidays`/`special_holidays` エンドポイントは `employee_id` パラメータを**無視**し、会社全体の休暇データを返す。このレスポンスをそのまま呼び出し元ユーザーに紐づけて保存すると、**他ユーザーの休暇データが全員に表示される情報漏えい**が発生する。

**発生状況**:
- `freee_leave_snapshot_service.go` で休暇同期APIを `employee_id` パラメータ付きで呼び出し
- freee APIは `employee_id` を無視し、全社員5名分のデータを返却
- レスポンス全件を呼び出し元ユーザーのIDで保存 → 他人の休暇が全員に表示

```go
// ❌ 危険: APIが全件返すのにフィルタリングなしで保存
result, err := s.apiClient.GetLeaveRequests(ctx, from, to)
for _, req := range result.PaidHolidays {
    snapshot := req.ToFreeeLeaveSnapshot(userID, syncedAt) // 全件がuserIDに紐づく
    snapshots = append(snapshots, *snapshot)
}

// ✅ 正しい: applicant_id（= freee_user_id）でフィルタリング
result, err := s.apiClient.GetLeaveRequests(ctx, from, to)
for _, req := range result.PaidHolidays {
    if freeeUserID > 0 && req.ApplicantID != freeeUserID {
        continue // 他ユーザーのデータをスキップ
    }
    snapshot := req.ToFreeeLeaveSnapshot(userID, syncedAt)
    snapshots = append(snapshots, *snapshot)
}
```

**教訓**:
1. **外部APIのフィルタリングパラメータを信用しない** - ドキュメントに記載があっても、実際に動作するか必ずテストする
2. **レスポンスに「誰のデータか」を示すフィールドがあるか確認** - freeeの場合は `applicant_id`
3. **IDシステムの違いに注意** - freeeには `employee_id`（従業員ID）と `user_id`（ユーザーID/applicant_id）の2系統がある。これらは完全に異なる値
4. **APIが全件返す場合はクライアント側フィルタリングを必ず実装** - フィルタリング用IDは事前に取得・保存しておく

**freee API固有の注意点**:
| エンドポイント | `employee_id`パラメータ | 実際の挙動 |
|---------------|----------------------|-----------|
| `GET /hr/api/v1/employees/{id}` | パスパラメータ | 正常動作 |
| `GET /hr/api/v1/paid_holidays` | クエリパラメータ | **無視される** |
| `GET /hr/api/v1/special_holidays` | クエリパラメータ | **無視される** |
| `GET /hr/api/v1/approval_requests/paid_holidays` | `applicant_id` | **400エラー** |

**関連ファイル**: `backend/internal/service/freee_leave_snapshot_service.go`, `backend/internal/service/freee_api_client.go`

---

### 6.2 外部APIのID体系を正しくマッピングする

**問題**: freee APIには `employee_id`（従業員ID）と `user_id`（ユーザーID）の2つの独立したID体系がある。休暇申請レスポンスの `applicant_id` は `user_id` に対応するが、システム内で保持していたのは `employee_id` のみだったため、フィルタリングができなかった。

**対策**:
1. 外部APIのIDマッピングを理解し、必要なIDをすべてDBに保存する
2. freeeの場合: `GET /hr/api/v1/employees/{id}` のレスポンスに含まれる `user_id` を `freee_employees.freee_user_id` として保存
3. 従業員同期時にIDを取得し、休暇同期時にフィルタリングに使用

```
freee_employee_id (2553841) ≠ freee_user_id (10302858)
                                    ↕ 対応
                            applicant_id (10302858) ← 休暇申請レスポンス
```

**関連メモリ**: `freee_oauth_implementation_guide`

---

## 7. テスト関連

### 7.1 エラーの詳細を確認せずに終了

**問題**: 401が返ってきたら「認証が必要で正常」と判断して終了してしまう

```bash
# ❌ 不十分
curl http://localhost:8080/api/v1/admin/engineers
# 結果: 401
# 判断: 「認証が必要で正常」→ 終了
```

**対策**: レスポンスボディとログを必ず確認

```bash
# ✅ 正しい
curl -v http://localhost:8080/api/v1/admin/engineers | jq .
# レスポンスの details を確認
# "unexpected signing method: RS256" など重要な情報が含まれる

docker compose logs backend | grep ERROR
```

**関連メモリ**: `testing-best-practices`

---

## チェックリスト

実装前に確認：

### APIクライアント（フロントエンド）
- [ ] `createPresetApiClient` を関数内で呼び出しているか
- [ ] baseURL を上書きしていないか
- [ ] `/api/v1` をハードコーディングしていないか

### Server Component / ISR（フロントエンド）
- [ ] Server Componentでフェッチするデータは**ユーザー固有でない**か確認したか
- [ ] `revalidate` や ISR を使用する場合、キャッシュされても問題ないデータか確認したか
- [ ] ユーザー固有データ（ダッシュボード、通知、経費等）はClient Component + React Queryでフェッチしているか
- [ ] `lib/api/server/` ディレクトリにファイルを追加していないか（このパターンは廃止済み）

### 外部API連携（バックエンド）
- [ ] 外部APIがユーザー単位でフィルタリングされたデータを返すか**実際にテスト**したか
- [ ] レスポンスに「誰のデータか」を示すフィールド（`applicant_id`等）を確認したか
- [ ] 全件返却の場合、クライアント側フィルタリングを実装しているか
- [ ] フィルタリング用のIDマッピング（`employee_id` ↔ `user_id`等）をDBに保存しているか

### トランザクション（バックエンド）
- [ ] トランザクション内で新規リポジトリを生成しているか
- [ ] エラー時の外部API（Cognito等）ロールバックを考慮しているか

### テスト
- [ ] エラーレスポンスの詳細まで確認したか
- [ ] ログ出力を確認したか
- [ ] 認証を通した状態でもテストしたか

---

## 新しい落とし穴を発見したら

このドキュメントはプロジェクトの知見を蓄積する場所です。

**追加すべきケース:**
- バグの根本原因が設計パターンの誤りだった場合
- 同じ問題が複数回発生した場合
- デバッグに時間がかかった問題

**追加時のフォーマット:**

```markdown
### X.X タイトル

**問題**: 何が起きたか

**対策**: どうすれば防げるか

**関連メモリ**: `memory_name`（Serenaメモリがあれば）
```

> **重要**: 将来の開発者が同じ問題で苦しまないよう、発見した落とし穴は積極的に追加してください。
