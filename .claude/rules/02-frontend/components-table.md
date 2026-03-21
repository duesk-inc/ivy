---
paths: frontend/src/components/**/*.tsx, frontend/src/app/**/*.tsx
description: AppTable を使用した一覧画面の実装パターン
---

# AppTable 実装パターン

`AppTable` を使用した一覧画面の実装における標準パターンを定義します。

---

## 行クリックとチェックボックスの独立性

AppTable では **行クリックとチェックボックス選択は独立したアクション** として設計されています。

| アクション | 結果 |
|-----------|------|
| 行クリック | `onRowClick` が発火（詳細画面への遷移等） |
| チェックボックスクリック | `selection.onSelectionChange` が発火（選択状態の変更） |

```typescript
<AppTable
  data={data}
  columns={columns}
  onRowClick={(row) => router.push(`/detail/${row.id}`)}  // 行クリック → 詳細遷移
  selection={{
    selectedIds,
    onSelectionChange: setSelectedIds,  // チェックボックス → 選択変更
  }}
/>
```

**ポイント:**
- 行クリックとチェックボックスは干渉しない（両方を同時に使用可能）
- チェックボックスは明示的にクリックした場合のみトグルされる
- 行全体をクリックしても選択状態は変わらない

---

## 主要カラムのリンク化

一覧から詳細画面へのナビゲーションは、**主要カラム（名前等）をリンク化**するパターンを推奨します。

```typescript
import Link from 'next/link';

const columns: TableColumn<Project>[] = [
  {
    id: 'project_name',
    label: '案件名',
    render: (value, row) => (
      <Link
        href={`/manage/business/projects/${row.id}`}
        onClick={(e) => e.stopPropagation()}  // 必須: 行クリックとの干渉防止
        style={{ textDecoration: 'none' }}
      >
        <Typography
          variant="body2"
          fontWeight="medium"
          sx={{
            color: 'primary.main',
            '&:hover': { textDecoration: 'underline' },
          }}
        >
          {value}
        </Typography>
      </Link>
    ),
  },
];
```

**ポイント:**
- `onClick={(e) => e.stopPropagation()}` は必須（行クリックイベントとの干渉防止）
- リンク色は `primary.main` を使用
- ホバー時にアンダーライン表示

---

## UserNameCell によるリンク化

ユーザー名カラムをリンク化する場合は、`UserNameCell` の `linkHref` prop を使用します。

```typescript
import { UserNameCell } from '@/components/common';

const columns: TableColumn<LeaveRequest>[] = [
  {
    id: 'user',
    header: '氏名',
    render: (_, row) => (
      <UserNameCell
        sei={row.user.sei}
        mei={row.user.mei}
        seiKana={row.user.seiKana}
        meiKana={row.user.meiKana}
        email={row.user.email}
        linkHref={`/manage/engineers/${row.user.id}`}
      />
    ),
  },
];
```

**ポイント:**
- `linkHref` を指定すると氏名部分がリンク化される
- `stopPropagation` は `UserNameCell` 内部で処理済み
- カナ・メール表示は維持され、氏名のみがリンク化される

---

## InlineSelect によるインライン編集

ステータス等を一覧から直接編集する場合は `InlineSelect` を使用します。

```typescript
import { InlineSelect } from '@/components/common';

// 楽観的UI更新のハンドラ
const handleStatusChange = async (id: string, newStatus: string) => {
  const previousData = [...data];

  // 楽観的更新（即座にUIに反映）
  setData((prev) =>
    prev.map((item) => (item.id === id ? { ...item, status: newStatus } : item))
  );

  try {
    await updateStatus(id, { status: newStatus });
    showSuccess('ステータスを更新しました');
  } catch (error) {
    // エラー時はロールバック
    setData(previousData);
    showError('更新に失敗しました');
  }
};

// カラム定義
const columns: TableColumn<Project>[] = [
  {
    id: 'status',
    label: 'ステータス',
    render: (value, row) => (
      <InlineSelect
        value={value}
        options={STATUS_LABELS}  // Record<string, string> または InlineSelectOption[]
        onChange={(newStatus) => handleStatusChange(row.id, newStatus)}
        aria-label="ステータス"
      />
    ),
  },
];
```

**ポイント:**
- `InlineSelect` は `stopPropagation` を内蔵済み
- 楽観的UI更新を実装（即座にUIに反映、エラー時はロールバック）
- `options` は `Record<string, string>` または `InlineSelectOption[]` を受け付ける

---

## RowActionsMenu の使用（必要な場合のみ）

編集・削除等、インライン編集では対応できない行操作がある場合に `RowActionsMenu` を使用します。

**RowActionsMenuが不要なケース:**
- 詳細遷移は主要カラムのリンクで対応
- ステータス変更は `InlineSelect` で対応
- 上記で操作が完結する場合は操作列自体が不要

**RowActionsMenuが必要なケース:**
- 削除、複製、アーカイブなどインライン編集できない操作がある
- 複数の操作をまとめて提供したい

```typescript
import { RowActionsMenu } from '@/components/common';

const columns: TableColumn<Item>[] = [
  // ... 他のカラム
  {
    id: 'actions',
    label: '',
    width: 50,
    render: (_, row) => (
      <RowActionsMenu
        actions={[
          {
            label: '複製',
            onClick: () => handleDuplicate(row.id),
          },
          {
            label: '削除',
            onClick: () => handleDelete(row.id),
            color: 'error',
          },
        ]}
      />
    ),
  },
];
```

**使用ガイドライン:**
- 詳細画面への遷移は含めない（主要カラムのリンクで対応）
- アイコンは使用しない（テキストのみ）
- 削除等の危険な操作は `color: 'error'` を指定

---

## stopPropagation の必須対応

テーブル行にクリックイベントがある場合、行内のインタラクティブ要素には `stopPropagation` が必須です。

```typescript
// ✅ 正しい: stopPropagationを実装
<Link onClick={(e) => e.stopPropagation()} href="...">...</Link>
<Button onClick={(e) => { e.stopPropagation(); handleAction(); }}>...</Button>
<Checkbox onClick={(e) => e.stopPropagation()} onChange={handleChange} />

// ❌ 誤り: stopPropagationがない
<Link href="...">...</Link>  // 行クリックも発火してしまう
```

**対象要素:**
- リンク（Next.js Link, MUI Link）
- ボタン（Button, IconButton）
- チェックボックス
- セレクトボックス（※ InlineSelect は内蔵済み）
- その他すべてのクリック可能な要素

---

## チェックリスト

AppTable 一覧画面を実装する際：

- [ ] 詳細画面への遷移は主要カラム（名前等）のリンクで対応しているか
- [ ] リンク要素に `stopPropagation` を実装しているか
- [ ] インライン編集には `InlineSelect` を使用しているか
- [ ] インライン編集で楽観的UI更新を実装しているか
- [ ] 削除・複製等が必要な場合のみ `RowActionsMenu` を使用しているか（不要な場合は操作列なしでOK）

---

## 関連規約

- 共通コンポーネント一覧 → [components-index.md](./components-index.md)
- コンポーネント作成パターン → [components-patterns.md](./components-patterns.md)
