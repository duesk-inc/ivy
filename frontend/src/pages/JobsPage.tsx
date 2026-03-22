import { useState, Fragment } from 'react';
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
  Collapse,
  Pagination,
  Alert,
  IconButton,
  Tooltip,
} from '@mui/material';
import KeyboardArrowDownIcon from '@mui/icons-material/KeyboardArrowDown';
import KeyboardArrowUpIcon from '@mui/icons-material/KeyboardArrowUp';
import PersonSearchIcon from '@mui/icons-material/PersonSearch';
import { useQuery } from '@tanstack/react-query';
import Layout from '../components/common/Layout';
import { ActionButton, SectionLoader } from '../components/common';
import { SimpleSelect } from '../components/common/forms';
import { SimpleTextField } from '../components/common/forms';
import { getJobs, matchJobToEngineers } from '../lib/api/client';
import MatchCandidatesDialog from '../components/matching/MatchCandidatesDialog';
import type { Job, BatchMatchingResponse } from '../types';

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString('ja-JP', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  });
}

function formatRate(job: Job): string {
  const min = job.parsed.rate_min;
  const max = job.parsed.rate_max;
  if (min && max) return `${min}~${max}万`;
  if (min) return `${min}万~`;
  if (max) return `~${max}万`;
  return '-';
}

interface JobRowProps {
  job: Job;
  onMatchEngineers: (job: Job) => void;
}

function JobRow({ job, onMatchEngineers }: JobRowProps) {
  const [open, setOpen] = useState(false);

  return (
    <Fragment>
      <TableRow
        hover
        sx={{ cursor: 'pointer', '& > *': { borderBottom: open ? 'unset' : undefined } }}
        onClick={() => setOpen(!open)}
      >
        <TableCell sx={{ width: 40, p: 0.5 }}>
          <IconButton size="small">
            {open ? <KeyboardArrowUpIcon /> : <KeyboardArrowDownIcon />}
          </IconButton>
        </TableCell>
        <TableCell>
          <Typography variant="body2" fontWeight={500}>
            {job.parsed.name || '(案件名なし)'}
          </Typography>
        </TableCell>
        <TableCell>
          <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.5 }}>
            {job.parsed.skills?.slice(0, 5).map((skill) => (
              <Chip key={skill} label={skill} size="small" variant="outlined" />
            ))}
            {(job.parsed.skills?.length ?? 0) > 5 && (
              <Chip label={`+${(job.parsed.skills?.length ?? 0) - 5}`} size="small" />
            )}
          </Box>
        </TableCell>
        <TableCell sx={{ whiteSpace: 'nowrap' }}>{formatRate(job)}</TableCell>
        <TableCell sx={{ whiteSpace: 'nowrap' }}>{job.start_month || job.parsed.start_month || '-'}</TableCell>
        <TableCell align="center">
          <Chip
            label={job.status === 'active' ? '有効' : 'アーカイブ'}
            size="small"
            color={job.status === 'active' ? 'success' : 'default'}
            variant={job.status === 'active' ? 'filled' : 'outlined'}
          />
        </TableCell>
        <TableCell sx={{ whiteSpace: 'nowrap' }}>{formatDate(job.created_at)}</TableCell>
        <TableCell align="center" onClick={(e) => e.stopPropagation()}>
          <Tooltip title="この案件にマッチする人材を探す">
            <span>
              <ActionButton
                buttonType="secondary"
                size="small"
                icon={<PersonSearchIcon />}
                onClick={() => onMatchEngineers(job)}
              >
                人材を探す
              </ActionButton>
            </span>
          </Tooltip>
        </TableCell>
      </TableRow>
      <TableRow>
        <TableCell sx={{ py: 0 }} colSpan={8}>
          <Collapse in={open} timeout="auto" unmountOnExit>
            <Box sx={{ py: 2, px: 1 }}>
              <Typography variant="subtitle2" sx={{ mb: 1 }}>
                元テキスト
              </Typography>
              <Typography
                variant="body2"
                sx={{
                  whiteSpace: 'pre-wrap',
                  bgcolor: 'grey.50',
                  p: 2,
                  borderRadius: 1,
                  maxHeight: 300,
                  overflow: 'auto',
                  fontFamily: 'monospace',
                  fontSize: '0.8rem',
                }}
              >
                {job.raw_text}
              </Typography>
            </Box>
          </Collapse>
        </TableCell>
      </TableRow>
    </Fragment>
  );
}

export default function JobsPage() {
  const [page, setPage] = useState(1);
  const [startMonth, setStartMonth] = useState('');
  const [status, setStatus] = useState('');
  const [matchDialogOpen, setMatchDialogOpen] = useState(false);
  const [matchDialogTitle, setMatchDialogTitle] = useState('');
  const [batchResponse, setBatchResponse] = useState<BatchMatchingResponse | null>(null);

  const handleMatchEngineers = async (job: Job) => {
    const jobName = job.parsed.name || '(案件名なし)';
    setMatchDialogTitle(`${jobName} にマッチする人材`);
    setMatchDialogOpen(true);
    setBatchResponse(null);
    try {
      const resp = await matchJobToEngineers(job.id);
      setBatchResponse(resp);
    } catch {
      setBatchResponse(null);
    }
  };

  const { data, isLoading, error } = useQuery({
    queryKey: ['jobs', page, startMonth, status],
    queryFn: () =>
      getJobs({
        page,
        page_size: 20,
        start_month: startMonth || undefined,
        status: status || undefined,
      }),
  });

  return (
    <Layout>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3, flexWrap: 'wrap', gap: 2 }}>
        <Typography variant="h5" sx={{ fontWeight: 600 }}>
          案件一覧
        </Typography>
        <Box sx={{ display: 'flex', gap: 2, alignItems: 'center' }}>
          <SimpleTextField
            value={startMonth}
            onChange={(v) => {
              setStartMonth(v);
              setPage(1);
            }}
            label="稼働開始月"
            placeholder="2026-04"
            size="small"
          />
          <SimpleSelect
            value={status}
            onChange={(v) => {
              setStatus(v);
              setPage(1);
            }}
            options={[
              { value: '', label: '全て' },
              { value: 'active', label: '有効' },
              { value: 'archived', label: 'アーカイブ' },
            ]}
            label="ステータス"
            size="small"
            sx={{ minWidth: 120 }}
            fullWidth={false}
          />
        </Box>
      </Box>

      {error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          データの取得に失敗しました
        </Alert>
      )}

      {isLoading ? (
        <SectionLoader />
      ) : (
        <Card>
          <TableContainer>
            <Table>
              <TableHead>
                <TableRow>
                  <TableCell sx={{ width: 40 }} />
                  <TableCell>案件名</TableCell>
                  <TableCell>スキル</TableCell>
                  <TableCell>単価</TableCell>
                  <TableCell>開始月</TableCell>
                  <TableCell align="center">ステータス</TableCell>
                  <TableCell>登録日</TableCell>
                  <TableCell align="center">アクション</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {data?.items?.length === 0 && (
                  <TableRow>
                    <TableCell colSpan={8} align="center" sx={{ py: 4 }}>
                      <Typography color="text.secondary">案件がありません</Typography>
                    </TableCell>
                  </TableRow>
                )}
                {data?.items?.map((job) => (
                  <JobRow key={job.id} job={job} onMatchEngineers={handleMatchEngineers} />
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
      <MatchCandidatesDialog
        open={matchDialogOpen}
        onClose={() => setMatchDialogOpen(false)}
        title={matchDialogTitle}
        batchResponse={batchResponse}
      />
    </Layout>
  );
}
