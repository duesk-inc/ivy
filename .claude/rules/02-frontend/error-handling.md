---
paths: frontend/src/**/*.ts, frontend/src/**/*.tsx
---

# フロントエンド エラーハンドリング規約

## handleApiError の使用

**すべてのAPI呼び出しで `handleApiError` を使用すること**

### 基本パターン

```typescript
import { handleApiError } from '@/lib/api/error';

export const getExpenses = async (params: ExpenseListParams): Promise<ExpenseListResponse> => {
  try {
    const client = createPresetApiClient('auth');
    const response = await client.get('/expenses', { params });
    return response.data;
  } catch (error) {
    throw handleApiError(error, '経費一覧取得');
  }
};
```

### オプション付き

```typescript
import { handleApiError, ErrorHandlingOptions } from '@/lib/api/error';

const options: ErrorHandlingOptions = {
  showNotification: true,   // 通知を表示（デフォルト: true）
  logError: true,           // ログに記録（デフォルト: true）
  throwError: true,         // エラーを再スロー（デフォルト: true）
  silent: false,            // サイレントモード
};

throw handleApiError(error, '操作名', options);
```

---

## StandardErrorResponse 型

すべてのAPIエラーは以下の統一形式に変換されます：

```typescript
interface StandardErrorResponse {
  error: {
    code: ApiErrorCode | string;
    message: string;
    details?: ErrorDetails;
  };
  status: number;
  timestamp: string;
}

// エラーコード
enum ApiErrorCode {
  UNAUTHORIZED = 'UNAUTHORIZED',        // 401
  FORBIDDEN = 'FORBIDDEN',              // 403
  NOT_FOUND = 'NOT_FOUND',              // 404
  VALIDATION_ERROR = 'VALIDATION_ERROR', // 400
  TIMEOUT = 'TIMEOUT',                  // タイムアウト
  NETWORK_ERROR = 'NETWORK_ERROR',      // ネットワークエラー
  CANCELLED = 'CANCELLED',              // キャンセル
  INTERNAL_ERROR = 'INTERNAL_ERROR',    // 500
}
```

---

## エラー表示パターン

### Toast通知

```typescript
import { useToast } from '@/hooks/common/useToast';

const { showSuccess, showError, showWarning, showInfo } = useToast();

// 成功
showSuccess('保存しました');

// エラー
showError('保存に失敗しました');

// 警告
showWarning('入力内容を確認してください');

// 情報
showInfo('処理を開始しました');
```

### React Query との連携

```typescript
useMutation({
  mutationFn: createExpense,
  onSuccess: () => {
    showSuccess('経費を登録しました');
  },
  onError: (error: any) => {
    // handleApiError で標準化されたエラー
    const message = error?.error?.message || error?.message || '登録に失敗しました';
    showError(message);
  },
});
```

### Alert コンポーネント

```typescript
import { Alert } from '@mui/material';

// インラインエラー表示
{error && (
  <Alert severity="error" sx={{ mb: 2 }}>
    {error.message || 'エラーが発生しました'}
  </Alert>
)}
```

---

## サイレントエラーハンドリング

ユーザーに通知せずにエラーを処理する場合：

```typescript
import { handleApiErrorSilently } from '@/lib/api/error';

try {
  const response = await client.get('/optional-data');
  return response.data;
} catch (error) {
  // ログのみ、通知なし
  handleApiErrorSilently(error);
  return null;  // またはデフォルト値
}
```

---

## リトライ処理

```typescript
import { handleRetryableApiError } from '@/lib/api/error';

try {
  const response = await client.get('/data');
  return response.data;
} catch (error) {
  handleRetryableApiError(error, async () => {
    // リトライロジック
    console.log('Retrying...');
    return await client.get('/data');
  });
}
```

---

## AbortError（キャンセル）処理

```typescript
export const getLeaveTypes = async (signal?: AbortSignal): Promise<LeaveType[]> => {
  const client = createPresetApiClient('auth');

  try {
    const response = await client.get('/leave/types', { signal });
    return response.data;
  } catch (error) {
    const handledError = handleApiError(error, '休暇種別取得');

    // キャンセルエラーは特別処理（通知不要）
    if (handledError instanceof AbortError) {
      throw handledError;  // 静かに伝播
    }

    throw handledError;
  }
};
```

### コンポーネントでの使用

