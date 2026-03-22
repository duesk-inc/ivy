import { type ReactNode, useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import {
  Box,
  Typography,
  Drawer,
  List,
  ListItem,
  ListItemButton,
  ListItemText,
  Avatar,
  Container,
  AppBar,
  Toolbar,
  IconButton,
  Collapse,
  Tooltip,
  useTheme,
  useMediaQuery,
} from '@mui/material';
import MenuIcon from '@mui/icons-material/Menu';
import ChevronLeftIcon from '@mui/icons-material/ChevronLeft';
import ChevronRightIcon from '@mui/icons-material/ChevronRight';
import ExpandLessIcon from '@mui/icons-material/ExpandLess';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import CompareArrowsIcon from '@mui/icons-material/CompareArrows';
import HistoryIcon from '@mui/icons-material/History';
import SettingsIcon from '@mui/icons-material/Settings';
import LogoutIcon from '@mui/icons-material/Logout';
import { useAuth } from '../../context/AuthContext';

interface LayoutProps {
  children: ReactNode;
}

// Monstera準拠の定数
const SIDEBAR_WIDTH = 246;
const SIDEBAR_WIDTH_COLLAPSED = 68;
const SIDEBAR_MARGIN = 12;

// サイドバーグラデーション（アイビーグリーン基調）
const SIDEBAR_BG =
  'linear-gradient(180deg, #2A8A67 0%, #2A8A67 35%, #36B083 70%, #36B083 100%)';

// ダーク背景上のスタイルトークン
const D = {
  TEXT: 'rgba(255, 255, 255, 0.95)',
  TEXT_SUB: 'rgba(255, 255, 255, 0.7)',
  HOVER: 'rgba(255, 255, 255, 0.12)',
  ACTIVE: 'rgba(255, 255, 255, 0.18)',
  DIVIDER: 'rgba(255, 255, 255, 0.15)',
  GLASS_BG: 'rgba(255, 255, 255, 0.1)',
  GLASS_BORDER: '1px solid rgba(255, 255, 255, 0.3)',
  GLASS_BLUR: 'blur(12px)',
};

export default function Layout({ children }: LayoutProps) {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [collapsed, setCollapsed] = useState(false);
  const [userMenuOpen, setUserMenuOpen] = useState(false);

  const currentWidth = collapsed ? SIDEBAR_WIDTH_COLLAPSED : SIDEBAR_WIDTH;

  const menuItems = [
    { title: '個別マッチング', path: '/', icon: <CompareArrowsIcon /> },
    { title: '履歴', path: '/history', icon: <HistoryIcon /> },
    ...(user?.role === 'admin'
      ? [{ title: '設定', path: '/settings', icon: <SettingsIcon /> }]
      : []),
  ];

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  // ────────────────────────────────────────
  // メニュー項目の描画（Monstera BaseSidebar準拠）
  // ────────────────────────────────────────
  const renderMenuItem = (item: typeof menuItems[0], forMobile: boolean) => {
    const isActive = location.pathname === item.path;

    const button = (
      <ListItem disablePadding sx={{ display: 'block', px: forMobile ? 0 : 1 }}>
        <ListItemButton
          selected={isActive}
          onClick={() => {
            navigate(item.path);
            if (forMobile) setDrawerOpen(false);
          }}
          sx={{
            minHeight: forMobile ? 56 : 48,
            justifyContent: collapsed && !forMobile ? 'center' : 'initial',
            px: collapsed && !forMobile ? 1 : 2.5,
            borderRadius: forMobile ? 0 : 1,
            mx: forMobile ? 0 : 1,
            my: forMobile ? 0 : 0.5,
            bgcolor: isActive ? (forMobile ? 'primary.50' : D.ACTIVE) : 'transparent',
            color: isActive ? (forMobile ? 'primary.main' : D.TEXT) : (forMobile ? 'text.primary' : D.TEXT),
            '&:hover': {
              bgcolor: isActive
                ? (forMobile ? 'primary.100' : 'rgba(255,255,255,0.25)')
                : (forMobile ? 'action.hover' : D.HOVER),
            },
            ...(forMobile && isActive && {
              borderLeft: '3px solid',
              borderLeftColor: 'primary.main',
            }),
            ...(forMobile && {
              borderBottom: '1px solid',
              borderBottomColor: 'divider',
            }),
            transition: 'all 0.15s ease',
          }}
        >
          {/* アイコン */}
          {item.icon && collapsed && !forMobile && (
            <Box sx={{ display: 'flex', justifyContent: 'center', width: '100%', color: isActive ? D.TEXT : D.TEXT_SUB }}>
              {item.icon}
            </Box>
          )}
          {item.icon && (!collapsed || forMobile) && (
            <Box sx={{ mr: forMobile ? 2.5 : 1.5, display: 'flex', alignItems: 'center', color: isActive ? (forMobile ? 'primary.main' : D.TEXT) : (forMobile ? 'text.secondary' : D.TEXT_SUB) }}>
              {item.icon}
            </Box>
          )}
          {(!collapsed || forMobile) && (
            <ListItemText
              primary={item.title}
              sx={{
                '& .MuiTypography-root': {
                  fontSize: forMobile ? '1.05rem' : '0.875rem',
                  fontWeight: isActive ? 600 : 400,
                },
              }}
            />
          )}
        </ListItemButton>
      </ListItem>
    );

    return collapsed && !forMobile ? (
      <Tooltip key={item.path} title={item.title} placement="right" arrow>
        {button}
      </Tooltip>
    ) : (
      <Box key={item.path}>{button}</Box>
    );
  };

  // ────────────────────────────────────────
  // ユーザーメニューセクション（Monstera UserMenuSection準拠）
  // ────────────────────────────────────────
  const renderUserSection = (forMobile: boolean) => (
    <Box sx={{ borderTop: '1px solid', borderColor: forMobile ? 'divider' : D.DIVIDER }}>
      <ListItem disablePadding>
        <ListItemButton
          onClick={collapsed && !forMobile ? undefined : () => setUserMenuOpen(!userMenuOpen)}
          sx={{
            minHeight: forMobile ? 56 : 64,
            px: collapsed && !forMobile ? 1 : forMobile ? 3 : 2.5,
            justifyContent: collapsed && !forMobile ? 'center' : 'space-between',
            bgcolor: 'transparent',
            '&:hover': { bgcolor: forMobile ? 'action.hover' : 'rgba(255,255,255,0.15)' },
          }}
        >
          {collapsed && !forMobile ? (
            <Tooltip title={user?.name || ''} placement="right">
              <Avatar sx={{ width: 32, height: 32, bgcolor: 'rgba(255,255,255,0.2)', fontSize: '0.875rem' }}>
                {user?.name?.charAt(0) || 'U'}
              </Avatar>
            </Tooltip>
          ) : (
            <>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                <Avatar sx={{ width: 32, height: 32, bgcolor: forMobile ? 'primary.main' : 'rgba(255,255,255,0.2)', fontSize: '0.875rem' }}>
                  {user?.name?.charAt(0) || 'U'}
                </Avatar>
                <Box>
                  <Typography variant="body2" fontWeight={500} sx={{ fontSize: forMobile ? '1rem' : '0.8rem', color: forMobile ? 'text.primary' : D.TEXT }}>
                    {user?.name}
                  </Typography>
                  <Typography variant="caption" sx={{ fontSize: forMobile ? '0.85rem' : '0.7rem', color: forMobile ? 'text.secondary' : D.TEXT_SUB }}>
                    {user?.role}
                  </Typography>
                </Box>
              </Box>
              <Box sx={{ color: forMobile ? 'text.secondary' : D.TEXT_SUB }}>
                {userMenuOpen ? <ExpandLessIcon /> : <ExpandMoreIcon />}
              </Box>
            </>
          )}
        </ListItemButton>
      </ListItem>

      {(!collapsed || forMobile) && (
        <Collapse in={userMenuOpen} timeout="auto" unmountOnExit>
          <List component="div" disablePadding>
            <ListItem disablePadding>
              <ListItemButton
                onClick={handleLogout}
                sx={{
                  pl: 3,
                  py: forMobile ? 1.5 : 1,
                  bgcolor: 'transparent',
                  '&:hover': { bgcolor: forMobile ? 'action.hover' : 'rgba(255,255,255,0.15)' },
                }}
              >
                <LogoutIcon fontSize="small" sx={{ mr: forMobile ? 2 : 1.5, color: '#ef5350 !important' }} />
                <Typography variant="body2" fontWeight={500} sx={{ color: '#ef5350 !important', fontSize: forMobile ? '0.95rem' : '0.8rem' }}>
                  ログアウト
                </Typography>
              </ListItemButton>
            </ListItem>
          </List>
        </Collapse>
      )}
    </Box>
  );

  // ────────────────────────────────────────
  // サイドバーコンテンツ
  // ────────────────────────────────────────
  const sidebarContent = (forMobile: boolean) => (
    <Box
      sx={{
        display: 'flex',
        flexDirection: 'column',
        height: '100%',
        background: forMobile ? undefined : SIDEBAR_BG,
        backdropFilter: forMobile ? undefined : D.GLASS_BLUR,
        WebkitBackdropFilter: forMobile ? undefined : D.GLASS_BLUR,
        borderRadius: forMobile ? 0 : '16px',
        border: forMobile ? undefined : '1px solid rgba(255,255,255,0.1)',
        boxShadow: forMobile ? undefined : 'inset 0 1px 0 rgba(255,255,255,0.05), 0 4px 24px rgba(0,0,0,0.06)',
        overflow: 'hidden',
        position: 'relative',
        color: forMobile ? 'text.primary' : D.TEXT,
      }}
    >

      {/* ヘッダー（Monstera準拠: 左寄せ、py:2, px:2.5） */}
      {!forMobile && (
        <Box sx={{ py: 2, px: collapsed ? 0 : 2.5, display: 'flex', alignItems: 'center', justifyContent: collapsed ? 'center' : 'flex-start', gap: 1.5, position: 'relative', zIndex: 1 }}>
          {!collapsed && (
            <Typography variant="h6" fontWeight={700} sx={{ whiteSpace: 'nowrap' }}>
              Ivy
            </Typography>
          )}
          {collapsed && (
            <Typography variant="h6" fontWeight={700}>I</Typography>
          )}
        </Box>
      )}

      {/* メニューアイテム（Monstera準拠: py:0） */}
      <List role="menu" sx={{ flexGrow: 1, py: 0, overflow: 'auto', position: 'relative', zIndex: 1 }}>
        {menuItems.map((item) => renderMenuItem(item, forMobile))}
      </List>

      {/* ユーザーセクション（Monstera UserMenuSection準拠） */}
      <Box sx={{ position: 'relative', zIndex: 1 }}>
        {renderUserSection(forMobile)}
      </Box>

      {/* フッター */}
      {!collapsed && !forMobile && (
        <Typography variant="caption" sx={{ textAlign: 'center', py: 1, color: D.TEXT_SUB, fontSize: '0.65rem', position: 'relative', zIndex: 1 }}>
          © 2026 Ivy
        </Typography>
      )}
    </Box>
  );

  // ────────────────────────────────────────
  // モバイルレイアウト
  // ────────────────────────────────────────
  if (isMobile) {
    return (
      <Box sx={{ minHeight: '100vh' }}>
        <AppBar position="fixed" elevation={0} sx={{ bgcolor: 'rgba(255,255,255,0.8)', backdropFilter: 'blur(20px)', WebkitBackdropFilter: 'blur(20px)', borderBottom: '1px solid', borderColor: 'divider' }}>
          <Toolbar sx={{ minHeight: 64 }}>
            <IconButton onClick={() => setDrawerOpen(true)} sx={{ color: 'text.primary' }} aria-label="メニュー">
              <MenuIcon />
            </IconButton>
            <Typography variant="h6" sx={{ flex: 1, textAlign: 'center', fontWeight: 700, color: 'text.primary' }}>
              Ivy
            </Typography>
            <Avatar sx={{ width: 32, height: 32, bgcolor: 'primary.main', fontSize: '0.8rem' }}>
              {user?.name?.charAt(0) || 'U'}
            </Avatar>
          </Toolbar>
        </AppBar>
        <Drawer open={drawerOpen} onClose={() => setDrawerOpen(false)} sx={{ '& .MuiDrawer-paper': { width: SIDEBAR_WIDTH, border: 'none' } }}>
          {sidebarContent(true)}
        </Drawer>
        <Box component="main" sx={{ pt: '80px', px: 2, pb: 3 }}>
          <Container maxWidth="lg">{children}</Container>
        </Box>
      </Box>
    );
  }

  // ────────────────────────────────────────
  // デスクトップレイアウト
  // ────────────────────────────────────────
  return (
    <Box sx={{ display: 'flex', minHeight: '100vh' }}>
      <Box component="nav" sx={{ width: currentWidth + SIDEBAR_MARGIN * 2, flexShrink: 0, transition: 'width 0.2s ease' }}>
        <Box sx={{
          position: 'fixed',
          top: SIDEBAR_MARGIN,
          left: SIDEBAR_MARGIN,
          width: currentWidth,
          height: `calc(100vh - ${SIDEBAR_MARGIN * 2}px)`,
          transition: 'width 0.2s ease',
        }}>
          {/* サイドバー本体（スクロール領域） */}
          <Box sx={{
            width: '100%',
            height: '100%',
            overflow: 'auto',
            '&::-webkit-scrollbar': { display: 'none' },
            msOverflowStyle: 'none',
            scrollbarWidth: 'none',
          }}>
            {sidebarContent(false)}
          </Box>

          {/* 開閉ボタン（Monstera準拠: absolute, top:60, right:-24） */}
          <IconButton
            onClick={() => setCollapsed(!collapsed)}
            sx={{
              position: 'absolute',
              top: 60,
              right: -24,
              background: '#2A8A67',
              backdropFilter: D.GLASS_BLUR,
              WebkitBackdropFilter: D.GLASS_BLUR,
              borderTop: '1px solid rgba(255,255,255,0.1)',
              borderRight: '1px solid rgba(255,255,255,0.1)',
              borderBottom: '1px solid rgba(255,255,255,0.1)',
              borderLeft: 'none',
              borderRadius: '0 8px 8px 0',
              width: 24,
              height: 40,
              padding: 0,
              boxShadow: '0 4px 24px rgba(0,0,0,0.06)',
              '&:hover': {
                background: '#2A8A67',
                filter: 'brightness(1.15)',
                '& svg': { transform: 'scale(1.1)' },
              },
              transition: 'all 0.2s',
              zIndex: 1200,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            {collapsed
              ? <ChevronRightIcon sx={{ fontSize: 16, color: D.TEXT, transition: 'transform 0.2s' }} />
              : <ChevronLeftIcon sx={{ fontSize: 16, color: D.TEXT, transition: 'transform 0.2s' }} />
            }
          </IconButton>
        </Box>
      </Box>

      <Box component="main" sx={{ flex: 1, p: 3, minWidth: 0 }}>
        <Container maxWidth="lg">{children}</Container>
      </Box>
    </Box>
  );
}
