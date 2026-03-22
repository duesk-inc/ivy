import { useState, useEffect, useRef, useCallback } from 'react';
import {
  Box,
  Typography,
  Card,
  CardContent,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Chip,
  Alert,
  LinearProgress,
  Grid,
} from '@mui/material';
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import PreviewIcon from '@mui/icons-material/Preview';
import Layout from '../components/common/Layout';
import { ActionButton } from '../components/common';
import { SimpleTextField } from '../components/common/forms';
import {
  previewBatchMatching,
  executeBatchMatching,
  getBatchMatchingStatus,
} from '../lib/api/client';
import { gradeColor } from '../utils/grade';
import type { BatchMatchingPreview, BatchMatchingResponse } from '../types';

function getDefaultMonth(): string {
  const now = new Date();
  const year = now.getFullYear();
  const month = String(now.getMonth() + 2).padStart(2, '0');
  if (Number(month) > 12) {
    return `${year + 1}-01`;
  }
  return `${year}-${month}`;
}

export default function BatchMatchingPage() {
  const [startMonthFrom, setStartMonthFrom] = useState(getDefaultMonth);
  const [startMonthTo, setStartMonthTo] = useState(getDefaultMonth);
  const [preview, setPreview] = useState<BatchMatchingPreview | null>(null);
  const [previewLoading, setPreviewLoading] = useState(false);
  const [executeLoading, setExecuteLoading] = useState(false);
  const [batchResult, setBatchResult] = useState<BatchMatchingResponse | null>(null);
  const [error, setError] = useState('');
  const pollingRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const stopPolling = useCallback(() => {
    if (pollingRef.current) {
      clearInterval(pollingRef.current);
      pollingRef.current = null;
    }
  }, []);

  useEffect(() => {
    return () => {
      stopPolling();
    };
  }, [stopPolling]);

  const handlePreview = async () => {
    if (!startMonthFrom || !startMonthTo) {
      setError('期間を指定してください');
      return;
    }
    setError('');
    setPreviewLoading(true);
    setPreview(null);
    setBatchResult(null);
    try {
      const result = await previewBatchMatching({
        start_month_from: startMonthFrom,
        start_month_to: startMonthTo,
      });
      setPreview(result);
    } catch (err: unknown) {
      const message =
        err instanceof Error
          ? err.message
          : (err as { response?: { data?: { error?: string } } })?.response?.data?.error ||
            'プレビューの取得に失敗しました';
      setError(message);
    } finally {
      setPreviewLoading(false);
    }
  };

  const handleExecute = async () => {
    if (!preview) return;
    setError('');
    setExecuteLoading(true);
    setBatchResult(null);
    try {
      const result = await executeBatchMatching({
        start_month_from: startMonthFrom,
        start_month_to: startMonthTo,
      });
      setBatchResult(result);

      if (result.status === 'running') {
        pollingRef.current = setInterval(async () => {
          try {
            const updated = await getBatchMatchingStatus(result.id);
            setBatchResult(updated);
            if (updated.status !== 'running') {
              stopPolling();
              setExecuteLoading(false);
            }
          } catch {
            stopPolling();
            setExecuteLoading(false);
          }
        }, 5000);
      } else {
        setExecuteLoading(false);
      }
    } catch (err: unknown) {
      const message =
        err instanceof Error
          ? err.message
          : (err as { response?: { data?: { error?: string } } })?.response?.data?.error ||
            '一括マッチングの実行に失敗しました';
      setError(message);
      setExecuteLoading(false);
    }
  };

  const progress =
    batchResult && batchResult.total_pairs > 0
      ? ((batchResult.success_count + batchResult.failure_count) / batchResult.total_pairs) * 100
      : 0;

  return (
    <Layout>
      <Typography variant="h5" sx={{ mb: 3, fontWeight: 600 }}>
        一括マッチング
      </Typography>

      {error && (
        <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError('')}>
          {error}
        </Alert>
      )}

      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Typography variant="subtitle1" sx={{ mb: 2, fontWeight: 600 }}>
            期間指定
          </Typography>
          <Grid container spacing={2} alignItems="center">
            <Grid size={{ xs: 12, sm: 4, md: 3 }}>
              <SimpleTextField
                value={startMonthFrom}
                onChange={setStartMonthFrom}
                label="開始月（から）"
                type="month"
                size="small"
              />
            </Grid>
            <Grid size="auto">
              <Typography variant="body1" sx={{ pt: 1 }}>
                ~
              </Typography>
            </Grid>
            <Grid size={{ xs: 12, sm: 4, md: 3 }}>
              <SimpleTextField
                value={startMonthTo}
                onChange={setStartMonthTo}
                label="開始月（まで）"
                type="month"
                size="small"
              />
            </Grid>
            <Grid size="auto">
              <ActionButton
                buttonType="secondary"
                icon={<PreviewIcon />}
                onClick={handlePreview}
                loading={previewLoading}
                size="small"
              >
                プレビュー
              </ActionButton>
            </Grid>
          </Grid>

          {preview && (
            <Box sx={{ mt: 3, p: 2, bgcolor: 'grey.50', borderRadius: 1 }}>
              <Grid container spacing={2}>
                <Grid size={{ xs: 6, sm: 3 }}>
                  <Typography variant="body2" color="text.secondary">
                    対象案件
                  </Typography>
                  <Typography variant="h6" fontWeight={600}>
                    {preview.total_jobs}件
                  </Typography>
                </Grid>
                <Grid size={{ xs: 6, sm: 3 }}>
                  <Typography variant="body2" color="text.secondary">
                    対象人材
                  </Typography>
                  <Typography variant="h6" fontWeight={600}>
                    {preview.total_engineers}件
                  </Typography>
                </Grid>
                <Grid size={{ xs: 6, sm: 3 }}>
                  <Typography variant="body2" color="text.secondary">
                    フィルタ後ペア数
                  </Typography>
                  <Typography variant="h6" fontWeight={600}>
                    {preview.pairs_after_filter}ペア
                  </Typography>
                </Grid>
                <Grid size={{ xs: 6, sm: 3 }}>
                  <Typography variant="body2" color="text.secondary">
                    推定コスト
                  </Typography>
                  <Typography variant="h6" fontWeight={600}>
                    ${preview.estimated_cost.toFixed(2)}
                  </Typography>
                </Grid>
              </Grid>
            </Box>
          )}

          {preview && (
            <Box sx={{ mt: 3, display: 'flex', justifyContent: 'center' }}>
              <ActionButton
                buttonType="primary"
                size="large"
                icon={<PlayArrowIcon />}
                onClick={handleExecute}
                loading={executeLoading}
                disabled={preview.pairs_after_filter === 0}
                sx={{ px: 6, py: 1.5 }}
              >
                {executeLoading
                  ? `実行中... (${batchResult?.success_count ?? 0}/${batchResult?.total_pairs ?? preview.pairs_after_filter})`
                  : `一括マッチング実行 (${preview.pairs_after_filter}ペア / $${preview.estimated_cost.toFixed(2)})`}
              </ActionButton>
            </Box>
          )}
        </CardContent>
      </Card>

      {executeLoading && batchResult && (
        <Box sx={{ mb: 3 }}>
          <LinearProgress variant="determinate" value={progress} sx={{ height: 8, borderRadius: 4 }} />
          <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5, textAlign: 'center' }}>
            {batchResult.success_count + batchResult.failure_count} / {batchResult.total_pairs} 完了
          </Typography>
        </Box>
      )}

      {batchResult && batchResult.status !== 'running' && (
        <Box sx={{ mb: 2 }}>
          <Alert
            severity={batchResult.status === 'completed' ? 'success' : 'error'}
            sx={{ mb: 2 }}
          >
            {batchResult.status === 'completed'
              ? `一括マッチング完了: 成功 ${batchResult.success_count}件 / 失敗 ${batchResult.failure_count}件`
              : '一括マッチングに失敗しました'}
          </Alert>
        </Box>
      )}

      {batchResult && batchResult.results && batchResult.results.length > 0 && (
        <Card>
          <CardContent>
            <Typography variant="subtitle1" sx={{ mb: 2, fontWeight: 600 }}>
              結果（上位{batchResult.results.length}件）
            </Typography>
            <TableContainer>
              <Table size="small">
                <TableHead>
                  <TableRow>
                    <TableCell sx={{ width: 50 }}>#</TableCell>
                    <TableCell>案件名</TableCell>
                    <TableCell>人材名</TableCell>
                    <TableCell align="center">スコア</TableCell>
                    <TableCell align="center">判定</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {batchResult.results.map((item, index) => (
                    <TableRow key={`${item.job_id}-${item.engineer_id}`} hover>
                      <TableCell>{index + 1}</TableCell>
                      <TableCell>{item.job_name}</TableCell>
                      <TableCell>{item.engineer_name}</TableCell>
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
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          </CardContent>
        </Card>
      )}
    </Layout>
  );
}
