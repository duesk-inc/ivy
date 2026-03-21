import { type ReactNode } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import {
  AppBar,
  Toolbar,
  Typography,
  Button,
  IconButton,
  Container,
  Box,
  Tooltip,
} from '@mui/material';
import SettingsIcon from '@mui/icons-material/Settings';
import LogoutIcon from '@mui/icons-material/Logout';
import { useAuth } from '../../context/AuthContext';

interface LayoutProps {
  children: ReactNode;
}

const navItems = [
  { label: 'マッチング', path: '/' },
  { label: '履歴', path: '/history' },
];

export default function Layout({ children }: LayoutProps) {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  const isAdmin = user?.role === 'admin';

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', minHeight: '100vh' }}>
      <AppBar position="sticky" color="primary">
        <Toolbar>
          <Typography
            variant="h6"
            component="div"
            sx={{
              fontWeight: 700,
              letterSpacing: 1,
              cursor: 'pointer',
              mr: 4,
            }}
            onClick={() => navigate('/')}
          >
            Ivy
          </Typography>

          <Box sx={{ display: 'flex', gap: 1, flexGrow: 1 }}>
            {navItems.map((item) => (
              <Button
                key={item.path}
                color="inherit"
                onClick={() => navigate(item.path)}
                sx={{
                  fontWeight: location.pathname === item.path ? 700 : 400,
                  borderBottom:
                    location.pathname === item.path
                      ? '2px solid white'
                      : '2px solid transparent',
                  borderRadius: 0,
                  px: 2,
                }}
              >
                {item.label}
              </Button>
            ))}
          </Box>

          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            {user && (
              <Typography variant="body2" sx={{ mr: 1 }}>
                {user.name}
              </Typography>
            )}

            {isAdmin && (
              <Tooltip title="設定">
                <IconButton color="inherit" onClick={() => navigate('/settings')}>
                  <SettingsIcon />
                </IconButton>
              </Tooltip>
            )}

            <Tooltip title="ログアウト">
              <IconButton color="inherit" onClick={handleLogout}>
                <LogoutIcon />
              </IconButton>
            </Tooltip>
          </Box>
        </Toolbar>
      </AppBar>

      <Container
        component="main"
        maxWidth="lg"
        sx={{ flex: 1, py: 3 }}
      >
        {children}
      </Container>
    </Box>
  );
}
