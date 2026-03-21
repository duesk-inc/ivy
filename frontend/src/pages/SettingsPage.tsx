import { useState } from 'react';
import {
  Box,
  Typography,
  Card,
  CardContent,
  Alert,
  Divider,
} from '@mui/material';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { SectionLoader } from '../components/common';
import { SimpleTextField, SimpleSelect, CurrencyTextField } from '../components/common/forms';
import Layout from '../components/common/Layout';
import { getSettings, updateSetting } from '../lib/api/client';

const AI_MODELS = [
  { value: 'claude-haiku-4-5-20251001', label: 'Claude Haiku 4.5 (高速・低コスト)' },
  { value: 'claude-sonnet-4-6', label: 'Claude Sonnet 4.6 (高精度・中コスト)' },
  { value: 'claude-opus-4-6', label: 'Claude Opus 4.6 (最高精度・高コスト)' },
];

export default function SettingsPage() {
  const queryClient = useQueryClient();
  const [success, setSuccess] = useState('');
  const [error, setError] = useState('');

  const { data, isLoading } = useQuery({
    queryKey: ['settings'],
    queryFn: getSettings,
  });

  const mutation = useMutation({
    mutationFn: ({ key, value }: { key: string; value: any }) => updateSetting(key, value),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['settings'] });
      setSuccess('設定を更新しました');
      setTimeout(() => setSuccess(''), 3000);
    },
    onError: (err: any) => {
      setError(err?.response?.data?.error || '設定の更新に失敗しました');
    },
  });

  const getSetting = (key: string) => {
    return data?.settings?.find((s) => s.key === key)?.value;
  };

  const margin = getSetting('margin') || { type: 'fixed', amount: 50000 };
  const aiModel = getSetting('ai_model') || { model: 'claude-haiku-4-5-20251001' };
  const dataRetention = getSetting('data_retention') || { jobs_days: 90, engineers_days: 180, matchings_days: 365 };

  const handleMarginUpdate = (amount: number) => {
    mutation.mutate({
      key: 'margin',
      value: { type: margin.type, amount },
    });
  };

  const handleMarginTypeUpdate = (type: string) => {
    mutation.mutate({
      key: 'margin',
      value: { type, amount: margin.amount },
    });
  };

  const handleModelUpdate = (model: string) => {
    mutation.mutate({
      key: 'ai_model',
      value: { model },
    });
  };

  if (isLoading) {
    return (
      <Layout>
        <SectionLoader />
      </Layout>
    );
  }

  return (
    <Layout>
      <Typography variant="h5" sx={{ mb: 3, fontWeight: 600 }}>
        設定
      </Typography>

      {success && <Alert severity="success" sx={{ mb: 2 }}>{success}</Alert>}
      {error && <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError('')}>{error}</Alert>}

      {/* マージン設定 */}
      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Typography variant="h6" sx={{ mb: 2 }}>
            マージン設定
          </Typography>
          <Box sx={{ display: 'flex', gap: 2, alignItems: 'center' }}>
            <SimpleSelect
              value={margin.type}
              onChange={(v) => handleMarginTypeUpdate(v)}
              options={[
                { value: 'fixed', label: '固定金額' },
                { value: 'percentage', label: 'パーセンテージ' },
              ]}
              label="種別"
              size="small"
              sx={{ minWidth: 150 }}
              fullWidth={false}
            />
            {margin.type === 'fixed' ? (
              <CurrencyTextField
                label="金額（万円）"
                size="small"
                value={margin.amount / 10000}
                onChange={(v) => handleMarginUpdate(v * 10000)}
                currencyPosition="end"
                sx={{ width: 150 }}
                fullWidth={false}
              />
            ) : (
              <SimpleTextField
                label="パーセンテージ（%）"
                type="number"
                size="small"
                value={margin.amount}
                onChange={(v) => handleMarginUpdate(Number(v))}
                sx={{ width: 150 }}
                fullWidth={false}
              />
            )}
            {margin.type === 'fixed' && (
              <Typography color="text.secondary">
                （{margin.amount.toLocaleString()}円）
              </Typography>
            )}
          </Box>
        </CardContent>
      </Card>

      {/* AIモデル設定 */}
      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Typography variant="h6" sx={{ mb: 2 }}>
            AIモデル設定
          </Typography>
          <SimpleSelect
            value={aiModel.model}
            onChange={(v) => handleModelUpdate(v)}
            options={AI_MODELS}
            label="モデル"
            size="small"
          />
        </CardContent>
      </Card>

      {/* データ保持期間 */}
      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Typography variant="h6" sx={{ mb: 2 }}>
            データ保持期間
          </Typography>
          <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap' }}>
            <SimpleTextField
              label="案件情報（日）"
              type="number"
              size="small"
              value={dataRetention.jobs_days}
              onChange={(v) => mutation.mutate({ key: 'data_retention', value: { ...dataRetention, jobs_days: Number(v) } })}
              sx={{ width: 150 }}
              fullWidth={false}
            />
            <SimpleTextField
              label="人材情報（日）"
              type="number"
              size="small"
              value={dataRetention.engineers_days}
              onChange={(v) => mutation.mutate({ key: 'data_retention', value: { ...dataRetention, engineers_days: Number(v) } })}
              sx={{ width: 150 }}
              fullWidth={false}
            />
            <SimpleTextField
              label="マッチング結果（日）"
              type="number"
              size="small"
              value={dataRetention.matchings_days}
              onChange={(v) => mutation.mutate({ key: 'data_retention', value: { ...dataRetention, matchings_days: Number(v) } })}
              sx={{ width: 180 }}
              fullWidth={false}
            />
          </Box>
        </CardContent>
      </Card>

      {/* API使用量 */}
      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Typography variant="h6" sx={{ mb: 1 }}>
            API使用量（今月）
          </Typography>
          <Divider sx={{ mb: 2 }} />
          <Typography color="text.secondary">
            Phase 2で実装予定です。matchingsテーブルのtokens_usedから月間集計を表示します。
          </Typography>
        </CardContent>
      </Card>

      {/* ユーザー管理 */}
      <Card>
        <CardContent>
          <Typography variant="h6" sx={{ mb: 1 }}>
            ユーザー管理
          </Typography>
          <Divider sx={{ mb: 2 }} />
          <Typography color="text.secondary">
            Monsteraの管理画面で一元管理しています。
          </Typography>
        </CardContent>
      </Card>
    </Layout>
  );
}
