---
paths: frontend/src/**/*.tsx, frontend/src/components/**/*.tsx
---

# セレクトボックス設計規約

## コンポーネント選択ガイド

選択肢の数と用途に応じて、適切なコンポーネントを選択すること。

| 選択肢の数 | 推奨コンポーネント | 理由 |
|-----------|------------------|------|
| **2個** | `FormRadioGroup` / `FormSwitch` | 1クリックで完了、全選択肢が常に見える |
| **3〜5個** | `FormRadioGroup` | スキャンしやすい、クリック数最小化 |
| **5〜15個** | `FormSelect` / `SimpleSelect` | スペース効率とスキャン性のバランス |
| **15個以上** | `FormAutocomplete` | 長いリストは検索機能が必須 |
| **複数選択** | `FormAutocomplete` (multiple) | チェックボックス+検索 |
| **テーブル行内** | `InlineSelect` | 専用の小型デザイン |

---

## コンポーネント一覧

### FormSelect（React Hook Form統合）

```typescript
import { FormSelect } from '@/components/common/forms';

<FormSelect
  name="department"
  control={control}
  label="部署"
  options={[
    { value: 'sales', label: '営業部' },
    { value: 'dev', label: '開発部' },
    { value: 'hr', label: '人事部', disabled: true },
  ]}
  required
  helperText="所属する部署を選択してください"
/>
```

**Props:**
| Prop | 型 | 必須 | 説明 |
|------|---|-----|------|
| `name` | `Path<T>` | ○ | フィールド名 |
| `control` | `Control<T>` | ○ | RHF control |
| `options` | `SelectOption[]` | ○ | 選択肢 |
| `label` | `string` | ○ | ラベル |
| `required` | `boolean` | - | 必須フラグ |
| `disabled` | `boolean` | - | 無効化 |
| `loading` | `boolean` | - | ローディング状態 |
| `placeholder` | `string` | - | プレースホルダー |
| `helperText` | `string` | - | 補足説明 |
| `size` | `'small' \| 'medium'` | - | サイズ（デフォルト: medium） |

---

### SimpleSelect（非React Hook Form）

```typescript
import { SimpleSelect } from '@/components/common/forms';

<SimpleSelect
  value={selectedYear}
  onChange={setSelectedYear}
  options={yearOptions}
  label="年"
/>
```

**Props:**
| Prop | 型 | 必須 | 説明 |
|------|---|-----|------|
| `value` | `T` | ○ | 現在の値 |
| `onChange` | `(value: T) => void` | ○ | 値変更コールバック |
| `options` | `SelectOption<T>[]` | ○ | 選択肢 |
| `label` | `string` | ○ | ラベル |
| `error` | `boolean` | - | エラー状態 |
| `helperText` | `string` | - | エラーメッセージ/補足 |

---

### FormAutocomplete（検索付き/複数選択）

```typescript
import { FormAutocomplete } from '@/components/common/forms';

<FormAutocomplete
  name="client"
  control={control}
  options={clients}
  label="取引先"
  placeholder="取引先を検索..."
  loading={isLoading}
/>
```

**Props:**
| Prop | 型 | 必須 | 説明 |
|------|---|-----|------|
| `name` | `Path<T>` | ○ | フィールド名 |
| `control` | `Control<T>` | ○ | RHF control |
| `options` | `SelectOption[]` | ○ | 選択肢 |
| `label` | `string` | ○ | ラベル |
| `multiple` | `boolean` | - | 複数選択 |
| `loading` | `boolean` | - | ローディング状態 |
| `groupBy` | `(option) => string` | - | グループ化関数 |

---

### InlineSelect（テーブル行内）

```typescript
import { InlineSelect } from '@/components/common';

<InlineSelect
  value={row.status}
  options={STATUS_OPTIONS}
  onChange={(newStatus) => handleStatusChange(row.id, newStatus)}
  aria-label="ステータス"
/>
```

**特徴:**
- 行クリックとの干渉防止（stopPropagation内蔵）
- 小型デザイン（28px高さ）
- 楽観的UI対応（loading状態）

---

## 統一型定義

```typescript
// @/types/forms.ts
interface SelectOption<T = string | number> {
  value: T;
  label: string;
  disabled?: boolean;
  description?: string;  // Autocompleteで使用
  group?: string;        // グループ化用
}
```

**使用例:**

