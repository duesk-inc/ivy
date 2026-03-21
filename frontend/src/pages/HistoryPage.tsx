import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Box,
  Typography,
  Card,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Chip,
  IconButton,
  Pagination,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  CircularProgress,
  Alert,
} from '@mui/material';
import DeleteIcon from '@mui/icons-material/Delete';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import Layout from '../components/common/Layout';
import { getMatchings, deleteMatching } from '../lib/api/client';
import { gradeColor, supplyChainColor } from '../utils/grade';
import { SUPPLY_CHAIN_LABELS } from '../types';

export default function HistoryPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [page, setPage] = useState(1);
  const [grade, setGrade] = useState('');

  const { data, isLoading, error } = useQuery({
    queryKey: ['matchings', page, grade],
    queryFn: () => getMatchings(page, 20, grade || undefined),
  });

  const deleteMutation = useMutation({
    mutationFn: deleteMatching,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['matchings'] });
    },
  });

  const handleDelete = (e: React.MouseEvent, id: string) => {
    e.stopPropagation();
    if (window.confirm('この結果を削除しますか？')) {
      deleteMutation.mutate(id);
    }
  };

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString('ja-JP', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  return (
    <Layout>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h5" sx={{ fontWeight: 600 }}>
          マッチング履歴
        </Typography>
        <FormControl size="small" sx={{ minWidth: 120 }}>
          <InputLabel>判定</InputLabel>
          <Select
            value={grade}
            label="判定"
            onChange={(e) => { setGrade(e.target.value); setPage(1); }}
          >
            <MenuItem value="">全て</MenuItem>
            <MenuItem value="A">A判定</MenuItem>
            <MenuItem value="B">B判定</MenuItem>
            <MenuItem value="C">C判定</MenuItem>
            <MenuItem value="D">D判定</MenuItem>
          </Select>
        </FormControl>
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
      ) : (
        <Card>
          <TableContainer>
            <Table>
              <TableHead>
                <TableRow>
                  <TableCell>日時</TableCell>
                  <TableCell>案件名</TableCell>
                  <TableCell>送信元</TableCell>
                  <TableCell align="center">商流</TableCell>
                  <TableCell align="center">スコア</TableCell>
                  <TableCell align="center">判定</TableCell>
                  <TableCell align="center">操作</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {data?.items?.length === 0 && (
                  <TableRow>
                    <TableCell colSpan={7} align="center" sx={{ py: 4 }}>
                      <Typography color="text.secondary">履歴がありません</Typography>
                    </TableCell>
                  </TableRow>
                )}
                {data?.items?.map((item) => (
                  <TableRow
                    key={item.id}
                    hover
                    sx={{ cursor: 'pointer' }}
                    onClick={() => navigate(`/history/${item.id}`)}
                  >
                    <TableCell sx={{ whiteSpace: 'nowrap' }}>
                      {formatDate(item.created_at)}
                    </TableCell>
                    <TableCell>
                      {item.job_summary?.name || '(案件名なし)'}
                    </TableCell>
                    <TableCell>
                      <Typography variant="body2">
                        {item.supply_chain_source || '-'}
                      </Typography>
                    </TableCell>
                    <TableCell align="center">
                      <Chip
                        label={SUPPLY_CHAIN_LABELS[item.supply_chain_level] || '不明'}
                        size="small"
                        sx={{
                          bgcolor: supplyChainColor(item.supply_chain_level),
                          color: 'white',
                          fontWeight: 600,
                        }}
                      />
                    </TableCell>
                    <TableCell align="center">
                      <Typography fontWeight={600}>{item.total_score}点</Typography>
                    </TableCell>
                    <TableCell align="center">
                      <Chip
                        label={`${item.grade} - ${item.grade_label}`}
                        size="small"
                        sx={{
                          bgcolor: gradeColor(item.grade),
                          color: 'white',
                          fontWeight: 600,
                        }}
                      />
                    </TableCell>
                    <TableCell align="center">
                      <IconButton
                        size="small"
                        onClick={(e) => handleDelete(e, item.id)}
                        disabled={deleteMutation.isPending}
                      >
                        <DeleteIcon fontSize="small" />
                      </IconButton>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
          {data && data.total_pages > 1 && (
            <Box sx={{ display: 'flex', justifyContent: 'center', py: 2 }}>
              <Pagination
                count={data.total_pages}
                page={page}
                onChange={(_, p) => setPage(p)}
              />
            </Box>
          )}
        </Card>
      )}
    </Layout>
  );
}
