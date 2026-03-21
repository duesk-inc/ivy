import { createTheme, alpha } from '@mui/material';

// Ivy Green（アイビーグリーン基調）
const PRIMARY_COLOR = {
  light: '#5A9A4A',
  main: '#3F7036',
  dark: '#2D5227',
};

// Text colors (Monstera monochrome style)
const TEXT_COLOR = {
  primary: '#1E293B',
  secondary: '#64748B',
};

// Border radius tokens (Monstera-aligned)
const BORDER_RADIUS = {
  SM: 4,
  MD: 8,
  LG: 12,
  XL: 12,
  FULL: 9999,
};

// Glass effect styles
const GLASS = {
  CARD_BG: 'rgba(255, 255, 255, 0.85)',
  CARD_BACKDROP: 'blur(20px) saturate(180%)',
  CARD_BORDER: '1px solid rgba(255, 255, 255, 0.3)',
  CARD_SHADOW: '0 4px 16px rgba(0, 0, 0, 0.06)',
  CARD_SHADOW_HOVER: '0 8px 24px rgba(0, 0, 0, 0.1)',
  CARD_INNER_HIGHLIGHT: 'inset 0 1px 0 rgba(255, 255, 255, 0.5)',
};

// App background (Monstera monochrome palette)
const APP_BACKGROUND = 'linear-gradient(135deg, #f6f6f6 0%, #f4f4f4 35%, #f3f3f3 70%, #f5f5f5 100%)';

