---
paths: frontend/src/hooks/**/*.ts, frontend/src/hooks/**/*.tsx
---

# React Query 使用規約

## キャッシュキー管理

**必ず `queryKeys` オブジェクトからキーを参照すること**

### queryKeys の場所

```typescript
// frontend/src/lib/tanstack-query.ts
import { queryKeys } from '@/lib/tanstack-query';
```

### 使用パターン

```typescript
// ✅ 正しい: queryKeys から参照
useQuery({
  queryKey: queryKeys.adminEngineers(params),
  queryFn: () => fetchEngineers(params),
});

// ❌ 禁止: 直接文字列を指定
useQuery({
  queryKey: ['engineers', params],
  queryFn: () => fetchEngineers(params),
});
```

### キー構造

```typescript
export const queryKeys = {
  // 管理者系
  adminDashboard: ['admin', 'dashboard'] as const,
  adminEngineers: (params?: unknown) => ['admin', 'engineers', params] as const,
  adminEngineerDetail: (id: string) => ['admin', 'engineers', id] as const,
  adminEngineerStatistics: ['admin', 'engineers', 'statistics'] as const,

  // 週報関連
  adminWeeklyReports: (params?: unknown) => ['admin', 'weeklyReports', params] as const,

  // エンジニア個人向け
  engineerProfile: ['engineer', 'profile'] as const,
  engineerWeeklyReports: (params?: unknown) => ['engineer', 'weeklyReports', params] as const,

  // 共通
  notifications: ['notifications'] as const,
  notificationUnreadCount: ['notifications', 'unreadCount'] as const,
};
```

---

## useQuery パターン

### 基本パターン

```typescript
import { useQuery } from '@tanstack/react-query';
import { queryKeys } from '@/lib/tanstack-query';
import { adminEngineerApi } from '@/lib/api/admin/engineer';

export const useEngineersQuery = (params?: GetEngineersParams) => {
  return useQuery({
    queryKey: queryKeys.adminEngineers(params),
    queryFn: async () => {
      const response = await adminEngineerApi.getEngineers(params || {});
      return response;
    },
  });
};
```

### enabled 制御

```typescript
// IDが存在する場合のみ実行
export const useEngineerDetailQuery = (id: string | undefined) => {
  return useQuery({
    queryKey: queryKeys.adminEngineerDetail(id!),
    queryFn: () => adminEngineerApi.getEngineerDetail(id!),
    enabled: !!id,  // idが存在する場合のみ実行
  });
};
```

### staleTime / cacheTime 設定

```typescript
// マスターデータなど変更頻度が低いもの
useQuery({
  queryKey: queryKeys.leaveTypes,
  queryFn: fetchLeaveTypes,
  staleTime: 1000 * 60 * 30,  // 30分間fresh
  cacheTime: 1000 * 60 * 60,  // 1時間キャッシュ保持
});

// 頻繁に更新されるデータ
useQuery({
  queryKey: queryKeys.notifications,
  queryFn: fetchNotifications,
  staleTime: 1000 * 30,  // 30秒
  refetchInterval: 1000 * 60,  // 1分ごとに自動更新
});
```

---

## useMutation パターン

### 基本パターン

```typescript
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { queryKeys } from '@/lib/tanstack-query';
import { useToast } from '@/hooks/common/useToast';

export const useCreateEngineer = () => {
  const queryClient = useQueryClient();
  const { showSuccess, showError } = useToast();

  return useMutation({
    mutationFn: (data: CreateEngineerRequest) => {
      return adminEngineerApi.createEngineer(data);
    },

    onSuccess: () => {
      // キャッシュ無効化
      queryClient.invalidateQueries({ queryKey: queryKeys.adminEngineers() });
      queryClient.invalidateQueries({ queryKey: queryKeys.adminEngineerStatistics });

      // 成功通知
      showSuccess('エンジニアを登録しました');
    },

    onError: (error: any) => {
      // エラー通知
      const message = error?.message || 'エンジニアの登録に失敗しました';
      showError(message);
    },
  });
};
```

