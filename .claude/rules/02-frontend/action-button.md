---
paths: frontend/src/**/*.tsx
---

# ActionButton 統一規約

## 基本原則

**新規UIでは必ず `ActionButton` を使用し、`@mui/material/Button` の直接使用は避けること**

例外が必要な場合は PR で根拠を明示し、`docs/06_standards/button_direct_usage_20250928.json` を更新する。

---

## buttonType 早見表

| buttonType | 用途 | 見た目 |
|------------|------|--------|
| `primary` | 新規作成、編集、保存、送信 | 青背景 |
| `secondary` | 追加、設定、一時保存 | 青枠線 |
| `cancel` | キャンセル、閉じる | グレー枠線 |
| `danger` | 削除確定（ダイアログ内） | 赤背景 |
| `dangerSecondary` | 削除ボタン（ヘッダー等）、却下 | 赤枠線 |
| `success` | 承認確定（ダイアログ内） | 緑背景 |
| `successSecondary` | 承認ボタン（一覧・ドロワー等） | 緑枠線 |
| `ghost` | 戻る、詳細を見る、再読み込み | テキストのみ |
| `tertiary` | 補助リンク風アクション | テキストのみ |

---

## 用途別パターン

### 新規作成ボタン

```tsx
<ActionButton
  buttonType="primary"
  icon={<AddIcon />}
  onClick={handleCreate}
>
  新規作成
</ActionButton>
```

### 編集ボタン

```tsx
<ActionButton
  buttonType="primary"
  icon={<EditIcon />}
  onClick={() => setIsEditing(true)}
>
  編集
</ActionButton>
```

### 保存ボタン（loading対応）

```tsx
<ActionButton
  buttonType="primary"
  type="submit"
  form="form-id"
  icon={<SaveIcon />}
  loading={isPending}
>
  {isPending ? '保存中...' : '保存'}
</ActionButton>
```

### 削除ボタン（ヘッダー等）

```tsx
<ActionButton
  buttonType="dangerSecondary"
  icon={<DeleteIcon />}
  onClick={handleDelete}
  loading={isDeleting}
>
  削除
</ActionButton>
```

### 削除確定ボタン（ダイアログ内）

```tsx
<ActionButton
  buttonType="danger"
  onClick={handleConfirmDelete}
  loading={isDeleting}
>
  削除する
</ActionButton>
```

### キャンセルボタン

```tsx
<ActionButton
  buttonType="cancel"
  onClick={handleCancel}
  disabled={isLoading}
>
  キャンセル
</ActionButton>
```

### 戻るボタン（同一機能内）

```tsx
<ActionButton
  buttonType="ghost"
  icon={<ArrowBackIcon />}
  onClick={() => router.push('/list')}
>
  一覧に戻る
</ActionButton>
```

### 追加ボタン（セクション内）

```tsx
<ActionButton
  buttonType="secondary"
  size="small"
  icon={<AddIcon />}
  onClick={handleAdd}
>
  アサイン追加
</ActionButton>
```

### 再読み込みボタン

```tsx
<ActionButton
  buttonType="ghost"
  size="small"
  onClick={handleRefresh}
>
  再読み込み
</ActionButton>
```

---

## 機能間ナビゲーションボタン

機能をまたぐページ遷移（例：休暇管理 → 付与管理）には、**`NavigationButton` コンポーネント**を使用する。

**特徴:**
- 枠なし、hover時に背景色がつく
- アイコンは矢印（`<` / `>`）のみ自動設定
- `href` と `direction` だけで使えるシンプルなAPI

### 別機能へ遷移するボタン（進む）

```tsx
import { NavigationButton } from '@/components/common';

<NavigationButton href="/manage/business/projects">
  案件管理
</NavigationButton>
```

### 元の機能へ戻るボタン

```tsx
import { NavigationButton } from '@/components/common';

<NavigationButton href="/manage/users" direction="back">
  メンバー一覧に戻る
</NavigationButton>
```

