import { useParams, useNavigate } from 'react-router-dom';
import {
  Box,
  Alert,
} from '@mui/material';
import ArrowBackIcon from '@mui/icons-material/ArrowBack';
import { useQuery } from '@tanstack/react-query';
import { ActionButton, SectionLoader } from '../components/common';
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
        <ActionButton
          buttonType="ghost"
          icon={<ArrowBackIcon />}
          onClick={() => navigate('/history')}
          sx={{ mb: 2 }}
        >
          履歴に戻る
        </ActionButton>
      </Box>

      {error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          データの取得に失敗しました
        </Alert>
      )}

      {isLoading ? (
        <SectionLoader />
      ) : data ? (
        <MatchingResult result={data} />
      ) : null}
    </Layout>
  );
}
