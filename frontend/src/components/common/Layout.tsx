import { type ReactNode, useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import {
  Box,
  Typography,
  Drawer,
  List,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  Avatar,
  Container,
  AppBar,
  Toolbar,
  IconButton,
  Divider,
  Tooltip,
  useTheme,
  useMediaQuery,
} from '@mui/material';
import MenuIcon from '@mui/icons-material/Menu';
import ChevronLeftIcon from '@mui/icons-material/ChevronLeft';
import ChevronRightIcon from '@mui/icons-material/ChevronRight';
import CompareArrowsIcon from '@mui/icons-material/CompareArrows';
import HistoryIcon from '@mui/icons-material/History';
import SettingsIcon from '@mui/icons-material/Settings';
import LogoutIcon from '@mui/icons-material/Logout';
import { useAuth } from '../../context/AuthContext';

interface LayoutProps {
  children: ReactNode;
}

const SIDEBAR_WIDTH = 246;
const SIDEBAR_WIDTH_COLLAPSED = 68;
const SIDEBAR_MARGIN = 12;

const SIDEBAR_GRADIENT =
  'linear-gradient(180deg, #1a1a1a 0%, #1a1a1a 35%, #3a3a3a 70%, #3a3a3a 100%)';

const GLASS_OVERLAY = {
  background: 'rgba(255, 255, 255, 0.1)',
  backdropFilter: 'blur(12px)',
  border: '1px solid rgba(255, 255, 255, 0.3)',
};

const DARK = {
  textPrimary: 'rgba(255, 255, 255, 0.95)',
  textSecondary: 'rgba(255, 255, 255, 0.7)',
  bgHover: 'rgba(255, 255, 255, 0.12)',
  bgActive: 'rgba(255, 255, 255, 0.18)',
  divider: 'rgba(255, 255, 255, 0.15)',
};

export default function Layout({ children }: LayoutProps) {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [collapsed, setCollapsed] = useState(false);

  const currentWidth = collapsed ? SIDEBAR_WIDTH_COLLAPSED : SIDEBAR_WIDTH;

  const menuItems = [
    { title: 'マッチング', path: '/', icon: <CompareArrowsIcon /> },
    { title: '履歴', path: '/history', icon: <HistoryIcon /> },
    ...(user?.role === 'admin'
      ? [{ title: '設定', path: '/settings', icon: <SettingsIcon /> }]
      : []),
  ];

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  const sidebarContent = (forMobile: boolean) => (
    <Box
      sx={{
        display: 'flex',
        flexDirection: 'column',
        height: '100%',
        background: SIDEBAR_GRADIENT,
        borderRadius: forMobile ? 0 : '16px',
        overflow: 'hidden',
        position: 'relative',
      }}
    >
      {/* Glass overlay */}
      {!forMobile && (
        <Box
          sx={{
            position: 'absolute',
            inset: 0,
            ...GLASS_OVERLAY,
            borderRadius: '16px',
            pointerEvents: 'none',
          }}
        />
      )}

      {/* Logo */}
      <Box
        sx={{
          p: collapsed && !forMobile ? 1.5 : 3,
          textAlign: 'center',
          position: 'relative',
          zIndex: 1,
        }}
      >
        <Typography
          variant="h6"
          sx={{ color: 'white', fontWeight: 700, letterSpacing: 1 }}
        >
          {collapsed && !forMobile ? 'I' : 'Ivy'}
        </Typography>
        {(!collapsed || forMobile) && (
          <Typography variant="caption" sx={{ color: 'rgba(255,255,255,0.6)' }}>
            SES マッチングツール
          </Typography>
        )}
      </Box>
      <Divider sx={{ borderColor: DARK.divider, position: 'relative', zIndex: 1 }} />

      {/* Menu Items */}
      <List sx={{ flex: 1, px: collapsed && !forMobile ? 0.5 : 1, py: 2, position: 'relative', zIndex: 1 }}>
        {menuItems.map((item) => {
          const isActive = location.pathname === item.path;
          const button = (
            <ListItemButton
              key={item.path}
              selected={isActive}
              onClick={() => {
                navigate(item.path);
                if (forMobile) setDrawerOpen(false);
              }}
              sx={{
                borderRadius: '4px',
                mx: collapsed && !forMobile ? 0.5 : 1,
                my: 0.5,
                minHeight: forMobile ? 56 : 48,
                justifyContent: collapsed && !forMobile ? 'center' : 'initial',
                px: collapsed && !forMobile ? 1 : 2.5,
                color: DARK.textSecondary,
                '&.Mui-selected': {
                  bgcolor: DARK.bgActive,
                  color: DARK.textPrimary,
                  '&:hover': { bgcolor: 'rgba(255,255,255,0.22)' },
                },
                '&:hover': { bgcolor: DARK.bgHover },
                transition: 'all 0.15s ease',
              }}
            >
              <ListItemIcon
                sx={{
                  color: 'inherit',
                  minWidth: collapsed && !forMobile ? 0 : 36,
                  justifyContent: 'center',
                  '& .MuiSvgIcon-root': { fontSize: 20 },
                }}
              >
                {item.icon}
              </ListItemIcon>
              {(!collapsed || forMobile) && (
                <ListItemText
                  primary={item.title}
                  primaryTypographyProps={{
                    fontSize: '0.875rem',
                    fontWeight: isActive ? 600 : 400,
                  }}
                />
              )}
            </ListItemButton>
          );

          return collapsed && !forMobile ? (
            <Tooltip key={item.path} title={item.title} placement="right" arrow>
              {button}
            </Tooltip>
          ) : (
            <Box key={item.path}>{button}</Box>
          );
        })}
      </List>

      <Divider sx={{ borderColor: DARK.divider, position: 'relative', zIndex: 1 }} />

      {/* User section */}
      <Box sx={{ p: collapsed && !forMobile ? 1 : 2, position: 'relative', zIndex: 1 }}>
        {collapsed && !forMobile ? (
          <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 1 }}>
            <Tooltip title={user?.name || ''} placement="right">
              <Avatar
                sx={{
                  width: 32,
                  height: 32,
                  bgcolor: 'rgba(255,255,255,0.2)',
                  fontSize: '0.8rem',
                  cursor: 'pointer',
                }}
              >
                {user?.name?.charAt(0) || 'U'}
              </Avatar>
            </Tooltip>
            <Tooltip title="ログアウト" placement="right">
              <IconButton
                onClick={handleLogout}
                size="small"
                sx={{ color: '#ef5350', '&:hover': { bgcolor: 'rgba(239,83,80,0.1)' } }}
              >
                <LogoutIcon fontSize="small" />
              </IconButton>
            </Tooltip>
          </Box>
        ) : (
          <>
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5, mb: 1 }}>
              <Avatar
                sx={{
                  width: 32,
                  height: 32,
                  bgcolor: 'rgba(255,255,255,0.2)',
                  fontSize: '0.8rem',
                }}
              >
                {user?.name?.charAt(0) || 'U'}
              </Avatar>
              <Box sx={{ overflow: 'hidden' }}>
                <Typography
                  sx={{
                    color: DARK.textPrimary,
                    fontSize: '0.8rem',
                    fontWeight: 600,
                    whiteSpace: 'nowrap',
                    textOverflow: 'ellipsis',
                    overflow: 'hidden',
                  }}
                >
                  {user?.name}
                </Typography>
                <Typography sx={{ color: 'rgba(255,255,255,0.5)', fontSize: '0.7rem' }}>
                  {user?.role}
                </Typography>
              </Box>
            </Box>
            <ListItemButton
              onClick={handleLogout}
              sx={{
                borderRadius: '4px',
                color: '#ef5350',
                py: 0.5,
                '&:hover': { bgcolor: 'rgba(239,83,80,0.1)' },
              }}
            >
              <ListItemIcon sx={{ color: 'inherit', minWidth: 36 }}>
                <LogoutIcon fontSize="small" />
              </ListItemIcon>
              <ListItemText
                primary="ログアウト"
                primaryTypographyProps={{ fontSize: '0.85rem' }}
              />
            </ListItemButton>
          </>
        )}
      </Box>
    </Box>
  );

  if (isMobile) {
    return (
      <Box sx={{ minHeight: '100vh' }}>
        {/* Mobile TopBar */}
        <AppBar
          position="fixed"
          elevation={0}
          sx={{
            bgcolor: 'rgba(255,255,255,0.8)',
            backdropFilter: 'blur(20px)',
            WebkitBackdropFilter: 'blur(20px)',
            borderBottom: '1px solid',
            borderColor: 'divider',
          }}
        >
          <Toolbar sx={{ minHeight: 64 }}>
            <IconButton
              onClick={() => setDrawerOpen(true)}
              sx={{ color: 'text.primary' }}
              aria-label="メニューを開閉"
            >
              <MenuIcon />
            </IconButton>
            <Typography
              variant="h6"
              sx={{ flex: 1, textAlign: 'center', fontWeight: 700, color: 'text.primary' }}
            >
              Ivy
            </Typography>
            <Avatar
              sx={{ width: 32, height: 32, bgcolor: 'primary.main', fontSize: '0.8rem' }}
            >
              {user?.name?.charAt(0) || 'U'}
            </Avatar>
          </Toolbar>
        </AppBar>

        {/* Mobile Drawer */}
        <Drawer
          open={drawerOpen}
          onClose={() => setDrawerOpen(false)}
          sx={{ '& .MuiDrawer-paper': { width: SIDEBAR_WIDTH, border: 'none' } }}
        >
          {sidebarContent(true)}
        </Drawer>

        {/* Main Content */}
        <Box component="main" sx={{ pt: '80px', px: 2, pb: 3 }}>
          <Container maxWidth="lg">{children}</Container>
        </Box>
      </Box>
    );
  }

  // Desktop
  return (
    <Box sx={{ display: 'flex', minHeight: '100vh' }}>
      {/* Permanent Sidebar with margin */}
      <Box
        component="nav"
        sx={{
          width: currentWidth + SIDEBAR_MARGIN * 2,
          flexShrink: 0,
          transition: 'width 0.2s ease',
        }}
      >
        <Box
          sx={{
            position: 'fixed',
            top: SIDEBAR_MARGIN,
            left: SIDEBAR_MARGIN,
            width: currentWidth,
            height: `calc(100vh - ${SIDEBAR_MARGIN * 2}px)`,
            overflow: 'auto',
            transition: 'width 0.2s ease',
            // Hide scrollbar
            '&::-webkit-scrollbar': { display: 'none' },
            msOverflowStyle: 'none',
            scrollbarWidth: 'none',
          }}
        >
          {sidebarContent(false)}

          {/* Collapse toggle - positioned at right edge, vertically centered */}
          <IconButton
            onClick={() => setCollapsed(!collapsed)}
            size="small"
            sx={{
              position: 'absolute',
              top: '50%',
              right: -14,
              transform: 'translateY(-50%)',
              width: 28,
              height: 28,
              bgcolor: '#1a1a1a',
              color: DARK.textSecondary,
              border: '2px solid',
              borderColor: 'background.default',
              boxShadow: '0 2px 8px rgba(0,0,0,0.15)',
              zIndex: 10,
              '&:hover': { bgcolor: '#2a2a2a', color: DARK.textPrimary },
            }}
          >
            {collapsed ? <ChevronRightIcon sx={{ fontSize: 16 }} /> : <ChevronLeftIcon sx={{ fontSize: 16 }} />}
          </IconButton>
        </Box>
      </Box>

      {/* Main Content */}
      <Box component="main" sx={{ flex: 1, p: 3, minWidth: 0 }}>
        <Container maxWidth="lg">{children}</Container>
      </Box>
    </Box>
  );
}
