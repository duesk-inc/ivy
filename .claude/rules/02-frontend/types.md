---
paths: frontend/src/types/**/*.ts
---

# 型定義規約

## ディレクトリ構成

```
frontend/src/types/
├── expense.ts           # 機能ドメイン単位
├── leave.ts
├── engineer.ts
├── weeklyReport.ts
├── notification.ts
├── common.ts            # 共通型
├── admin/               # 管理者向け（サブディレクトリ）
│   ├── index.ts         # re-export
│   ├── invoice.ts
│   ├── client.ts
│   └── weeklyReport.ts
└── index.ts             # 全体re-export
```

---

## 型の分類と命名規則

### 1. Backend API型（snake_case）

APIレスポンスをそのまま受け取る型。JSONのsnake_caseを維持。

```typescript
// ✅ Backend API型: [Domain]BackendResponse suffix
interface ExpenseBackendResponse {
  id: string;
  user_id: string;           // snake_case のまま
  title: string;
  category: string;
  amount: number;
  expense_date: string;      // snake_case のまま
  created_at: string;        // snake_case のまま
  updated_at: string;
}

interface ExpenseListBackendResponse {
  items: ExpenseBackendResponse[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;       // snake_case のまま
}
```

### 2. Frontend内部型（camelCase）

コンポーネントやフックで使用する型。camelCaseに変換済み。

```typescript
// ✅ Frontend内部型: [Domain]Data または [Domain] suffix
interface ExpenseData {
  id: string;
  userId: string;            // camelCase
  title: string;
  category: ExpenseCategoryType;
  amount: number;
  expenseDate: Date;         // camelCase + Date型
  createdAt: Date;           // camelCase + Date型
  updatedAt: Date;
}

// 関連データを含む詳細型
interface ExpenseDetail extends ExpenseData {
  user?: UserMinimal;
  receipts?: ReceiptData[];
  approver?: UserMinimal;
}
```

### 3. リクエストパラメータ型

APIに送信するパラメータの型。

```typescript
// ✅ リクエスト型: [Domain]Params または [Domain]Request suffix
interface ExpenseListParams {
  page?: number;
  limit?: number;
  status?: ExpenseStatusType;
  category?: ExpenseCategoryType;
  startDate?: string;        // ISO8601形式の文字列
  endDate?: string;
  sortBy?: ExpenseSortField;
  sortOrder?: SortDirection;
}

interface CreateExpenseRequest {
  title: string;
  category: ExpenseCategoryType;
  amount: number;
  expenseDate: string;
  description?: string;
  receiptUrls?: string[];
}

interface UpdateExpenseRequest {
  title?: string;
  category?: ExpenseCategoryType;
  amount?: number;
  description?: string;
}
```

### 4. フロント用レスポンス型

API層で変換後の型。

```typescript
// ✅ レスポンス型: [Domain]Response または [Domain]ListResponse suffix
interface ExpenseResponse {
  items: ExpenseData[];
  total: number;
  page: number;
  limit: number;
  totalPages: number;        // camelCase
}
```

---

## Union型と列挙型

### ステータス・種別の定義

```typescript
// ✅ Union型: [Domain][Type]Type suffix
type ExpenseStatusType =
  | 'draft'
  | 'submitted'
  | 'approved'
  | 'rejected'
  | 'paid'
  | 'cancelled';

type ExpenseCategoryType =
  | 'transport'
  | 'entertainment'
  | 'supplies'
  | 'communication'
  | 'other';

type LeaveRequestStatusType =
  | 'pending'
  | 'approved'
  | 'rejected'
  | 'cancelled';
```

### ソート関連

```typescript
// ソート方向（共通）
type SortDirection = 'asc' | 'desc';

// ソート可能フィールド（ドメイン別）
type ExpenseSortField =
  | 'expense_date'
  | 'amount'
  | 'created_at'
  | 'status';

type EngineerSortField =
  | 'name'
  | 'email'
  | 'hire_date'
  | 'status';
```

---

## フィルター・ページネーション

### フィルター型

```typescript
// ✅ フィルター型: [Domain]Filters suffix
interface ExpenseFilters {
  status?: ExpenseStatusType | ExpenseStatusType[];
  category?: ExpenseCategoryType;
  startDate?: Date;
  endDate?: Date;
  minAmount?: number;
  maxAmount?: number;
  userId?: string;
}
```

### ページネーション型

```typescript
// ✅ 共通ページネーション型
interface Pagination {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}

// ページ変更ハンドラ
type PageChangeHandler = (page: number) => void;

// ページサイズ変更ハンドラ
type PageSizeChangeHandler = (size: number) => void;
```

---

## snake_case → camelCase 変換

### API層での一元変換