### 使用方法

```typescript
const { mutateAsync: createEngineer, isPending } = useCreateEngineer();

const handleSubmit = async (data: FormData) => {
  try {
    await createEngineer(data);
    // 成功後の処理（onSuccessで通知済み）
  } catch (error) {
    // エラー処理（onErrorで通知済み）
  }
};
```

---

## キャッシュ操作

### 無効化（Invalidation）

```typescript
const queryClient = useQueryClient();

// 特定のクエリを無効化
queryClient.invalidateQueries({ queryKey: queryKeys.adminEngineerDetail(id) });

// プレフィックスで複数無効化
queryClient.invalidateQueries({ queryKey: ['admin', 'engineers'] });

// すべて無効化（非推奨）
queryClient.invalidateQueries();
```

### プリフェッチ

```typescript
// 次ページをプリフェッチ
const prefetchNextPage = useCallback(() => {
  queryClient.prefetchQuery({
    queryKey: queryKeys.adminEngineers({ ...params, page: page + 1 }),
    queryFn: () => fetchEngineers({ ...params, page: page + 1 }),
  });
}, [queryClient, params, page]);
```

### 楽観的更新

```typescript
useMutation({
  mutationFn: updateEngineer,

  onMutate: async (newData) => {
    // 進行中のクエリをキャンセル
    await queryClient.cancelQueries({ queryKey: queryKeys.adminEngineerDetail(id) });

    // 現在のデータをバックアップ
    const previousData = queryClient.getQueryData(queryKeys.adminEngineerDetail(id));

    // 楽観的に更新
    queryClient.setQueryData(queryKeys.adminEngineerDetail(id), newData);

    return { previousData };
  },

  onError: (err, newData, context) => {
    // エラー時はロールバック
    queryClient.setQueryData(queryKeys.adminEngineerDetail(id), context?.previousData);
  },

  onSettled: () => {
    // 完了後に再フェッチ
    queryClient.invalidateQueries({ queryKey: queryKeys.adminEngineerDetail(id) });
  },
});
```

---

## 複雑な状態管理パターン

フィルター、ソート、ページネーションを含むフック：

```typescript
export const useExpenses = ({
  initialFilters = {},
  initialPage = 1,
  autoFetch = true,
}: UseExpensesParams = {}): UseExpensesReturn => {
  const queryClient = useQueryClient();

  // ─ Local State
  const [filters, setFiltersState] = useState<ExpenseFilters>(initialFilters);
  const [sort, setSortState] = useState<ExpenseSort>(DEFAULT_SORT);
  const [page, setPageState] = useState(initialPage);

  // ─ Query
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['expenses', filters, sort, page],
    queryFn: () => getExpenseList({
      ...filters,
      page,
      sortField: sort.field,
      sortDirection: sort.direction,
    }),
    enabled: autoFetch,
  });

  // ─ Filter Helpers
  const setFilters = useCallback((newFilters: Partial<ExpenseFilters>) => {
    setFiltersState(prev => ({ ...prev, ...newFilters }));
    setPageState(1);  // フィルター変更時はページをリセット
  }, []);

  const toggleSort = useCallback((field: SortableFieldType) => {
    setSortState(prev => ({
      field,
      direction: prev.field === field && prev.direction === 'asc' ? 'desc' : 'asc',
    }));
  }, []);

  // ─ Cache Operations
  const invalidateCache = useCallback(() => {
    queryClient.invalidateQueries({ queryKey: ['expenses'] });
  }, [queryClient]);

  // ─ Return
  return {
    expenses: data?.items || [],
    pagination: { page, total: data?.total || 0 },
    filters,
    sort,
    isLoading,
    error,
    setFilters,
    setSort: setSortState,
    setPage: setPageState,
    toggleSort,
    invalidateCache,
    refetch,
  };
};
```

