---
paths: frontend/src/**/*.tsx, frontend/src/components/**/*.tsx
---

# テキストフィールド設計規約

## コンポーネント選択ガイド

用途とフォーム管理方式に応じて、適切なコンポーネントを選択すること。

| 用途 | フォーム管理 | 推奨コンポーネント |
|------|------------|------------------|
| 一般テキスト入力 | React Hook Form | `FormTextField` |
| 一般テキスト入力 | useState | `SimpleTextField` |
| 金額入力（¥ / 円） | useState | `CurrencyTextField` |
| 金額入力（¥ / 円） | React Hook Form | `FormTextField` + `startAdornment` |

### 判断フローチャート

```
テキスト入力が必要
    │
    ├─ React Hook Form を使用？
    │   ├─ Yes ─────────────────→ FormTextField
    │   │                           └─ 金額入力？ → startAdornment={<span>¥</span>}
    │   │
    │   └─ No（useState管理）
    │       │
    │       ├─ 金額入力？
    │       │   ├─ Yes ─────────→ CurrencyTextField
    │       │   └─ No ──────────→ SimpleTextField
    │       │
    │       └─ 複数行テキスト？
    │           └─ Yes ─────────→ SimpleTextField (multiline)
```

---

## MUI TextField 直接使用の禁止

**新規UIでは MUI `TextField` の直接使用は避け、以下の共通コンポーネントを使用すること**

### 例外として直接使用が許容されるケース

| ケース | 理由 |
|-------|------|
| `register` パターン（RHF） | Controller ベースではないため |
| `Autocomplete` 内部の TextField | MUI Autocomplete の仕様 |
| 郵便番号など特殊な正規化ロジック | カスタム onChange が複雑な場合 |

例外が必要な場合は PR で根拠を明示すること。

---

## FormTextField（React Hook Form統合）

React Hook Form の Controller パターンで使用するテキストフィールド。

```typescript
import { FormTextField } from '@/components/common/forms';

// 基本的な使用方法
<FormTextField
  name="email"
  control={control}
  label="メールアドレス"
  type="email"
  required
  placeholder="example@example.com"
/>

// 日付入力
<FormTextField
  name="birthDate"
  control={control}
  label="生年月日"
  type="date"
/>

// 複数行入力
<FormTextField
  name="description"
  control={control}
  label="説明"
  multiline
  rows={4}
  maxLength={1000}
/>

// 金額入力（RHFの場合）
<FormTextField
  name="price"
  control={control}
  label="価格"
  type="number"
  startAdornment={<span>¥</span>}
/>
```

**Props:**

| Prop | 型 | 必須 | 説明 |
|------|---|-----|------|
| `name` | `Path<T>` | ○ | フィールド名 |
| `control` | `Control<T>` | ○ | RHF control |
| `label` | `string` | ○ | ラベル |
| `type` | `HTMLInputTypeAttribute` | - | 入力タイプ（デフォルト: text） |
| `required` | `boolean` | - | 必須フラグ |
| `disabled` | `boolean` | - | 無効化 |
| `placeholder` | `string` | - | プレースホルダー |
| `size` | `'small' \| 'medium' \| 'large'` | - | サイズ（36/48/56px） |
| `multiline` | `boolean` | - | 複数行入力 |
| `rows` | `number` | - | 複数行の行数 |
| `maxLength` | `number` | - | 最大文字数 |
| `startAdornment` | `ReactNode` | - | 先頭装飾（¥など） |
| `endAdornment` | `ReactNode` | - | 末尾装飾 |
| `helperText` | `string` | - | 補足説明 |
| `error` | `FieldError` | - | バリデーションエラー |
| `rules` | `RegisterOptions` | - | バリデーションルール |

---

## SimpleTextField（非React Hook Form）

useState で値を管理する場合のテキストフィールド。

```typescript
import { SimpleTextField } from '@/components/common/forms';

// 基本的な使用方法
<SimpleTextField
  value={searchKeyword}
  onChange={setSearchKeyword}
  label="検索キーワード"
/>

// 日付入力
<SimpleTextField
  value={selectedDate}
  onChange={setSelectedDate}
  label="日付"
  type="date"
/>

// 複数行入力（文字数表示）
<SimpleTextField
  value={remarks}
  onChange={setRemarks}
  label="備考"
  multiline
  rows={4}
  maxLength={1000}
  helperText={`${remarks.length}/1000文字`}
/>
```

**Props:**

| Prop | 型 | 必須 | 説明 |
|------|---|-----|------|
| `value` | `string \| number` | ○ | 現在の値 |
| `onChange` | `(value: string) => void` | ○ | 値変更コールバック |
| `label` | `string` | ○ | ラベル |
| `type` | `HTMLInputTypeAttribute` | - | 入力タイプ（デフォルト: text） |
| `required` | `boolean` | - | 必須フラグ |
| `disabled` | `boolean` | - | 無効化 |
| `placeholder` | `string` | - | プレースホルダー |
| `size` | `'small' \| 'medium' \| 'large'` | - | サイズ（36/48/56px） |
| `multiline` | `boolean` | - | 複数行入力 |
| `rows` | `number` | - | 複数行の行数 |
| `maxLength` | `number` | - | 最大文字数 |
| `startAdornment` | `ReactNode` | - | 先頭装飾 |
| `endAdornment` | `ReactNode` | - | 末尾装飾 |
| `error` | `boolean` | - | エラー状態 |
| `helperText` | `string` | - | エラーメッセージ/補足 |