```typescript
// lib/api/expense.ts
import { toCamelCase } from '@/utils/caseConverter';

export const getExpenses = async (params: ExpenseListParams): Promise<ExpenseResponse> => {
  const client = createPresetApiClient('auth');
  const response = await client.get<ExpenseListBackendResponse>('/expenses', { params });

  // API層で変換を一元化
  return {
    items: response.data.items.map(convertExpenseFromBackend),
    total: response.data.total,
    page: response.data.page,
    limit: response.data.limit,
    totalPages: response.data.total_pages,
  };
};

// 変換関数
const convertExpenseFromBackend = (expense: ExpenseBackendResponse): ExpenseData => ({
  id: expense.id,
  userId: expense.user_id,
  title: expense.title,
  category: expense.category as ExpenseCategoryType,
  amount: expense.amount,
  expenseDate: new Date(expense.expense_date),
  createdAt: new Date(expense.created_at),
  updatedAt: new Date(expense.updated_at),
});
```

### コンポーネントでの使用

```typescript
// コンポーネントではcamelCase型のみ使用
const ExpenseList: React.FC = () => {
  const { data } = useExpenses();

  return (
    <div>
      {data?.items.map((expense: ExpenseData) => (
        <ExpenseCard
          key={expense.id}
          userId={expense.userId}          // camelCase
          expenseDate={expense.expenseDate} // Date型
        />
      ))}
    </div>
  );
};
```

---

## Propsの型定義

### コンポーネントProps

```typescript
// ✅ Props型: [Component]Props suffix
interface ExpenseCardProps {
  /** 経費データ */
  expense: ExpenseData;
  /** 編集可能か */
  editable?: boolean;
  /** 編集クリック時 */
  onEdit?: (expense: ExpenseData) => void;
  /** 削除クリック時 */
  onDelete?: (id: string) => void;
  /** スタイル拡張 */
  sx?: SxProps<Theme>;
}

// ✅ 使用例
const ExpenseCard: React.FC<ExpenseCardProps> = ({
  expense,
  editable = false,
  onEdit,
  onDelete,
  sx,
}) => {
  // ...
};
```

### フォームProps

```typescript
// フォーム値の型
interface ExpenseFormValues {
  title: string;
  category: ExpenseCategoryType;
  amount: number;
  expenseDate: Date | null;
  description: string;
  receiptUrls: string[];
}

// フォームProps
interface ExpenseFormProps {
  /** 初期値（編集時） */
  initialValues?: Partial<ExpenseFormValues>;
  /** 送信時 */
  onSubmit: (values: ExpenseFormValues) => Promise<void>;
  /** キャンセル時 */
  onCancel: () => void;
  /** 送信中か */
  isSubmitting?: boolean;
}
```

---

## 共通型

### ユーザー関連

```typescript
// 最小限のユーザー情報
interface UserMinimal {
  id: string;
  name: string;
  email: string;
}

// ユーザー詳細
interface UserData extends UserMinimal {
  firstName: string;
  lastName: string;
  role: UserRoleType;
  department?: string;
  position?: string;
}

type UserRoleType = 'admin' | 'manager' | 'engineer';
```

### API共通

```typescript
// APIエラーレスポンス
interface ApiErrorResponse {
  error: {
    code: string;
    message: string;
    details?: Record<string, string>;
  };
  status: number;
}

// 一覧レスポンスの共通構造
interface ListResponse<T> {
  items: T[];
  total: number;
  page: number;
  limit: number;
  totalPages: number;
}
```

---

## re-export パターン

### index.ts での集約

```typescript
// types/index.ts
export * from './expense';
export * from './leave';
export * from './engineer';
export * from './common';
export * from './admin';
```

### サブディレクトリの index.ts

```typescript
// types/admin/index.ts
export * from './invoice';
export * from './client';
export * from './weeklyReport';
```

### 使用側

```typescript
// ✅ 推奨: index.ts 経由でインポート
import type { ExpenseData, ExpenseFilters, UserMinimal } from '@/types';

// ✅ 許容: 直接インポート（大きなファイルの場合）
import type { ExpenseData } from '@/types/expense';
```

---

## チェックリスト

新しい型を定義する際：

- [ ] 適切なsuffixを使用しているか
  - Backend API型: `*BackendResponse`
  - Frontend内部型: `*Data` または `*Detail`
  - リクエスト型: `*Params` または `*Request`
  - レスポンス型: `*Response`
  - フィルター型: `*Filters`
  - Props型: `*Props`
- [ ] Union型は `*Type` suffixになっているか
- [ ] snake_case/camelCase の変換はAPI層で一元化されているか
- [ ] JSDocコメントを付けているか（特にProps）
- [ ] index.ts でre-exportしているか

Backend API型を追加した場合：

- [ ] 対応するFrontend内部型も定義したか
- [ ] 変換関数を実装したか
- [ ] 変換をAPI層に配置したか
