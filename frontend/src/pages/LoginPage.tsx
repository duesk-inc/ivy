import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Alert,
} from '@mui/material';
import { ActionButton } from '../components/common';
import { SimpleTextField } from '../components/common/forms';
import { useAuth } from '../context/AuthContext';

export default function LoginPage() {
  const navigate = useNavigate();
  const { login } = useAuth();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      await login(email, password);
      navigate('/');
    } catch (err: any) {
      setError(err?.response?.data?.error || 'ログインに失敗しました');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Box
      sx={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        bgcolor: '#f5f5f5',
      }}
    >
      <Card sx={{ maxWidth: 400, width: '100%', mx: 2 }}>
        <CardContent sx={{ p: 4 }}>
          <Typography variant="h4" align="center" gutterBottom sx={{ color: 'primary.main', fontWeight: 700 }}>
            Ivy
          </Typography>
          <Typography variant="body2" align="center" color="text.secondary" sx={{ mb: 3 }}>
            SES マッチングツール
          </Typography>

          {error && (
            <Alert severity="error" sx={{ mb: 2 }}>
              {error}
            </Alert>
          )}

          <Box component="form" onSubmit={handleSubmit}>
            <SimpleTextField
              label="メールアドレス"
              type="email"
              fullWidth
              required
              value={email}
              onChange={setEmail}
              sx={{ mb: 2 }}
              autoFocus
            />
            <SimpleTextField
              label="パスワード"
              type="password"
              fullWidth
              required
              value={password}
              onChange={setPassword}
              sx={{ mb: 3 }}
            />
            <ActionButton
              buttonType="primary"
              type="submit"
              fullWidth
              size="large"
              loading={loading}
            >
              ログイン
            </ActionButton>
          </Box>
        </CardContent>
      </Card>
    </Box>
  );
}
