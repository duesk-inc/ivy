import { useState, useRef } from 'react';
import {
  Box,
  Grid,
  TextField,
  Button,
  Typography,
  Card,
  CardContent,
  CircularProgress,
  Alert,
  MenuItem,
  Select,
  FormControl,
  InputLabel,
  Divider,
  Accordion,
  AccordionSummary,
  AccordionDetails,
} from '@mui/material';
import SearchIcon from '@mui/icons-material/Search';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import LinkIcon from '@mui/icons-material/Link';
import Layout from '../components/common/Layout';
import MatchingResult from '../components/matching/MatchingResult';
import FileUpload from '../components/matching/FileUpload';
import { executeMatching, parseFile, createJobGroup } from '../lib/api/client';
import type { MatchingRequest, MatchingResponse, SupplementInfo } from '../types';

export default function MatchingPage() {
  const [jobText, setJobText] = useState('');
  const [engineerText, setEngineerText] = useState('');
  const [engineerFileKey, setEngineerFileKey] = useState('');
  const [supplement, setSupplement] = useState<SupplementInfo>({});
  const [result, setResult] = useState<MatchingResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const resultRef = useRef<HTMLDivElement>(null);

  const handleFileUpload = async (file: File) => {
    try {
      const response = await parseFile(file);
      setEngineerText((prev) => {
        if (prev) return response.text + '\n\n--- 補足 ---\n' + prev;
        return response.text;
      });
      setEngineerFileKey(response.file_key);
      if (response.parse_warnings.length > 0) {
        setError(response.parse_warnings.join('\n'));
      }
    } catch (err: any) {
      setError(err?.response?.data?.error || 'ファイルの読み取りに失敗しました');
    }
  };

  const handleExecute = async () => {
    if (!jobText.trim()) {
      setError('案件情報を入力してください');
      return;
    }
    if (!engineerText.trim() && !engineerFileKey) {
      setError('エンジニア情報（テキストまたはファイル）を入力してください');
      return;
    }

    setError('');
    setLoading(true);
    setResult(null);

    try {
      const req: MatchingRequest = {
        job_text: jobText,
        engineer_text: engineerText,
        engineer_file_key: engineerFileKey || undefined,
        supplement: Object.keys(supplement).length > 0 ? supplement : undefined,
      };
      const response = await executeMatching(req);
      setResult(response);
      setTimeout(() => {
        resultRef.current?.scrollIntoView({ behavior: 'smooth', block: 'start' });
      }, 100);
    } catch (err: any) {
      setError(err?.response?.data?.error || 'マッチング処理に失敗しました');
    } finally {
      setLoading(false);
    }
  };

  const updateSupplement = (key: keyof SupplementInfo, value: any) => {
    setSupplement((prev) => {
      if (value === '' || value === undefined) {
        const next = { ...prev };
        delete next[key];
        return next;
      }
      return { ...prev, [key]: value };
    });
  };

  return (
    <Layout>
      <Typography variant="h5" sx={{ mb: 3, fontWeight: 600 }}>
        マッチング実行
      </Typography>

      {error && (
        <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError('')}>
          {error}
        </Alert>
      )}

      <Grid container spacing={3}>
        {/* 案件情報 */}
        <Grid size={{ xs: 12, md: 6 }}>
          <Card sx={{ height: '100%' }}>
            <CardContent>
              <Typography variant="subtitle1" sx={{ mb: 1, fontWeight: 600 }}>
                案件情報
              </Typography>
              <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
                メール本文をそのまま貼り付けてください
              </Typography>
              <TextField
                multiline
                rows={10}
                fullWidth
                placeholder="案件名、必須スキル、単価、勤務地、開始時期など..."
                value={jobText}
                onChange={(e) => setJobText(e.target.value)}
              />
            </CardContent>
          </Card>
        </Grid>

        {/* エンジニア情報 */}
        <Grid size={{ xs: 12, md: 6 }}>
          <Card sx={{ height: '100%' }}>
            <CardContent>
              <Typography variant="subtitle1" sx={{ mb: 1, fontWeight: 600 }}>
                エンジニア情報
              </Typography>
              <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
                配信メールを貼り付け、またはファイルをアップロード
              </Typography>
              <TextField
                multiline
                rows={10}
                fullWidth
                placeholder="スキル、経験年数、希望単価、稼働時期など..."
                value={engineerText}
                onChange={(e) => setEngineerText(e.target.value)}
                sx={{ mb: 2 }}
              />
              <FileUpload onUpload={handleFileUpload} />
            </CardContent>
          </Card>
        </Grid>

        {/* 補足情報 */}
        <Grid size={12}>
          <Accordion defaultExpanded variant="outlined" sx={{ borderRadius: 2, '&:before': { display: 'none' } }}>
            <AccordionSummary expandIcon={<ExpandMoreIcon />}>
              <Typography variant="subtitle1" sx={{ fontWeight: 600 }}>
                補足情報
              </Typography>
            </AccordionSummary>
            <AccordionDetails>
              <Grid container spacing={2}>
                <Grid size={{ xs: 6, sm: 2 }}>
                  <FormControl fullWidth size="small">
                    <InputLabel>所属</InputLabel>
                    <Select
                      value={supplement.affiliation_type || ''}
                      label="所属"
                      onChange={(e) => updateSupplement('affiliation_type', e.target.value)}
                    >
                      <MenuItem value="">未選択</MenuItem>
                      <MenuItem value="duesk">デュスク社員</MenuItem>
                      <MenuItem value="partner">パートナー</MenuItem>
                    </Select>
                  </FormControl>
                </Grid>
                {supplement.affiliation_type === 'partner' && (
                  <Grid size={{ xs: 6, sm: 2 }}>
                    <TextField
                      label="パートナー会社名"
                      size="small"
                      fullWidth
                      value={supplement.affiliation_name || ''}
                      onChange={(e) => updateSupplement('affiliation_name', e.target.value)}
                    />
                  </Grid>
                )}
                <Grid size={{ xs: 6, sm: 2 }}>
                  <TextField
                    label="希望単価（万円）"
                    type="number"
                    size="small"
                    fullWidth
                    value={supplement.rate || ''}
                    onChange={(e) => updateSupplement('rate', e.target.value ? Number(e.target.value) : undefined)}
                  />
                </Grid>
                <Grid size={{ xs: 6, sm: 2 }}>
                  <FormControl fullWidth size="small">
                    <InputLabel>国籍</InputLabel>
                    <Select
                      value={supplement.nationality || ''}
                      label="国籍"
                      onChange={(e) => updateSupplement('nationality', e.target.value)}
                    >
                      <MenuItem value="">未選択</MenuItem>
                      <MenuItem value="japanese">日本</MenuItem>
                      <MenuItem value="other">外国籍</MenuItem>
                    </Select>
                  </FormControl>
                </Grid>
                <Grid size={{ xs: 6, sm: 2 }}>
                  <FormControl fullWidth size="small">
                    <InputLabel>雇用形態</InputLabel>
                    <Select
                      value={supplement.employment_type || ''}
                      label="雇用形態"
                      onChange={(e) => updateSupplement('employment_type', e.target.value)}
                    >
                      <MenuItem value="">未選択</MenuItem>
                      <MenuItem value="employee">正社員</MenuItem>
                      <MenuItem value="freelance">フリーランス</MenuItem>
                    </Select>
                  </FormControl>
                </Grid>
                <Grid size={{ xs: 6, sm: 2 }}>
                  <TextField
                    label="稼働可能日"
                    placeholder="2026-04"
                    size="small"
                    fullWidth
                    value={supplement.available_from || ''}
                    onChange={(e) => updateSupplement('available_from', e.target.value)}
                  />
                </Grid>
                <Grid size={{ xs: 6, sm: 3 }}>
                  <TextField
                    label="送信元会社名"
                    size="small"
                    fullWidth
                    value={supplement.supply_chain_source || ''}
                    onChange={(e) => updateSupplement('supply_chain_source', e.target.value)}
                  />
                </Grid>
                <Grid size={{ xs: 6, sm: 2 }}>
                  <FormControl fullWidth size="small">
                    <InputLabel>商流</InputLabel>
                    <Select
                      value={supplement.supply_chain_level ?? ''}
                      label="商流"
                      onChange={(e) => {
                        const v = e.target.value as string | number;
                        updateSupplement('supply_chain_level', v === '' ? undefined : Number(v));
                      }}
                    >
                      <MenuItem value="">未選択</MenuItem>
                      <MenuItem value={1}>エンド直</MenuItem>
                      <MenuItem value={2}>1次請け</MenuItem>
                      <MenuItem value={3}>2次請け</MenuItem>
                      <MenuItem value={4}>3次以上</MenuItem>
                    </Select>
                  </FormControl>
                </Grid>
              </Grid>
            </AccordionDetails>
          </Accordion>
        </Grid>

        {/* 実行ボタン */}
        <Grid size={12}>
          <Box sx={{ display: 'flex', justifyContent: 'center' }}>
            <Button
              variant="contained"
              size="large"
              startIcon={loading ? <CircularProgress size={20} color="inherit" /> : <SearchIcon />}
              onClick={handleExecute}
              disabled={loading}
              sx={{ px: 6, py: 1.5 }}
            >
              {loading ? 'AIが分析中です（最大1分程度かかります）' : 'マッチング実行'}
            </Button>
          </Box>
        </Grid>

        {/* 結果表示 */}
        {result && (
          <Grid size={12} ref={resultRef}>
            <Divider sx={{ my: 2 }} />
            <Box sx={{ display: 'flex', justifyContent: 'flex-end', mb: 2 }}>
              <Button
                variant="outlined"
                size="small"
                startIcon={<LinkIcon />}
                onClick={async () => {
                  const name = window.prompt('案件グループ名を入力してください');
                  if (name) {
                    try {
                      await createJobGroup(name, result.id);
                      alert('案件グループを作成しました');
                    } catch {
                      alert('作成に失敗しました');
                    }
                  }
                }}
              >
                案件グループを作成
              </Button>
            </Box>
            <MatchingResult result={result} />
          </Grid>
        )}
      </Grid>
    </Layout>
  );
}
