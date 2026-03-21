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
  useTheme,
  useMediaQuery,
} from '@mui/material';
import MenuIcon from '@mui/icons-material/Menu';
import CompareArrowsIcon from '@mui/icons-material/CompareArrows';
import HistoryIcon from '@mui/icons-material/History';
import SettingsIcon from '@mui/icons-material/Settings';
import LogoutIcon from '@mui/icons-material/Logout';
import { useAuth } from '../../context/AuthContext';

interface LayoutProps {
  children: ReactNode;
}

const SIDEBAR_WIDTH = 246;

export default function Layout({ children }: LayoutProps) {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));
  const [drawerOpen, setDrawerOpen] = useState(false);

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

  const sidebarContent = (
    <Box
      sx={{
        display: 'flex',
        flexDirection: 'column',
        height: '100%',
        background:
          'linear-gradient(180deg, #1a1a1a 0%, #1a1a1a 35%, #3a3a3a 70%, #3a3a3a 100%)',
      }}
    >
      {/* Logo */}
      <Box sx={{ p: 3, textAlign: 'center' }}>
        <Typography
          variant="h6"
          sx={{ color: 'white', fontWeight: 700, letterSpacing: 1 }}
        >
          Ivy
        </Typography>
        <Typography
          variant="caption"
          sx={{ color: 'rgba(255,255,255,0.6)' }}
        >
          SES マッチングツール
        </Typography>
      </Box>
      <Divider sx={{ borderColor: 'rgba(255,255,255,0.15)' }} />

      {/* Menu Items */}
      <List sx={{ flex: 1, px: 1, py: 2 }}>
        {menuItems.map((item) => (
          <ListItemButton
            key={item.path}
            selected={location.pathname === item.path}
            onClick={() => {
              navigate(item.path);
              if (isMobile) setDrawerOpen(false);
            }}
            sx={{
              borderRadius: '4px',
              mx: 1,
              my: 0.5,
              color: 'rgba(255,255,255,0.7)',
              '&.Mui-selected': {
                bgcolor: 'rgba(255,255,255,0.18)',
                color: 'rgba(255,255,255,0.95)',
                fontWeight: 600,
                '&:hover': { bgcolor: 'rgba(255,255,255,0.22)' },
              },
              '&:hover': { bgcolor: 'rgba(255,255,255,0.12)' },
              transition: 'all 0.15s ease',
            }}
          >
            <ListItemIcon sx={{ color: 'inherit', minWidth: 36 }}>
              {item.icon}
            </ListItemIcon>
            <ListItemText primary={item.title} />
          </ListItemButton>
        ))}
      </List>

      {/* User section at bottom */}
      <Box sx={{ p: 2, borderTop: '1px solid rgba(255,255,255,0.15)' }}>
        <Box
          sx={{ display: 'flex', alignItems: 'center', gap: 1.5, mb: 1 }}
        >
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
          <Box>
            <Typography
              sx={{
                color: 'rgba(255,255,255,0.95)',
                fontSize: '0.8rem',
                fontWeight: 600,
              }}
            >
              {user?.name}
            </Typography>
            <Typography
              sx={{ color: 'rgba(255,255,255,0.5)', fontSize: '0.7rem' }}
            >
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
      </Box>
    </Box>
  );

  if (isMobile) {
    return (
      <Box sx={{ minHeight: '100vh', bgcolor: 'background.default' }}>
        {/* Mobile TopBar */}
        <AppBar
          position="fixed"
          elevation={0}
          sx={{
            bgcolor: 'transparent',
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
              sx={{
                flex: 1,
                textAlign: 'center',
                fontWeight: 700,
                color: 'primary.main',
              }}
            >
              Ivy
            </Typography>
            <Avatar
              sx={{
                width: 32,
                height: 32,
                bgcolor: 'primary.main',
                fontSize: '0.8rem',
              }}
            >
              {user?.name?.charAt(0) || 'U'}
            </Avatar>
          </Toolbar>
        </AppBar>

        {/* Mobile Drawer */}
        <Drawer
          open={drawerOpen}
          onClose={() => setDrawerOpen(false)}
          sx={{
            '& .MuiDrawer-paper': {
              width: SIDEBAR_WIDTH,
              border: 'none',
            },
          }}
        >
          {sidebarContent}
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
    <Box
      sx={{ display: 'flex', minHeight: '100vh', bgcolor: 'background.default' }}
    >
      {/* Permanent Sidebar */}
      <Box component="nav" sx={{ width: SIDEBAR_WIDTH, flexShrink: 0 }}>
        <Box
          sx={{
            position: 'fixed',
            width: SIDEBAR_WIDTH,
            height: '100vh',
            overflow: 'auto',
          }}
        >
          {sidebarContent}
        </Box>
      </Box>

      {/* Main Content */}
      <Box component="main" sx={{ flex: 1, p: 3, minWidth: 0 }}>
        <Container maxWidth="lg">{children}</Container>
      </Box>
    </Box>
  );
}
