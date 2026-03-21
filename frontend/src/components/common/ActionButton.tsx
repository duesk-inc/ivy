import React from 'react';
import {
  Button,
  ButtonProps,
  CircularProgress,
  useTheme,
  useMediaQuery,
} from '@mui/material';
import type { SxProps, Theme } from '@mui/material/styles';

// ボタンのバリエーション
export type ActionButtonVariant =
  | 'primary'
  | 'secondary'
  | 'tertiary'
  | 'cancel'
  | 'danger'
  | 'dangerSecondary'
  | 'success'
  | 'successSecondary'
  | 'ghost';

// ボタンサイズ（レスポンシブ対応）
export type ActionButtonSize = 'small' | 'medium' | 'large';
export type ResponsiveSize =
  | ActionButtonSize
  | {
      xs?: ActionButtonSize;
      sm?: ActionButtonSize;
      md?: ActionButtonSize;
      lg?: ActionButtonSize;
      xl?: ActionButtonSize;
    };

export interface ActionButtonProps
  extends Omit<ButtonProps, 'variant' | 'size' | 'fullWidth'> {
  buttonType?: ActionButtonVariant;
  /** ボタンのtype属性（デフォルト: "button"、フォーム送信時のみ"submit"を指定） */
  type?: 'button' | 'submit' | 'reset';
  icon?: React.ReactNode;
  endIcon?: React.ReactNode;
  loading?: boolean;
  loadingIndicatorSize?: number;
  loadingPosition?: 'start' | 'center';
  hideLabelWhenLoading?: boolean;
  size?: ResponsiveSize;
  fullWidth?:
    | boolean
    | {
        xs?: boolean;
        sm?: boolean;
        md?: boolean;
        lg?: boolean;
        xl?: boolean;
      };
}

// ═══════════════════════════════════════════════════════
// バリアントプリセット（MUI標準のvariant/colorマッピング）
// ═══════════════════════════════════════════════════════

type VariantPreset = {
  variant: ButtonProps['variant'];
  color: ButtonProps['color'];
};

const VARIANT_PRESETS: Record<ActionButtonVariant, VariantPreset> = {
  primary: { variant: 'contained', color: 'primary' },
  secondary: { variant: 'outlined', color: 'primary' },
  tertiary: { variant: 'text', color: 'primary' },
  cancel: { variant: 'outlined', color: 'inherit' },
  danger: { variant: 'contained', color: 'error' },
  dangerSecondary: { variant: 'outlined', color: 'error' },
  success: { variant: 'contained', color: 'success' },
  successSecondary: { variant: 'outlined', color: 'success' },
  ghost: { variant: 'text', color: 'inherit' },
};

const ensureSxArray = (base: SxProps<Theme> | undefined): SxProps<Theme>[] => {
  if (!base) {
    return [];
  }
  return Array.isArray(base) ? base : [base];
};

/**
 * アプリケーション全体で使用される統一されたアクションボタンコンポーネント
 *
 * @param buttonType - ボタンの種類（primary, secondary, tertiary, cancel, danger, dangerSecondary, success, successSecondary, ghost）
 * @param icon - ボタンのアイコン
 * @param loading - ローディング状態
 * @param size - ボタンのサイズ（small, medium, large）またはレスポンシブサイズオブジェクト
 * @param fullWidth - 幅を100%にするかどうか（ブール値またはレスポンシブオブジェクト）
 * @param children - ボタンのテキスト
 */
const ActionButton: React.FC<ActionButtonProps> = ({
  buttonType = 'primary',
  type = 'button',
  icon,
  endIcon,
  loading = false,
  loadingIndicatorSize = 16,
  loadingPosition = 'start',
  hideLabelWhenLoading = false,
  size = 'medium',
  fullWidth = false,
  children,
  sx: sxProp,
  disabled: disabledProp,
  ...restProps
}) => {
  const theme = useTheme();
  const isXs = useMediaQuery(theme.breakpoints.only('xs'));
  const isSm = useMediaQuery(theme.breakpoints.only('sm'));
  const isMd = useMediaQuery(theme.breakpoints.only('md'));
  const isLg = useMediaQuery(theme.breakpoints.only('lg'));
  const isXl = useMediaQuery(theme.breakpoints.only('xl'));

  const preset = VARIANT_PRESETS[buttonType] ?? VARIANT_PRESETS.primary;

  // レスポンシブサイズの解決
  const getResponsiveSize = (): ActionButtonSize => {
    if (typeof size === 'string') {
      return size;
    }

    const responsiveSize = size as {
      xs?: ActionButtonSize;
      sm?: ActionButtonSize;
      md?: ActionButtonSize;
      lg?: ActionButtonSize;
      xl?: ActionButtonSize;
    };

    if (isXl && responsiveSize.xl) return responsiveSize.xl;
    if (isLg && responsiveSize.lg) return responsiveSize.lg;
    if (isMd && responsiveSize.md) return responsiveSize.md;
    if (isSm && responsiveSize.sm) return responsiveSize.sm;
    if (isXs && responsiveSize.xs) return responsiveSize.xs;

    return 'medium';
  };

  // レスポンシブfullWidthの解決
  const getResponsiveFullWidth = (): boolean => {
    if (typeof fullWidth === 'boolean') {
      return fullWidth;
    }

    const responsiveFullWidth = fullWidth as {
      xs?: boolean;
      sm?: boolean;
      md?: boolean;
      lg?: boolean;
      xl?: boolean;
    };

    if (isXl && responsiveFullWidth.xl !== undefined)
      return responsiveFullWidth.xl;
    if (isLg && responsiveFullWidth.lg !== undefined)
      return responsiveFullWidth.lg;
    if (isMd && responsiveFullWidth.md !== undefined)
      return responsiveFullWidth.md;
    if (isSm && responsiveFullWidth.sm !== undefined)
      return responsiveFullWidth.sm;
    if (isXs && responsiveFullWidth.xs !== undefined)
      return responsiveFullWidth.xs;

    return false;
  };

  // ローディング時のアイコン表示
  const displayIcon =
    loading && loadingPosition === 'start' ? (
      <CircularProgress size={loadingIndicatorSize} />
    ) : (
      icon
    );

  const sxArray: SxProps<Theme> = [
    {
      position: 'relative',
      textTransform: 'none',
      fontWeight: 600,
    },
    ...ensureSxArray(sxProp),
  ] as SxProps<Theme>;

  const disabled = loading || disabledProp;

  return (
    <Button
      {...restProps}
      type={type}
      variant={preset.variant}
      color={preset.color}
      size={getResponsiveSize()}
      startIcon={displayIcon}
      endIcon={endIcon}
      disabled={disabled}
      fullWidth={getResponsiveFullWidth()}
      sx={sxArray}
    >
      {/* ラベル */}
      <span style={{ opacity: loading && hideLabelWhenLoading ? 0 : 1 }}>
        {children}
      </span>

      {/* センター配置のローディングインジケータ */}
      {loading && loadingPosition === 'center' && (
        <span
          aria-hidden
          style={{
            position: 'absolute',
            inset: 0,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }}
        >
          <CircularProgress size={loadingIndicatorSize} />
        </span>
      )}
    </Button>
  );
};

export default ActionButton;