### NavigationButton Props

| prop | 型 | デフォルト | 説明 |
|------|---|----------|------|
| `href` | `string` | 必須 | 遷移先のパス |
| `direction` | `'forward' \| 'back'` | `'forward'` | 遷移方向（矢印の向きを決定） |
| `children` | `ReactNode` | 必須 | ボタンのラベル |
| `disabled` | `boolean` | `false` | 無効化 |

**使い分けの基準:**

| パターン | 用途 | コンポーネント |
|---------|------|--------------|
| 同一機能内の戻る | 一覧・詳細間の移動 | `ActionButton` + `icon={<ArrowBackIcon />}` |
| 機能間ナビゲーション | 別機能への遷移 | `NavigationButton` |

---

## 承認・却下ボタン（ワークフロー）

承認・却下ボタンはセマンティックカラーを使用し、ユーザーの直感に合わせる。
- **承認**: 緑（success系） - 肯定的なアクション
- **却下**: 赤（danger系） - 否定的なアクション

参考: [SAP Fiori Design Guidelines](https://www.sap.com/design-system/fiori-design-web/v1-136/foundations/best-practices/ui-elements/how-to-use-semantic-colors)

### 一覧画面・ドロワーの承認・却下

```tsx
<Stack direction="row" spacing={1}>
  <ActionButton
    buttonType="successSecondary"
    size="small"
    onClick={() => handleApprove(id)}
    loading={isApproving}
  >
    承認
  </ActionButton>
  <ActionButton
    buttonType="dangerSecondary"
    size="small"
    onClick={() => handleReject(id)}
    loading={isRejecting}
  >
    却下
  </ActionButton>
</Stack>
```

### 承認確認ダイアログ

```tsx
<DialogActions>
  <ActionButton buttonType="cancel" onClick={onClose} disabled={loading}>
    キャンセル
  </ActionButton>
  <ActionButton buttonType="success" onClick={onConfirm} loading={loading}>
    承認する
  </ActionButton>
</DialogActions>
```

---

## ダイアログのボタン配置

```tsx
<DialogActions>
  <ActionButton buttonType="cancel" onClick={onClose} disabled={loading}>
    キャンセル
  </ActionButton>
  <ActionButton buttonType="primary" onClick={onConfirm} loading={loading}>
    保存
  </ActionButton>
</DialogActions>
```

削除確認ダイアログの場合:

```tsx
<DialogActions>
  <ActionButton buttonType="cancel" onClick={onClose} disabled={loading}>
    キャンセル
  </ActionButton>
  <ActionButton buttonType="danger" onClick={onConfirm} loading={loading}>
    削除する
  </ActionButton>
</DialogActions>
```

---

## 禁止パターン

```tsx
// MUI Button の直接使用
<Button variant="contained">保存</Button>

// startIcon（icon を使用）
<ActionButton startIcon={<SaveIcon />}>保存</ActionButton>

// variant/color の直接指定（buttonType を使用）
<ActionButton variant="outlined" color="error">削除</ActionButton>
```

---

## MUI Button からの移行

| MUI Button | ActionButton |
|------------|--------------|
| `variant="contained"` | `buttonType="primary"` |
| `variant="outlined"` | `buttonType="secondary"` |
| `variant="outlined" color="error"` | `buttonType="dangerSecondary"` |
| `variant="contained" color="error"` | `buttonType="danger"` |
| `variant="outlined" color="success"` | `buttonType="successSecondary"` |
| `variant="contained" color="success"` | `buttonType="success"` |
| `variant="text"` | `buttonType="ghost"` |
| `startIcon={<Icon />}` | `icon={<Icon />}` |

---

## チェックリスト

新しいボタンを追加する際:

- [ ] `ActionButton` を使用しているか
- [ ] 適切な `buttonType` を選択しているか
- [ ] `startIcon` ではなく `icon` を使用しているか
- [ ] loading 状態を `loading` prop で制御しているか
- [ ] MUI Button を直接使用していないか