```typescript
import { SelectOption } from '@/types/forms';

const yearOptions: SelectOption<number>[] = [
  { value: 2024, label: '2024年' },
  { value: 2025, label: '2025年' },
];

const statusOptions: SelectOption<string>[] = [
  { value: 'active', label: '有効' },
  { value: 'inactive', label: '無効', disabled: true },
];
```

---

## 禁止事項

### MUI Select の直接使用

```typescript
// ❌ 禁止
import { Select, MenuItem } from '@mui/material';

<FormControl>
  <Select value={value} onChange={handleChange}>
    <MenuItem value="a">A</MenuItem>
  </Select>
</FormControl>

// ✅ 代わりに
import { FormSelect } from '@/components/common/forms';

<FormSelect
  name="field"
  control={control}
  options={[{ value: 'a', label: 'A' }]}
  label="ラベル"
/>
```

### SelectOption型の独自定義

```typescript
// ❌ 禁止: 独自定義
interface MyOption {
  val: string;
  text: string;
}

// ✅ 代わりに
import { SelectOption } from '@/types/forms';
```

### 2択にセレクトボックス使用

```typescript
// ❌ 禁止: 2択にSelect
<FormSelect
  options={[
    { value: 'yes', label: 'はい' },
    { value: 'no', label: 'いいえ' },
  ]}
/>

// ✅ 代わりに
<FormSwitch name="isEnabled" label="有効にする" />
// または
<FormRadioGroup
  name="answer"
  options={[
    { value: 'yes', label: 'はい' },
    { value: 'no', label: 'いいえ' },
  ]}
/>
```

---

## アクセシビリティ要件（WCAG 2.1準拠）

| 要件 | 実装 | WCAG基準 |
|-----|------|----------|
| ラベル関連付け | `labelId` + `InputLabel` | 1.3.1, 4.1.2 |
| エラー関連付け | `FormHelperText` | 3.3.1 |
| 必須表示 | `required` prop | 3.3.2 |
| キーボード操作 | MUI標準対応 | 2.1.1 |
| コントラスト比 | テーマで管理 | 1.4.3 |
| 無効オプション | `disabled: true`（削除しない） | 予測可能性 |

---

## サイズ基準

| サイズ | 高さ | 用途 |
|-------|------|------|
| `small` | 40px | テーブル内、フィルター、コンパクトUI |
| `medium` | 56px | 通常フォーム（デフォルト） |

---

## MenuProps（統一設定）

全セレクトコンポーネントで以下の設定を適用:

```typescript
const UNIFIED_MENU_PROPS = {
  PaperProps: {
    sx: {
      maxHeight: 280,  // 約7項目表示
      mt: 0.5,
    },
  },
};
```

---

## 使用ケース別ガイド

### フォーム内の単一選択

```typescript
// React Hook Formを使用している場合
<FormSelect name="status" control={control} options={options} label="ステータス" />

// React Hook Formを使用していない場合
<SimpleSelect value={status} onChange={setStatus} options={options} label="ステータス" />
```

### 検索可能な選択（15件以上）

```typescript
<FormAutocomplete
  name="client"
  control={control}
  options={clients}
  label="取引先"
  placeholder="検索..."
/>
```

### 複数選択

```typescript
<FormAutocomplete
  name="tags"
  control={control}
  options={tags}
  label="タグ"
  multiple
/>
```

### テーブル行内の選択

```typescript
<InlineSelect
  value={row.status}
  options={statusOptions}
  onChange={(val) => updateStatus(row.id, val)}
  aria-label="ステータス"
  loading={isUpdating}
/>
```

---

## チェックリスト

新しいセレクトボックスを実装する際:

- [ ] 選択肢の数に応じた適切なコンポーネントを選択したか
- [ ] `SelectOption` 型を使用しているか（`@/types/forms`）
- [ ] MUI `Select` を直接使用していないか
- [ ] 2択の場合は `Radio` / `Switch` を検討したか
- [ ] `label` を必ず指定しているか
- [ ] テーブル内では `InlineSelect` を使用しているか
- [ ] 15件以上では `FormAutocomplete` を使用しているか

---

## 関連ドキュメント

- [コンポーネント設計規約](./components.md)
- [フォームエラーハンドリング](./error-handling.md)
- [既知の落とし穴](../05-pitfalls/known-issues.md)
