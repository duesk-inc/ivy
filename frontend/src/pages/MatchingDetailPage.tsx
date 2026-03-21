import { useParams, useNavigate } from 'react-router-dom';
import {
  Box,
  Button,
  CircularProgress,
  Alert,
} from '@mui/material';
import ArrowBackIcon from '@mui/icons-material/ArrowBack';
import { useQuery } from '@tanstack/react-query';
import Layout from '../components/common/Layout';
import MatchingResult from '../components/matching/MatchingResult';
import { getMatchingDetail } from '../lib/api/client';

export default function MatchingDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();

  const { data, isLoading, error } = useQuery({
    queryKey: ['matching', id],
    queryFn: () => getMatchingDetail(id!),
    enabled: !!id,
  });

  return (
    <Layout>
      <Box sx={{ mb: 3 }}>
        <Button
          startIcon={<ArrowBackIcon />}
          onClick={() => navigate('/history')}
          sx={{ mb: 2 }}
        >
          履歴に戻る
        </Button>
      </Box>

      {error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          データの取得に失敗しました
        </Alert>
      )}

      {isLoading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
          <CircularProgress />
        </Box>
      ) : data ? (
        <MatchingResult result={data} />
      ) : null}
    </Layout>
  );
}