---

## CurrencyTextField（金額入力専用）

金額入力に特化したコンポーネント。数値型への自動変換機能付き。

```typescript
import { CurrencyTextField } from '@/components/common/forms';

// ¥ prefix（デフォルト）
<CurrencyTextField
  value={amount}
  onChange={setAmount}
  label="単価"
/>

// 円 suffix
<CurrencyTextField
  value={amount}
  onChange={setAmount}
  label="金額"
  currencyPosition="end"
/>

// 範囲指定
<CurrencyTextField
  value={unitPrice}
  onChange={setUnitPrice}
  label="単価"
  min={0}
  max={10000000}
  step={100}
/>
```

**Props:**

| Prop | 型 | 必須 | 説明 |
|------|---|-----|------|
| `value` | `number \| string` | ○ | 現在の値 |
| `onChange` | `(value: number) => void` | ○ | 値変更コールバック（**数値を返す**） |
| `label` | `string` | - | ラベル |
| `currencyPosition` | `'start' \| 'end'` | - | 通貨記号位置（start: ¥, end: 円） |
| `min` | `number` | - | 最小値（デフォルト: 0） |
| `max` | `number` | - | 最大値 |
| `step` | `number` | - | ステップ値（デフォルト: 1） |
| `required` | `boolean` | - | 必須フラグ |
| `disabled` | `boolean` | - | 無効化 |
| `placeholder` | `string` | - | プレースホルダー（デフォルト: "0"） |
| `size` | `'small' \| 'medium' \| 'large'` | - | サイズ（36/48/56px） |
| `error` | `boolean` | - | エラー状態 |
| `helperText` | `string` | - | エラーメッセージ/補足 |

### CurrencyTextField の特徴

1. **数値変換の自動化**: `onChange` は `number` 型を返す（文字列変換不要）
2. **通貨記号の統一**: `currencyPosition` で ¥ / 円 を切り替え
3. **スピンボタン非表示**: 数値入力のスピンボタンはCSSで非表示
4. **空入力時の挙動**: 空文字や不正値は `0` を返す

---

## 高さの統一規約

**すべてのフォーム入力コンポーネントは以下の高さ規約に従う。**

| size | 高さ | 用途 |
|------|-----|------|
| `small` | 36px | テーブル内、コンパクトなフォーム |
| `medium` | 48px | 標準フォーム（デフォルト） |
| `large` | 56px | 大きなフォーム、アクセシビリティ重視 |

### 対象コンポーネント

この高さ規約は以下のすべてのコンポーネントに適用される：

| コンポーネント | 対応状況 |
|--------------|---------|
| `FormTextField` | ✅ 対応済み |
| `SimpleTextField` | ✅ 対応済み |
| `CurrencyTextField` | ✅ 対応済み |
| `FormSelect` | ✅ 対応済み |
| `SimpleSelect` | ✅ 対応済み |
| `FormAutocomplete` | ✅ 対応済み |
| `FormDatePicker` | ✅ 対応済み |
| `FormTimePicker` | ✅ 対応済み |
| `UserSelectField` | ✅ 対応済み |

### 例外

| コンポーネント | 高さ | 理由 |
|--------------|-----|------|
| `InlineSelect` | 26px | テーブル行内での使用に特化（意図的な例外） |

### 実装パターン

フォーム入力コンポーネントで統一された高さを実現するためのCSSパターン：

```typescript
// TextField系
sx={{
  '& .MuiInputBase-root': {
    height: size === 'small' ? 36 : size === 'large' ? 56 : 48,
  },
}}

// Select系（追加で内部要素の上書きが必要）
sx={{
  '& .MuiInputBase-root': {
    height: size === 'small' ? 36 : size === 'large' ? 56 : 48,
  },
  '& .MuiSelect-select': {
    minHeight: 'unset',
    height: '100%',
    display: 'flex',
    alignItems: 'center',
  },
}}
```

---

## 禁止パターン

```typescript
// ❌ MUI TextField の直接使用
import { TextField } from '@mui/material';
<TextField label="名前" value={name} onChange={(e) => setName(e.target.value)} />

// ❌ SimpleTextField で金額入力（CurrencyTextField を使用）
<SimpleTextField
  value={amount}
  onChange={setAmount}
  type="number"
  startAdornment={<span>¥</span>}
/>

// ❌ onChange で文字列を受け取り手動変換（CurrencyTextField を使用）
<SimpleTextField
  value={String(amount)}
  onChange={(v) => setAmount(parseInt(v, 10) || 0)}
  type="number"
/>
```

---

## 移行ガイド

### MUI TextField → 共通コンポーネント

| 元のパターン | 移行先 |
|-------------|--------|
| `<TextField value={v} onChange={e => set(e.target.value)} />` | `SimpleTextField` |
| `<Controller ... render={...TextField...} />` | `FormTextField` |
| `<TextField type="number" InputProps={{startAdornment: ¥}} />` | `CurrencyTextField` |
| `<TextField {...register('field')} />` | 移行対象外（registerパターン） |

---

## チェックリスト

新しいテキスト入力を追加する際：

- [ ] React Hook Form を使用しているか確認したか
- [ ] 金額入力の場合は `CurrencyTextField` を使用しているか
- [ ] MUI TextField を直接使用していないか
- [ ] 適切な `size` を選択しているか（デフォルト: medium）
- [ ] `required` / `error` / `helperText` を適切に設定しているか