```typescript
useEffect(() => {
  const controller = new AbortController();

  const fetchData = async () => {
    try {
      const data = await getLeaveTypes(controller.signal);
      setLeaveTypes(data);
    } catch (error) {
      if (error instanceof AbortError) {
        // アンマウント時のキャンセル、何もしない
        return;
      }
      setError(error);
    }
  };

  fetchData();

  return () => {
    controller.abort();  // クリーンアップ時にキャンセル
  };
}, []);
```

---

## フォームエラーハンドリング

### React Hook Form との連携

```typescript
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';

const {
  control,
  handleSubmit,
  setError,
  formState: { errors },
} = useForm({
  resolver: zodResolver(schema),
});

const onSubmit = async (data: FormData) => {
  try {
    await createExpense(data);
  } catch (error: any) {
    // バリデーションエラーの場合、フィールドにエラーを設定
    if (error?.error?.code === 'VALIDATION_ERROR' && error?.error?.details) {
      const details = error.error.details;
      Object.keys(details).forEach((field) => {
        setError(field as keyof FormData, {
          type: 'server',
          message: details[field],
        });
      });
    } else {
      // 一般的なエラー
      showError(error?.message || 'エラーが発生しました');
    }
  }
};
```

---

## エラーバウンダリ

### QueryErrorBoundary

```typescript
import { QueryErrorBoundary } from '@/components/common/QueryErrorBoundary';

<QueryErrorBoundary
  fallback={<ErrorFallback />}
  onError={(error) => {
    DebugLogger.error({ category: 'UI', operation: 'Render' }, 'Error', error);
  }}
>
  <MyComponent />
</QueryErrorBoundary>
```

### ErrorBoundary

```typescript
import { ErrorBoundary } from '@/components/common/ErrorBoundary';

<ErrorBoundary
  fallback={<FullScreenErrorDisplay message="エラーが発生しました" />}
>
  <App />
</ErrorBoundary>
```

---

## DebugLogger

開発時のデバッグログ出力：

```typescript
import { DebugLogger } from '@/utils/debugLogger';

// API呼び出し前
DebugLogger.info(
  { category: 'API', operation: 'GetExpenses' },
  'Fetching expenses',
  { params }
);

// API成功時
DebugLogger.info(
  { category: 'API', operation: 'GetExpenses' },
  'Expenses fetched successfully',
  { count: data.length }
);

// APIエラー時
DebugLogger.error(
  { category: 'API', operation: 'GetExpenses' },
  'Failed to fetch expenses',
  error
);

// 専用メソッド
DebugLogger.apiRequest({ category: 'EXPENSE', operation: 'Create' }, { data });
DebugLogger.apiResponse({ category: 'EXPENSE', operation: 'Create' }, { response });
DebugLogger.apiError({ category: 'EXPENSE', operation: 'Create' }, { error });
```

---

## HTTPステータス別の処理

```typescript
try {
  const response = await client.get('/data');
  return response.data;
} catch (error: any) {
  const status = error?.response?.status || error?.status;

  switch (status) {
    case 401:
      // 認証エラー → ログイン画面へ
      router.push('/login');
      break;

    case 403:
      // 権限エラー
      showError('この操作を行う権限がありません');
      break;

    case 404:
      // 未発見
      showError('データが見つかりませんでした');
      break;

    case 409:
      // 競合
      showError('データが更新されています。再読み込みしてください');
      break;

    case 422:
      // バリデーションエラー
      // フィールドエラーとして処理
      break;

    default:
      // その他
      showError('エラーが発生しました');
  }

  throw handleApiError(error, '操作名');
}
```

---

## チェックリスト

### API呼び出し

- [ ] `handleApiError` でエラーをラッピングしているか
- [ ] 適切な操作名を指定しているか
- [ ] サイレント処理が必要な場合は `handleApiErrorSilently` を使用しているか

### 通知

- [ ] `useToast` フックを使用しているか
- [ ] 成功/エラー/警告を適切に使い分けているか
- [ ] ユーザーにとって意味のあるメッセージを表示しているか

### フォーム

- [ ] サーバーバリデーションエラーをフィールドに反映しているか
- [ ] 送信中は再送信を防いでいるか（isPending）

### AbortSignal

- [ ] 長時間の処理にはキャンセル機能を実装しているか
- [ ] useEffectのクリーンアップでabortしているか
- [ ] AbortErrorは静かに処理しているか

### デバッグ

- [ ] DebugLoggerで適切にログ出力しているか
- [ ] 開発時のトラブルシューティングに役立つ情報を含めているか
