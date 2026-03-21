---
paths: frontend/src/lib/api/**/*.ts, frontend/src/api/**/*.ts, frontend/src/hooks/**/*.ts
---

# APIクライアント規約

## 絶対ルール

**createPresetApiClient は関数内で毎回生成すること**

これはMonsteraプロジェクトで最も重要なルールです。過去に多くの問題がこのルール違反から発生しています。

---

## 正しいパターン

```typescript
import { createPresetApiClient } from '@/lib/api';
import { handleApiError } from '@/lib/api/error';

export const getExpenses = async (params: ExpenseListParams): Promise<ExpenseListResponse> => {
  try {
    // 関数内でクライアントを生成（毎回新規）
    const client = createPresetApiClient('auth');

    const response = await client.get('/expenses', { params });
    return response.data;
  } catch (error) {
    throw handleApiError(error, '経費一覧取得');
  }
};
```

---

## 禁止パターン

### 1. モジュールレベルでのクライアント定義

```typescript
// ❌ 禁止: キャッシュ問題・認証状態不整合の原因
const client = createPresetApiClient('auth');

export const getExpenses = async () => {
  return await client.get('/expenses');
};
```

### 2. baseURLの上書き

```typescript
// ❌ 禁止: プリセット設定を破壊する
const client = createPresetApiClient('auth', {
  baseURL: API_BASE_URL
});
```

### 3. /api/v1 のハードコーディング

```typescript
// ❌ 禁止: 自動付与されるため不要
await client.get('/api/v1/users');

// ✅ 正しい: パスのみ指定
await client.get('/users');
```

### 4. 不要な設定の追加

```typescript
// ❌ 禁止: プリセットで既に設定済み
const client = createPresetApiClient('auth', {
  timeout: 30000,
  withCredentials: true
});
```

---

## プリセット一覧と使い分け

| プリセット | 用途 | タイムアウト | 特徴 |
|-----------|------|-------------|------|
| `auth` | 一般認証API | 30秒 | 標準プリセット（最も使用） |
| `admin` | 管理者API | 30秒 | `X-Admin-Request: true` ヘッダー自動付与 |
| `upload` | ファイルアップロード | 120秒 | `Content-Type: multipart/form-data` 自動設定 |
| `batch` | バッチ処理・エクスポート | 300秒 | リトライ回数増加（最大5回） |
| `public` | 公開API（認証不要） | 30秒 | Cookie無効、認証ヘッダーなし |
| `realtime` | リアルタイム通信 | 5秒 | リトライなし |

### 使い分け例

```typescript
// 一般的なデータ取得
const client = createPresetApiClient('auth');
await client.get('/expenses');

// 管理者向けAPI
const adminClient = createPresetApiClient('admin');
await adminClient.get('/engineers');

// ファイルアップロード
const uploadClient = createPresetApiClient('upload');
await uploadClient.post('/expenses/receipts/upload', formData);

// CSV出力等の長時間処理
const batchClient = createPresetApiClient('batch');
await batchClient.post('/exports', { type: 'csv' });

// ログイン前のデータ取得
const publicClient = createPresetApiClient('public');
await publicClient.get('/public/holidays');
```

---

## API関数の標準構造

```typescript
import { createPresetApiClient } from '@/lib/api';
import { handleApiError } from '@/lib/api/error';
import { DebugLogger } from '@/utils/debugLogger';

export const fetchEntityList = async (params: ListParams): Promise<ListResponse> => {
  try {
    // 1. デバッグログ（開始）
    DebugLogger.info(
      { category: 'API', operation: 'FetchEntityList' },
      'Fetching entity list',
      { params }
    );

    // 2. クライアント生成（関数内で毎回）
    const client = createPresetApiClient('auth');

    // 3. API呼び出し
    const response = await client.get('/entities', { params });

    // 4. デバッグログ（成功）
    DebugLogger.info(
      { category: 'API', operation: 'FetchEntityList' },
      'Entity list fetched successfully',
      { count: response.data.items?.length }
    );

    // 5. レスポンス返却
    return response.data;
  } catch (error) {
    // 6. エラーログ
    DebugLogger.error(
      { category: 'API', operation: 'FetchEntityList' },
      'Failed to fetch entity list',
      error
    );

    // 7. 統一エラーハンドリング
    throw handleApiError(error, 'エンティティ一覧取得');
  }
};
```

---

## エラーハンドリング

### 基本パターン

```typescript
import { handleApiError } from '@/lib/api/error';

try {
  const client = createPresetApiClient('auth');
  const response = await client.get('/endpoint');
  return response.data;
} catch (error) {
  throw handleApiError(error, '操作名');
}
```

### AbortSignal対応

```typescript
export const getLeaveTypes = async (signal?: AbortSignal): Promise<LeaveType[]> => {
  const client = createPresetApiClient('auth');
  try {
    const response = await client.get('/leave/types', { signal });
    return response.data;
  } catch (error) {
    const handledError = handleApiError(error, '休暇種別取得');

    // キャンセルエラーは特別処理
    if (handledError instanceof AbortError) {
      throw handledError;
    }

    throw handledError;
  }
};
```

---

## Admin API オブジェクトパターン

管理者APIは以下のようにオブジェクト形式で定義することも可能：

```typescript
const adminGet = async <T>(path: string, params?: any): Promise<T> => {
  const client = createPresetApiClient('admin');
  const response = await client.get(path, { params });
  return response.data;
};

const adminPost = async <T>(path: string, data: any): Promise<T> => {
  const client = createPresetApiClient('admin');
  const response = await client.post(path, data);
  return response.data;
};

export const adminEngineerApi = {
  getEngineers: (params?: GetEngineersParams) =>
    adminGet<GetEngineersResponse>('/engineers', params),

  getEngineerDetail: (id: string) =>
    adminGet<EngineerDetail>(`/engineers/${id}`),

  createEngineer: (data: CreateEngineerRequest) =>
    adminPost<Engineer>('/engineers', data),
};
```

---

## チェックリスト

新しいAPI関数を作成する際：

- [ ] `createPresetApiClient` を関数内で呼び出している
- [ ] 適切なプリセットを選択している
- [ ] baseURL や不要な設定を追加していない
- [ ] `/api/v1` をハードコーディングしていない
- [ ] `handleApiError` でエラーハンドリングしている
- [ ] 戻り値の型を明示している
- [ ] DebugLogger でログ出力している（推奨）
