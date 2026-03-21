---
paths: frontend/src/__tests__/**/*.ts, frontend/src/__tests__/**/*.tsx
description: React フロントエンドテスト規約
---

# Frontend テスト規約（TypeScript）

## 基本原則

- **TDD実践**: テストファーストを推奨
- **カバレッジ目標**: 80%以上
- **テストの独立性**: 各テストは他のテストに依存しない
- **テストの可読性**: テスト名から何をテストしているか明確にわかる

---

## ディレクトリ構成

```
frontend/src/
├── __tests__/
│   ├── expense/
│   │   ├── components/
│   │   │   ├── ExpenseForm.test.tsx
│   │   │   └── ExpenseList.test.tsx
│   │   ├── hooks/
│   │   │   └── useExpenses.test.ts
│   │   ├── utils/
│   │   │   └── expenseMockData.ts
│   │   └── integration/
│   │       └── ExpenseFlow.test.tsx
│   ├── weeklyReport/
│   │   └── ...
│   └── setup.ts                  # グローバルセットアップ
├── components/
│   └── expense/
│       └── ExpenseForm.tsx       # テスト対象
└── hooks/
    └── expense/
        └── useExpenses.ts        # テスト対象
```

## ファイル命名規則

| パターン | 例 |
|---------|-----|
| `*.test.tsx` | `ExpenseForm.test.tsx`（コンポーネント） |
| `*.test.ts` | `useExpenses.test.ts`（フック・ユーティリティ） |
| `*MockData.ts` | `expenseMockData.ts`（モックデータ） |

---

## テスト構造（Jest + React Testing Library）

```typescript
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ExpenseForm } from '@/components/expense/ExpenseForm';
import { createMockExpenseFormProps } from '../utils/expenseMockData';

describe('ExpenseForm', () => {
  // グループ: 基本レンダリング
  describe('基本レンダリング', () => {
    test('新規作成モードで正しくレンダリングされる', () => {
      // Arrange
      const props = createMockExpenseFormProps();

      // Act
      render(<ExpenseForm {...props} />);

      // Assert
      expect(screen.getByLabelText('タイトル')).toBeInTheDocument();
      expect(screen.getByLabelText('カテゴリ')).toBeInTheDocument();
      expect(screen.getByLabelText('金額')).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '保存' })).toBeInTheDocument();
    });

    test('編集モードで初期値が表示される', () => {
      const props = createMockExpenseFormProps({
        initialValues: {
          title: 'テスト経費',
          amount: 5000,
        },
      });

      render(<ExpenseForm {...props} />);

      expect(screen.getByDisplayValue('テスト経費')).toBeInTheDocument();
      expect(screen.getByDisplayValue('5000')).toBeInTheDocument();
    });
  });

  // グループ: ユーザーインタラクション
  describe('ユーザーインタラクション', () => {
    test('フォーム送信時にonSubmitが呼ばれる', async () => {
      const user = userEvent.setup();
      const onSubmit = jest.fn().mockResolvedValue(undefined);
      const props = createMockExpenseFormProps({ onSubmit });

      render(<ExpenseForm {...props} />);

      // フォーム入力
      await user.type(screen.getByLabelText('タイトル'), 'テスト経費');
      await user.type(screen.getByLabelText('金額'), '5000');

      // 送信
      await user.click(screen.getByRole('button', { name: '保存' }));

      // 検証
      await waitFor(() => {
        expect(onSubmit).toHaveBeenCalledTimes(1);
        expect(onSubmit).toHaveBeenCalledWith(
          expect.objectContaining({
            title: 'テスト経費',
            amount: 5000,
          })
        );
      });
    });
  });

  // グループ: バリデーション
  describe('バリデーション', () => {
    test('必須項目が空の場合エラーが表示される', async () => {
      const user = userEvent.setup();
      const props = createMockExpenseFormProps();

      render(<ExpenseForm {...props} />);

      // 空のまま送信
      await user.click(screen.getByRole('button', { name: '保存' }));

      await waitFor(() => {
        expect(screen.getByText('タイトルは必須です')).toBeInTheDocument();
      });
    });
  });
});
```

---

## モックデータ

