import React from 'react';
import { Box, CircularProgress, Typography } from '@mui/material';

export interface SectionLoaderProps {
  message?: string;
  size?: 'small' | 'medium' | 'large';
  padding?: number | string;
  minHeight?: number | string;
  centerContent?: boolean;
}

/**
 * セクション内のローディング状態を表示する統一コンポーネント
 * カード内やリスト内など、画面の一部分のローディングに使用
 *
 * @param message - 表示するローディングメッセージ（オプション）
 * @param size - ローディングアイコンのサイズ（デフォルト: 'medium'）
 * @param padding - 内側の余白（デフォルト: 3）
 * @param minHeight - 最小高さ（オプション）
 * @param centerContent - コンテンツを中央揃えにするかどうか（デフォルト: true）
 */
const SectionLoader: React.FC<SectionLoaderProps> = ({
  message,
  size = 'medium',
  padding = 3,
  minHeight,
  centerContent = true,
}) => {
  const getCircularProgressSize = () => {
    switch (size) {
      case 'small':
        return 16;
      case 'medium':
        return 24;
      case 'large':
        return 32;
      default:
        return 24;
    }
  };

  const getTypographyVariant = () => {
    switch (size) {
      case 'small':
        return 'body2' as const;
      case 'medium':
        return 'body1' as const;
      case 'large':
        return 'h6' as const;
      default:
        return 'body1' as const;
    }
  };

  return (
    <Box
      sx={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: centerContent ? 'center' : 'flex-start',
        justifyContent: centerContent ? 'center' : 'flex-start',
        p: padding,
        minHeight,
      }}
    >
      <CircularProgress size={getCircularProgressSize()} color="primary" />
      {message && (
        <Typography
          variant={getTypographyVariant()}
          sx={{
            mt: size === 'small' ? 1 : 2,
            color: 'text.secondary',
            textAlign: centerContent ? 'center' : 'left',
          }}
        >
          {message}
        </Typography>
      )}
    </Box>
  );
};

export default SectionLoader;
