import { useState, useEffect, useCallback } from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Box,
  Typography,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Chip,
  LinearProgress,
  Alert,
} from '@mui/material';
import SearchIcon from '@mui/icons-material/Search';
import { ActionButton } from '../common';
import { getBatchMatchingStatus } from '../../lib/api/client';
import type { BatchMatchingResponse, BatchMatchingResultItem } from '../../types';

interface MatchCandidatesDialogProps {
  open: boolean;
  onClose: () => void;
  title: string;
  batchResponse: BatchMatchingResponse | null;
}

function gradeColor(grade: string): 'success' | 'info' | 'warning' | 'error' {
  switch (grade) {
    case 'A': return 'success';
    case 'B': return 'info';
    case 'C': return 'warning';
    default: return 'error';
  }
}

export default function MatchCandidatesDialog({
  open,
  onClose,
  title,
  batchResponse,
}: MatchCandidatesDialogProps) {
  const [status, setStatus] = useState<BatchMatchingResponse | null>(batchResponse);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setStatus(batchResponse);
    setError(null);
  }, [batchResponse]);

  // ポーリング
  useEffect(() => {
    if (!open || !status || status.status !== 'running') return;

    const interval = setInterval(async () => {
      try {
        const updated = await getBatchMatchingStatus(status.id);
        setStatus(updated);
        if (updated.status !== 'running') {
          clearInterval(interval);
        }
      } catch {
        setError('ステータスの取得に失敗しました');
        clearInterval(interval);
      }
    }, 3000);

    return () => clearInterval(interval);
  }, [open, status?.id, status?.status]);

  const results: BatchMatchingResultItem[] = (() => {
    if (!status?.results) return [];
    if (Array.isArray(status.results)) return status.results;
    try {
      return JSON.parse(status.results as unknown as string);
    } catch {
      return [];
    }
  })();

  const isRunning = status?.status === 'running';
  const isCompleted = status?.status === 'completed';
  const isFailed = status?.status === 'failed';
  const progress = status && status.total_pairs > 0
    ? ((status.success_count + status.failure_count) / status.total_pairs) * 100
    : 0;

  const handleClose = useCallback(() => {
    onClose();
  }, [onClose]);

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="md" fullWidth>
      <DialogTitle>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          <SearchIcon />
          {title}
        </Box>
      </DialogTitle>
      <DialogContent>
        {error && (
          <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>
        )}

        {isRunning && (
          <Box sx={{ mb: 2 }}>
            <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
              AIが分析中です... ({status.success_count + status.failure_count} / {status.total_pairs})
            </Typography>
            <LinearProgress variant="determinate" value={progress} />
          </Box>
        )}

        {isFailed && (
          <Alert severity="error" sx={{ mb: 2 }}>
            マッチング処理が失敗しました
          </Alert>
        )}

        {status && status.total_pairs === 0 && !isRunning && (
          <Alert severity="info">
            マッチング候補が見つかりませんでした（プレフィルタで全て除外されました）
          </Alert>
        )}

        {isCompleted && status.failure_count > 0 && (
          <Alert severity="warning" sx={{ mb: 2 }}>
            成功: {status.success_count}件 / 失敗: {status.failure_count}件
          </Alert>
        )}

        {results.length > 0 && (
          <TableContainer>
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell sx={{ width: 40 }}>#</TableCell>
                  <TableCell>案件</TableCell>
                  <TableCell>人材</TableCell>
                  <TableCell align="center">スコア</TableCell>
                  <TableCell align="center">判定</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {results.map((item, idx) => (
                  <TableRow key={`${item.job_id}-${item.engineer_id}`} hover>
                    <TableCell>{idx + 1}</TableCell>
                    <TableCell>
                      <Typography variant="body2">{item.job_name || '(案件名なし)'}</Typography>
                    </TableCell>
                    <TableCell>
                      <Typography variant="body2">{item.engineer_name || '(名前なし)'}</Typography>
                    </TableCell>
                    <TableCell align="center">
                      <Typography variant="body2" fontWeight={600}>{item.total_score}点</Typography>
                    </TableCell>
                    <TableCell align="center">
                      <Chip
                        label={`${item.grade} ${item.grade_label}`}
                        size="small"
                        color={gradeColor(item.grade)}
                      />
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        )}
      </DialogContent>
      <DialogActions>
        <ActionButton buttonType="cancel" onClick={handleClose}>
          閉じる
        </ActionButton>
      </DialogActions>
    </Dialog>
  );
}