```typescript
// __tests__/expense/utils/expenseMockData.ts
import { ExpenseData, ExpenseFormProps, ExpenseFormValues } from '@/types';

export function createMockExpenseData(overrides?: Partial<ExpenseData>): ExpenseData {
  return {
    id: 'test-expense-id',
    userId: 'test-user-id',
    title: 'テスト経費',
    category: 'entertainment',
    amount: 5000,
    status: 'draft',
    expenseDate: new Date('2024-01-15'),
    createdAt: new Date('2024-01-15T10:00:00'),
    updatedAt: new Date('2024-01-15T10:00:00'),
    ...overrides,
  };
}

export function createMockExpenseFormProps(
  overrides?: Partial<ExpenseFormProps>
): ExpenseFormProps {
  return {
    onSubmit: jest.fn().mockResolvedValue(undefined),
    onCancel: jest.fn(),
    isSubmitting: false,
    ...overrides,
  };
}

export function createMockExpenseList(count: number = 5): ExpenseData[] {
  return Array.from({ length: count }, (_, i) =>
    createMockExpenseData({
      id: `expense-${i + 1}`,
      title: `テスト経費 ${i + 1}`,
      amount: (i + 1) * 1000,
    })
  );
}
```

---

## MSW（API モック）

```typescript
// __tests__/setup.ts
import { setupServer } from 'msw/node';
import { rest } from 'msw';

export const handlers = [
  rest.get('/api/v1/expenses', (req, res, ctx) => {
    return res(
      ctx.json({
        items: createMockExpenseList(10),
        total: 10,
        page: 1,
        limit: 20,
        total_pages: 1,
      })
    );
  }),

  rest.post('/api/v1/expenses', (req, res, ctx) => {
    return res(
      ctx.status(201),
      ctx.json(createMockExpenseData())
    );
  }),
];

export const server = setupServer(...handlers);

// テストファイルでの使用
beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

// 特定テストでのハンドラー上書き
test('APIエラー時にエラーメッセージが表示される', async () => {
  server.use(
    rest.get('/api/v1/expenses', (req, res, ctx) => {
      return res(ctx.status(500), ctx.json({ error: 'Server Error' }));
    })
  );

  // テスト実行...
});
```

---

## フックのテスト

```typescript
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useExpenses } from '@/hooks/expense/useExpenses';

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  });

  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  );
};

describe('useExpenses', () => {
  test('経費一覧を取得できる', async () => {
    const { result } = renderHook(() => useExpenses(), {
      wrapper: createWrapper(),
    });

    // 初期状態
    expect(result.current.isLoading).toBe(true);

    // データ取得完了後
    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.expenses).toHaveLength(10);
    expect(result.current.error).toBeNull();
  });
});
```

---

## テスト命名規約

### 日本語を使用

```typescript
// ✅ 推奨: 日本語で何をテストしているか明確に
describe('ExpenseForm', () => {
  test('新規作成モードで正しくレンダリングされる', () => { });
  test('必須項目が空の場合エラーが表示される', () => { });
  test('送信中は保存ボタンが無効になる', () => { });
});
```

### カテゴリ分け

| カテゴリ | 説明 |
|---------|------|
| 正常系 | 期待通りの入力で期待通りの結果 |
| 異常系 | 不正な入力でエラーが発生 |
| 境界値 | 境界条件でのテスト |
| 権限 | 権限に基づくアクセス制御 |

---

## テスト実行コマンド

```bash
# 全テスト実行
cd frontend && npm test

# 監視モード
cd frontend && npm test -- --watch

# カバレッジ
cd frontend && npm test -- --coverage

# 特定ファイル
cd frontend && npm test -- ExpenseForm.test.tsx

# E2Eテスト（Playwright）
cd frontend && npm run test:e2e
cd frontend && npm run test:e2e:smoke
```

---

## チェックリスト

- [ ] `*.test.tsx` / `*.test.ts` ファイル名になっているか
- [ ] `describe` / `test` で階層化しているか
- [ ] `@testing-library/react` を使用しているか
- [ ] モックデータは `createMock*` 関数化されているか
- [ ] MSW でAPIモックを設定しているか（必要な場合）
- [ ] `userEvent` を使用しているか（ユーザー操作）
- [ ] `waitFor` を使用しているか（非同期処理）
- [ ] テストは独立して実行可能か
- [ ] テストの実行順序に依存していないか
- [ ] カバレッジ目標（80%）を達成しているか
