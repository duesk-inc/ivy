/**
 * フォーム関連の共通型定義
 */

/**
 * セレクトボックス共通オプション型
 * 全てのSelect系コンポーネントで使用
 */
export interface SelectOption<T = string | number> {
  /** 値 */
  value: T;
  /** 表示ラベル */
  label: string;
  /** 無効化フラグ */
  disabled?: boolean;
  /** 補足説明 */
  description?: string;
  /** グループ名（optgroup相当） */
  group?: string;
}