---

## エラーハンドリング

### onError コールバック

```typescript
useMutation({
  mutationFn: createExpense,
  onError: (error: any) => {
    // handleApiError で標準化されたエラーを受け取る
    const message = error?.error?.message || error?.message || 'エラーが発生しました';
    showError(message);

    // 必要に応じてログ
    DebugLogger.error({ category: 'EXPENSE', operation: 'Create' }, message, error);
  },
});
```

### グローバルエラーハンドリング

```typescript
// QueryClient 設定
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: (failureCount, error: any) => {
        // 認証エラーはリトライしない
        if (error?.status === 401 || error?.status === 403) {
          return false;
        }
        return failureCount < 3;
      },
    },
    mutations: {
      onError: (error: any) => {
        // グローバルエラー通知
        console.error('Mutation error:', error);
      },
    },
  },
});
```

---

## Server Component と React Query の使い分け

### 重要: ユーザー固有データは必ず React Query を使用

**Monsteraプロジェクトでは、ユーザー固有のデータは必ずクライアントサイドでReact Queryを使用してフェッチすること。**

Server Component での ISR（`revalidate`）はユーザー単位ではなくルート単位でキャッシュするため、ユーザー固有データをキャッシュすると**情報漏えいリスク**が発生します。

### React Query を使用すべきケース（ユーザー固有データ）

| データ種別 | フェッチ方法 |
|-----------|-------------|
| ダッシュボード統計 | `useDashboard()` |
| 通知一覧 | `useNotifications()` |
| 自分の経費 | `useExpenses()` |
| 自分の週報 | `useWeeklyReports()` |
| プロフィール情報 | `useProfile()` |
| 休暇申請履歴 | `useLeaveRequests()` |

```typescript
// ✅ 正しい: Client ComponentでReact Queryを使用
'use client';

export function DashboardClient() {
  const { data, isLoading, error } = useDashboard();
  // ...
}
```

### Server Component でフェッチして良いケース（公開データのみ）

| データ種別 | 条件 |
|-----------|------|
| 祝日マスタ | ログイン不要で取得可能 |
| 設定マスタ | 全ユーザー共通 |
| 公開お知らせ | 認証不要 |

```typescript
// ✅ 許可: ユーザー固有でないデータ
export default async function HolidaysPage() {
  const holidays = await getPublicHolidays({ revalidate: 3600 });
  return <HolidayList holidays={holidays} />;
}
```

### 禁止パターン

```typescript
// ❌ 禁止: Server ComponentでユーザーデータをISRキャッシュ
export default async function DashboardPage() {
  const data = await getDashboardData({ revalidate: 60 }); // 危険！
  return <DashboardClient initialData={data} />;
}
```

> **参照**: 詳細は `05-pitfalls/known-issues.md` の「3.4 Server Component で ISR を使用したユーザー固有データのキャッシュ」を参照

---

## チェックリスト

### useQuery

- [ ] `queryKeys` からキーを参照しているか
- [ ] 条件付き実行は `enabled` を使用しているか
- [ ] 適切な `staleTime` / `cacheTime` を設定しているか

### useMutation

- [ ] `onSuccess` でキャッシュを無効化しているか
- [ ] `onSuccess` / `onError` で適切な通知を表示しているか
- [ ] `useToast` フックを使用しているか

### 全般

- [ ] キャッシュキーが一貫しているか
- [ ] 不要なリフェッチが発生していないか
- [ ] エラーハンドリングが実装されているか
- [ ] **ユーザー固有データはClient Component + React Queryでフェッチしているか**

### queryKeys に新しいキーを追加した場合

- [ ] `frontend/src/lib/tanstack-query.ts` の `queryKeys` オブジェクトに追加したか
- [ ] 命名規則（`adminXxx`, `engineerXxx` 等）に従っているか

> **注意**: 新しいクエリキーを追加したら `tanstack-query.ts` を必ず更新してください。
