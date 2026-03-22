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
import WorkIcon from '@mui/icons-material/Work';
import { useQuery } from '@tanstack/react-query';
import Layout from '../components/common/Layout';
import { ActionButton, SectionLoader } from '../components/common';
import { SimpleSelect } from '../components/common/forms';
import { SimpleTextField } from '../components/common/forms';
import { getEngineerProfiles, matchEngineerToJobs } from '../lib/api/client';
import MatchCandidatesDialog from '../components/matching/MatchCandidatesDialog';
import type { EngineerProfile, BatchMatchingResponse } from '../types';

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString('ja-JP', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  });
}

interface EngineerRowProps {
  engineer: EngineerProfile;
  onMatchJobs: (engineer: EngineerProfile) => void;
}

function EngineerRow({ engineer, onMatchJobs }: EngineerRowProps) {
  const [open, setOpen] = useState(false);
  const p = engineer.parsed;

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
            {p.initials || '(名前なし)'}
          </Typography>
        </TableCell>
        <TableCell sx={{ whiteSpace: 'nowrap' }}>{p.age ? `${p.age}歳` : '-'}</TableCell>
        <TableCell>
          <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.5 }}>
            {p.skills?.slice(0, 5).map((skill) => (
              <Chip key={skill} label={skill} size="small" variant="outlined" />
            ))}
            {(p.skills?.length ?? 0) > 5 && (
              <Chip label={`+${(p.skills?.length ?? 0) - 5}`} size="small" />
            )}
          </Box>
        </TableCell>
        <TableCell sx={{ whiteSpace: 'nowrap' }}>{p.rate ? `${p.rate}万` : '-'}</TableCell>
        <TableCell sx={{ whiteSpace: 'nowrap' }}>
          {engineer.start_month || p.start_month || '-'}
        </TableCell>
        <TableCell sx={{ whiteSpace: 'nowrap' }}>{p.nationality || '-'}</TableCell>
        <TableCell align="center">
          <Chip
            label={engineer.status === 'active' ? '有効' : 'アーカイブ'}
            size="small"
            color={engineer.status === 'active' ? 'success' : 'default'}
            variant={engineer.status === 'active' ? 'filled' : 'outlined'}
          />
        </TableCell>
        <TableCell sx={{ whiteSpace: 'nowrap' }}>{formatDate(engineer.created_at)}</TableCell>
        <TableCell align="center" onClick={(e) => e.stopPropagation()}>
          <Tooltip title="この人材にマッチする案件を探す">
            <span>
              <ActionButton
                buttonType="secondary"
                size="small"
                icon={<WorkIcon />}
                onClick={() => onMatchJobs(engineer)}
              >
                案件を探す
              </ActionButton>
            </span>
          </Tooltip>
        </TableCell>
      </TableRow>
      <TableRow>
        <TableCell sx={{ py: 0 }} colSpan={10}>
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
                {engineer.raw_text}
              </Typography>
            </Box>
          </Collapse>
        </TableCell>
      </TableRow>
    </Fragment>
  );
}

export default function EngineersPage() {
  const [page, setPage] = useState(1);
  const [startMonth, setStartMonth] = useState('');
  const [status, setStatus] = useState('');
  const [matchDialogOpen, setMatchDialogOpen] = useState(false);
  const [matchDialogTitle, setMatchDialogTitle] = useState('');
  const [batchResponse, setBatchResponse] = useState<BatchMatchingResponse | null>(null);

  const handleMatchJobs = async (engineer: EngineerProfile) => {
    const name = engineer.parsed.initials || '(名前なし)';
    setMatchDialogTitle(`${name} にマッチする案件`);
    setMatchDialogOpen(true);
    setBatchResponse(null);
    try {
      const resp = await matchEngineerToJobs(engineer.id);
      setBatchResponse(resp);
    } catch {
      setBatchResponse(null);
    }
  };

  const { data, isLoading, error } = useQuery({
    queryKey: ['engineer-profiles', page, startMonth, status],
    queryFn: () =>
      getEngineerProfiles({
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
          人材一覧
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
                  <TableCell>イニシャル</TableCell>
                  <TableCell>年齢</TableCell>
                  <TableCell>スキル</TableCell>
                  <TableCell>単価</TableCell>
                  <TableCell>開始月</TableCell>
                  <TableCell>国籍</TableCell>
                  <TableCell align="center">ステータス</TableCell>
                  <TableCell>登録日</TableCell>
                  <TableCell align="center">アクション</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {data?.items?.length === 0 && (
                  <TableRow>
                    <TableCell colSpan={10} align="center" sx={{ py: 4 }}>
                      <Typography color="text.secondary">人材がありません</Typography>
                    </TableCell>
                  </TableRow>
                )}
                {data?.items?.map((engineer) => (
                  <EngineerRow key={engineer.id} engineer={engineer} onMatchJobs={handleMatchJobs} />
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