const theme = createTheme({
  palette: {
    primary: {
      light: PRIMARY_COLOR.light,
      main: PRIMARY_COLOR.main,
      dark: PRIMARY_COLOR.dark,
      contrastText: '#FFFFFF',
    },
    secondary: {
      light: '#B2F5EA',
      main: '#0EA5A9',
      dark: '#0D8A8D',
      contrastText: '#FFFFFF',
    },
    error: {
      light: '#ef5350',
      main: '#d32f2f',
      dark: '#c62828',
    },
    background: {
      default: '#f6f6f6',
      paper: '#ffffff',
    },
    text: {
      primary: TEXT_COLOR.primary,
      secondary: TEXT_COLOR.secondary,
    },
    divider: 'rgba(100, 116, 139, 0.12)',
  },
  typography: {
    fontSize: 14,
    fontFamily: [
      '"Noto Sans JP"',
      'Hiragino Sans',
      'Hiragino Kaku Gothic ProN',
      'Meiryo',
      'sans-serif',
    ].join(','),
    h1: { fontSize: '1.75rem', fontWeight: 700, lineHeight: 1.3 },
    h2: { fontSize: '1.5rem', fontWeight: 700, lineHeight: 1.35 },
    h3: { fontSize: '1.25rem', fontWeight: 600, lineHeight: 1.4 },
    h4: { fontSize: '1.125rem', fontWeight: 600, lineHeight: 1.4 },
    h5: { fontSize: '1rem', fontWeight: 600, lineHeight: 1.5 },
    h6: { fontSize: '0.875rem', fontWeight: 600, lineHeight: 1.5 },
    body1: { fontSize: '0.875rem', lineHeight: 1.6 },
    body2: { fontSize: '0.8125rem', lineHeight: 1.5 },
  },
  shape: {
    borderRadius: BORDER_RADIUS.MD,
  },
  components: {
    MuiCssBaseline: {
      styleOverrides: {
        body: {
          background: APP_BACKGROUND,
          backgroundAttachment: 'fixed',
          minHeight: '100vh',
          colorScheme: 'light',
        },
        'a, button, [role="button"], .clickable': {
          cursor: 'pointer',
        },
      },
    },
    MuiButton: {
      styleOverrides: {
        root: {
          borderRadius: BORDER_RADIUS.XL,
          textTransform: 'none' as const,
          fontWeight: 600,
          boxShadow: 'none',
          transition: 'all 0.2s ease-in-out',
        },
        containedPrimary: {
          '&:hover': {
            boxShadow: '0 4px 12px rgba(30, 41, 59, 0.25)',
            backgroundColor: '#334155',
          },
        },
        outlinedPrimary: {
          borderColor: alpha(PRIMARY_COLOR.main, 0.5),
          '&:hover': {
            borderColor: PRIMARY_COLOR.main,
            backgroundColor: alpha(PRIMARY_COLOR.main, 0.04),
          },
        },
      },
    },
    MuiCard: {
      styleOverrides: {
        root: {
          borderRadius: BORDER_RADIUS.LG,
          backgroundColor: GLASS.CARD_BG,
          backdropFilter: GLASS.CARD_BACKDROP,
          WebkitBackdropFilter: GLASS.CARD_BACKDROP,
          boxShadow: `${GLASS.CARD_SHADOW}, ${GLASS.CARD_INNER_HIGHLIGHT}`,
          border: GLASS.CARD_BORDER,
          transition: 'box-shadow 0.2s ease-in-out, transform 0.2s ease-in-out',
          '&:hover': {
            boxShadow: `${GLASS.CARD_SHADOW_HOVER}, ${GLASS.CARD_INNER_HIGHLIGHT}`,
          },
        },
      },
    },
    MuiPaper: {
      styleOverrides: {
        root: {
          borderRadius: BORDER_RADIUS.MD,
          backgroundColor: GLASS.CARD_BG,
          backdropFilter: GLASS.CARD_BACKDROP,
          WebkitBackdropFilter: GLASS.CARD_BACKDROP,
          border: GLASS.CARD_BORDER,
        },
        elevation1: {
          boxShadow: `${GLASS.CARD_SHADOW}, ${GLASS.CARD_INNER_HIGHLIGHT}`,
        },
        elevation2: {
          boxShadow: `${GLASS.CARD_SHADOW_HOVER}, ${GLASS.CARD_INNER_HIGHLIGHT}`,
        },
      },
    },
    MuiOutlinedInput: {
      styleOverrides: {
        root: {
          borderRadius: BORDER_RADIUS.MD,
          transition: 'box-shadow 0.2s ease-in-out',
          '&.Mui-focused .MuiOutlinedInput-notchedOutline': {
            borderColor: PRIMARY_COLOR.main,
            borderWidth: 2,
          },
          '&.Mui-focused': {
            boxShadow: `0 0 0 3px ${alpha(PRIMARY_COLOR.main, 0.15)}`,
          },
        },
        notchedOutline: {
          transition: 'border-color 0.2s ease-in-out',
        },
      },
    },
    MuiTextField: {
      styleOverrides: {
        root: {
          '& .MuiOutlinedInput-root.Mui-focused .MuiOutlinedInput-notchedOutline': {
            borderColor: PRIMARY_COLOR.main,
          },
          '& .MuiInputLabel-root.Mui-focused': {
            color: PRIMARY_COLOR.main,
          },
        },
      },
    },
    MuiAppBar: {
      styleOverrides: {
        root: {
          boxShadow: 'none',
        },
      },
    },
    MuiListItemButton: {
      styleOverrides: {
        root: {
          borderRadius: BORDER_RADIUS.XL,
          margin: '4px 8px',
          transition: 'background-color 0.15s ease-in-out',
        },
      },
    },
    MuiChip: {
      styleOverrides: {
        root: {
          borderRadius: BORDER_RADIUS.FULL,
          fontWeight: 600,
        },
      },
    },
    MuiAlert: {
      styleOverrides: {
        root: {
          borderRadius: BORDER_RADIUS.LG,
          border: 'none',
          backdropFilter: 'blur(8px)',
          WebkitBackdropFilter: 'blur(8px)',
        },
      },
    },
    MuiTooltip: {
      styleOverrides: {
        tooltip: {
          backgroundColor: 'rgba(30, 41, 59, 0.9)',
          backdropFilter: 'blur(8px)',
          WebkitBackdropFilter: 'blur(8px)',
          borderRadius: BORDER_RADIUS.MD,
        },
      },
    },
    MuiTableCell: {
      styleOverrides: {
        head: {
          backgroundColor: '#fafafa',
          fontWeight: 600,
        },
      },
    },
    MuiTableContainer: {
      styleOverrides: {
        root: {
          borderRadius: BORDER_RADIUS.MD,
        },
      },
    },
    MuiDrawer: {
      styleOverrides: {
        paper: {
          borderRight: 'none',
        },
      },
    },
    MuiIconButton: {
      styleOverrides: {
        root: {
          '&:hover': {
            backgroundColor: alpha(PRIMARY_COLOR.main, 0.04),
          },
        },
      },
    },
    MuiLink: {
      styleOverrides: {
        root: {
          color: PRIMARY_COLOR.main,
          textDecoration: 'none',
          fontWeight: 500,
          transition: 'color 0.15s ease-in-out',
          '&:hover': {
            color: PRIMARY_COLOR.dark,
          },
        },
      },
    },
  },
});

export default theme;
