---
paths: frontend/src/components/**/*.tsx
description: コンポーネント作成時の設計パターン
---

# コンポーネント設計パターン

## 基本構造

```typescript
'use client';  // MUI/hooks使用時は必須

import React, { useState, useCallback, useMemo } from 'react';
import { Box, Typography, Button } from '@mui/material';
import type { SxProps, Theme } from '@mui/material/styles';

// ═══════════════════════════════════════════════════════
// 型定義
// ═══════════════════════════════════════════════════════
interface ComponentNameProps {
  /** 必須プロパティの説明 */
  requiredProp: string;
  /** オプショナルプロパティの説明 */
  optionalProp?: boolean;
  /** コールバック関数 */
  onAction?: (value: string) => void;
  /** スタイル拡張用 */
  sx?: SxProps<Theme>;
  /** テストID */
  'data-testid'?: string;
}

// ═══════════════════════════════════════════════════════
// スタイル定義
// ═══════════════════════════════════════════════════════
const styles = {
  container: {
    display: 'flex',
    flexDirection: 'column',
    gap: 2,
  },
  header: {
    display: 'flex',
    alignItems: 'center',
    gap: 1,
  },
} as const;

// ═══════════════════════════════════════════════════════
// コンポーネント
// ═══════════════════════════════════════════════════════
export const ComponentName: React.FC<ComponentNameProps> = ({
  requiredProp,
  optionalProp = false,  // デフォルト値を明示
  onAction,
  sx,
  'data-testid': testId,
}) => {
  // ─ State
  const [localState, setLocalState] = useState(false);

  // ─ Handlers
  const handleClick = useCallback(() => {
    onAction?.(requiredProp);
  }, [onAction, requiredProp]);

  // ─ Computed
  const displayValue = useMemo(
    () => requiredProp.toUpperCase(),
    [requiredProp]
  );

  // ─ Render
  return (
    <Box sx={{ ...styles.container, ...sx }} data-testid={testId}>
      <Typography variant="h6">{displayValue}</Typography>
      <Button onClick={handleClick} disabled={optionalProp}>
        Action
      </Button>
    </Box>
  );
};

export default ComponentName;
```

---

## Props型定義規約

### 必須ルール

```typescript
interface Props {
  /** JSDocコメントは必須 */
  requiredProp: string;

  /** オプショナルは ? で明示 */
  optionalProp?: boolean;

  /** コールバックは on + ActionName */
  onSubmit?: (data: FormData) => void;
  onChange?: (value: string) => void;

  /** スタイル拡張用sxを受け取る */
  sx?: SxProps<Theme>;

  /** テストID対応 */
  'data-testid'?: string;

  /** 子要素 */
  children?: React.ReactNode;
}
```

### MUI Propsの継承

```typescript
// 既存のPropsを拡張する場合
interface DetailDrawerProps extends Omit<DrawerProps, 'open' | 'onClose'> {
  open: boolean;
  onClose: () => void;
  title?: React.ReactNode;
  // 追加のprops...
}
```

---

## イベントハンドラ命名規約

| 場面 | パターン | 例 |
|------|---------|-----|
| Props定義 | `on + Action` | `onClose`, `onSubmit`, `onChange` |
| 実装 | `handle + Action` | `handleClose`, `handleSubmit`, `handleChange` |

```typescript
interface Props {
  onSubmit: (data: FormData) => void;  // Props
}

const Component: React.FC<Props> = ({ onSubmit }) => {
  const handleSubmit = (data: FormData) => {  // 実装
    // 追加処理...
    onSubmit(data);
  };

  return <form onSubmit={handleSubmit}>...</form>;
};
```

---

## スタイリング規約

### sx prop を優先

```typescript
// ✅ 推奨: sx prop
<Box
  sx={{
    display: 'flex',
    gap: 2,
    bgcolor: 'background.paper',
    borderRadius: 2,
    // レスポンシブ
    mb: { xs: 2, md: 3 },
    // 外部sxをマージ
    ...sx,
  }}
>
```

### スタイル定数化

```typescript
// コンポーネント上部で定義
const styles = {
  container: { mb: 4 },
  header: { display: 'flex', gap: 1 },
  content: { p: 2 },
} as const;

// 使用
<Box sx={styles.container}>
  <Box sx={styles.header}>...</Box>
</Box>
```

### テーマ値の参照

```typescript
// ✅ テーマカラーを参照
color: 'primary.main'
bgcolor: 'background.paper'
borderColor: 'divider'

// ❌ ハードコーディング
color: '#1976d2'
backgroundColor: '#ffffff'
```

---

## 状態管理規約

| 用途 | 使用するもの |
|------|-------------|
| UI状態（open/close等） | `useState` |
| フォーム状態 | React Hook Form (`useForm`, `Controller`) |
| APIデータ | React Query (`useQuery`, `useMutation`) |
| イベントハンドラ | `useCallback` |
| 計算結果 | `useMemo` |
| グローバル状態 | Context API（AuthContext等） |

---

## MUIコンポーネント使用パターン

### よく使うコンポーネント

| コンポーネント | 用途 |
|--------------|------|
| `Box` | レイアウト基盤（最も多用） |
| `Stack` | 方向別配置（spacing活用） |
| `Typography` | テキスト（variant必須） |
| `Button` / `IconButton` | アクション |
| `Paper` | カード背景 |
| `Divider` | セクション区切り |
| `Skeleton` | ローディング表示 |
| `Alert` | エラー・警告・情報表示 |

### Stack vs Box

```typescript
// ✅ 縦並び・横並びにはStack
<Stack spacing={2} direction="row">
  <Button>A</Button>
  <Button>B</Button>
</Stack>

// ✅ 複雑なレイアウトにはBox
<Box sx={{ display: 'grid', gridTemplateColumns: '1fr 2fr', gap: 2 }}>
  ...
</Box>
```

---

## チェックリスト

新しいコンポーネントを作成する際：

- [ ] 共通コンポーネント一覧を確認し、既存で対応できないか検討したか
- [ ] `'use client'` ディレクティブが必要か確認したか
- [ ] `React.FC<Props>` 型定義をしているか
- [ ] Props に JSDoc コメントを付けているか
- [ ] `sx` prop を受け取り、外部スタイルをマージできるか
- [ ] `data-testid` prop に対応しているか
- [ ] イベントハンドラは `on + Action` / `handle + Action` 規約に従っているか
- [ ] `useCallback` / `useMemo` で適切にメモ化しているか
- [ ] テーマ値を参照しているか（ハードコーディングしていないか）

---

## 関連規約

- 共通コンポーネント一覧 → [components-index.md](./components-index.md)
- AppTable 実装パターン → [components-table.md](./components-table.md)
